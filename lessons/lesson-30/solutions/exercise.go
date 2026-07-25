// Package solutions obsahuje referenční řešení lekce 30.
package solutions

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand/v2"
	"net"
	"net/http"
	"time"
)

// MaxBodyBytes je horní limit velikosti těla odpovědi, které FetchJSON přečte.
const MaxBodyBytes int64 = 1 << 20 // 1 MiB

// ShutdownGracePeriod je čas, který RunServer dá rozpracovaným požadavkům na dokončení.
const ShutdownGracePeriod = 5 * time.Second

// maxBackoff je strop pro exponenciální backoff, aby čekání nerostlo do nekonečna.
const maxBackoff = 30 * time.Second

// ErrBodyTooLarge znamená, že odpověď překročila MaxBodyBytes.
var ErrBodyTooLarge = errors.New("response body too large")

// ErrNoAttempts znamená, že Retry dostal nekladný počet pokusů.
var ErrNoAttempts = errors.New("attempts must be at least 1")

// StatusError je chyba pro odpověď mimo rozsah 2xx.
type StatusError struct {
	StatusCode int
	Status     string
}

// Error implementuje error.
func (e *StatusError) Error() string {
	return fmt.Sprintf("unexpected status %s", e.Status)
}

// PermanentError obaluje chybu, kterou nemá smysl opakovat.
type PermanentError struct {
	Err error
}

// Error implementuje error.
func (e *PermanentError) Error() string {
	return "permanent: " + e.Err.Error()
}

// Unwrap zpřístupní obalenou chybu pro errors.Is a errors.As.
func (e *PermanentError) Unwrap() error {
	return e.Err
}

// Permanent označí chybu jako neopakovatelnou.
func Permanent(err error) error {
	return &PermanentError{Err: err}
}

// NewHTTPClient vrátí klienta s timeoutem a vlastním transportem.
func NewHTTPClient(timeout time.Duration) *http.Client {
	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   5 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		MaxIdleConns:          100,
		MaxIdleConnsPerHost:   10,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   5 * time.Second,
		ExpectContinueTimeout: time.Second,
		ForceAttemptHTTP2:     true,
	}
	return &http.Client{
		Timeout:   timeout,
		Transport: transport,
	}
}

// FetchJSON stáhne URL a rozparsuje odpověď do T.
func FetchJSON[T any](ctx context.Context, c *http.Client, url string) (T, error) {
	var zero T

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return zero, fmt.Errorf("fetch %s: %w", url, err)
	}
	req.Header.Set("Accept", "application/json")

	resp, err := c.Do(req)
	if err != nil {
		return zero, fmt.Errorf("fetch %s: %w", url, err)
	}
	// Tělo je potřeba dočíst, jinak se spojení nevrátí do poolu a naváže se nové.
	defer func() {
		_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, MaxBodyBytes))
		_ = resp.Body.Close()
	}()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return zero, fmt.Errorf("fetch %s: %w", url, &StatusError{
			StatusCode: resp.StatusCode,
			Status:     resp.Status,
		})
	}

	// O jeden bajt víc, než smíme přijmout — přetečení pak poznáme podle délky.
	data, err := io.ReadAll(io.LimitReader(resp.Body, MaxBodyBytes+1))
	if err != nil {
		return zero, fmt.Errorf("fetch %s: %w", url, err)
	}
	if int64(len(data)) > MaxBodyBytes {
		return zero, fmt.Errorf("fetch %s: %w", url, ErrBodyTooLarge)
	}

	var out T
	if err := json.Unmarshal(data, &out); err != nil {
		return zero, fmt.Errorf("fetch %s: %w", url, err)
	}
	return out, nil
}

// Retry volá fn dokud neuspěje, nejvýš attempts krát, s exponenciálním backoffem.
func Retry(ctx context.Context, attempts int, base time.Duration, fn func(ctx context.Context) error) error {
	if attempts < 1 {
		return fmt.Errorf("attempts=%d: %w", attempts, ErrNoAttempts)
	}

	var lastErr error
	for attempt := 0; attempt < attempts; attempt++ {
		if err := ctx.Err(); err != nil {
			return joinErrors(lastErr, err)
		}

		err := fn(ctx)
		if err == nil {
			return nil
		}

		var permanent *PermanentError
		if errors.As(err, &permanent) {
			return err
		}
		lastErr = err

		if attempt == attempts-1 {
			break
		}
		if err := wait(ctx, backoff(base, attempt)); err != nil {
			return joinErrors(lastErr, err)
		}
	}
	return fmt.Errorf("po %d pokusech: %w", attempts, lastErr)
}

// joinErrors spojí poslední chybu s důvodem ukončení; nil operandy zahodí.
func joinErrors(last, reason error) error {
	if last == nil {
		return reason
	}
	return errors.Join(last, reason)
}

// backoff spočítá exponenciální prodlevu s jitterem, aby klienti nešli v zákrytu.
func backoff(base time.Duration, attempt int) time.Duration {
	if base <= 0 {
		return 0
	}
	d := base
	for i := 0; i < attempt && d < maxBackoff; i++ {
		d *= 2
	}
	if d > maxBackoff {
		d = maxBackoff
	}
	half := d / 2
	return half + rand.N(half+1)
}

// wait počká danou dobu, nebo skončí dřív se zrušeným kontextem.
func wait(ctx context.Context, d time.Duration) error {
	if d <= 0 {
		return ctx.Err()
	}
	timer := time.NewTimer(d)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

// RunServer obsluhuje ln, dokud se nezruší ctx, pak server elegantně ukončí.
func RunServer(ctx context.Context, srv *http.Server, ln net.Listener) error {
	serveErr := make(chan error, 1)
	go func() {
		serveErr <- srv.Serve(ln)
	}()

	select {
	case err := <-serveErr:
		// Server spadl sám od sebe — ErrServerClosed sem přijde jen po Close.
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	case <-ctx.Done():
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), ShutdownGracePeriod)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("shutdown: %w", err)
	}
	if err := <-serveErr; err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

// Package solutions obsahuje referenční řešení lekce 30.
package solutions

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"time"
)

// ShutdownGracePeriod je čas, který RunServer dá rozpracovaným požadavkům na dokončení.
const ShutdownGracePeriod = 5 * time.Second

// PermanentError obaluje chybu, kterou nemá smysl opakovat.
type PermanentError struct {
	Err error
}

// --- Stupeň: jednoduchý ---

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

// --- Stupeň: střední ---

// NewHTTPClient vrátí klienta s timeoutem a vlastním transportem.
func NewHTTPClient(timeout time.Duration) *http.Client {
	transport := &http.Transport{
		Proxy:                 http.ProxyFromEnvironment,
		MaxIdleConns:          100,
		MaxIdleConnsPerHost:   10,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   5 * time.Second,
		ExpectContinueTimeout: time.Second,
	}
	return &http.Client{
		Timeout:   timeout,
		Transport: transport,
	}
}

// --- Stupeň: obtížný ---

// RunServer obsluhuje ln, dokud se nezruší ctx, pak server elegantně ukončí.
func RunServer(ctx context.Context, srv *http.Server, ln net.Listener) error {
	serveErr := make(chan error, 1)
	go func() {
		serveErr <- srv.Serve(ln)
	}()

	select {
	case err := <-serveErr:
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

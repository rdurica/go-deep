// Package exercise obsahuje cvičení lekce 30.
package exercise

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"time"
)

// MaxBodyBytes je horní limit velikosti těla odpovědi, které FetchJSON přečte.
const MaxBodyBytes int64 = 1 << 20 // 1 MiB

// ShutdownGracePeriod je čas, který RunServer dá rozpracovaným požadavkům na dokončení.
const ShutdownGracePeriod = 5 * time.Second

// ErrBodyTooLarge znamená, že odpověď překročila MaxBodyBytes.
var ErrBodyTooLarge = errors.New("response body too large")

// ErrNoAttempts znamená, že Retry dostal nekladný počet pokusů.
var ErrNoAttempts = errors.New("attempts must be at least 1")

// StatusError je chyba pro odpověď mimo rozsah 2xx.
type StatusError struct {
	StatusCode int
	Status     string
}

// --- Stupeň: jednoduchý ---
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

// --- Stupeň: střední ---
// Unwrap zpřístupní obalenou chybu pro errors.Is a errors.As.
func (e *PermanentError) Unwrap() error {
	return e.Err
}

// Permanent označí chybu jako neopakovatelnou.
func Permanent(err error) error {
	return &PermanentError{Err: err}
}

// --- Stupeň: obtížný ---
// NewHTTPClient vrátí klienta s timeoutem a vlastním transportem.
// Transport s kladným MaxIdleConnsPerHost.
func NewHTTPClient(timeout time.Duration) *http.Client {
	// TODO
	return nil
}

// FetchJSON stáhne URL a rozparsuje odpověď do T.
// GET s kontextem; tělo v defer dočti a zavři; mimo 2xx → *StatusError (errors.As).
// Tělo max MaxBodyBytes; překročení → chyba obalující ErrBodyTooLarge. Při chybě nulové T.
func FetchJSON[T any](ctx context.Context, c *http.Client, url string) (T, error) {
	// TODO
	return *new(T), nil
}

// Retry volá fn dokud neuspěje, nejvýš attempts krát, s exponenciálním backoffem.
// attempts < 1 → ErrNoAttempts; ctx zrušený → nevolej fn; *PermanentError → bez dalšího pokusu.
// Backoff base, 2*base, 4*base… s jitterem; po vyčerpání obal poslední chybu %w.
func Retry(ctx context.Context, attempts int, base time.Duration, fn func(ctx context.Context) error) error {
	// TODO
	return nil
}

// RunServer obsluhuje ln, dokud se nezruší ctx, pak server elegantně ukončí.
// Serve v goroutině; ErrServerClosed → nil; při ctx.Done() Shutdown s novým kontextem ShutdownGracePeriod.
func RunServer(ctx context.Context, srv *http.Server, ln net.Listener) error {
	// TODO
	return nil
}

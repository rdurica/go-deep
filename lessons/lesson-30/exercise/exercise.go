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
	panic("TODO: úkol A")
}

// FetchJSON stáhne URL a rozparsuje odpověď do T.
func FetchJSON[T any](ctx context.Context, c *http.Client, url string) (T, error) {
	panic("TODO: úkol A")
}

// Retry volá fn dokud neuspěje, nejvýš attempts krát, s exponenciálním backoffem.
func Retry(ctx context.Context, attempts int, base time.Duration, fn func(ctx context.Context) error) error {
	panic("TODO: úkol B")
}

// RunServer obsluhuje ln, dokud se nezruší ctx, pak server elegantně ukončí.
func RunServer(ctx context.Context, srv *http.Server, ln net.Listener) error {
	panic("TODO: úkol C")
}

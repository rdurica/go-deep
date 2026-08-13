// Package exercise obsahuje cvičení lekce 30.
package exercise

import (
	"context"
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
//
// POZOR: kód níže je ZÁMĚRNĚ VADNÝ. Vrací nil místo obalené chyby.
// Najdi chybu a oprav — testy před opravou padají.
func (e *PermanentError) Unwrap() error {
	return nil
}

// Permanent označí chybu jako neopakovatelnou.
func Permanent(err error) error {
	return &PermanentError{Err: err}
}

// --- Stupeň: střední ---

// NewHTTPClient vrátí klienta s timeoutem a vlastním transportem.
// Transport s kladným MaxIdleConnsPerHost.
func NewHTTPClient(timeout time.Duration) *http.Client {
	// TODO
	return nil
}

// --- Stupeň: obtížný ---

// RunServer obsluhuje ln, dokud se nezruší ctx, pak server elegantně ukončí.
// Serve v goroutině; ErrServerClosed → nil; při ctx.Done() Shutdown s novým kontextem ShutdownGracePeriod.
func RunServer(ctx context.Context, srv *http.Server, ln net.Listener) error {
	// TODO
	return nil
}

// Package exercise obsahuje cvičení lekce 37.
package exercise

import (
	"context"
	"errors"
	"net/http"
)

// Middleware je obálka nad http.Handlerem.
type Middleware func(http.Handler) http.Handler

// Chyby parsování hlavičky Authorization.
var (
	// ErrMissingAuthorization znamená, že hlavička chybí nebo je prázdná.
	ErrMissingAuthorization = errors.New("chybí hlavička Authorization")
	// ErrUnsupportedScheme znamená, že schéma není Bearer.
	ErrUnsupportedScheme = errors.New("nepodporované autentizační schéma")
	// ErrMissingToken znamená, že za schématem Bearer není token.
	ErrMissingToken = errors.New("chybí token")
	// ErrMalformedAuthorization znamená, že hlavička má víc částí, než má mít.
	ErrMalformedAuthorization = errors.New("neplatný tvar hlavičky Authorization")
)

// WWWAuthenticate je hodnota hlavičky WWW-Authenticate, kterou vrací 401.
const WWWAuthenticate = `Bearer realm="api"`

// RouteUnknown je náhradní hodnota labelu route, když vzor cesty není známý.
const RouteUnknown = "unknown"

// ParseBearer vytáhne token z hodnoty hlavičky Authorization.
func ParseBearer(header string) (string, error) {
	// TODO: úkol A
	return "", nil
}

// HashPassword spočítá otisk hesla se solí.
//
// POZOR: SHA-256 je na hesla ŠPATNĚ. Je záměrně rychlá, takže se dá hrubou
// silou zkoušet miliardkrát za sekundu. Tady je jen proto, aby cvičení
// zůstalo ve standardní knihovně. V produkci patří bcrypt nebo argon2
// (golang.org/x/crypto/bcrypt, golang.org/x/crypto/argon2).
func HashPassword(password, salt string) string {
	// TODO: úkol A
	return ""
}

// VerifyPassword porovná otisk hesla v konstantním čase.
func VerifyPassword(hash, password, salt string) bool {
	// TODO: úkol A
	return false
}

// Authenticate vrací middleware, který ověří Bearer token proti mapě token → uživatel.
func Authenticate(tokens map[string]string) Middleware {
	// TODO: úkol B
	return *new(Middleware)
}

// UserFrom vrací jméno ověřeného uživatele z kontextu.
func UserFrom(ctx context.Context) (string, bool) {
	// TODO: úkol B
	return "", false
}

// Stat je souhrn jedné metrické řady.
type Stat struct {
	Count int
	Sum   float64
	Min   float64
	Max   float64
}

// Metrics je registr metrik bezpečný pro souběžné použití.
type Metrics struct {
	// TODO: doplň pole (zámek a mapa řad).
}

// NewMetrics vytvoří prázdný registr metrik.
func NewMetrics() *Metrics {
	// TODO: úkol B
	return nil
}

// SeriesKey složí jméno řady se seřazenými labely, například
// `http_requests_total{method="GET",route="/items/{id}"}`.
func SeriesKey(name string, labels map[string]string) string {
	// TODO: úkol B
	return ""
}

// Inc zvýší čítač o jedna.
func (m *Metrics) Inc(name string, labels map[string]string) {
	// TODO: úkol B
}

// Observe zaznamená jedno pozorování do řady name.
func (m *Metrics) Observe(name string, v float64) {
	// TODO: úkol B
}

// Snapshot vrací kopii všech řad.
func (m *Metrics) Snapshot() map[string]Stat {
	// TODO: úkol B
	return nil
}

// Text serializuje registr do deterministického textového formátu.
func (m *Metrics) Text() string {
	// TODO: úkol B
	return ""
}

// WithRoute uloží vzor cesty do kontextu požadavku.
func WithRoute(route string, next http.Handler) http.Handler {
	// TODO: úkol C
	return *new(http.Handler)
}

// RouteFrom vrací vzor cesty z kontextu, nebo RouteUnknown.
func RouteFrom(ctx context.Context) string {
	// TODO: úkol C
	return ""
}

// Instrument vrací middleware, který měří počet požadavků a jejich dobu.
func Instrument(m *Metrics) Middleware {
	// TODO: úkol C
	return *new(Middleware)
}

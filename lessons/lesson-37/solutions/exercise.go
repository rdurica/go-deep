// Package solutions obsahuje referenční řešení lekce 37.
package solutions

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
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

// --- Stupeň: jednoduchý ---
// ParseBearer vytáhne token z hodnoty hlavičky Authorization.
func ParseBearer(header string) (string, error) {
	// strings.Fields sjednotí libovolné množství mezer i tabulátorů,
	// takže "  bearer   abc  " projde stejně jako "Bearer abc".
	fields := strings.Fields(header)
	switch {
	case len(fields) == 0:
		return "", ErrMissingAuthorization
	case !strings.EqualFold(fields[0], "bearer"):
		return "", ErrUnsupportedScheme
	case len(fields) == 1:
		return "", ErrMissingToken
	case len(fields) > 2:
		return "", ErrMalformedAuthorization
	}
	return fields[1], nil
}

// HashPassword spočítá otisk hesla se solí.
//
// POZOR: SHA-256 je na hesla ŠPATNĚ. Je záměrně rychlá, takže se dá hrubou
// silou zkoušet miliardkrát za sekundu. Tady je jen proto, aby cvičení
// zůstalo ve standardní knihovně. V produkci patří bcrypt nebo argon2
// (golang.org/x/crypto/bcrypt, golang.org/x/crypto/argon2).
func HashPassword(password, salt string) string {
	sum := sha256.Sum256([]byte(salt + ":" + password))
	return hex.EncodeToString(sum[:])
}

// VerifyPassword porovná otisk hesla v konstantním čase.
func VerifyPassword(hash, password, salt string) bool {
	got := HashPassword(password, salt)
	return subtle.ConstantTimeCompare([]byte(got), []byte(hash)) == 1
}

// contextKey je vlastní typ klíče, aby nemohlo dojít ke kolizi s klíči
// jiných balíčků. Klíč typu string je v kontextu vždy chyba.
type contextKey int

const (
	userKey contextKey = iota
	routeKey
)

// Authenticate vrací middleware, který ověří Bearer token proti mapě token → uživatel.
func Authenticate(tokens map[string]string) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			presented, err := ParseBearer(r.Header.Get("Authorization"))
			if err != nil {
				unauthorized(w)
				return
			}

			// Projdeme vždy všechny tokeny a nikdy nekončíme dřív.
			// Ani délka smyčky, ani délka porovnání tak neprozradí,
			// jak blízko byl útočník správnému tokenu.
			var user string
			var found int
			for token, name := range tokens {
				if subtle.ConstantTimeCompare([]byte(token), []byte(presented)) == 1 {
					user = name
					found = 1
				}
			}
			if found != 1 {
				unauthorized(w)
				return
			}

			next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), userKey, user)))
		})
	}
}

func unauthorized(w http.ResponseWriter) {
	w.Header().Set("WWW-Authenticate", WWWAuthenticate)
	http.Error(w, "unauthorized", http.StatusUnauthorized)
}

// UserFrom vrací jméno ověřeného uživatele z kontextu.
func UserFrom(ctx context.Context) (string, bool) {
	user, ok := ctx.Value(userKey).(string)
	return user, ok
}

// --- Stupeň: střední ---
// Stat je souhrn jedné metrické řady.
type Stat struct {
	Count int
	Sum   float64
	Min   float64
	Max   float64
}

// Metrics je registr metrik bezpečný pro souběžné použití.
type Metrics struct {
	mu     sync.Mutex
	series map[string]Stat
}

// NewMetrics vytvoří prázdný registr metrik.
func NewMetrics() *Metrics {
	return &Metrics{series: make(map[string]Stat)}
}

// SeriesKey složí jméno řady se seřazenými labely, například
// `http_requests_total{method="GET",route="/items/{id}"}`.
func SeriesKey(name string, labels map[string]string) string {
	if len(labels) == 0 {
		return name
	}
	names := make([]string, 0, len(labels))
	for label := range labels {
		names = append(names, label)
	}
	sort.Strings(names)

	var b strings.Builder
	b.WriteString(name)
	b.WriteByte('{')
	for i, label := range names {
		if i > 0 {
			b.WriteByte(',')
		}
		b.WriteString(label)
		b.WriteByte('=')
		b.WriteString(strconv.Quote(labels[label]))
	}
	b.WriteByte('}')
	return b.String()
}

// Inc zvýší čítač o jedna.
func (m *Metrics) Inc(name string, labels map[string]string) {
	m.record(SeriesKey(name, labels), 1)
}

// Observe zaznamená jedno pozorování do řady name.
func (m *Metrics) Observe(name string, v float64) {
	m.record(name, v)
}

func (m *Metrics) record(key string, v float64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	stat, ok := m.series[key]
	if !ok {
		m.series[key] = Stat{Count: 1, Sum: v, Min: v, Max: v}
		return
	}
	stat.Count++
	stat.Sum += v
	if v < stat.Min {
		stat.Min = v
	}
	if v > stat.Max {
		stat.Max = v
	}
	m.series[key] = stat
}

// Snapshot vrací kopii všech řad.
func (m *Metrics) Snapshot() map[string]Stat {
	m.mu.Lock()
	defer m.mu.Unlock()

	out := make(map[string]Stat, len(m.series))
	for key, stat := range m.series {
		out[key] = stat
	}
	return out
}

// --- Stupeň: obtížný ---
// Text serializuje registr do textu — řady seřazené abecedně, deterministický výpis.
// Každá řada: "<SeriesKey> count=%d sum=%g min=%g max=%g\n".
func (m *Metrics) Text() string {
	snapshot := m.Snapshot()
	keys := make([]string, 0, len(snapshot))
	for key := range snapshot {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	var b strings.Builder
	for _, key := range keys {
		stat := snapshot[key]
		b.WriteString(key)
		b.WriteString(" count=")
		b.WriteString(strconv.Itoa(stat.Count))
		b.WriteString(" sum=")
		b.WriteString(formatFloat(stat.Sum))
		b.WriteString(" min=")
		b.WriteString(formatFloat(stat.Min))
		b.WriteString(" max=")
		b.WriteString(formatFloat(stat.Max))
		b.WriteByte('\n')
	}
	return b.String()
}

func formatFloat(v float64) string {
	return strconv.FormatFloat(v, 'g', -1, 64)
}

// WithRoute uloží vzor cesty do kontextu požadavku.
func WithRoute(route string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), routeKey, route)))
	})
}

// RouteFrom vrací vzor cesty z kontextu, nebo RouteUnknown.
func RouteFrom(ctx context.Context) string {
	route, ok := ctx.Value(routeKey).(string)
	if !ok || route == "" {
		return RouteUnknown
	}
	return route
}

// statusRecorder si pamatuje status kód, protože http.ResponseWriter
// ho po zápisu nijak nezpřístupňuje.
type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (s *statusRecorder) WriteHeader(status int) {
	s.status = status
	s.ResponseWriter.WriteHeader(status)
}

func (s *statusRecorder) Write(b []byte) (int, error) {
	if s.status == 0 {
		s.status = http.StatusOK
	}
	return s.ResponseWriter.Write(b)
}

// Instrument vrací middleware, který měří počet požadavků a jejich dobu.
func Instrument(m *Metrics) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			rec := &statusRecorder{ResponseWriter: w}
			next.ServeHTTP(rec, r)
			if rec.status == 0 {
				rec.status = http.StatusOK
			}

			// Label route je VZOR cesty, ne r.URL.Path. Kdyby tu bylo
			// /items/1, /items/2, …, vyrobí každý požadavek novou řadu
			// a metrický backend se udusí kardinalitou.
			m.Inc("http_requests_total", map[string]string{
				"method": r.Method,
				"route":  RouteFrom(r.Context()),
				"status": strconv.Itoa(rec.status),
			})
			m.Observe("http_request_duration_seconds", time.Since(start).Seconds())
		})
	}
}

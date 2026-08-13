// Package exercise obsahuje cvičení lekce 37.
package exercise

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

type Middleware func(http.Handler) http.Handler

var (
	ErrMissingAuthorization   = errors.New("chybí hlavička Authorization")
	ErrUnsupportedScheme      = errors.New("nepodporované autentizační schéma")
	ErrMissingToken           = errors.New("chybí token")
	ErrMalformedAuthorization = errors.New("neplatný tvar hlavičky Authorization")
)

const WWWAuthenticate = `Bearer realm="api"`
const RouteUnknown = "unknown"

// --- Stupeň: jednoduchý ---

// ParseBearer vytáhne token z Authorization. Schéma case-insensitive.
func ParseBearer(header string) (string, error) {
	// TODO
	return "", nil
}

// HashPassword spočítá hex SHA-256 nad solí a heslem.
func HashPassword(password, salt string) string {
	sum := sha256.Sum256([]byte(salt + ":" + password))
	return hex.EncodeToString(sum[:])
}

// VerifyPassword porovná otisk hesla.
//
// POZOR: kód níže je ZÁMĚRNĚ VADNÝ. Používá == místo ConstantTimeCompare.
func VerifyPassword(hash, password, salt string) bool {
	got := HashPassword(password, salt)
	_ = subtle.ConstantTimeCompare // nápověda: použij místo ==
	return got == hash
}

type contextKey int

const userKey contextKey = iota

// Authenticate vrací middleware: Bearer token, konstantní čas, uživatel v kontextu.
// Selhání → 401 s WWW-Authenticate.
func Authenticate(tokens map[string]string) Middleware {
	// TODO
	return *new(Middleware)
}

func UserFrom(ctx context.Context) (string, bool) {
	u, ok := ctx.Value(userKey).(string)
	return u, ok
}

// --- Stupeň: střední ---

type Stat struct {
	Count int
	Sum   float64
	Min   float64
	Max   float64
}

type Metrics struct {
	mu     sync.Mutex
	series map[string]Stat
}

func NewMetrics() *Metrics {
	return &Metrics{series: make(map[string]Stat)}
}

func SeriesKey(name string, labels map[string]string) string {
	if len(labels) == 0 {
		return name
	}
	keys := make([]string, 0, len(labels))
	for k := range labels {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var b strings.Builder
	b.WriteString(name)
	b.WriteByte('{')
	for i, k := range keys {
		if i > 0 {
			b.WriteByte(',')
		}
		b.WriteString(k)
		b.WriteByte('=')
		b.WriteString(strconv.Quote(labels[k]))
	}
	b.WriteByte('}')
	return b.String()
}

func (m *Metrics) Inc(name string, labels map[string]string) {
	key := SeriesKey(name, labels)
	m.mu.Lock()
	defer m.mu.Unlock()
	s := m.series[key]
	s.Count++
	m.series[key] = s
}

func (m *Metrics) Observe(name string, v float64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	s := m.series[name]
	s.Count++
	s.Sum += v
	if s.Count == 1 || v < s.Min {
		s.Min = v
	}
	if s.Count == 1 || v > s.Max {
		s.Max = v
	}
	m.series[name] = s
}

func (m *Metrics) Snapshot() map[string]Stat {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make(map[string]Stat, len(m.series))
	for k, v := range m.series {
		out[k] = v
	}
	return out
}

func WithRoute(route string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), routeKey, route)))
	})
}

type routeKeyType int

const routeKey routeKeyType = 0

func RouteFrom(ctx context.Context) string {
	if v, ok := ctx.Value(routeKey).(string); ok && v != "" {
		return v
	}
	return RouteUnknown
}

// --- Stupeň: obtížný ---

// Instrument měří http_requests_total a http_request_duration_seconds.
// route je vzor cesty z kontextu; implicitní status 200.
func Instrument(m *Metrics) Middleware {
	// TODO
	return *new(Middleware)
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (s *statusRecorder) WriteHeader(code int) {
	s.status = code
	s.ResponseWriter.WriteHeader(code)
}

func since(start time.Time) float64 { return time.Since(start).Seconds() }

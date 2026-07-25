// Package solutions obsahuje referenční řešení lekce 60.
package solutions

import (
	"errors"
	"io"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"
)

// RateLimiter je token bucket omezovač bezpečný pro souběžné použití.
type RateLimiter struct {
	mu       sync.Mutex
	capacity int
	refill   time.Duration
	now      func() time.Time
	buckets  map[string]*bucket
}

// bucket je stav jednoho klíče.
type bucket struct {
	tokens int
	last   time.Time
	seen   time.Time
}

// NewRateLimiter vytvoří omezovač s kapacitou capacity a doplňováním jednoho tokenu
// za refill. Hodiny se předávají jako závislost; nil znamená time.Now.
func NewRateLimiter(capacity int, refill time.Duration, now func() time.Time) *RateLimiter {
	if capacity < 1 {
		capacity = 1
	}
	if refill <= 0 {
		refill = time.Second
	}
	if now == nil {
		now = time.Now
	}
	return &RateLimiter{
		capacity: capacity,
		refill:   refill,
		now:      now,
		buckets:  make(map[string]*bucket),
	}
}

// Allow odebere token pro daný klíč. Vrací false, když je kbelík prázdný.
func (rl *RateLimiter) Allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	t := rl.now()
	b, ok := rl.buckets[key]
	if !ok {
		b = &bucket{tokens: rl.capacity, last: t}
		rl.buckets[key] = b
	}

	if elapsed := t.Sub(b.last); elapsed >= rl.refill {
		add := int(elapsed / rl.refill)
		b.tokens += add
		if b.tokens > rl.capacity {
			b.tokens = rl.capacity
		}
		b.last = b.last.Add(time.Duration(add) * rl.refill)
	}
	b.seen = t

	if b.tokens <= 0 {
		return false
	}
	b.tokens--
	return true
}

// Cleanup zahodí kbelíky, které se nepoužily aspoň idle, a vrátí jejich počet.
func (rl *RateLimiter) Cleanup(idle time.Duration) int {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	t := rl.now()
	removed := 0
	for key, b := range rl.buckets {
		if t.Sub(b.seen) >= idle {
			delete(rl.buckets, key)
			removed++
		}
	}
	return removed
}

// Len vrací počet sledovaných klíčů.
func (rl *RateLimiter) Len() int {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	return len(rl.buckets)
}

// HardenOptions konfiguruje middleware chain funkce Harden.
type HardenOptions struct {
	MaxBodyBytes int64
	Timeout      time.Duration
	Limiter      *RateLimiter
	KeyFunc      func(*http.Request) string
}

// SecurityHeaders jsou hlavičky, které Harden přidá do každé odpovědi.
var SecurityHeaders = map[string]string{
	"X-Content-Type-Options":  "nosniff",
	"X-Frame-Options":         "DENY",
	"Referrer-Policy":         "no-referrer",
	"Content-Security-Policy": "default-src 'none'",
}

// Harden složí produkční middleware chain: recovery, bezpečnostní hlavičky,
// rate limiting, limit velikosti těla a timeout.
func Harden(h http.Handler, opts HardenOptions) http.Handler {
	next := h

	if opts.Timeout > 0 {
		next = http.TimeoutHandler(next, opts.Timeout, "timeout")
	}
	if opts.MaxBodyBytes > 0 {
		next = bodyLimit(next, opts.MaxBodyBytes)
	}
	if opts.Limiter != nil {
		next = rateLimit(next, opts.Limiter, opts.KeyFunc)
	}
	next = securityHeaders(next)
	return recovery(next)
}

// recovery zachytí paniku v handleru a odpoví 500.
func recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				w.WriteHeader(http.StatusInternalServerError)
				io.WriteString(w, "internal server error\n")
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// securityHeaders přidá bezpečnostní hlavičky před zavoláním handleru.
func securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for name, value := range SecurityHeaders {
			w.Header().Set(name, value)
		}
		next.ServeHTTP(w, r)
	})
}

// rateLimit odmítne požadavek nad limit s kódem 429.
func rateLimit(next http.Handler, limiter *RateLimiter, keyFunc func(*http.Request) string) http.Handler {
	if keyFunc == nil {
		keyFunc = func(r *http.Request) string {
			host, _, found := strings.Cut(r.RemoteAddr, ":")
			if !found {
				return r.RemoteAddr
			}
			return host
		}
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !limiter.Allow(keyFunc(r)) {
			w.WriteHeader(http.StatusTooManyRequests)
			io.WriteString(w, "too many requests\n")
			return
		}
		next.ServeHTTP(w, r)
	})
}

// bodyLimit odmítne příliš velké tělo a zbytek omezí přes MaxBytesReader.
func bodyLimit(next http.Handler, max int64) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.ContentLength > max {
			w.WriteHeader(http.StatusRequestEntityTooLarge)
			io.WriteString(w, "request body too large\n")
			return
		}
		if r.Body != nil {
			r.Body = http.MaxBytesReader(w, r.Body, max)
		}
		next.ServeHTTP(w, r)
	})
}

// Chyby produkčního auditu.
var (
	ErrNoChecks       = errors.New("audit: prázdný checklist")
	ErrDuplicateCheck = errors.New("audit: duplicitní ID kontroly")
	ErrInvalidCheck   = errors.New("audit: kontrola bez ID")
)

// Check je jedna položka production checklistu.
type Check struct {
	ID       string
	Area     string
	Done     bool
	Critical bool
}

// Report je vyhodnocení production checklistu.
type Report struct {
	Total          int
	Passed         int
	CriticalFailed int
	Score          float64
	Ready          bool
	Missing        []string
}

// AuditReport vyhodnotí připravenost k nasazení podle production checklistu.
func AuditReport(checks []Check) (Report, error) {
	if len(checks) == 0 {
		return Report{}, ErrNoChecks
	}

	seen := make(map[string]struct{}, len(checks))
	report := Report{Total: len(checks)}
	for _, c := range checks {
		if strings.TrimSpace(c.ID) == "" {
			return Report{}, ErrInvalidCheck
		}
		if _, dup := seen[c.ID]; dup {
			return Report{}, ErrDuplicateCheck
		}
		seen[c.ID] = struct{}{}

		if c.Done {
			report.Passed++
			continue
		}
		report.Missing = append(report.Missing, c.ID)
		if c.Critical {
			report.CriticalFailed++
		}
	}

	sort.Strings(report.Missing)
	report.Score = float64(report.Passed) / float64(report.Total)
	report.Ready = report.CriticalFailed == 0 && report.Score >= 0.8
	return report, nil
}

// LessonResult je výsledek jedné lekce kurzu.
type LessonResult struct {
	Lesson int
	Phase  int
	Passed bool
}

// PhaseStat je pokrok v jedné fázi kurzu.
type PhaseStat struct {
	Total  int
	Passed int
}

// Summary je souhrn pokroku celým kurzem.
type Summary struct {
	Total        int
	Passed       int
	Percent      float64
	ByPhase      map[int]PhaseStat
	WeakestPhase int
}

// Coverage spočítá pokrok kurzem a najde nejslabší fázi.
func Coverage(lessons []LessonResult) Summary {
	summary := Summary{ByPhase: make(map[int]PhaseStat)}
	if len(lessons) == 0 {
		return summary
	}

	for _, l := range lessons {
		summary.Total++
		stat := summary.ByPhase[l.Phase]
		stat.Total++
		if l.Passed {
			summary.Passed++
			stat.Passed++
		}
		summary.ByPhase[l.Phase] = stat
	}
	summary.Percent = float64(summary.Passed) / float64(summary.Total) * 100

	phases := make([]int, 0, len(summary.ByPhase))
	for phase := range summary.ByPhase {
		phases = append(phases, phase)
	}
	sort.Ints(phases)

	worst := 2.0
	for _, phase := range phases {
		stat := summary.ByPhase[phase]
		ratio := float64(stat.Passed) / float64(stat.Total)
		if ratio < worst {
			worst = ratio
			summary.WeakestPhase = phase
		}
	}
	return summary
}

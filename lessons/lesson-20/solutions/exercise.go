// Package solutions obsahuje referenční řešení lekce 20.
package solutions

import (
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"
)

// Výchozí hodnoty klienta, které se použijí, když je nepřepíše žádná Option.
const (
	DefaultTimeout   = 5 * time.Second
	DefaultRetries   = 3
	DefaultUserAgent = "go-deep/1.0"
)

// Chyby vracené konstruktory.
var (
	ErrMissingAddr    = errors.New("missing server address")
	ErrInvalidAddr    = errors.New("invalid server address")
	ErrMissingLogger  = errors.New("missing logger")
	ErrMissingBaseURL = errors.New("missing base url")
	ErrInvalidBaseURL = errors.New("invalid base url")
	ErrInvalidTimeout = errors.New("invalid timeout")
	ErrInvalidRetries = errors.New("invalid retries")
	ErrEmptyUserAgent = errors.New("empty user agent")
)

// Server má povinné závislosti, takže bez konstruktoru nedává smysl.
type Server struct {
	addr   string
	logger *slog.Logger
}

// Client je HTTP klient konfigurovaný přes functional options.
type Client struct {
	baseURL   string
	timeout   time.Duration
	retries   int
	userAgent string
}

// Option mění volitelnou část konfigurace klienta.
type Option func(*Client)

// WithTimeout nastaví timeout klienta jako Option.
func WithTimeout(d time.Duration) Option {
	return func(c *Client) { c.timeout = d }
}

// WithRetries nastaví počet opakování jako Option.
func WithRetries(n int) Option {
	return func(c *Client) { c.retries = n }
}

// WithUserAgent nastaví hlavičku User-Agent jako Option.
func WithUserAgent(ua string) Option {
	return func(c *Client) { c.userAgent = ua }
}

// --- Stupeň: jednoduchý ---

// NewClient vrací klienta s výchozí konfigurací přepsanou zadanými options.
func NewClient(baseURL string, opts ...Option) (*Client, error) {
	if baseURL == "" {
		return nil, ErrMissingBaseURL
	}
	if !strings.HasPrefix(baseURL, "http://") && !strings.HasPrefix(baseURL, "https://") {
		return nil, fmt.Errorf("%w: %q", ErrInvalidBaseURL, baseURL)
	}

	c := &Client{
		baseURL:   strings.TrimRight(baseURL, "/"),
		timeout:   DefaultTimeout,
		retries:   DefaultRetries,
		userAgent: DefaultUserAgent,
	}
	for _, opt := range opts {
		if opt != nil {
			opt(c)
		}
	}

	if c.timeout <= 0 {
		return nil, fmt.Errorf("%w: %v", ErrInvalidTimeout, c.timeout)
	}
	if c.retries < 0 {
		return nil, fmt.Errorf("%w: %d", ErrInvalidRetries, c.retries)
	}
	if c.userAgent == "" {
		return nil, ErrEmptyUserAgent
	}
	return c, nil
}

// BaseURL vrací základní adresu bez koncového lomítka.
func (c *Client) BaseURL() string { return c.baseURL }

// Timeout vrací nastavený timeout klienta.
func (c *Client) Timeout() time.Duration { return c.timeout }

// Retries vrací nastavený počet opakování.
func (c *Client) Retries() int { return c.retries }

// UserAgent vrací nastavenou hlavičku User-Agent.
func (c *Client) UserAgent() string { return c.userAgent }

// --- Stupeň: střední ---

// NewServer ověří povinné závislosti a vrátí připravený server.
func NewServer(addr string, logger *slog.Logger) (*Server, error) {
	if addr == "" {
		return nil, ErrMissingAddr
	}
	if !strings.Contains(addr, ":") {
		return nil, fmt.Errorf("%w: %q", ErrInvalidAddr, addr)
	}
	if logger == nil {
		return nil, ErrMissingLogger
	}
	return &Server{addr: addr, logger: logger}, nil
}

// MustNewServer je varianta NewServer, která při chybě panikuje.
func MustNewServer(addr string, logger *slog.Logger) *Server {
	s, err := NewServer(addr, logger)
	if err != nil {
		panic(err)
	}
	return s
}

// Addr vrací adresu, na které server poslouchá.
func (s *Server) Addr() string { return s.addr }

// Logger vrací logger serveru.
func (s *Server) Logger() *slog.Logger { return s.logger }

// --- Stupeň: obtížný ---

// Registry je typ s užitečnou zero value — funguje i bez konstruktoru.
type Registry struct {
	entries map[string]string
}

// Set uloží hodnotu pod klíč.
func (r *Registry) Set(key, value string) {
	if r.entries == nil {
		r.entries = make(map[string]string)
	}
	r.entries[key] = value
}

// Lookup vrací hodnotu a informaci, jestli klíč existuje.
func (r *Registry) Lookup(key string) (string, bool) {
	v, ok := r.entries[key]
	return v, ok
}

// Len vrací počet uložených klíčů v Registry.
func (r *Registry) Len() int { return len(r.entries) }

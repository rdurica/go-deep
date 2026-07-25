// Package solutions obsahuje referenční řešení lekce 20.
package solutions

import (
	"errors"
	"fmt"
	"log/slog"
	"sort"
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
	ErrMissingStore   = errors.New("missing store")
	ErrEmptyRecordID  = errors.New("empty record id")
)

// Server má povinné závislosti, takže bez konstruktoru nedává smysl.
type Server struct {
	addr   string
	logger *slog.Logger
}

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

// Client je HTTP klient konfigurovaný přes functional options.
type Client struct {
	baseURL   string
	timeout   time.Duration
	retries   int
	userAgent string
}

// Option mění volitelnou část konfigurace klienta.
type Option func(*Client)

// WithTimeout nastaví timeout klienta.
func WithTimeout(d time.Duration) Option {
	return func(c *Client) { c.timeout = d }
}

// WithRetries nastaví počet opakování.
func WithRetries(n int) Option {
	return func(c *Client) { c.retries = n }
}

// WithUserAgent nastaví hlavičku User-Agent.
func WithUserAgent(ua string) Option {
	return func(c *Client) { c.userAgent = ua }
}

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
		if opt == nil {
			continue
		}
		opt(c)
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

// Timeout vrací nastavený timeout.
func (c *Client) Timeout() time.Duration { return c.timeout }

// Retries vrací nastavený počet opakování.
func (c *Client) Retries() int { return c.retries }

// UserAgent vrací nastavenou hlavičku User-Agent.
func (c *Client) UserAgent() string { return c.userAgent }

// Record je jeden uložený záznam.
type Record struct {
	ID    string
	Value string
}

// Store je minimální port, který Service potřebuje. Definuje ho konzument.
type Store interface {
	Save(Record) error
	All() []Record
}

// Service je konkrétní typ postavený nad libovolným Store.
type Service struct {
	store Store
}

// NewService ověří závislost a vrátí službu.
func NewService(store Store) (*Service, error) {
	if store == nil {
		return nil, ErrMissingStore
	}
	return &Service{store: store}, nil
}

// Add uloží nový záznam.
func (s *Service) Add(id, value string) error {
	if id == "" {
		return ErrEmptyRecordID
	}
	if err := s.store.Save(Record{ID: id, Value: value}); err != nil {
		return fmt.Errorf("save record %q: %w", id, err)
	}
	return nil
}

// Count vrací počet záznamů ve Store.
func (s *Service) Count() int { return len(s.store.All()) }

// Values vrací hodnoty všech záznamů v pořadí, v jakém je vrátil Store.
func (s *Service) Values() []string {
	all := s.store.All()
	values := make([]string, 0, len(all))
	for _, r := range all {
		values = append(values, r.Value)
	}
	return values
}

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

// Len vrací počet uložených klíčů.
func (r *Registry) Len() int { return len(r.entries) }

// Keys vrací klíče seřazené vzestupně.
func (r *Registry) Keys() []string {
	keys := make([]string, 0, len(r.entries))
	for k := range r.entries {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

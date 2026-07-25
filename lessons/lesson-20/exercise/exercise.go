// Package exercise obsahuje cvičení lekce 20.
package exercise

import (
	"errors"
	"log/slog"
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
	// TODO: úkol A
	return nil, nil
}

// MustNewServer je varianta NewServer, která při chybě panikuje.
func MustNewServer(addr string, logger *slog.Logger) *Server {
	// TODO: úkol A
	return nil
}

// Addr vrací adresu, na které server poslouchá.
func (s *Server) Addr() string {
	// TODO: úkol A
	return ""
}

// Logger vrací logger serveru.
func (s *Server) Logger() *slog.Logger {
	// TODO: úkol A
	return nil
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

// WithTimeout nastaví timeout klienta.
func WithTimeout(d time.Duration) Option {
	// TODO: úkol B
	return *new(Option)
}

// WithRetries nastaví počet opakování.
func WithRetries(n int) Option {
	// TODO: úkol B
	return *new(Option)
}

// WithUserAgent nastaví hlavičku User-Agent.
func WithUserAgent(ua string) Option {
	// TODO: úkol B
	return *new(Option)
}

// NewClient vrací klienta s výchozí konfigurací přepsanou zadanými options.
func NewClient(baseURL string, opts ...Option) (*Client, error) {
	// TODO: úkol B
	return nil, nil
}

// BaseURL vrací základní adresu bez koncového lomítka.
func (c *Client) BaseURL() string {
	// TODO: úkol B
	return ""
}

// Timeout vrací nastavený timeout.
func (c *Client) Timeout() time.Duration {
	// TODO: úkol B
	return *new(time.Duration)
}

// Retries vrací nastavený počet opakování.
func (c *Client) Retries() int {
	// TODO: úkol B
	return 0
}

// UserAgent vrací nastavenou hlavičku User-Agent.
func (c *Client) UserAgent() string {
	// TODO: úkol B
	return ""
}

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
	// TODO: úkol C
	return nil, nil
}

// Add uloží nový záznam.
func (s *Service) Add(id, value string) error {
	// TODO: úkol C
	return nil
}

// Count vrací počet záznamů ve Store.
func (s *Service) Count() int {
	// TODO: úkol C
	return 0
}

// Values vrací hodnoty všech záznamů v pořadí, v jakém je vrátil Store.
func (s *Service) Values() []string {
	// TODO: úkol C
	return nil
}

// Registry je typ s užitečnou zero value — funguje i bez konstruktoru.
type Registry struct {
	entries map[string]string
}

// Set uloží hodnotu pod klíč.
func (r *Registry) Set(key, value string) {
	// TODO: úkol C
}

// Lookup vrací hodnotu a informaci, jestli klíč existuje.
func (r *Registry) Lookup(key string) (string, bool) {
	// TODO: úkol C
	return "", false
}

// Len vrací počet uložených klíčů.
func (r *Registry) Len() int {
	// TODO: úkol C
	return 0
}

// Keys vrací klíče seřazené vzestupně.
func (r *Registry) Keys() []string {
	// TODO: úkol C
	return nil
}

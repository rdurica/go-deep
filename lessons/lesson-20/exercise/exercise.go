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
	panic("TODO: úkol A")
}

// MustNewServer je varianta NewServer, která při chybě panikuje.
func MustNewServer(addr string, logger *slog.Logger) *Server {
	panic("TODO: úkol A")
}

// Addr vrací adresu, na které server poslouchá.
func (s *Server) Addr() string {
	panic("TODO: úkol A")
}

// Logger vrací logger serveru.
func (s *Server) Logger() *slog.Logger {
	panic("TODO: úkol A")
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
	panic("TODO: úkol B")
}

// WithRetries nastaví počet opakování.
func WithRetries(n int) Option {
	panic("TODO: úkol B")
}

// WithUserAgent nastaví hlavičku User-Agent.
func WithUserAgent(ua string) Option {
	panic("TODO: úkol B")
}

// NewClient vrací klienta s výchozí konfigurací přepsanou zadanými options.
func NewClient(baseURL string, opts ...Option) (*Client, error) {
	panic("TODO: úkol B")
}

// BaseURL vrací základní adresu bez koncového lomítka.
func (c *Client) BaseURL() string {
	panic("TODO: úkol B")
}

// Timeout vrací nastavený timeout.
func (c *Client) Timeout() time.Duration {
	panic("TODO: úkol B")
}

// Retries vrací nastavený počet opakování.
func (c *Client) Retries() int {
	panic("TODO: úkol B")
}

// UserAgent vrací nastavenou hlavičku User-Agent.
func (c *Client) UserAgent() string {
	panic("TODO: úkol B")
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
	panic("TODO: úkol C")
}

// Add uloží nový záznam.
func (s *Service) Add(id, value string) error {
	panic("TODO: úkol C")
}

// Count vrací počet záznamů ve Store.
func (s *Service) Count() int {
	panic("TODO: úkol C")
}

// Values vrací hodnoty všech záznamů v pořadí, v jakém je vrátil Store.
func (s *Service) Values() []string {
	panic("TODO: úkol C")
}

// Registry je typ s užitečnou zero value — funguje i bez konstruktoru.
type Registry struct {
	entries map[string]string
}

// Set uloží hodnotu pod klíč.
func (r *Registry) Set(key, value string) {
	panic("TODO: úkol C")
}

// Lookup vrací hodnotu a informaci, jestli klíč existuje.
func (r *Registry) Lookup(key string) (string, bool) {
	panic("TODO: úkol C")
}

// Len vrací počet uložených klíčů.
func (r *Registry) Len() int {
	panic("TODO: úkol C")
}

// Keys vrací klíče seřazené vzestupně.
func (r *Registry) Keys() []string {
	panic("TODO: úkol C")
}

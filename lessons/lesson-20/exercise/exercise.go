// Package exercise obsahuje cvičení lekce 20.
package exercise

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
// Validace baseURL před vytvořením: prázdný → ErrMissingBaseURL; bez http(s) → ErrInvalidBaseURL.
// Po options: timeout > 0, retries >= 0, user agent neprázdný. Při chybě nil klient.
//
// POZOR: kód níže je ZÁMĚRNĚ VADNÝ. Validuje stav před aplikací options.
// Najdi chybu a oprav — testy před opravou padají.
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

	// Špatně: validace výchozích hodnot před options — option může je rozbít.
	if c.timeout <= 0 {
		return nil, fmt.Errorf("%w: %v", ErrInvalidTimeout, c.timeout)
	}
	if c.retries < 0 {
		return nil, fmt.Errorf("%w: %d", ErrInvalidRetries, c.retries)
	}
	if c.userAgent == "" {
		return nil, ErrEmptyUserAgent
	}

	for _, opt := range opts {
		if opt != nil {
			opt(c)
		}
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
// Prázdná addr → ErrMissingAddr; addr bez ':' → chyba obalující ErrInvalidAddr.
// Nil logger → ErrMissingLogger. Jinak připravený *Server.
func NewServer(addr string, logger *slog.Logger) (*Server, error) {
	// TODO
	return nil, nil
}

// MustNewServer je varianta NewServer, která při chybě panikuje.
// Panikuje s chybou (panic(err)), ne panic(err.Error()).
func MustNewServer(addr string, logger *slog.Logger) *Server {
	// TODO
	return nil
}

// Addr vrací adresu, na které server poslouchá (getter bez prefixu Get).
func (s *Server) Addr() string {
	// TODO
	return ""
}

// Logger vrací logger serveru (getter bez prefixu Get).
func (s *Server) Logger() *slog.Logger {
	// TODO
	return nil
}

// --- Stupeň: obtížný ---

// Registry je typ s užitečnou zero value — funguje i bez konstruktoru.
type Registry struct {
	entries map[string]string
}

// Set uloží hodnotu pod klíč v Registry.
// Líná inicializace mapy; přepis existujícího klíče je povolený.
func (r *Registry) Set(key, value string) {
	// TODO
}

// Lookup vrací hodnotu a informaci, jestli klíč existuje.
// Musí fungovat i na nulové Registry (čtení z nil mapy bez paniky).
func (r *Registry) Lookup(key string) (string, bool) {
	// TODO
	return "", false
}

// Len vrací počet uložených klíčů v Registry.
// Nulová Registry (nil mapa) vrací 0 bez paniky.
func (r *Registry) Len() int {
	// TODO
	return 0
}

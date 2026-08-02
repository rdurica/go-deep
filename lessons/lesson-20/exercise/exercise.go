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

// --- Stupeň: jednoduchý ---
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
// Options se aplikují v pořadí; poslední vyhrává. Nil option přeskoč.
func WithTimeout(d time.Duration) Option {
	// TODO
	return *new(Option)
}

// WithRetries nastaví počet opakování jako Option.
// Hodnota 0 je platná; záporná se validuje až v NewClient po aplikaci options.
func WithRetries(n int) Option {
	// TODO
	return *new(Option)
}

// --- Stupeň: střední ---
// WithUserAgent nastaví hlavičku User-Agent jako Option.
// Prázdný agent po aplikaci všech options je chyba v NewClient (ErrEmptyUserAgent).
func WithUserAgent(ua string) Option {
	// TODO
	return *new(Option)
}

// NewClient vrací klienta s výchozí konfigurací přepsanou zadanými options.
// Validace baseURL před vytvořením: prázdný → ErrMissingBaseURL; bez http(s) → ErrInvalidBaseURL.
// Po options: timeout > 0, retries >= 0, user agent neprázdný. Při chybě nil klient.
func NewClient(baseURL string, opts ...Option) (*Client, error) {
	// TODO
	return nil, nil
}

// BaseURL vrací základní adresu bez koncového lomítka.
// Uložená hodnota je normalizovaná (https://x.com/ → https://x.com).
func (c *Client) BaseURL() string {
	// TODO
	return ""
}

// Timeout vrací nastavený timeout klienta (výchozí DefaultTimeout).
func (c *Client) Timeout() time.Duration {
	// TODO
	return *new(time.Duration)
}

// Retries vrací nastavený počet opakování (výchozí DefaultRetries; 0 je platná hodnota).
func (c *Client) Retries() int {
	// TODO
	return 0
}

// UserAgent vrací nastavenou hlavičku User-Agent (výchozí DefaultUserAgent).
func (c *Client) UserAgent() string {
	// TODO
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

// NewService ověří závislost a vrátí službu nad portem Store.
// Nil store → ErrMissingStore. Vrací *Service (konkrétní typ), ne interface.
func NewService(store Store) (*Service, error) {
	// TODO
	return nil, nil
}

// --- Stupeň: obtížný ---
// Add uloží nový záznam do Store.
// Prázdné id → ErrEmptyRecordID a nic se neukládá.
// Chybu Save obal: fmt.Errorf("save record %q: %w", id, err).
func (s *Service) Add(id, value string) error {
	// TODO
	return nil
}

// Count vrací počet záznamů ve Store podle store.All().
// Prázdný store → 0, ne panika.
func (s *Service) Count() int {
	// TODO
	return 0
}

// Values vrací hodnoty všech záznamů v pořadí store.All(), ne abecedně.
// Prázdný store → prázdný slice (len 0), ne panika.
func (s *Service) Values() []string {
	// TODO
	return nil
}

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

// Keys vrací klíče seřazené vzestupně. Na prázdné i nulové Registry prázdný slice.
// Funguje bez paniky i na nil mapě. Vrací nový slice (ne sdílený interní stav).
func (r *Registry) Keys() []string {
	// TODO
	return nil
}

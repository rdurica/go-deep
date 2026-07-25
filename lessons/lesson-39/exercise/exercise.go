// Package exercise obsahuje kumulativní cvičení lekce 39.
//
// Balíček hraje roli domény "booking" i jejích adaptérů dohromady —
// v reálném projektu by to byly balíčky booking, store a httpapi.
package exercise

import (
	"context"
	"errors"
	"net/http"
	"sync"
	"time"
)

// Chyby hodnotových typů.
var (
	// ErrEmptyRoomID vrací ParseRoomID pro prázdný vstup.
	ErrEmptyRoomID = errors.New("číslo pokoje je prázdné")
	// ErrInvalidRoomID vrací ParseRoomID pro vstup mimo tvar "A-101".
	ErrInvalidRoomID = errors.New("číslo pokoje má neplatný tvar")
	// ErrInvalidRange vrací NewDateRange, když konec není po začátku.
	ErrInvalidRange = errors.New("konec pobytu musí být po začátku")
	// ErrRangeTooLong vrací NewDateRange pro pobyt delší než MaxNights.
	ErrRangeTooLong = errors.New("pobyt je příliš dlouhý")
)

// Doménové chyby případu užití.
var (
	// ErrRoomTaken říká, že pokoj je v daném termínu obsazený.
	ErrRoomTaken = errors.New("pokoj je v tomto termínu obsazený")
	// ErrDuplicateRef říká, že rezervace s touto referencí už existuje.
	ErrDuplicateRef = errors.New("rezervace s touto referencí už existuje")
)

// MaxNights je nejdelší povolený pobyt.
const MaxNights = 30

// DateLayout je formát data na hranici systému.
const DateLayout = "2006-01-02"

// Kódy validačních chyb.
const (
	CodeRequired = "required"
	CodeFormat   = "format"
	CodeRange    = "range"
)

// Názvy metrik. Jednotka i sufix _total jsou součástí názvu.
const (
	MetricCreated  = "bookings_created_total"
	MetricRejected = "bookings_rejected_total"
)

// RoomID je ověřené číslo pokoje ve tvaru "A-101".
type RoomID string

// ParseRoomID normalizuje a ověří číslo pokoje.
// Tvar je jedno písmeno, pomlčka a tři číslice; "A-000" neexistuje.
func ParseRoomID(s string) (RoomID, error) {
	// TODO: úkol A
	return *new(RoomID), nil
}

// DateRange je polootevřený interval pobytu <From, To).
type DateRange struct {
	from time.Time
	to   time.Time
}

// NewDateRange ověří termín a zarovná ho na celé dny v UTC.
func NewDateRange(from, to time.Time) (DateRange, error) {
	// TODO: úkol A
	return *new(DateRange), nil
}

// From vrací první noc pobytu.
func (d DateRange) From() time.Time { return d.from }

// To vrací den odjezdu; ten se už nepočítá.
func (d DateRange) To() time.Time { return d.to }

// Nights vrací počet nocí.
func (d DateRange) Nights() int {
	// TODO: úkol A
	return 0
}

// Overlaps vrací true, pokud se termíny překrývají.
// Interval je polootevřený, takže odjezd v den příjezdu dalšího hosta je v pořádku.
func (d DateRange) Overlaps(other DateRange) bool {
	// TODO: úkol A
	return false
}

// String vrací termín ve tvaru "2024-05-17..2024-05-20".
func (d DateRange) String() string {
	return d.from.Format(DateLayout) + ".." + d.to.Format(DateLayout)
}

// Booking je rezervace pokoje.
type Booking struct {
	Ref   string
	Room  RoomID
	Stay  DateRange
	Guest string
	Total int64
}

// FieldError popisuje jednu zamítnutou položku vstupu.
type FieldError struct {
	Field   string `json:"field"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

// ValidationErrors je seznam zamítnutých položek; implementuje error.
type ValidationErrors []FieldError

// Error implementuje error.
func (v ValidationErrors) Error() string {
	// TODO: úkol C
	return ""
}

// Get vrátí chybu pro dané pole, pokud tam je.
func (v ValidationErrors) Get(field string) (FieldError, bool) {
	// TODO: úkol C
	return *new(FieldError), false
}

// CreateBookingRequest je DTO na hranici systému.
type CreateBookingRequest struct {
	Ref         string `json:"ref"`
	Room        string `json:"room"`
	Guest       string `json:"guest"`
	From        string `json:"from"`
	To          string `json:"to"`
	NightlyRate int64  `json:"nightly_rate"`
}

// Validate ověří všechna pole najednou a vrátí ValidationErrors,
// nebo nil, pokud je požadavek v pořádku.
func (r CreateBookingRequest) Validate() error {
	// TODO: úkol C
	return nil
}

// Repository je port persistence, definovaný u svého konzumenta.
type Repository interface {
	Save(ctx context.Context, b Booking) error
	ByRoom(ctx context.Context, room RoomID) ([]Booking, error)
}

// MemoryRepo je in-memory adaptér portu Repository.
type MemoryRepo struct {
	mu       sync.RWMutex
	bookings map[string]Booking
}

// NewMemoryRepo vytvoří prázdné úložiště.
func NewMemoryRepo() *MemoryRepo {
	// TODO: úkol B
	return nil
}

// Save uloží rezervaci. Duplicitní reference vrací ErrDuplicateRef.
func (r *MemoryRepo) Save(ctx context.Context, b Booking) error {
	// TODO: úkol B
	return nil
}

// ByRoom vrátí rezervace pokoje seřazené podle začátku pobytu.
func (r *MemoryRepo) ByRoom(ctx context.Context, room RoomID) ([]Booking, error) {
	// TODO: úkol B
	return nil, nil
}

// Metrics je minimální registr čítačů. Nulová hodnota je použitelná.
type Metrics struct {
	mu       sync.Mutex
	counters map[string]int64
}

// Inc zvýší čítač o jedna.
func (m *Metrics) Inc(name string) {
	// TODO: úkol B
}

// Snapshot vrátí kopii čítačů.
func (m *Metrics) Snapshot() map[string]int64 {
	// TODO: úkol B
	return nil
}

// Service je aplikační služba rezervací.
type Service struct {
	repo    Repository
	metrics *Metrics
}

// NewService složí službu nad portem. Metriky jsou volitelné.
func NewService(repo Repository, metrics *Metrics) *Service {
	if metrics == nil {
		metrics = &Metrics{}
	}
	return &Service{repo: repo, metrics: metrics}
}

// Metrics vrací registr metrik služby.
func (s *Service) Metrics() *Metrics { return s.metrics }

// Book ověří požadavek, zkontroluje obsazenost a rezervaci uloží.
func (s *Service) Book(ctx context.Context, req CreateBookingRequest) (Booking, error) {
	// TODO: úkol B
	return *new(Booking), nil
}

// ProblemContentType je Content-Type chybové odpovědi podle RFC 7807.
const ProblemContentType = "application/problem+json"

// Problem je zjednodušené tělo chybové odpovědi podle RFC 7807.
type Problem struct {
	Type   string           `json:"type"`
	Title  string           `json:"title"`
	Status int              `json:"status"`
	Errors ValidationErrors `json:"errors,omitempty"`
}

// Handler je HTTP adaptér nad službou.
func Handler(svc *Service) http.Handler {
	// TODO: úkol C
	return *new(http.Handler)
}

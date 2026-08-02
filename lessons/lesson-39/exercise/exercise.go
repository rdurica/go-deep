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

// --- Stupeň: jednoduchý ---
// ParseRoomID ořízne, převede na velká písmena, ověří tvar písmeno-pomlčka-3 číslice.
// Prázdný → ErrEmptyRoomID; A-000 a jiný tvar → ErrInvalidRoomID. Při chybě nulová RoomID.
func ParseRoomID(s string) (RoomID, error) {
	// TODO
	return *new(RoomID), nil
}

// DateRange je polootevřený interval pobytu <From, To).
type DateRange struct {
	from time.Time
	to   time.Time
}

// NewDateRange zarovná oba časy na celý den UTC. Konec musí být po začátku;
// max MaxNights nocí. ErrInvalidRange / ErrRangeTooLong.
func NewDateRange(from, to time.Time) (DateRange, error) {
	// TODO
	return *new(DateRange), nil
}

// From vrací první noc pobytu.
// Součást polootevřeného intervalu <From, To).
func (d DateRange) From() time.Time { return d.from }

// To vrací den odjezdu; ten se už nepočítá.
func (d DateRange) To() time.Time { return d.to }

// Nights vrací počet nocí mezi From a To.
// Počítá celé dny v UTC po zarovnání v NewDateRange.
func (d DateRange) Nights() int {
	// TODO
	return 0
}

// Overlaps vrací true při překryvu polootevřených intervalů.
// Odjezd v den příjezdu dalšího hosta se nepřekrývá.
func (d DateRange) Overlaps(other DateRange) bool {
	// TODO
	return false
}

// --- Stupeň: střední ---
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

// Error implementuje error pro ValidationErrors.
// Vrací souhrnný text; detailní pole získáš přes Get(field).
func (v ValidationErrors) Error() string {
	// TODO
	return ""
}

// Get vrátí chybu pro dané pole, pokud existuje v seznamu validačních chyb.
// Neexistující pole → zero value FieldError a false.
func (v ValidationErrors) Get(field string) (FieldError, bool) {
	// TODO
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

// Validate sbírá všechny chyby v pořadí ref, room, guest, from, to, nightly_rate.
// Kódy CodeRequired, CodeFormat, CodeRange. Chyba termínu patří k poli to.
// Datum ve tvaru DateLayout. Vrať nil, ne typovaný nil slice.
func (r CreateBookingRequest) Validate() error {
	// TODO
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

// NewMemoryRepo vytvoří prázdné úložiště. Bezpečné pro souběžné použití.
// Duplicitní Ref při Save vrací ErrDuplicateRef.
func NewMemoryRepo() *MemoryRepo {
	// TODO
	return nil
}

// Save uloží rezervaci do úložiště.
// Duplicitní Ref → ErrDuplicateRef. Zrušený kontext vrací chybu z ctx.Err().
func (r *MemoryRepo) Save(ctx context.Context, b Booking) error {
	// TODO
	return nil
}

// --- Stupeň: obtížný ---
// ByRoom vrátí rezervace pokoje seřazené podle začátku pobytu.
// Zrušený kontext vrací chybu z ctx.Err().
func (r *MemoryRepo) ByRoom(ctx context.Context, room RoomID) ([]Booking, error) {
	// TODO
	return nil, nil
}

// Metrics je minimální registr čítačů. Nulová hodnota je použitelná.
type Metrics struct {
	mu       sync.Mutex
	counters map[string]int64
}

// Inc zvýší čítač name o jedna pod mutexem.
// Nulová hodnota Metrics je použitelná: var m Metrics a rovnou Inc.
func (m *Metrics) Inc(name string) {
	// TODO
}

// Snapshot vrátí kopii čítačů jako novou mapu (ne sdílenou referenci na interní mapu).
func (m *Metrics) Snapshot() map[string]int64 {
	// TODO
	return nil
}

// Service je aplikační služba rezervací.
type Service struct {
	repo    Repository
	metrics *Metrics
}

// NewService složí službu nad portem. Metriky volitelné (nil → prázdné Metrics).
func NewService(repo Repository, metrics *Metrics) *Service {
	if metrics == nil {
		metrics = &Metrics{}
	}
	return &Service{repo: repo, metrics: metrics}
}

// Metrics vrací registr metrik služby.
// Nikdy nevrací nil — nil v konstruktoru se nahradí prázdným Metrics.
func (s *Service) Metrics() *Metrics { return s.metrics }

// Book zvaliduje req, převede na typy, zkontroluje překryv (ErrRoomTaken),
// spočítá nightly_rate × noci a uloží. Metriky: MetricCreated po úspěchu,
// MetricRejected při validaci, překryvu i duplicitní Ref.
func (s *Service) Book(ctx context.Context, req CreateBookingRequest) (Booking, error) {
	// TODO
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

// Handler je HTTP adaptér: POST /bookings (201, Location, JSON s nights a total),
// GET /metrics. Mapování 400/422/409/500, problem+json.
func Handler(svc *Service) http.Handler {
	// TODO
	return *new(http.Handler)
}

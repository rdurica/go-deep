// Package solutions obsahuje referenční řešení lekce 39.
//
// Balíček hraje roli domény "booking" i jejích adaptérů dohromady —
// v reálném projektu by to byly balíčky booking, store a httpapi.
package solutions

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
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
	s = strings.ToUpper(strings.TrimSpace(s))
	if s == "" {
		return "", ErrEmptyRoomID
	}
	if len(s) != 5 || s[1] != '-' {
		return "", fmt.Errorf("%w: %q", ErrInvalidRoomID, s)
	}
	if s[0] < 'A' || s[0] > 'Z' {
		return "", fmt.Errorf("%w: %q nezačíná písmenem", ErrInvalidRoomID, s)
	}
	for i := 2; i < 5; i++ {
		if s[i] < '0' || s[i] > '9' {
			return "", fmt.Errorf("%w: %q", ErrInvalidRoomID, s)
		}
	}
	if s[2:] == "000" {
		return "", fmt.Errorf("%w: pokoj %q neexistuje", ErrInvalidRoomID, s)
	}
	return RoomID(s), nil
}

// DateRange je polootevřený interval pobytu <From, To).
type DateRange struct {
	from time.Time
	to   time.Time
}

// NewDateRange ověří termín a zarovná ho na celé dny v UTC.
func NewDateRange(from, to time.Time) (DateRange, error) {
	f, t := truncateDay(from), truncateDay(to)
	if !t.After(f) {
		return DateRange{}, fmt.Errorf("%w: %s–%s", ErrInvalidRange, f.Format(DateLayout), t.Format(DateLayout))
	}
	nights := int(t.Sub(f) / (24 * time.Hour))
	if nights > MaxNights {
		return DateRange{}, fmt.Errorf("%w: %d nocí, maximum je %d", ErrRangeTooLong, nights, MaxNights)
	}
	return DateRange{from: f, to: t}, nil
}

func truncateDay(t time.Time) time.Time {
	t = t.UTC()
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
}

// From vrací první noc pobytu.
func (d DateRange) From() time.Time { return d.from }

// To vrací den odjezdu; ten se už nepočítá.
func (d DateRange) To() time.Time { return d.to }

// Nights vrací počet nocí.
func (d DateRange) Nights() int {
	return int(d.to.Sub(d.from) / (24 * time.Hour))
}

// Overlaps vrací true, pokud se termíny překrývají.
// Interval je polootevřený, takže odjezd v den příjezdu dalšího hosta je v pořádku.
func (d DateRange) Overlaps(other DateRange) bool {
	return d.from.Before(other.to) && other.from.Before(d.to)
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
	if len(v) == 0 {
		return "validace prošla"
	}
	parts := make([]string, 0, len(v))
	for _, fe := range v {
		parts = append(parts, fmt.Sprintf("%s: %s (%s)", fe.Field, fe.Message, fe.Code))
	}
	return "validace selhala: " + strings.Join(parts, "; ")
}

// Get vrátí chybu pro dané pole, pokud tam je.
func (v ValidationErrors) Get(field string) (FieldError, bool) {
	for _, fe := range v {
		if fe.Field == field {
			return fe, true
		}
	}
	return FieldError{}, false
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
	var errs ValidationErrors

	if strings.TrimSpace(r.Ref) == "" {
		errs = append(errs, FieldError{"ref", CodeRequired, "reference je povinná"})
	}

	switch _, err := ParseRoomID(r.Room); {
	case errors.Is(err, ErrEmptyRoomID):
		errs = append(errs, FieldError{"room", CodeRequired, "pokoj je povinný"})
	case err != nil:
		errs = append(errs, FieldError{"room", CodeFormat, "pokoj musí mít tvar A-101"})
	}

	guest := strings.TrimSpace(r.Guest)
	switch {
	case guest == "":
		errs = append(errs, FieldError{"guest", CodeRequired, "jméno hosta je povinné"})
	case len(guest) < 2 || len(guest) > 40:
		errs = append(errs, FieldError{"guest", CodeFormat, "jméno hosta smí mít 2–40 znaků"})
	}

	from, fromErr := parseDate("from", r.From, &errs)
	to, toErr := parseDate("to", r.To, &errs)
	if fromErr == nil && toErr == nil {
		if _, err := NewDateRange(from, to); err != nil {
			errs = append(errs, FieldError{"to", CodeRange,
				fmt.Sprintf("termín musí končit po začátku a trvat nejvýš %d nocí", MaxNights)})
		}
	}

	if r.NightlyRate <= 0 {
		errs = append(errs, FieldError{"nightly_rate", CodeRange, "cena za noc musí být kladná"})
	}

	if len(errs) == 0 {
		return nil
	}
	return errs
}

func parseDate(field, value string, errs *ValidationErrors) (time.Time, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		*errs = append(*errs, FieldError{field, CodeRequired, "datum je povinné"})
		return time.Time{}, errors.New("prázdné datum")
	}
	t, err := time.Parse(DateLayout, value)
	if err != nil {
		*errs = append(*errs, FieldError{field, CodeFormat, "datum musí být ve tvaru RRRR-MM-DD"})
		return time.Time{}, err
	}
	return t, nil
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
	return &MemoryRepo{bookings: make(map[string]Booking)}
}

// Save uloží rezervaci. Duplicitní reference vrací ErrDuplicateRef.
func (r *MemoryRepo) Save(ctx context.Context, b Booking) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.bookings[b.Ref]; ok {
		return fmt.Errorf("%w: %q", ErrDuplicateRef, b.Ref)
	}
	r.bookings[b.Ref] = b
	return nil
}

// ByRoom vrátí rezervace pokoje seřazené podle začátku pobytu.
func (r *MemoryRepo) ByRoom(ctx context.Context, room RoomID) ([]Booking, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	r.mu.RLock()
	defer r.mu.RUnlock()

	var out []Booking
	for _, b := range r.bookings {
		if b.Room == room {
			out = append(out, b)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Stay.From().Before(out[j].Stay.From()) })
	return out, nil
}

// Metrics je minimální registr čítačů. Nulová hodnota je použitelná.
type Metrics struct {
	mu       sync.Mutex
	counters map[string]int64
}

// Inc zvýší čítač o jedna.
func (m *Metrics) Inc(name string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.counters == nil {
		m.counters = make(map[string]int64)
	}
	m.counters[name]++
}

// Snapshot vrátí kopii čítačů.
func (m *Metrics) Snapshot() map[string]int64 {
	m.mu.Lock()
	defer m.mu.Unlock()

	out := make(map[string]int64, len(m.counters))
	for k, v := range m.counters {
		out[k] = v
	}
	return out
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
	if err := req.Validate(); err != nil {
		s.metrics.Inc(MetricRejected)
		return Booking{}, err
	}

	// Po validaci konstruktory selhat nemohou — to je smysl parse-don't-validate.
	room, err := ParseRoomID(req.Room)
	if err != nil {
		return Booking{}, err
	}
	from, err := time.Parse(DateLayout, strings.TrimSpace(req.From))
	if err != nil {
		return Booking{}, err
	}
	to, err := time.Parse(DateLayout, strings.TrimSpace(req.To))
	if err != nil {
		return Booking{}, err
	}
	stay, err := NewDateRange(from, to)
	if err != nil {
		return Booking{}, err
	}

	existing, err := s.repo.ByRoom(ctx, room)
	if err != nil {
		return Booking{}, fmt.Errorf("načtení rezervací pokoje %s: %w", room, err)
	}
	for _, b := range existing {
		if b.Stay.Overlaps(stay) {
			s.metrics.Inc(MetricRejected)
			return Booking{}, fmt.Errorf("%w: %s v termínu %s", ErrRoomTaken, room, stay)
		}
	}

	booking := Booking{
		Ref:   strings.TrimSpace(req.Ref),
		Room:  room,
		Stay:  stay,
		Guest: strings.TrimSpace(req.Guest),
		Total: req.NightlyRate * int64(stay.Nights()),
	}
	if err := s.repo.Save(ctx, booking); err != nil {
		if errors.Is(err, ErrDuplicateRef) {
			s.metrics.Inc(MetricRejected)
		}
		return Booking{}, fmt.Errorf("uložení rezervace %q: %w", booking.Ref, err)
	}

	s.metrics.Inc(MetricCreated)
	return booking, nil
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

// errMalformedJSON je interní chyba dekódování těla požadavku.
var errMalformedJSON = errors.New("neplatné JSON tělo")

// Handler je HTTP adaptér nad službou.
func Handler(svc *Service) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /bookings", func(w http.ResponseWriter, r *http.Request) {
		var req CreateBookingRequest
		if err := decodeJSON(r, &req); err != nil {
			writeProblem(w, http.StatusBadRequest, "Neplatné tělo požadavku", nil)
			return
		}

		booking, err := svc.Book(r.Context(), req)
		if err != nil {
			var errs ValidationErrors
			switch {
			case errors.As(err, &errs):
				writeProblem(w, http.StatusUnprocessableEntity, "Neplatná data rezervace", errs)
			case errors.Is(err, ErrRoomTaken):
				writeProblem(w, http.StatusConflict, "Pokoj je obsazený", nil)
			case errors.Is(err, ErrDuplicateRef):
				writeProblem(w, http.StatusConflict, "Rezervace už existuje", nil)
			default:
				writeProblem(w, http.StatusInternalServerError, "Interní chyba serveru", nil)
			}
			return
		}

		w.Header().Set("Location", "/bookings/"+booking.Ref)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"ref":    booking.Ref,
			"room":   string(booking.Room),
			"guest":  booking.Guest,
			"from":   booking.Stay.From().Format(DateLayout),
			"to":     booking.Stay.To().Format(DateLayout),
			"nights": booking.Stay.Nights(),
			"total":  booking.Total,
		})
	})

	mux.HandleFunc("GET /metrics", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(svc.Metrics().Snapshot())
	})

	return mux
}

func writeProblem(w http.ResponseWriter, status int, title string, errs ValidationErrors) {
	w.Header().Set("Content-Type", ProblemContentType)
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(Problem{
		Type:   "about:blank",
		Title:  title,
		Status: status,
		Errors: errs,
	})
}

func decodeJSON(r *http.Request, dst any) error {
	dec := json.NewDecoder(io.LimitReader(r.Body, 1<<20))
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		return errMalformedJSON
	}
	if err := dec.Decode(new(json.RawMessage)); !errors.Is(err, io.EOF) {
		return errMalformedJSON
	}
	return nil
}

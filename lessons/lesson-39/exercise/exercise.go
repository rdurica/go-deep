// Package exercise obsahuje kumulativní checkpoint fáze 4.
package exercise

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

var (
	ErrEmptyRoomID   = errors.New("číslo pokoje je prázdné")
	ErrInvalidRoomID = errors.New("číslo pokoje má neplatný tvar")
	ErrInvalidRange  = errors.New("konec pobytu musí být po začátku")
	ErrRangeTooLong  = errors.New("pobyt je příliš dlouhý")
	ErrRoomTaken     = errors.New("pokoj je v tomto termínu obsazený")
	ErrDuplicateRef  = errors.New("rezervace s touto referencí už existuje")
)

const (
	MaxNights          = 30
	DateLayout         = "2006-01-02"
	CodeRequired       = "required"
	CodeFormat         = "format"
	CodeRange          = "range"
	MetricCreated      = "bookings_created_total"
	MetricRejected     = "bookings_rejected_total"
	ProblemContentType = "application/problem+json"
)

type RoomID string

// --- Stupeň: jednoduchý ---

// ParseRoomID ořízne, převede na velká písmena, ověří tvar písmeno-pomlčka-3 číslice.
func ParseRoomID(s string) (RoomID, error) {
	// TODO
	return *new(RoomID), nil
}

type DateRange struct {
	from time.Time
	to   time.Time
}

func NewDateRange(from, to time.Time) (DateRange, error) {
	f, t := truncateDay(from), truncateDay(to)
	if !t.After(f) {
		return DateRange{}, ErrInvalidRange
	}
	if int(t.Sub(f)/(24*time.Hour)) > MaxNights {
		return DateRange{}, ErrRangeTooLong
	}
	return DateRange{from: f, to: t}, nil
}

func truncateDay(t time.Time) time.Time {
	t = t.UTC()
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
}

func (d DateRange) From() time.Time { return d.from }
func (d DateRange) To() time.Time   { return d.to }
func (d DateRange) Nights() int     { return int(d.to.Sub(d.from) / (24 * time.Hour)) }

// Overlaps vrací true při překryvu polootevřených intervalů <From, To).
func (d DateRange) Overlaps(other DateRange) bool {
	// TODO
	return false
}

// --- Stupeň: střední ---

func (d DateRange) String() string {
	return d.from.Format(DateLayout) + ".." + d.to.Format(DateLayout)
}

type Booking struct {
	Ref   string
	Room  RoomID
	Stay  DateRange
	Guest string
	Total int64
}

type FieldError struct {
	Field   string `json:"field"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

type ValidationErrors []FieldError

func (v ValidationErrors) Error() string {
	if len(v) == 0 {
		return "validace prošla"
	}
	var parts []string
	for _, fe := range v {
		parts = append(parts, fe.Field+": "+fe.Message)
	}
	return strings.Join(parts, "; ")
}

func (v ValidationErrors) Get(field string) (FieldError, bool) {
	for _, fe := range v {
		if fe.Field == field {
			return fe, true
		}
	}
	return FieldError{}, false
}

type CreateBookingRequest struct {
	Ref         string `json:"ref"`
	Room        string `json:"room"`
	Guest       string `json:"guest"`
	From        string `json:"from"`
	To          string `json:"to"`
	NightlyRate int64  `json:"nightly_rate"`
}

// Validate sbírá všechny chyby. Kódy CodeRequired, CodeFormat, CodeRange.
// Vrať nil, ne typovaný nil slice.
func (r CreateBookingRequest) Validate() error {
	// TODO
	return nil
}

type Repository interface {
	Save(ctx context.Context, b Booking) error
	ByRoom(ctx context.Context, room RoomID) ([]Booking, error)
}

type MemoryRepo struct {
	mu       sync.RWMutex
	bookings map[string]Booking
}

func NewMemoryRepo() *MemoryRepo {
	return &MemoryRepo{bookings: make(map[string]Booking)}
}

func (r *MemoryRepo) Save(ctx context.Context, b Booking) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, dup := r.bookings[b.Ref]; dup {
		return ErrDuplicateRef
	}
	r.bookings[b.Ref] = b
	return nil
}

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
	return out, nil
}

// --- Stupeň: obtížný ---

type Metrics struct {
	mu       sync.Mutex
	counters map[string]int64
}

func (m *Metrics) Inc(name string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.counters == nil {
		m.counters = make(map[string]int64)
	}
	m.counters[name]++
}

func (m *Metrics) Snapshot() map[string]int64 {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make(map[string]int64, len(m.counters))
	for k, v := range m.counters {
		out[k] = v
	}
	return out
}

type Service struct {
	repo    Repository
	metrics *Metrics
}

func NewService(repo Repository, metrics *Metrics) *Service {
	if metrics == nil {
		metrics = &Metrics{}
	}
	return &Service{repo: repo, metrics: metrics}
}

func (s *Service) Metrics() *Metrics { return s.metrics }

// Book zvaliduje req, zkontroluje překryv (ErrRoomTaken), spočítá total a uloží.
// Metriky: MetricCreated po úspěchu, MetricRejected při chybě validace/překryvu/duplicitě.
func (s *Service) Book(ctx context.Context, req CreateBookingRequest) (Booking, error) {
	// TODO
	return *new(Booking), nil
}

// Handler je hotový HTTP adaptér — neměň (checkpoint ověřuje Book, ne routing).
func Handler(svc *Service) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /bookings", func(w http.ResponseWriter, r *http.Request) {
		var req CreateBookingRequest
		dec := json.NewDecoder(io.LimitReader(r.Body, 1<<20))
		dec.DisallowUnknownFields()
		if err := dec.Decode(&req); err != nil {
			writeProblem(w, 400, "Neplatný požadavek")
			return
		}
		b, err := svc.Book(r.Context(), req)
		if err != nil {
			status := 500
			switch {
			case errors.As(err, new(ValidationErrors)):
				status = 422
			case errors.Is(err, ErrRoomTaken), errors.Is(err, ErrDuplicateRef):
				status = 409
			}
			if status >= 500 {
				svc.metrics.Inc(MetricRejected)
			}
			writeProblem(w, status, err.Error())
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"ref": b.Ref, "room": string(b.Room), "guest": b.Guest,
			"nights": b.Stay.Nights(), "total": b.Total,
		})
	})
	return mux
}

func writeProblem(w http.ResponseWriter, status int, title string) {
	w.Header().Set("Content-Type", ProblemContentType)
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"type": "about:blank", "title": title, "status": status,
	})
}

var _ = fmt.Sprintf

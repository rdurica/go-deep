// Package httpapi je vstupní adaptér: překládá HTTP na volání use-casů.
package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/rdurica/go-deep/lessons/lesson-38/solutions/app"
	"github.com/rdurica/go-deep/lessons/lesson-38/solutions/order"
)

// ProblemContentType je Content-Type chybové odpovědi podle RFC 7807.
const ProblemContentType = "application/problem+json"

// maxBody je strop pro tělo požadavku.
const maxBody = 1 << 20

// lineDTO je položka na hranici. Doménový order.Line schválně nemá JSON
// tagy — formát drátu je věc adaptéru, ne domény.
type lineDTO struct {
	SKU            string `json:"sku"`
	Quantity       int    `json:"quantity"`
	UnitPriceCents int64  `json:"unit_price_cents"`
}

type placeRequest struct {
	Lines []lineDTO `json:"lines"`
}

type orderResponse struct {
	ID         string    `json:"id"`
	Status     string    `json:"status"`
	TotalCents int64     `json:"total_cents"`
	Lines      []lineDTO `json:"lines"`
}

type problemDetails struct {
	Type   string `json:"type"`
	Title  string `json:"title"`
	Status int    `json:"status"`
	Detail string `json:"detail,omitempty"`
}

type handler struct {
	svc *app.Service
}

// --- Stupeň: jednoduchý ---
// NewHandler složí router služby nad aplikační vrstvou.
func NewHandler(svc *app.Service) http.Handler {
	h := &handler{svc: svc}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", h.health)
	mux.HandleFunc("POST /orders", h.place)
	mux.HandleFunc("GET /orders/{id}", h.get)
	mux.HandleFunc("POST /orders/{id}/pay", h.transition(svc.Pay))
	mux.HandleFunc("POST /orders/{id}/ship", h.transition(svc.Ship))
	mux.HandleFunc("POST /orders/{id}/cancel", h.transition(svc.Cancel))
	return mux
}

func (h *handler) health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *handler) place(w http.ResponseWriter, r *http.Request) {
	var req placeRequest
	dec := json.NewDecoder(io.LimitReader(r.Body, maxBody))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		writeProblem(w, http.StatusBadRequest, "Neplatný požadavek", "tělo požadavku není platný JSON")
		return
	}

	lines := make([]order.Line, 0, len(req.Lines))
	for _, l := range req.Lines {
		lines = append(lines, order.Line{
			SKU:            l.SKU,
			Quantity:       l.Quantity,
			UnitPriceCents: l.UnitPriceCents,
		})
	}

	o, err := h.svc.Place(r.Context(), lines)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, toResponse(o))
}

func (h *handler) get(w http.ResponseWriter, r *http.Request) {
	o, err := h.svc.Get(r.Context(), r.PathValue("id"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toResponse(o))
}

// transition vyrobí handler pro use-case, který mění stav podle ID v cestě.
func (h *handler) transition(useCase func(context.Context, string) (order.Order, error)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		o, err := useCase(r.Context(), r.PathValue("id"))
		if err != nil {
			writeError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, toResponse(o))
	}
}

func toResponse(o order.Order) orderResponse {
	lines := make([]lineDTO, 0, len(o.Lines))
	for _, l := range o.Lines {
		lines = append(lines, lineDTO{
			SKU:            l.SKU,
			Quantity:       l.Quantity,
			UnitPriceCents: l.UnitPriceCents,
		})
	}
	return orderResponse{
		ID:         o.ID,
		Status:     o.Status.String(),
		TotalCents: o.TotalCents(),
		Lines:      lines,
	}
}

// writeError je JEDINÉ místo, kde se doménová chyba překládá na HTTP status.
// Doména o statusech neví a vědět nemá.
func writeError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, order.ErrNotFound):
		writeProblem(w, http.StatusNotFound, "Nenalezeno", "objednávka neexistuje")
	case errors.Is(err, order.ErrInvalidTransition):
		writeProblem(w, http.StatusConflict, "Konflikt", "operace není v současném stavu objednávky povolená")
	case errors.Is(err, order.ErrEmptyOrder),
		errors.Is(err, order.ErrInvalidLine),
		errors.Is(err, order.ErrMissingID):
		writeProblem(w, http.StatusUnprocessableEntity, "Neplatná data", "objednávka porušuje pravidla domény")
	default:
		writeProblem(w, http.StatusInternalServerError, "Vnitřní chyba serveru", "požadavek se nepodařilo zpracovat")
	}
}

func writeProblem(w http.ResponseWriter, status int, title, detail string) {
	w.Header().Set("Content-Type", ProblemContentType)
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(status)
	// Status i hlavičky už jsou odeslané, s chybou zápisu se nedá nic dělat.
	_ = json.NewEncoder(w).Encode(problemDetails{
		Type:   "about:blank",
		Title:  title,
		Status: status,
		Detail: detail,
	})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

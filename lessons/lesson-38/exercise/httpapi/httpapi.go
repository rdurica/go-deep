// Package httpapi je vstupní adaptér — jediné místo, kde student píše kód lekce 38.
package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/rdurica/go-deep/lessons/lesson-38/exercise/app"
	"github.com/rdurica/go-deep/lessons/lesson-38/exercise/order"
)

// ProblemContentType je Content-Type chybové odpovědi podle RFC 7807.
const ProblemContentType = "application/problem+json"

const maxBody = 1 << 20

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

// WriteJSONForTest exportuje writeJSON pro testy (PART1).
func WriteJSONForTest(w http.ResponseWriter, status int, v any) {
	writeJSON(w, status, v)
}

// --- Stupeň: jednoduchý ---

// writeJSON zapíše JSON odpověď. Hlavičky před WriteHeader, tělo až potom.
//
// POZOR: kód níže je ZÁMĚRNĚ VADNÝ. Zapíše tělo před nastavením statusu.
func writeJSON(w http.ResponseWriter, status int, v any) {
	_ = json.NewEncoder(w).Encode(v)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
}

// --- Stupeň: střední ---

// writeError mapuje doménové chyby na HTTP statusy na jednom místě.
// ErrNotFound→404, ErrInvalidTransition→409, invarianty→422, ostatní→500 bez detailu.
func writeError(w http.ResponseWriter, err error) {
	// TODO
}

func writeProblem(w http.ResponseWriter, status int, title, detail string) {
	w.Header().Set("Content-Type", ProblemContentType)
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(problemDetails{
		Type:   "about:blank",
		Title:  title,
		Status: status,
		Detail: detail,
	})
}

// --- Stupeň: obtížný ---

// NewHandler složí router: GET /healthz, POST /orders, GET /orders/{id},
// POST /orders/{id}/{pay,ship,cancel}. Tělo přísně (LimitReader, DisallowUnknownFields).
func NewHandler(svc *app.Service) http.Handler {
	// TODO
	return *new(http.Handler)
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
			SKU: l.SKU, Quantity: l.Quantity, UnitPriceCents: l.UnitPriceCents,
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
		lines = append(lines, lineDTO{SKU: l.SKU, Quantity: l.Quantity, UnitPriceCents: l.UnitPriceCents})
	}
	return orderResponse{ID: o.ID, Status: o.Status.String(), TotalCents: o.TotalCents(), Lines: lines}
}

// unused import guard until writeError is implemented
var _ = errors.Is

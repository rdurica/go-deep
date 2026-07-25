// Package httpapi je vstupní adaptér: překládá HTTP na volání use-casů.
//
// Vlastní DTO s JSON tagy jsou tady schválně. Formát na drátě je veřejný
// kontrakt, který se mění jinou rychlostí než doména; kdyby tagy nesla
// doména, přejmenování pole rozbije klienty.
package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/rdurica/go-deep/projects/p03-hex-service/internal/app"
	"github.com/rdurica/go-deep/projects/p03-hex-service/internal/order"
)

// maxBody je strop pro tělo požadavku. Bez něj by útočník poslal
// nekonečný stream a server by ho poslušně dekódoval do paměti.
const maxBody = 1 << 20

// errMalformedJSON značí nepoužitelné tělo požadavku.
var errMalformedJSON = errors.New("neplatné JSON tělo")

type lineRequest struct {
	SKU            string `json:"sku"`
	Quantity       int    `json:"quantity"`
	UnitPriceCents int64  `json:"unit_price_cents"`
}

type placeRequest struct {
	Customer string        `json:"customer"`
	Currency string        `json:"currency"`
	Lines    []lineRequest `json:"lines"`
}

type lineResponse struct {
	SKU            string `json:"sku"`
	Quantity       int    `json:"quantity"`
	UnitPriceCents int64  `json:"unit_price_cents"`
	TotalCents     int64  `json:"total_cents"`
}

type orderResponse struct {
	ID         string         `json:"id"`
	Customer   string         `json:"customer"`
	Status     string         `json:"status"`
	Currency   string         `json:"currency"`
	TotalCents int64          `json:"total_cents"`
	Total      string         `json:"total"`
	PlacedAt   time.Time      `json:"placed_at"`
	Lines      []lineResponse `json:"lines"`
}

type handler struct {
	svc *app.Service
}

// NewHandler složí router služby nad aplikační vrstvou.
func NewHandler(svc *app.Service) http.Handler {
	h := &handler{svc: svc}

	mux := http.NewServeMux()
	// Liveness nesahá na závislosti: kdyby ano, výpadek databáze by
	// orchestrátoru řekl, ať restartuje celou flotilu.
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})
	mux.HandleFunc("GET /readyz", h.ready)
	mux.HandleFunc("POST /orders", h.place)
	mux.HandleFunc("GET /orders", h.list)
	mux.HandleFunc("GET /orders/{id}", h.get)
	mux.HandleFunc("POST /orders/{id}/pay", h.transition(svc.Pay))
	mux.HandleFunc("POST /orders/{id}/ship", h.transition(svc.Ship))
	mux.HandleFunc("POST /orders/{id}/cancel", h.transition(svc.Cancel))
	return mux
}

// ready se na rozdíl od healthz ptá i závislostí.
func (h *handler) ready(w http.ResponseWriter, r *http.Request) {
	if _, err := h.svc.List(r.Context()); err != nil {
		writeProblemStatus(w, http.StatusServiceUnavailable, "Nedostupné", "úložiště neodpovídá")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ready"})
}

func (h *handler) place(w http.ResponseWriter, r *http.Request) {
	var req placeRequest
	dec := json.NewDecoder(io.LimitReader(r.Body, maxBody))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		writeError(w, fmt.Errorf("%w: %v", errMalformedJSON, err))
		return
	}

	cmd := app.PlaceCommand{
		Customer: req.Customer,
		Currency: req.Currency,
		Lines:    make([]app.LineCommand, 0, len(req.Lines)),
	}
	for _, l := range req.Lines {
		cmd.Lines = append(cmd.Lines, app.LineCommand{
			SKU:            l.SKU,
			Quantity:       l.Quantity,
			UnitPriceCents: l.UnitPriceCents,
		})
	}

	o, err := h.svc.Place(r.Context(), cmd)
	if err != nil {
		writeError(w, err)
		return
	}
	h.respond(w, http.StatusCreated, o)
}

func (h *handler) get(w http.ResponseWriter, r *http.Request) {
	o, err := h.svc.Get(r.Context(), r.PathValue("id"))
	if err != nil {
		writeError(w, err)
		return
	}
	h.respond(w, http.StatusOK, o)
}

func (h *handler) list(w http.ResponseWriter, r *http.Request) {
	orders, err := h.svc.List(r.Context())
	if err != nil {
		writeError(w, err)
		return
	}
	out := make([]orderResponse, 0, len(orders))
	for _, o := range orders {
		body, err := toResponse(o)
		if err != nil {
			writeError(w, err)
			return
		}
		out = append(out, body)
	}
	writeJSON(w, http.StatusOK, map[string]any{"orders": out})
}

// transition vyrobí handler pro use-case měnící stav podle ID v cestě.
func (h *handler) transition(useCase func(context.Context, string) (order.Order, error)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		o, err := useCase(r.Context(), r.PathValue("id"))
		if err != nil {
			writeError(w, err)
			return
		}
		h.respond(w, http.StatusOK, o)
	}
}

func (h *handler) respond(w http.ResponseWriter, status int, o order.Order) {
	body, err := toResponse(o)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, status, body)
}

func toResponse(o order.Order) (orderResponse, error) {
	total, err := o.Total()
	if err != nil {
		return orderResponse{}, err
	}

	lines := make([]lineResponse, 0, len(o.Lines))
	for i, l := range o.Lines {
		lineTotal, err := l.Total()
		if err != nil {
			return orderResponse{}, fmt.Errorf("položka %d: %w", i, err)
		}
		lines = append(lines, lineResponse{
			SKU:            l.SKU,
			Quantity:       l.Quantity,
			UnitPriceCents: l.UnitPrice.Cents(),
			TotalCents:     lineTotal.Cents(),
		})
	}

	return orderResponse{
		ID:         o.ID,
		Customer:   o.Customer,
		Status:     o.Status.String(),
		Currency:   total.Currency(),
		TotalCents: total.Cents(),
		Total:      total.String(),
		PlacedAt:   o.PlacedAt,
		Lines:      lines,
	}, nil
}

func writeProblemStatus(w http.ResponseWriter, status int, title, detail string) {
	w.Header().Set("Content-Type", ProblemContentType)
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(ProblemDetails{
		Type:   "about:blank",
		Title:  title,
		Status: status,
		Detail: detail,
	})
}

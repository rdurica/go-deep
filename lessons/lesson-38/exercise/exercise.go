// Package exercise je fasáda nad balíčky cvičení lekce 38.
package exercise

import (
	"net/http"

	"github.com/rdurica/go-deep/lessons/lesson-38/exercise/app"
	"github.com/rdurica/go-deep/lessons/lesson-38/exercise/httpapi"
	"github.com/rdurica/go-deep/lessons/lesson-38/exercise/memstore"
	"github.com/rdurica/go-deep/lessons/lesson-38/exercise/order"
)

type (
	Order      = order.Order
	Line       = order.Line
	Status     = order.Status
	Repository = app.Repository
	IDGen      = app.IDGen
	Service    = app.Service
)

const (
	StatusUnknown   = order.StatusUnknown
	StatusNew       = order.StatusNew
	StatusPaid      = order.StatusPaid
	StatusShipped   = order.StatusShipped
	StatusCancelled = order.StatusCancelled
)

var (
	ErrMissingID         = order.ErrMissingID
	ErrEmptyOrder        = order.ErrEmptyOrder
	ErrInvalidLine       = order.ErrInvalidLine
	ErrInvalidTransition = order.ErrInvalidTransition
	ErrNotFound          = order.ErrNotFound
)

const ProblemContentType = httpapi.ProblemContentType

func NewOrder(id string, lines []Line) (Order, error) { return order.New(id, lines) }
func NewMemoryRepository() Repository                 { return memstore.New() }
func NewService(repo Repository, ids IDGen) *Service  { return app.NewService(repo, ids) }
func NewHandler(svc *Service) http.Handler            { return httpapi.NewHandler(svc) }

// WriteJSON je export pro test PART1 — volá interní writeJSON přes health handler.
// Testy ověřují pořadí hlaviček přes httptest.
func WriteJSON(w http.ResponseWriter, status int, v any) {
	httpapi.WriteJSONForTest(w, status, v)
}

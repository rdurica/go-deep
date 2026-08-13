// Package solutions je fasáda nad balíčky lekce 38.
package solutions

import (
	"net/http"

	"github.com/rdurica/go-deep/lessons/lesson-38/solutions/app"
	"github.com/rdurica/go-deep/lessons/lesson-38/solutions/httpapi"
	"github.com/rdurica/go-deep/lessons/lesson-38/solutions/memstore"
	"github.com/rdurica/go-deep/lessons/lesson-38/solutions/order"
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

func NewOrder(id string, lines []Line) (Order, error)    { return order.New(id, lines) }
func NewMemoryRepository() Repository                    { return memstore.New() }
func NewService(repo Repository, ids IDGen) *Service     { return app.NewService(repo, ids) }
func NewHandler(svc *Service) http.Handler               { return httpapi.NewHandler(svc) }
func WriteJSON(w http.ResponseWriter, status int, v any) { httpapi.WriteJSONForTest(w, status, v) }

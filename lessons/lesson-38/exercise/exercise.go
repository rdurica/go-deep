// Package exercise je fasáda nad balíčky cvičení lekce 38.
//
// Skutečný kód píšeš v podbalíčcích order/, app/, memstore/ a httpapi/.
// Tenhle soubor jen zpřístupňuje jejich API pod jedním jménem, aby testy
// mohly zůstat v jediném balíčku. Nic tady neimplementuj — je hotový.
package exercise

import (
	"net/http"

	"github.com/rdurica/go-deep/lessons/lesson-38/exercise/app"
	"github.com/rdurica/go-deep/lessons/lesson-38/exercise/httpapi"
	"github.com/rdurica/go-deep/lessons/lesson-38/exercise/memstore"
	"github.com/rdurica/go-deep/lessons/lesson-38/exercise/order"
)

// Aliasy doménových a aplikačních typů. Alias (znaménko =) není nový typ,
// jen druhé jméno pro ten samý.
type (
	Order      = order.Order
	Line       = order.Line
	Status     = order.Status
	Repository = app.Repository
	IDGen      = app.IDGen
	Service    = app.Service
)

// Stavy objednávky.
const (
	StatusUnknown   = order.StatusUnknown
	StatusNew       = order.StatusNew
	StatusPaid      = order.StatusPaid
	StatusShipped   = order.StatusShipped
	StatusCancelled = order.StatusCancelled
)

// Doménové chyby.
var (
	ErrMissingID         = order.ErrMissingID
	ErrEmptyOrder        = order.ErrEmptyOrder
	ErrInvalidLine       = order.ErrInvalidLine
	ErrInvalidTransition = order.ErrInvalidTransition
	ErrNotFound          = order.ErrNotFound
)

// ProblemContentType je Content-Type chybové odpovědi podle RFC 7807.
const ProblemContentType = httpapi.ProblemContentType

// --- Stupeň: jednoduchý ---
// NewOrder je fasáda nad order.New. Hotová, neimplementuj ji.
func NewOrder(id string, lines []Line) (Order, error) { return order.New(id, lines) }

// --- Stupeň: střední ---
// NewMemoryRepository je fasáda nad memstore.New. Hotová, neimplementuj ji.
func NewMemoryRepository() Repository { return memstore.New() }

// --- Stupeň: obtížný ---
// NewService je fasáda nad app.NewService. Hotová, neimplementuj ji.
func NewService(repo Repository, ids IDGen) *Service { return app.NewService(repo, ids) }

// NewHandler je fasáda nad httpapi.NewHandler. Hotová, neimplementuj ji.
func NewHandler(svc *Service) http.Handler { return httpapi.NewHandler(svc) }

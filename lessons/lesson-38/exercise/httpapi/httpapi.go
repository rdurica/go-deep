// Package httpapi je vstupní adaptér: překládá HTTP na volání use-casů.
package httpapi

import (
	"net/http"

	"github.com/rdurica/go-deep/lessons/lesson-38/exercise/app"
)

// ProblemContentType je Content-Type chybové odpovědi podle RFC 7807.
const ProblemContentType = "application/problem+json"

// NewHandler složí router: GET /healthz, POST /orders (201), GET /orders/{id},
// POST /orders/{id}/{pay,ship,cancel}. Tělo přísně (LimitReader, DisallowUnknownFields).
// Mapování na jednom místě: ErrNotFound→404, ErrInvalidTransition→409,
// invarianty→422, nečitelné tělo→400, ostatní→500 bez detailu. problem+json.
func NewHandler(svc *app.Service) http.Handler {
	// TODO
	return *new(http.Handler)
}

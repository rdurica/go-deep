// Package httpapi je vstupní adaptér: překládá HTTP na volání use-casů.
package httpapi

import (
	"net/http"

	"github.com/rdurica/go-deep/lessons/lesson-38/exercise/app"
)

// ProblemContentType je Content-Type chybové odpovědi podle RFC 7807.
const ProblemContentType = "application/problem+json"

// NewHandler složí router služby nad aplikační vrstvou.
func NewHandler(svc *app.Service) http.Handler {
	panic("TODO: úkol C")
}

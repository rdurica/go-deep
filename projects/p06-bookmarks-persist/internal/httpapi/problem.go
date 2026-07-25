package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/rdurica/go-deep/projects/p06-bookmarks-persist/internal/bookmark"
)

// ProblemContentType je Content-Type chybové odpovědi podle RFC 7807.
const ProblemContentType = "application/problem+json"

// ProblemDetails je tělo chybové odpovědi podle RFC 7807.
type ProblemDetails struct {
	Type   string `json:"type"`
	Title  string `json:"title"`
	Status int    `json:"status"`
	Detail string `json:"detail,omitempty"`
}

var errMalformedJSON = errors.New("neplatné JSON tělo")

// StatusFor překládá chybu na HTTP status a problem-details.
func StatusFor(err error) (int, ProblemDetails) {
	problem := func(status int, title, detail string) (int, ProblemDetails) {
		return status, ProblemDetails{
			Type:   "about:blank",
			Title:  title,
			Status: status,
			Detail: detail,
		}
	}

	switch {
	case errors.Is(err, errMalformedJSON):
		return problem(http.StatusBadRequest, "Neplatný požadavek", "tělo požadavku není platný JSON")
	case errors.Is(err, bookmark.ErrNotFound):
		return problem(http.StatusNotFound, "Nenalezeno", "záložka neexistuje")
	case errors.Is(err, bookmark.ErrDuplicateURL), errors.Is(err, bookmark.ErrDuplicateID):
		return problem(http.StatusConflict, "Konflikt", "záložka s touto URL nebo ID už existuje")
	case errors.Is(err, bookmark.ErrInvalidURL),
		errors.Is(err, bookmark.ErrEmptyTitle),
		errors.Is(err, bookmark.ErrTitleTooLong),
		errors.Is(err, bookmark.ErrInvalidTag),
		errors.Is(err, bookmark.ErrTooManyTags),
		errors.Is(err, bookmark.ErrDuplicateTag),
		errors.Is(err, bookmark.ErrEmptyID),
		errors.Is(err, bookmark.ErrInvalidQuery),
		errors.Is(err, bookmark.ErrInvalidCursor):
		return problem(http.StatusBadRequest, "Neplatný požadavek", err.Error())
	default:
		return problem(http.StatusInternalServerError, "Vnitřní chyba serveru", "požadavek se nepodařilo zpracovat")
	}
}

func writeError(w http.ResponseWriter, err error) {
	status, body := StatusFor(err)
	w.Header().Set("Content-Type", ProblemContentType)
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

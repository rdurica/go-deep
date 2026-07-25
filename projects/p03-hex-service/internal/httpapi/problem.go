package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/rdurica/go-deep/projects/p03-hex-service/internal/order"
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

// StatusFor je JEDINÉ místo, kde se doménová chyba překládá na HTTP status.
// Doména o statusech neví a vědět nemá.
//
// Nerozpoznaná chyba končí jako holá pětistovka: text interní chyby se
// ke klientovi nesmí dostat, patří do logu s request ID.
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
		return problem(http.StatusBadRequest, "Neplatný požadavek",
			"tělo požadavku není platný JSON")
	case errors.Is(err, order.ErrNotFound):
		return problem(http.StatusNotFound, "Nenalezeno",
			"objednávka neexistuje")
	case errors.Is(err, order.ErrInvalidTransition):
		return problem(http.StatusConflict, "Konflikt",
			"operace není v současném stavu objednávky povolená")
	case errors.Is(err, order.ErrEmptyOrder),
		errors.Is(err, order.ErrInvalidLine),
		errors.Is(err, order.ErrMissingID),
		errors.Is(err, order.ErrMissingCustomer),
		errors.Is(err, order.ErrMissingTimestamp),
		errors.Is(err, order.ErrInvalidCurrency),
		errors.Is(err, order.ErrCurrencyMismatch),
		errors.Is(err, order.ErrNegativeAmount),
		errors.Is(err, order.ErrAmountOverflow):
		return problem(http.StatusUnprocessableEntity, "Neplatná data",
			"požadavek porušuje pravidla domény")
	default:
		return problem(http.StatusInternalServerError, "Vnitřní chyba serveru",
			"požadavek se nepodařilo zpracovat")
	}
}

func writeError(w http.ResponseWriter, err error) {
	status, body := StatusFor(err)
	w.Header().Set("Content-Type", ProblemContentType)
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(status)
	// Status i hlavičky jsou na drátě; s chybou zápisu se nedá nic dělat.
	_ = json.NewEncoder(w).Encode(body)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

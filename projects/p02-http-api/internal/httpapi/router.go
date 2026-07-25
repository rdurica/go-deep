// Package httpapi vystavuje doménu úkolů jako REST API nad net/http.
package httpapi

import (
	"log/slog"
	"net/http"
	"strings"

	"github.com/rdurica/go-deep/projects/p02-http-api/internal/task"
)

// NewRouter sestaví kompletní handler API včetně middleware chainu.
func NewRouter(store *task.Store, logger *slog.Logger) http.Handler {
	h := &handlers{store: store}
	mux := http.NewServeMux()

	mux.HandleFunc("GET /healthz", h.health)

	mux.HandleFunc("GET /tasks", h.listTasks)
	mux.HandleFunc("POST /tasks", h.createTask)

	mux.HandleFunc("GET /tasks/{id}", h.getTask)
	mux.HandleFunc("PUT /tasks/{id}", h.updateTask)
	mux.HandleFunc("DELETE /tasks/{id}", h.deleteTask)

	// Vzory bez metody jsou méně specifické než ty s metodou, takže se uplatní
	// až když cesta sedí a metoda ne. Díky nim má i 405 stejný tvar JSONu jako
	// ostatní chyby — vestavěná odpověď ServeMuxu je prostý text.
	mux.HandleFunc("/healthz", methodNotAllowed(http.MethodGet))
	mux.HandleFunc("/tasks", methodNotAllowed(http.MethodGet, http.MethodPost))
	mux.HandleFunc("/tasks/{id}", methodNotAllowed(http.MethodGet, http.MethodPut, http.MethodDelete))
	mux.HandleFunc("/", notFound)

	return Chain(mux, RequestID, Logging(logger), Recovery(logger))
}

// methodNotAllowed vrací handler pro 405 včetně hlavičky Allow.
func methodNotAllowed(allowed ...string) http.HandlerFunc {
	allow := strings.Join(allowed, ", ")
	return func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Allow", allow)
		writeError(w, http.StatusMethodNotAllowed, CodeMethodNotAllowed, "method not allowed for this resource")
	}
}

// notFound odpovídá na neznámé cesty ve stejném tvaru jako ostatní chyby.
func notFound(w http.ResponseWriter, _ *http.Request) {
	writeError(w, http.StatusNotFound, CodeNotFound, "unknown endpoint")
}

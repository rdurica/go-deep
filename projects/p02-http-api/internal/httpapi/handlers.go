package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/rdurica/go-deep/projects/p02-http-api/internal/task"
)

// maxBodyBytes je strop pro tělo požadavku. Bez něj rozhoduje o spotřebě
// paměti klient, ne server.
const maxBodyBytes = 64 << 10 // 64 KiB

// taskResponse je reprezentace úkolu na hranici API. Doménový typ zůstává bez
// JSON tagů, takže změna formátu odpovědi se nedotkne domény.
type taskResponse struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func newTaskResponse(t task.Task) taskResponse {
	return taskResponse{
		ID:        t.ID,
		Title:     t.Title,
		Status:    string(t.Status),
		CreatedAt: t.CreatedAt,
		UpdatedAt: t.UpdatedAt,
	}
}

// taskRequest je vstupní tělo pro POST a PUT.
type taskRequest struct {
	Title  string `json:"title"`
	Status string `json:"status"`
}

// listResponse obaluje seznam, aby šlo odpověď v budoucnu rozšířit o stránkování.
type listResponse struct {
	Tasks []taskResponse `json:"tasks"`
}

// handlers drží závislosti HTTP vrstvy.
type handlers struct {
	store *task.Store
}

func (h *handlers) health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *handlers) listTasks(w http.ResponseWriter, _ *http.Request) {
	tasks := h.store.List()
	out := make([]taskResponse, 0, len(tasks))
	for _, t := range tasks {
		out = append(out, newTaskResponse(t))
	}
	writeJSON(w, http.StatusOK, listResponse{Tasks: out})
}

func (h *handlers) createTask(w http.ResponseWriter, r *http.Request) {
	in, ok := decodeTaskRequest(w, r)
	if !ok {
		return
	}

	created, err := h.store.Create(in.Title, task.Status(in.Status))
	if err != nil {
		writeDomainError(w, err)
		return
	}

	w.Header().Set("Location", "/tasks/"+created.ID)
	writeJSON(w, http.StatusCreated, newTaskResponse(created))
}

func (h *handlers) getTask(w http.ResponseWriter, r *http.Request) {
	found, err := h.store.Get(r.PathValue("id"))
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, newTaskResponse(found))
}

func (h *handlers) updateTask(w http.ResponseWriter, r *http.Request) {
	in, ok := decodeTaskRequest(w, r)
	if !ok {
		return
	}

	updated, err := h.store.Update(r.PathValue("id"), in.Title, task.Status(in.Status))
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, newTaskResponse(updated))
}

func (h *handlers) deleteTask(w http.ResponseWriter, r *http.Request) {
	if err := h.store.Delete(r.PathValue("id")); err != nil {
		writeDomainError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// decodeTaskRequest ověří Content-Type a rozparsuje tělo požadavku.
// Při chybě sám odešle odpověď a vrátí ok=false.
func decodeTaskRequest(w http.ResponseWriter, r *http.Request) (taskRequest, bool) {
	if !hasJSONContentType(r) {
		writeError(w, http.StatusUnsupportedMediaType, CodeUnsupportedMediaType,
			"Content-Type must be application/json")
		return taskRequest{}, false
	}

	var in taskRequest
	dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, maxBodyBytes))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&in); err != nil {
		writeError(w, http.StatusBadRequest, CodeBadRequest, "request body is not valid JSON")
		return taskRequest{}, false
	}
	// Druhé dekódování musí narazit na EOF — jinak přišlo víc JSON dokumentů.
	if err := dec.Decode(new(json.RawMessage)); err == nil {
		writeError(w, http.StatusBadRequest, CodeBadRequest, "request body must contain a single JSON object")
		return taskRequest{}, false
	}
	return in, true
}

// hasJSONContentType ověří, že tělo je deklarované jako JSON.
func hasJSONContentType(r *http.Request) bool {
	ct := r.Header.Get("Content-Type")
	if i := strings.IndexByte(ct, ';'); i >= 0 {
		ct = ct[:i]
	}
	return strings.EqualFold(strings.TrimSpace(ct), "application/json")
}

// writeDomainError přeloží chybu domény na HTTP status a kód.
func writeDomainError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, task.ErrNotFound):
		writeError(w, http.StatusNotFound, CodeNotFound, err.Error())
	case errors.Is(err, task.ErrEmptyTitle),
		errors.Is(err, task.ErrTitleTooLong),
		errors.Is(err, task.ErrInvalidStatus):
		writeError(w, http.StatusBadRequest, CodeValidationFailed, err.Error())
	default:
		writeError(w, http.StatusInternalServerError, CodeInternalError, "internal server error")
	}
}

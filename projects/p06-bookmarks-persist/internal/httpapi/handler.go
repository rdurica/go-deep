// Package httpapi je HTTP adaptér nad portem BookmarkStore.
package httpapi

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/rdurica/go-deep/projects/p06-bookmarks-persist/internal/app"
	"github.com/rdurica/go-deep/projects/p06-bookmarks-persist/internal/bookmark"
)

// IDGen vyrábí identifikátory záložek.
type IDGen func() string

// Clock vrací aktuální čas.
type Clock func() time.Time

// Server je HTTP API záložek.
type Server struct {
	store app.BookmarkStore
	idGen IDGen
	now   Clock
}

// NewServer sestaví handler včetně middleware.
func NewServer(st app.BookmarkStore, opts Options) http.Handler {
	return NewServerWith(st, opts, nil, nil)
}

// NewServerWith umožní injektovat generátor ID a hodiny (pro testy).
func NewServerWith(st app.BookmarkStore, opts Options, idGen IDGen, now Clock) http.Handler {
	if idGen == nil {
		idGen = randomID
	}
	if now == nil {
		now = time.Now
	}
	s := &Server{store: st, idGen: idGen, now: now}
	return Wrap(s.routes(), opts)
}

func (s *Server) routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", s.healthz)
	mux.HandleFunc("GET /readyz", s.readyz)
	mux.HandleFunc("POST /bookmarks", s.create)
	mux.HandleFunc("GET /bookmarks/{id}", s.get)
	mux.HandleFunc("DELETE /bookmarks/{id}", s.delete)
	mux.HandleFunc("GET /bookmarks", s.search)
	return mux
}

type createRequest struct {
	URL   string   `json:"url"`
	Title string   `json:"title"`
	Tags  []string `json:"tags"`
}

type bookmarkResponse struct {
	ID        string    `json:"id"`
	URL       string    `json:"url"`
	Title     string    `json:"title"`
	Tags      []string  `json:"tags"`
	CreatedAt time.Time `json:"created_at"`
}

type searchResponse struct {
	Items      []bookmarkResponse `json:"items"`
	NextCursor string             `json:"next_cursor,omitempty"`
	Total      int                `json:"total"`
}

func (s *Server) healthz(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) readyz(w http.ResponseWriter, r *http.Request) {
	if s.store == nil {
		writeError(w, fmt.Errorf("úložiště není připravené"))
		return
	}
	if err := s.store.Ready(r.Context()); err != nil {
		writeError(w, fmt.Errorf("úložiště není připravené: %w", err))
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ready"})
}

func (s *Server) create(w http.ResponseWriter, r *http.Request) {
	var req createRequest
	dec := json.NewDecoder(io.LimitReader(r.Body, 1<<20))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		writeError(w, fmt.Errorf("%w: %v", errMalformedJSON, err))
		return
	}

	b, err := bookmark.New(s.idGen(), req.URL, req.Title, req.Tags, s.now().UTC())
	if err != nil {
		writeError(w, err)
		return
	}
	if err := s.store.Add(r.Context(), b); err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, toResponse(b))
}

func (s *Server) get(w http.ResponseWriter, r *http.Request) {
	b, err := s.store.Get(r.Context(), r.PathValue("id"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toResponse(b))
}

func (s *Server) delete(w http.ResponseWriter, r *http.Request) {
	if err := s.store.Delete(r.Context(), r.PathValue("id")); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) search(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	limit := 0
	if raw := q.Get("limit"); raw != "" {
		n, err := strconv.Atoi(raw)
		if err != nil {
			writeError(w, bookmark.ErrInvalidQuery)
			return
		}
		limit = n
	}

	page, err := s.store.Search(r.Context(), app.Query{
		Text:   q.Get("q"),
		Tag:    q.Get("tag"),
		Limit:  limit,
		Cursor: q.Get("cursor"),
	})
	if err != nil {
		writeError(w, err)
		return
	}

	out := searchResponse{
		Items:      make([]bookmarkResponse, 0, len(page.Items)),
		NextCursor: page.NextCursor,
		Total:      page.Total,
	}
	for _, b := range page.Items {
		out.Items = append(out.Items, toResponse(b))
	}
	writeJSON(w, http.StatusOK, out)
}

func toResponse(b bookmark.Bookmark) bookmarkResponse {
	tags := b.Tags
	if tags == nil {
		tags = []string{}
	}
	return bookmarkResponse{
		ID:        b.ID,
		URL:       b.URL,
		Title:     b.Title,
		Tags:      tags,
		CreatedAt: b.CreatedAt,
	}
}

func randomID() string {
	var buf [8]byte
	if _, err := rand.Read(buf[:]); err != nil {
		panic("crypto/rand: " + err.Error())
	}
	return "bm_" + hex.EncodeToString(buf[:])
}

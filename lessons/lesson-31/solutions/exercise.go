// Package solutions obsahuje referenční řešení checkpointu fáze 3 (lekce 31).
package solutions

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

// RequestIDHeader je hlavička, ve které se přenáší identifikátor požadavku.
const RequestIDHeader = "X-Request-ID"

// Limity a výchozí hodnoty služby.
const (
	defaultAddr            = "127.0.0.1:8080"
	defaultShutdownTimeout = 5 * time.Second
	maxBodyBytes           = 64 << 10 // 64 KiB
	maxNoteLength          = 500
)

// ErrInvalid označuje hodnotu konfigurace, kterou se nepodařilo zpracovat.
var ErrInvalid = errors.New("invalid value")

// Config je konfigurace HTTP služby načtená z prostředí.
type Config struct {
	Addr            string
	LogLevel        slog.Level
	ShutdownTimeout time.Duration
}

// Note je jedna poznámka uložená v paměti.
type Note struct {
	ID   string `json:"id"`
	Text string `json:"text"`
}

// LoadConfig sestaví Config z getenv a posbírá všechny chyby najednou.
func LoadConfig(getenv func(string) string) (Config, error) {
	var errs []error

	cfg := Config{
		Addr:            defaultAddr,
		LogLevel:        slog.LevelInfo,
		ShutdownTimeout: defaultShutdownTimeout,
	}

	if raw := getenv("ADDR"); raw != "" {
		cfg.Addr = raw
	}

	if raw := getenv("LOG_LEVEL"); raw != "" {
		var level slog.Level
		if err := level.UnmarshalText([]byte(raw)); err != nil {
			errs = append(errs, fmt.Errorf("LOG_LEVEL=%q: %w", raw, ErrInvalid))
		} else {
			cfg.LogLevel = level
		}
	}

	if raw := getenv("SHUTDOWN_TIMEOUT"); raw != "" {
		d, err := time.ParseDuration(raw)
		switch {
		case err != nil:
			errs = append(errs, fmt.Errorf("SHUTDOWN_TIMEOUT=%q: %w", raw, ErrInvalid))
		case d <= 0:
			errs = append(errs, fmt.Errorf("SHUTDOWN_TIMEOUT=%q musí být kladný: %w", raw, ErrInvalid))
		default:
			cfg.ShutdownTimeout = d
		}
	}

	if len(errs) > 0 {
		return Config{}, errors.Join(errs...)
	}
	return cfg, nil
}

// Chain složí middleware kolem handleru; první uvedená je nejvíc vně.
func Chain(h http.Handler, mws ...func(http.Handler) http.Handler) http.Handler {
	for i := len(mws) - 1; i >= 0; i-- {
		h = mws[i](h)
	}
	return h
}

// ctxKey je privátní typ klíče, aby se hodnota v kontextu nedala přepsat cizím balíčkem.
type ctxKey struct{}

var requestIDKey ctxKey

// RequestIDMiddleware zajistí, že každý požadavek má identifikátor v kontextu i hlavičce.
func RequestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get(RequestIDHeader)
		if id == "" {
			id = newRequestID()
		}
		w.Header().Set(RequestIDHeader, id)

		ctx := context.WithValue(r.Context(), requestIDKey, id)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequestIDFromContext vytáhne identifikátor požadavku z kontextu.
func RequestIDFromContext(ctx context.Context) (string, bool) {
	id, ok := ctx.Value(requestIDKey).(string)
	return id, ok && id != ""
}

// newRequestID vyrobí náhodný identifikátor požadavku.
func newRequestID() string {
	var b [8]byte
	if _, err := rand.Read(b[:]); err != nil {
		// crypto/rand v praxi neselhává; kdyby ano, radši unikátní čas než prázdný řetězec.
		return strconv.FormatInt(time.Now().UnixNano(), 16)
	}
	return hex.EncodeToString(b[:])
}

// noteStore je in-memory úložiště poznámek bezpečné pro souběžný přístup.
type noteStore struct {
	mu    sync.RWMutex
	notes map[string]Note
	order []string
	seq   int
}

func newNoteStore() *noteStore {
	return &noteStore{notes: make(map[string]Note)}
}

func (s *noteStore) create(text string) Note {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.seq++
	n := Note{ID: strconv.Itoa(s.seq), Text: text}
	s.notes[n.ID] = n
	s.order = append(s.order, n.ID)
	return n
}

func (s *noteStore) get(id string) (Note, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	n, ok := s.notes[id]
	return n, ok
}

func (s *noteStore) list() []Note {
	s.mu.RLock()
	defer s.mu.RUnlock()

	out := make([]Note, 0, len(s.order))
	for _, id := range s.order {
		out = append(out, s.notes[id])
	}
	return out
}

func (s *noteStore) delete(id string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.notes[id]; !ok {
		return false
	}
	delete(s.notes, id)
	for i, existing := range s.order {
		if existing == id {
			s.order = append(s.order[:i], s.order[i+1:]...)
			break
		}
	}
	return true
}

// NewServer sestaví router s middleware chainem a in-memory úložištěm poznámek.
func NewServer(logger *slog.Logger) http.Handler {
	store := newNoteStore()
	mux := http.NewServeMux()

	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})
	mux.HandleFunc("GET /notes", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{"notes": store.list()})
	})
	mux.HandleFunc("POST /notes", func(w http.ResponseWriter, r *http.Request) {
		createNote(w, r, store)
	})
	mux.HandleFunc("GET /notes/{id}", func(w http.ResponseWriter, r *http.Request) {
		n, ok := store.get(r.PathValue("id"))
		if !ok {
			writeError(w, http.StatusNotFound, "not_found", "note not found")
			return
		}
		writeJSON(w, http.StatusOK, n)
	})
	mux.HandleFunc("DELETE /notes/{id}", func(w http.ResponseWriter, r *http.Request) {
		if !store.delete(r.PathValue("id")) {
			writeError(w, http.StatusNotFound, "not_found", "note not found")
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})

	// Vzory bez metody jsou méně specifické, takže se uplatní až když se metoda
	// neshoduje. Díky nim má i 405 stejný JSON tvar jako ostatní chyby.
	mux.HandleFunc("/healthz", methodNotAllowed(http.MethodGet))
	mux.HandleFunc("/notes", methodNotAllowed(http.MethodGet, http.MethodPost))
	mux.HandleFunc("/notes/{id}", methodNotAllowed(http.MethodGet, http.MethodDelete))
	mux.HandleFunc("/", func(w http.ResponseWriter, _ *http.Request) {
		writeError(w, http.StatusNotFound, "not_found", "unknown endpoint")
	})

	return Chain(mux, RequestIDMiddleware, RecoveryMiddleware(logger))
}

// createNote zpracuje POST /notes včetně validace vstupu.
func createNote(w http.ResponseWriter, r *http.Request, store *noteStore) {
	if !hasJSONContentType(r) {
		writeError(w, http.StatusUnsupportedMediaType, "unsupported_media_type", "expected application/json")
		return
	}

	var in struct {
		Text string `json:"text"`
	}
	dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, maxBodyBytes))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&in); err != nil {
		writeError(w, http.StatusBadRequest, "bad_request", "invalid JSON body")
		return
	}

	text := strings.TrimSpace(in.Text)
	switch {
	case text == "":
		writeError(w, http.StatusBadRequest, "validation_failed", "text must not be empty")
		return
	case len(text) > maxNoteLength:
		writeError(w, http.StatusBadRequest, "validation_failed", "text is too long")
		return
	}

	n := store.create(text)
	w.Header().Set("Location", "/notes/"+n.ID)
	writeJSON(w, http.StatusCreated, n)
}

// methodNotAllowed vrací handler pro 405 s hlavičkou Allow.
func methodNotAllowed(allowed ...string) http.HandlerFunc {
	allow := strings.Join(allowed, ", ")
	return func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Allow", allow)
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed")
	}
}

// hasJSONContentType ověří, že tělo je deklarované jako JSON.
func hasJSONContentType(r *http.Request) bool {
	ct := r.Header.Get("Content-Type")
	if i := strings.IndexByte(ct, ';'); i >= 0 {
		ct = ct[:i]
	}
	return strings.EqualFold(strings.TrimSpace(ct), "application/json")
}

// writeJSON pošle odpověď v JSONu se správnou hlavičkou.
func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

// writeError pošle chybu v konzistentním tvaru {"error":{"code","message"}}.
func writeError(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, map[string]any{
		"error": map[string]string{"code": code, "message": message},
	})
}

// RecoveryMiddleware zachytí paniku v handleru, zaloguje ji a vrátí 500 v JSONu.
func RecoveryMiddleware(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				rec := recover()
				if rec == nil {
					return
				}
				requestID, _ := RequestIDFromContext(r.Context())
				logger.LogAttrs(r.Context(), slog.LevelError, "panic recovered",
					slog.String("request_id", requestID),
					slog.String("method", r.Method),
					slog.String("path", r.URL.Path),
					slog.String("panic", fmt.Sprint(rec)),
				)
				writeError(w, http.StatusInternalServerError, "internal_error", "internal server error")
			}()
			next.ServeHTTP(w, r)
		})
	}
}

// Run obsluhuje ln, dokud se nezruší ctx, pak server elegantně ukončí.
func Run(ctx context.Context, cfg Config, h http.Handler, ln net.Listener) error {
	srv := &http.Server{
		Handler:           h,
		ReadHeaderTimeout: 5 * time.Second,
	}

	serveErr := make(chan error, 1)
	go func() {
		serveErr <- srv.Serve(ln)
	}()

	select {
	case err := <-serveErr:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	case <-ctx.Done():
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("shutdown: %w", err)
	}
	if err := <-serveErr; err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

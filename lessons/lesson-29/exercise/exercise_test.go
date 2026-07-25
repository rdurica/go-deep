package exercise_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	exercise "github.com/rdurica/go-deep/lessons/lesson-29/exercise"
)

// safeBuffer je buffer chráněný mutexem. Handler běží v goroutině HTTP serveru,
// takže bez zámku by -race hlásil závod při čtení z testu.
type safeBuffer struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (b *safeBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.Write(p)
}

func (b *safeBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.String()
}

// records rozparsuje výstup JSON handleru — jeden řádek = jeden záznam.
func records(t *testing.T, out string) []map[string]any {
	t.Helper()

	var recs []map[string]any
	for _, line := range strings.Split(strings.TrimSpace(out), "\n") {
		if line == "" {
			continue
		}
		var m map[string]any
		if err := json.Unmarshal([]byte(line), &m); err != nil {
			t.Fatalf("řádek %q není platný JSON: %v", line, err)
		}
		recs = append(recs, m)
	}
	return recs
}

// singleRecord očekává právě jeden záznam a vrátí ho.
func singleRecord(t *testing.T, out string) map[string]any {
	t.Helper()

	recs := records(t, out)
	if len(recs) != 1 {
		t.Fatalf("chci právě 1 záznam, mám %d; výstup:\n%s", len(recs), out)
	}
	return recs[0]
}

func wantString(t *testing.T, rec map[string]any, key, want string) {
	t.Helper()

	got, ok := rec[key].(string)
	if !ok {
		t.Errorf("záznam nemá řetězcové pole %q; záznam: %v", key, rec)
		return
	}
	if got != want {
		t.Errorf("pole %q = %q, chci %q", key, got, want)
	}
}

func wantNumber(t *testing.T, rec map[string]any, key string, want float64) {
	t.Helper()

	got, ok := rec[key].(float64)
	if !ok {
		t.Errorf("záznam nemá číselné pole %q; záznam: %v", key, rec)
		return
	}
	if got != want {
		t.Errorf("pole %q = %v, chci %v", key, got, want)
	}
}

func TestNewLoggerWritesJSON(t *testing.T) {
	t.Parallel()

	var buf safeBuffer
	logger := exercise.NewLogger(&buf, slog.LevelInfo)
	logger.Info("hotovo", "count", 3)

	rec := singleRecord(t, buf.String())
	wantString(t, rec, slog.MessageKey, "hotovo")
	wantString(t, rec, slog.LevelKey, "INFO")
	wantNumber(t, rec, "count", 3)
	if _, ok := rec[slog.TimeKey]; !ok {
		t.Error("záznam nemá pole time")
	}
}

func TestNewLoggerRespectsLevel(t *testing.T) {
	t.Parallel()

	var buf safeBuffer
	logger := exercise.NewLogger(&buf, slog.LevelWarn)

	logger.Debug("neviditelné")
	logger.Info("taky neviditelné")
	if out := buf.String(); out != "" {
		t.Fatalf("pod nastavenou úrovní se nesmí nic zapsat, mám:\n%s", out)
	}

	logger.Error("viditelné")
	rec := singleRecord(t, buf.String())
	wantString(t, rec, slog.LevelKey, "ERROR")
}

func TestLogRequest(t *testing.T) {
	t.Parallel()

	var buf safeBuffer
	logger := exercise.NewLogger(&buf, slog.LevelInfo)
	exercise.LogRequest(logger, http.MethodPost, "/tasks", http.StatusCreated, 1500*time.Millisecond)

	rec := singleRecord(t, buf.String())
	wantString(t, rec, slog.LevelKey, "INFO")
	wantString(t, rec, "method", http.MethodPost)
	wantString(t, rec, "path", "/tasks")
	wantNumber(t, rec, "status", float64(http.StatusCreated))
	// slog.Duration se v JSONu serializuje jako počet nanosekund.
	wantNumber(t, rec, "duration", float64((1500 * time.Millisecond).Nanoseconds()))
}

func TestServiceLogsSuccess(t *testing.T) {
	t.Parallel()

	var buf safeBuffer
	svc := exercise.NewService(exercise.NewLogger(&buf, slog.LevelDebug))

	if err := svc.Process("task-42"); err != nil {
		t.Fatalf("Process(\"task-42\") = %v, chci nil", err)
	}

	rec := singleRecord(t, buf.String())
	wantString(t, rec, slog.LevelKey, "INFO")
	wantString(t, rec, "component", "service")
	wantString(t, rec, "id", "task-42")
}

func TestServiceLogsFailure(t *testing.T) {
	t.Parallel()

	var buf safeBuffer
	svc := exercise.NewService(exercise.NewLogger(&buf, slog.LevelDebug))

	err := svc.Process("")
	if !errors.Is(err, exercise.ErrEmptyID) {
		t.Fatalf("Process(\"\") = %v, chci ErrEmptyID", err)
	}

	rec := singleRecord(t, buf.String())
	wantString(t, rec, slog.LevelKey, "ERROR")
	wantString(t, rec, "component", "service")
	if _, ok := rec["error"]; !ok {
		t.Errorf("chybový záznam nemá pole error; záznam: %v", rec)
	}
}

func TestRedactingHandlerTopLevel(t *testing.T) {
	t.Parallel()

	var buf safeBuffer
	logger := slog.New(exercise.NewRedactingHandler(slog.NewJSONHandler(&buf, nil)))
	logger.Info("login", "user", "radek", "password", "hunter2", "token", "t0k3n")

	out := buf.String()
	if strings.Contains(out, "hunter2") || strings.Contains(out, "t0k3n") {
		t.Fatalf("tajemství uniklo do logu:\n%s", out)
	}

	rec := singleRecord(t, out)
	wantString(t, rec, "user", "radek")
	wantString(t, rec, "password", exercise.Redacted)
	wantString(t, rec, "token", exercise.Redacted)
}

func TestRedactingHandlerNestedGroup(t *testing.T) {
	t.Parallel()

	var buf safeBuffer
	logger := slog.New(exercise.NewRedactingHandler(slog.NewJSONHandler(&buf, nil)))
	logger.Info("call",
		slog.Group("auth",
			slog.String("user", "radek"),
			slog.String("api_key", "sk-live-secret"),
			slog.Group("inner", slog.String("password", "hunter2")),
		),
	)

	out := buf.String()
	if strings.Contains(out, "sk-live-secret") || strings.Contains(out, "hunter2") {
		t.Fatalf("tajemství ve vnořené skupině uniklo:\n%s", out)
	}

	rec := singleRecord(t, out)
	auth, ok := rec["auth"].(map[string]any)
	if !ok {
		t.Fatalf("záznam nemá skupinu auth; záznam: %v", rec)
	}
	wantString(t, auth, "user", "radek")
	wantString(t, auth, "api_key", exercise.Redacted)

	inner, ok := auth["inner"].(map[string]any)
	if !ok {
		t.Fatalf("skupina auth nemá vnořenou skupinu inner; auth: %v", auth)
	}
	wantString(t, inner, "password", exercise.Redacted)
}

func TestRedactingHandlerWithAttrs(t *testing.T) {
	t.Parallel()

	var buf safeBuffer
	base := slog.New(exercise.NewRedactingHandler(slog.NewJSONHandler(&buf, nil)))
	logger := base.With("api_key", "sk-live-secret", "component", "worker")
	logger.Info("start")

	out := buf.String()
	if strings.Contains(out, "sk-live-secret") {
		t.Fatalf("tajemství z With uniklo do logu:\n%s", out)
	}

	rec := singleRecord(t, out)
	wantString(t, rec, "component", "worker")
	wantString(t, rec, "api_key", exercise.Redacted)
}

func TestRedactingHandlerWithGroup(t *testing.T) {
	t.Parallel()

	var buf safeBuffer
	base := slog.New(exercise.NewRedactingHandler(slog.NewJSONHandler(&buf, nil)))
	base.WithGroup("req").Info("start", "token", "t0k3n", "path", "/tasks")

	out := buf.String()
	if strings.Contains(out, "t0k3n") {
		t.Fatalf("tajemství ve skupině z WithGroup uniklo:\n%s", out)
	}

	rec := singleRecord(t, out)
	req, ok := rec["req"].(map[string]any)
	if !ok {
		t.Fatalf("záznam nemá skupinu req; záznam: %v", rec)
	}
	wantString(t, req, "token", exercise.Redacted)
	wantString(t, req, "path", "/tasks")
}

func TestLoggingMiddleware(t *testing.T) {
	t.Parallel()

	var buf safeBuffer
	logger := exercise.NewLogger(&buf, slog.LevelInfo)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(time.Millisecond)
		w.WriteHeader(http.StatusTeapot)
		_, _ = w.Write([]byte("čajník"))
	})

	srv := httptest.NewServer(exercise.LoggingMiddleware(logger)(handler))
	resp, err := srv.Client().Get(srv.URL + "/tasks/7")
	if err != nil {
		t.Fatalf("požadavek selhal: %v", err)
	}
	_ = resp.Body.Close()
	if resp.StatusCode != http.StatusTeapot {
		t.Errorf("status = %d, chci %d — middleware nesmí měnit odpověď", resp.StatusCode, http.StatusTeapot)
	}
	// Close počká na dokončení všech požadavků, takže log je už zapsaný.
	srv.Close()

	rec := singleRecord(t, buf.String())
	wantString(t, rec, "method", http.MethodGet)
	wantString(t, rec, "path", "/tasks/7")
	wantNumber(t, rec, "status", float64(http.StatusTeapot))

	dur, ok := rec["duration"].(float64)
	if !ok {
		t.Fatalf("záznam nemá pole duration; záznam: %v", rec)
	}
	if dur <= 0 {
		t.Errorf("duration = %v, chci kladnou hodnotu", dur)
	}
}

func TestLoggingMiddlewareDefaultStatus(t *testing.T) {
	t.Parallel()

	var buf safeBuffer
	logger := exercise.NewLogger(&buf, slog.LevelInfo)

	// Handler, který WriteHeader nezavolá — na drátě to je 200.
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("ok"))
	})

	srv := httptest.NewServer(exercise.LoggingMiddleware(logger)(handler))
	resp, err := srv.Client().Get(srv.URL + "/healthz")
	if err != nil {
		t.Fatalf("požadavek selhal: %v", err)
	}
	_ = resp.Body.Close()
	srv.Close()

	rec := singleRecord(t, buf.String())
	wantNumber(t, rec, "status", float64(http.StatusOK))
}

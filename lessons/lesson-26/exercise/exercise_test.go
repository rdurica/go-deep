package exercise_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	exercise "github.com/rdurica/go-deep/lessons/lesson-26/exercise"
)

const testTagHeader = "X-Test-Tag"

// okHandler odpoví 200 a zadaným tělem.
func okHandler(body string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, body)
	})
}

// trace vrací middleware, který zapisuje do events před a po zavolání dalšího handleru.
func trace(events *[]string, name string) exercise.Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			*events = append(*events, name+":pred")
			next.ServeHTTP(w, r)
			*events = append(*events, name+":po")
		})
	}
}

func TestStatusRecorderDefaultStatusIs200(t *testing.T) {
	rec := httptest.NewRecorder()
	sr := exercise.NewStatusRecorder(rec)

	fmt.Fprint(sr, "ahoj")

	if got := sr.RecordedStatus(); got != http.StatusOK {
		t.Errorf("RecordedStatus = %d, chci %d (handler WriteHeader nezavolal)", got, http.StatusOK)
	}
	if got := sr.RecordedBytes(); got != 4 {
		t.Errorf("RecordedBytes = %d, chci %d", got, 4)
	}
}

func TestStatusRecorderIgnoresSecondWriteHeader(t *testing.T) {
	rec := httptest.NewRecorder()
	sr := exercise.NewStatusRecorder(rec)

	sr.WriteHeader(http.StatusCreated)
	sr.WriteHeader(http.StatusInternalServerError)
	fmt.Fprint(sr, "ok")

	if got := sr.RecordedStatus(); got != http.StatusCreated {
		t.Errorf("RecordedStatus = %d, chci %d", got, http.StatusCreated)
	}
	if rec.Code != http.StatusCreated {
		t.Errorf("status na odpovědi = %d, chci %d", rec.Code, http.StatusCreated)
	}
}

func TestChainOrder(t *testing.T) {
	var events []string

	handler := exercise.Chain(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			events = append(events, "handler")
		}),
		trace(&events, "A"),
		trace(&events, "B"),
	)

	handler.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))

	want := []string{"A:pred", "B:pred", "handler", "B:po", "A:po"}
	if strings.Join(events, ",") != strings.Join(want, ",") {
		t.Errorf("pořadí = %v, chci %v (první middleware je nejvíc vnější)", events, want)
	}
}

func TestChainWithoutMiddleware(t *testing.T) {
	rec := httptest.NewRecorder()

	exercise.Chain(okHandler("holy handler")).
		ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	if got := rec.Body.String(); got != "holy handler" {
		t.Errorf("tělo = %q, chci %q", got, "holy handler")
	}
}

// runWithLogger prožene handler middlewarem Logging a vrátí rozparsovaný log záznam.
func runWithLogger(t *testing.T, h http.Handler, method, target string) (map[string]any, *httptest.ResponseRecorder) {
	t.Helper()

	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, nil))

	rec := httptest.NewRecorder()
	exercise.Logging(logger)(h).ServeHTTP(rec, httptest.NewRequest(method, target, nil))

	line := strings.TrimSpace(buf.String())
	if line == "" {
		t.Fatal("Logging nic nezalogoval")
	}
	if i := strings.IndexByte(line, '\n'); i >= 0 {
		line = line[:i]
	}

	var entry map[string]any
	if err := json.Unmarshal([]byte(line), &entry); err != nil {
		t.Fatalf("log není platný JSON: %v (%q)", err, line)
	}
	return entry, rec
}

func TestLoggingWritesFields(t *testing.T) {
	entry, rec := runWithLogger(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTeapot)
		fmt.Fprint(w, "12345")
	}), http.MethodPost, "/orders/7")

	if got, ok := entry["msg"].(string); !ok || got != "request" {
		t.Errorf("msg = %v, chci %q", entry["msg"], "request")
	}
	if got, ok := entry["method"].(string); !ok || got != http.MethodPost {
		t.Errorf("method = %v, chci %q", entry["method"], http.MethodPost)
	}
	if got, ok := entry["path"].(string); !ok || got != "/orders/7" {
		t.Errorf("path = %v, chci %q", entry["path"], "/orders/7")
	}
	if got, ok := entry["status"].(float64); !ok || int(got) != http.StatusTeapot {
		t.Errorf("status = %v, chci %d", entry["status"], http.StatusTeapot)
	}
	if got, ok := entry["bytes"].(float64); !ok || int(got) != 5 {
		t.Errorf("bytes = %v, chci %d", entry["bytes"], 5)
	}

	if rec.Code != http.StatusTeapot {
		t.Errorf("status na odpovědi = %d, chci %d", rec.Code, http.StatusTeapot)
	}
	if rec.Body.String() != "12345" {
		t.Errorf("tělo odpovědi = %q, chci %q", rec.Body.String(), "12345")
	}
}

func TestLoggingDefaultStatusIs200(t *testing.T) {
	entry, _ := runWithLogger(t, okHandler("ahoj"), http.MethodGet, "/")

	if got, ok := entry["status"].(float64); !ok || int(got) != http.StatusOK {
		t.Errorf("status = %v, chci %d (handler WriteHeader nezavolal)", entry["status"], http.StatusOK)
	}
	if got, ok := entry["bytes"].(float64); !ok || int(got) != 4 {
		t.Errorf("bytes = %v, chci %d", entry["bytes"], 4)
	}
}

func TestRecoveryTurnsPanicInto500(t *testing.T) {
	handler := exercise.Recovery()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("něco se pokazilo")
	}))

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, chci %d", rec.Code, http.StatusInternalServerError)
	}
	if ct := rec.Header().Get("Content-Type"); !strings.HasPrefix(ct, "application/json") {
		t.Errorf("Content-Type = %q, chci application/json", ct)
	}

	var body exercise.ErrorResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("tělo není platný JSON: %v (%q)", err, rec.Body.String())
	}
	if body.Error == "" {
		t.Error("tělo má prázdné pole error")
	}
	if strings.Contains(body.Error, "něco se pokazilo") {
		t.Error("chybová odpověď nemá klientovi prozrazovat text paniky")
	}
}

func TestRecoveryPassesThroughWithoutPanic(t *testing.T) {
	rec := httptest.NewRecorder()

	exercise.Recovery()(okHandler("v pohodě")).
		ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, chci %d", rec.Code, http.StatusOK)
	}
	if rec.Body.String() != "v pohodě" {
		t.Errorf("tělo = %q, chci %q", rec.Body.String(), "v pohodě")
	}
}

func TestRecoveryDoesNotSwallowErrAbortHandler(t *testing.T) {
	defer func() {
		recovered := recover()
		if recovered != http.ErrAbortHandler {
			t.Errorf("recover() = %v, chci http.ErrAbortHandler (Recovery ho musí poslat dál)", recovered)
		}
	}()

	handler := exercise.Recovery()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic(http.ErrAbortHandler)
	}))
	handler.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))

	t.Error("Recovery paniku spolkla, ale http.ErrAbortHandler má propustit")
}

func TestChainMultipleMiddlewareTogether(t *testing.T) {
	var events []string
	var loggedStatus int

	setTag := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			events = append(events, "tag:pred")
			w.Header().Set(testTagHeader, "req-9")
			next.ServeHTTP(w, r)
			events = append(events, "tag:po")
		})
	}
	observeStatus := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			events = append(events, "log:pred")
			rec := httptest.NewRecorder()
			next.ServeHTTP(rec, r)
			loggedStatus = rec.Code
			for k, vv := range rec.Header() {
				for _, v := range vv {
					w.Header().Add(k, v)
				}
			}
			w.WriteHeader(rec.Code)
			_, _ = w.Write(rec.Body.Bytes())
			events = append(events, "log:po")
		})
	}
	recoverPanic := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			events = append(events, "rec:pred")
			defer func() {
				if recover() != nil {
					w.WriteHeader(http.StatusInternalServerError)
					_, _ = w.Write([]byte(`{"error":"internal"}`))
				}
			}()
			next.ServeHTTP(w, r)
			events = append(events, "rec:po")
		})
	}

	handler := exercise.Chain(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			panic("boom")
		}),
		setTag,
		observeStatus,
		recoverPanic,
	)

	req := httptest.NewRequest(http.MethodGet, "/pad", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, chci %d", rec.Code, http.StatusInternalServerError)
	}
	if got := rec.Header().Get(testTagHeader); got != "req-9" {
		t.Errorf("hlavička %s = %q, chci %q", testTagHeader, got, "req-9")
	}
	if loggedStatus != http.StatusInternalServerError {
		t.Errorf("status viděný prostředním middlewarem = %d, chci %d (recovery je vnitřnější)",
			loggedStatus, http.StatusInternalServerError)
	}
	want := []string{"tag:pred", "log:pred", "rec:pred", "log:po", "tag:po"}
	if strings.Join(events, ",") != strings.Join(want, ",") {
		t.Errorf("pořadí = %v, chci %v", events, want)
	}
}

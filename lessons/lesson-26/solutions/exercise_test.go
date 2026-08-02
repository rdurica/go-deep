package solutions_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	exercise "github.com/rdurica/go-deep/lessons/lesson-26/solutions"
)

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

	// obalený writer musí odpověď stále doručit beze změny
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

func TestLoggingSecondWriteHeaderDoesNotOverwriteStatus(t *testing.T) {
	entry, rec := runWithLogger(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		w.WriteHeader(http.StatusInternalServerError) // nadbytečné volání, musí se ignorovat
		fmt.Fprint(w, "ok")
	}), http.MethodGet, "/")

	if got, ok := entry["status"].(float64); !ok || int(got) != http.StatusCreated {
		t.Errorf("zalogovaný status = %v, chci %d", entry["status"], http.StatusCreated)
	}
	if rec.Code != http.StatusCreated {
		t.Errorf("status na odpovědi = %d, chci %d", rec.Code, http.StatusCreated)
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

// idCapturingHandler uloží ID požadavku z kontextu do got.
func idCapturingHandler(got *string, found *bool) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		*got, *found = exercise.RequestIDFrom(r.Context())
	})
}

func TestRequestIDPropagatesHeader(t *testing.T) {
	var fromCtx string
	var found bool

	handler := exercise.RequestID()(idCapturingHandler(&fromCtx, &found))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(exercise.RequestIDHeader, "abc-123")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if !found {
		t.Fatal("RequestIDFrom(ctx) nenašlo ID")
	}
	if fromCtx != "abc-123" {
		t.Errorf("ID v kontextu = %q, chci %q", fromCtx, "abc-123")
	}
	if got := rec.Header().Get(exercise.RequestIDHeader); got != "abc-123" {
		t.Errorf("hlavička v odpovědi = %q, chci %q", got, "abc-123")
	}
}

func TestRequestIDGeneratesUniqueID(t *testing.T) {
	var fromCtx string
	var found bool
	handler := exercise.RequestID()(idCapturingHandler(&fromCtx, &found))

	seen := make(map[string]bool)
	for i := 0; i < 20; i++ {
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

		if !found || fromCtx == "" {
			t.Fatalf("iterace %d: middleware nevygeneroval ID", i)
		}
		if got := rec.Header().Get(exercise.RequestIDHeader); got != fromCtx {
			t.Fatalf("iterace %d: hlavička = %q, kontext = %q, chci shodu", i, got, fromCtx)
		}
		if seen[fromCtx] {
			t.Fatalf("iterace %d: ID %q se zopakovalo", i, fromCtx)
		}
		seen[fromCtx] = true
	}
}

func TestRequestIDFromEmptyContext(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	if id, ok := exercise.RequestIDFrom(req.Context()); ok {
		t.Errorf("RequestIDFrom bez middlewaru = (%q, true), chci (\"\", false)", id)
	}
}

func TestTimeoutFastHandlerPasses(t *testing.T) {
	rec := httptest.NewRecorder()

	exercise.Timeout(time.Second)(okHandler("hotovo")).
		ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, chci %d", rec.Code, http.StatusOK)
	}
	if rec.Body.String() != "hotovo" {
		t.Errorf("tělo = %q, chci %q", rec.Body.String(), "hotovo")
	}
}

func TestTimeoutSetsDeadline(t *testing.T) {
	var hasDeadline bool

	exercise.Timeout(time.Second)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, hasDeadline = r.Context().Deadline()
	})).ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))

	if !hasDeadline {
		t.Error("handler nedostal kontext s deadline")
	}
}

func TestTimeoutSlowHandlerGets503(t *testing.T) {
	done := make(chan struct{})

	slow := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-r.Context().Done() // handler korektně čeká na zrušení, ne na pevný sleep
		close(done)
	})

	rec := httptest.NewRecorder()
	exercise.Timeout(50*time.Millisecond)(slow).
		ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, chci %d", rec.Code, http.StatusServiceUnavailable)
	}
	if !strings.Contains(rec.Body.String(), "timeout") {
		t.Errorf("tělo = %q, chci aby obsahovalo %q", rec.Body.String(), "timeout")
	}

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Error("handler po vypršení deadline neskončil")
	}
}

func TestChainMultipleMiddlewareTogether(t *testing.T) {
	var events []string
	var loggedStatus int

	// Anonymní middleware — ne Logging/Recovery/RequestID stuby.
	setID := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			events = append(events, "id:pred")
			w.Header().Set(exercise.RequestIDHeader, "req-9")
			next.ServeHTTP(w, r)
			events = append(events, "id:po")
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
		setID,
		observeStatus,
		recoverPanic,
	)

	req := httptest.NewRequest(http.MethodGet, "/pad", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, chci %d", rec.Code, http.StatusInternalServerError)
	}
	if got := rec.Header().Get(exercise.RequestIDHeader); got != "req-9" {
		t.Errorf("hlavička %s = %q, chci %q", exercise.RequestIDHeader, got, "req-9")
	}
	if loggedStatus != http.StatusInternalServerError {
		t.Errorf("status viděný prostředním middlewarem = %d, chci %d (recovery je vnitřnější)",
			loggedStatus, http.StatusInternalServerError)
	}
	want := []string{"id:pred", "log:pred", "rec:pred", "log:po", "id:po"}
	if strings.Join(events, ",") != strings.Join(want, ",") {
		t.Errorf("pořadí = %v, chci %v", events, want)
	}
}

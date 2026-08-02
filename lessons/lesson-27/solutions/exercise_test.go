package solutions_test

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	exercise "github.com/rdurica/go-deep/lessons/lesson-27/solutions"
)

var testUsers = map[string]exercise.User{
	"tok-radek": {ID: "1", Name: "Radek"},
	"tok-eva":   {ID: "2", Name: "Eva"},
}

func TestWithUserAndUserFrom(t *testing.T) {
	want := exercise.User{ID: "42", Name: "Kdosi"}

	ctx := exercise.WithUser(context.Background(), want)

	got, ok := exercise.UserFrom(ctx)
	if !ok {
		t.Fatal("UserFrom nenašlo uživatele v kontextu, který ho má obsahovat")
	}
	if got != want {
		t.Errorf("UserFrom(ctx) = %+v, chci %+v", got, want)
	}
}

func TestUserFromEmptyContext(t *testing.T) {
	got, ok := exercise.UserFrom(context.Background())
	if ok {
		t.Errorf("UserFrom(Background()) = (%+v, true), chci (User{}, false)", got)
	}
	if got != (exercise.User{}) {
		t.Errorf("UserFrom(Background()) vrátilo %+v, chci zero value", got)
	}
}

func TestWithUserDoesNotMutateOriginalContext(t *testing.T) {
	parent := context.Background()

	_ = exercise.WithUser(parent, exercise.User{ID: "1", Name: "Radek"})

	if _, ok := exercise.UserFrom(parent); ok {
		t.Error("WithUser zmutoval rodičovský kontext, má vracet kopii")
	}
}

func TestWithUserCanNest(t *testing.T) {
	ctx := exercise.WithUser(context.Background(), exercise.User{ID: "1", Name: "Radek"})
	ctx = exercise.WithUser(ctx, exercise.User{ID: "2", Name: "Eva"})

	got, ok := exercise.UserFrom(ctx)
	if !ok || got.Name != "Eva" {
		t.Errorf("UserFrom(ctx) = (%+v, %v), chci uživatele Eva", got, ok)
	}
}

func TestWhoAmIWithoutUserInContext(t *testing.T) {
	rec := httptest.NewRecorder()

	exercise.WhoAmI().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/me", nil))

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, chci %d", rec.Code, http.StatusInternalServerError)
	}
}

func TestAuthenticateRejects(t *testing.T) {
	handler := exercise.Authenticate(testUsers)(exercise.WhoAmI())

	tests := []struct {
		name   string
		header string
	}{
		{"missing header", ""},
		{"other scheme", "Basic cmFkZWs6aGVzbG8="},
		{"empty token", "Bearer "},
		{"unknown token", "Bearer tok-nikdo"},
		{"token without scheme", "tok-radek"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/me", nil)
			if tt.header != "" {
				req.Header.Set("Authorization", tt.header)
			}
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			if rec.Code != http.StatusUnauthorized {
				t.Fatalf("Authorization %q dalo status %d, chci %d", tt.header, rec.Code, http.StatusUnauthorized)
			}
			if ct := rec.Header().Get("Content-Type"); !strings.HasPrefix(ct, "application/json") {
				t.Errorf("Content-Type = %q, chci application/json", ct)
			}

			var body exercise.ErrorResponse
			if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
				t.Fatalf("tělo není platný JSON: %v (%q)", err, rec.Body.String())
			}
			if body.Error == "" {
				t.Error("401 odpověď má prázdné pole error")
			}
		})
	}
}

func TestAuthenticateAllowsAndInjectsUser(t *testing.T) {
	srv := httptest.NewServer(exercise.Authenticate(testUsers)(exercise.WhoAmI()))
	defer srv.Close()

	tests := []struct {
		name   string
		header string
		want   exercise.User
	}{
		{"valid token", "Bearer tok-radek", exercise.User{ID: "1", Name: "Radek"}},
		{"second user", "Bearer tok-eva", exercise.User{ID: "2", Name: "Eva"}},
		{"scheme is case-insensitive", "bearer tok-radek", exercise.User{ID: "1", Name: "Radek"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodGet, srv.URL+"/me", nil)
			if err != nil {
				t.Fatalf("nepovedlo se sestavit požadavek: %v", err)
			}
			req.Header.Set("Authorization", tt.header)

			res, err := srv.Client().Do(req)
			if err != nil {
				t.Fatalf("požadavek selhal: %v", err)
			}
			defer res.Body.Close()

			if res.StatusCode != http.StatusOK {
				t.Fatalf("status = %d, chci %d", res.StatusCode, http.StatusOK)
			}

			raw, _ := io.ReadAll(res.Body)
			var got exercise.User
			if err := json.Unmarshal(raw, &got); err != nil {
				t.Fatalf("tělo není platný JSON: %v (%q)", err, string(raw))
			}
			if got != tt.want {
				t.Errorf("WhoAmI vrátilo %+v, chci %+v", got, tt.want)
			}
		})
	}
}

func TestFetchWithTimeoutSuccess(t *testing.T) {
	got, err := exercise.FetchWithTimeout(context.Background(), func(ctx context.Context) (string, error) {
		return "hotovo", nil
	}, time.Second)

	if err != nil {
		t.Fatalf("FetchWithTimeout vrátil chybu %v, chci nil", err)
	}
	if got != "hotovo" {
		t.Errorf("FetchWithTimeout = %q, chci %q", got, "hotovo")
	}
}

func TestFetchWithTimeoutForwardsError(t *testing.T) {
	wantErr := errors.New("selhalo to")

	got, err := exercise.FetchWithTimeout(context.Background(), func(ctx context.Context) (string, error) {
		return "", wantErr
	}, time.Second)

	if !errors.Is(err, wantErr) {
		t.Errorf("FetchWithTimeout err = %v, chci %v", err, wantErr)
	}
	if got != "" {
		t.Errorf("FetchWithTimeout = %q, chci prázdný řetězec", got)
	}
}

func TestFetchWithTimeoutDeadline(t *testing.T) {
	started := make(chan struct{})
	finished := make(chan error, 1)

	got, err := exercise.FetchWithTimeout(context.Background(), func(ctx context.Context) (string, error) {
		close(started)
		<-ctx.Done() // funkce spolupracuje: čeká na zrušení, ne na pevný sleep
		finished <- ctx.Err()
		return "", ctx.Err()
	}, 30*time.Millisecond)

	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("FetchWithTimeout err = %v, chci context.DeadlineExceeded", err)
	}
	if got != "" {
		t.Errorf("FetchWithTimeout = %q, chci prázdný řetězec", got)
	}

	select {
	case <-started:
	case <-time.After(2 * time.Second):
		t.Fatal("fn se vůbec nespustila")
	}

	select {
	case innerErr := <-finished:
		if !errors.Is(innerErr, context.DeadlineExceeded) {
			t.Errorf("fn viděla chybu %v, chci context.DeadlineExceeded (kontext musí propadnout dovnitř)", innerErr)
		}
	case <-time.After(2 * time.Second):
		t.Error("fn nedostala signál o zrušení")
	}
}

func TestFetchWithTimeoutCanceledParent(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := exercise.FetchWithTimeout(ctx, func(ctx context.Context) (string, error) {
		<-ctx.Done()
		return "", ctx.Err()
	}, time.Hour)

	if !errors.Is(err, context.Canceled) {
		t.Errorf("FetchWithTimeout err = %v, chci context.Canceled", err)
	}
}

func TestSlowHandlerCompletesWork(t *testing.T) {
	rec := httptest.NewRecorder()

	exercise.SlowHandler(20*time.Millisecond).
		ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/slow", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, chci %d", rec.Code, http.StatusOK)
	}

	var body exercise.StatusResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("tělo není platný JSON: %v (%q)", err, rec.Body.String())
	}
	if body.Status != "done" {
		t.Errorf("status v těle = %q, chci %q", body.Status, "done")
	}
}

func TestSlowHandlerStopsOnClientDisconnect(t *testing.T) {
	exited := make(chan error, 1)

	handler := exercise.SlowHandlerWithHook(10*time.Second, func(err error) {
		exited <- err
	})

	srv := httptest.NewServer(handler)
	defer srv.Close()

	client := &http.Client{Timeout: 100 * time.Millisecond}
	res, err := client.Get(srv.URL + "/slow")
	if err == nil {
		res.Body.Close()
		t.Fatal("klient dostal odpověď, ale měl vypršet timeout")
	}

	// Handler musí skončit krátce po odpojení klienta, ne až po 10 vteřinách práce.
	select {
	case exitErr := <-exited:
		if exitErr == nil {
			t.Error("handler skončil s nil chybou, chci důvod zrušení z kontextu")
		}
	case <-time.After(3 * time.Second):
		t.Fatal("handler po odpojení klienta neskončil — nekontroluje ctx.Done()")
	}
}

func TestSlowHandlerHookOnSuccess(t *testing.T) {
	exited := make(chan error, 1)

	handler := exercise.SlowHandlerWithHook(20*time.Millisecond, func(err error) {
		exited <- err
	})

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/slow", nil))

	select {
	case exitErr := <-exited:
		if exitErr != nil {
			t.Errorf("hook dostal %v, chci nil při dokončení práce", exitErr)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("hook se nezavolal")
	}
	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, chci %d", rec.Code, http.StatusOK)
	}
}

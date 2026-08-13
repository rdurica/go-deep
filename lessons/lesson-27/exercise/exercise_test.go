package exercise_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	exercise "github.com/rdurica/go-deep/lessons/lesson-27/exercise"
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
	ctx, cancel := context.WithCancel(context.Background())
	req := httptest.NewRequest(http.MethodGet, "/slow", nil).WithContext(ctx)

	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	start := time.Now()
	rec := httptest.NewRecorder()
	exercise.SlowHandler(10*time.Second).ServeHTTP(rec, req)

	if elapsed := time.Since(start); elapsed > 2*time.Second {
		t.Fatalf("handler běžel %v po zrušení kontextu, chci skončit do ~2 s", elapsed)
	}
	if rec.Body.Len() > 0 {
		t.Error("po zrušení kontextu handler nesmí psát do odpovědi")
	}
}

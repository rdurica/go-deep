package exercise_test

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	exercise "github.com/rdurica/go-deep/lessons/lesson-30/exercise"
)

type user struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

func TestNewHTTPClient(t *testing.T) {
	t.Parallel()

	c := exercise.NewHTTPClient(3 * time.Second)
	if c == nil {
		t.Fatal("NewHTTPClient vrátil nil")
	}
	if c.Timeout != 3*time.Second {
		t.Errorf("Timeout = %v, chci %v", c.Timeout, 3*time.Second)
	}
	if c.Transport == nil {
		t.Fatal("Transport je nil — klient by použil http.DefaultTransport sdílený s celým procesem")
	}
	tr, ok := c.Transport.(*http.Transport)
	if !ok {
		t.Fatalf("Transport má typ %T, chci *http.Transport", c.Transport)
	}
	if tr.MaxIdleConnsPerHost <= 0 {
		t.Errorf("MaxIdleConnsPerHost = %d, chci kladnou hodnotu", tr.MaxIdleConnsPerHost)
	}
}

func TestFetchJSONHappyPath(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"id":7,"name":"radek"}`)
	}))
	defer srv.Close()

	got, err := exercise.FetchJSON[user](context.Background(), srv.Client(), srv.URL)
	if err != nil {
		t.Fatalf("FetchJSON vrátilo chybu %v", err)
	}
	want := user{ID: 7, Name: "radek"}
	if got != want {
		t.Errorf("FetchJSON = %+v, chci %+v", got, want)
	}
}

func TestFetchJSONStatusError(t *testing.T) {
	t.Parallel()

	tests := []int{http.StatusInternalServerError, http.StatusNotFound, http.StatusMovedPermanently}
	for _, status := range tests {
		t.Run(fmt.Sprintf("status-%d", status), func(t *testing.T) {
			t.Parallel()

			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(status)
				_, _ = io.WriteString(w, `{"id":1,"name":"x"}`)
			}))
			defer srv.Close()

			// Přesměrování nechceme následovat automaticky, ať test měří status.
			client := srv.Client()
			client.CheckRedirect = func(*http.Request, []*http.Request) error {
				return http.ErrUseLastResponse
			}

			_, err := exercise.FetchJSON[user](context.Background(), client, srv.URL)
			if err == nil {
				t.Fatalf("FetchJSON při statusu %d vrátilo nil, chci chybu", status)
			}
			var statusErr *exercise.StatusError
			if !errors.As(err, &statusErr) {
				t.Fatalf("chyba %v není *StatusError", err)
			}
			if statusErr.StatusCode != status {
				t.Errorf("StatusCode = %d, chci %d", statusErr.StatusCode, status)
			}
		})
	}
}

func TestFetchJSONInvalidJSON(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, `{"id":`)
	}))
	defer srv.Close()

	if _, err := exercise.FetchJSON[user](context.Background(), srv.Client(), srv.URL); err == nil {
		t.Fatal("FetchJSON nad nevalidním JSONem vrátilo nil, chci chybu")
	}
}

func TestFetchJSONBodyTooLarge(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, `{"id":1,"name":"`)
		_, _ = io.WriteString(w, strings.Repeat("a", int(exercise.MaxBodyBytes)+16))
		_, _ = io.WriteString(w, `"}`)
	}))
	defer srv.Close()

	_, err := exercise.FetchJSON[user](context.Background(), srv.Client(), srv.URL)
	if !errors.Is(err, exercise.ErrBodyTooLarge) {
		t.Fatalf("FetchJSON = %v, chci ErrBodyTooLarge", err)
	}
}

func TestFetchJSONTimeout(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		<-r.Context().Done() // odpověď nikdy nepřijde
	}))
	defer srv.Close()

	client := exercise.NewHTTPClient(50 * time.Millisecond)
	start := time.Now()
	_, err := exercise.FetchJSON[user](context.Background(), client, srv.URL)
	elapsed := time.Since(start)

	if err == nil {
		t.Fatal("FetchJSON proti zaseknutému serveru vrátilo nil, chci chybu")
	}
	if elapsed > 2*time.Second {
		t.Errorf("FetchJSON trvalo %v — timeout klienta se neuplatnil", elapsed)
	}
}

func TestFetchJSONContextCancel(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		<-r.Context().Done()
	}))
	defer srv.Close()

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(20 * time.Millisecond)
		cancel()
	}()
	defer cancel()

	_, err := exercise.FetchJSON[user](ctx, srv.Client(), srv.URL)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("FetchJSON = %v, chci chybu obalující context.Canceled", err)
	}
}

func TestRetrySucceedsFirstTime(t *testing.T) {
	t.Parallel()

	var calls atomic.Int32
	err := exercise.Retry(context.Background(), 3, time.Millisecond, func(context.Context) error {
		calls.Add(1)
		return nil
	})
	if err != nil {
		t.Fatalf("Retry = %v, chci nil", err)
	}
	if got := calls.Load(); got != 1 {
		t.Errorf("fn zavolána %dx, chci 1x", got)
	}
}

func TestRetrySucceedsAfterFailures(t *testing.T) {
	t.Parallel()

	var calls atomic.Int32
	start := time.Now()
	err := exercise.Retry(context.Background(), 5, time.Millisecond, func(context.Context) error {
		if calls.Add(1) < 3 {
			return errors.New("dočasná chyba")
		}
		return nil
	})
	if err != nil {
		t.Fatalf("Retry = %v, chci nil", err)
	}
	if got := calls.Load(); got != 3 {
		t.Errorf("fn zavolána %dx, chci 3x", got)
	}
	if elapsed := time.Since(start); elapsed > time.Second {
		t.Errorf("Retry trvalo %v — backoff vychází z base, nemá být pevný", elapsed)
	}
}

func TestRetryExhausted(t *testing.T) {
	t.Parallel()

	sentinel := errors.New("pořád to padá")
	var calls atomic.Int32
	start := time.Now()

	err := exercise.Retry(context.Background(), 4, time.Millisecond, func(context.Context) error {
		calls.Add(1)
		return sentinel
	})
	if err == nil {
		t.Fatal("Retry po vyčerpání pokusů vrátil nil, chci chybu")
	}
	if !errors.Is(err, sentinel) {
		t.Errorf("chyba %v neobaluje poslední chybu z fn", err)
	}
	if got := calls.Load(); got != 4 {
		t.Errorf("fn zavolána %dx, chci 4x", got)
	}
	if elapsed := time.Since(start); elapsed > time.Second {
		t.Errorf("Retry trvalo %v, chci výrazně méně", elapsed)
	}
}

func TestRetryBackoffRoste(t *testing.T) {
	t.Parallel()

	// Osm pokusů se základem 5 ms nemůže při exponenciálním růstu trvat
	// tak málo jako při konstantní prodlevě.
	var calls atomic.Int32
	start := time.Now()
	_ = exercise.Retry(context.Background(), 5, 5*time.Millisecond, func(context.Context) error {
		calls.Add(1)
		return errors.New("nope")
	})
	elapsed := time.Since(start)

	// Součet minimálních prodlev (poloviny 5+10+20+40 ms) je 37,5 ms.
	if elapsed < 30*time.Millisecond {
		t.Errorf("Retry trvalo jen %v — čekáš mezi pokusy vůbec?", elapsed)
	}
	if elapsed > 3*time.Second {
		t.Errorf("Retry trvalo %v, backoff nemá růst takhle rychle", elapsed)
	}
}

func TestRetryPermanentError(t *testing.T) {
	t.Parallel()

	sentinel := errors.New("400 bad request")
	var calls atomic.Int32

	err := exercise.Retry(context.Background(), 5, time.Millisecond, func(context.Context) error {
		calls.Add(1)
		return exercise.Permanent(sentinel)
	})
	if err == nil {
		t.Fatal("Retry vrátil nil, chci chybu")
	}
	if got := calls.Load(); got != 1 {
		t.Errorf("fn zavolána %dx, chci 1x — trvalá chyba se neopakuje", got)
	}
	if !errors.Is(err, sentinel) {
		t.Errorf("chyba %v neobaluje původní chybu", err)
	}
	var permanent *exercise.PermanentError
	if !errors.As(err, &permanent) {
		t.Errorf("chyba %v se nedá přečíst jako *PermanentError", err)
	}
}

func TestRetryCanceledContext(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	var calls atomic.Int32
	err := exercise.Retry(ctx, 5, time.Millisecond, func(context.Context) error {
		calls.Add(1)
		return errors.New("nemělo se volat")
	})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("Retry = %v, chci chybu obalující context.Canceled", err)
	}
	if got := calls.Load(); got != 0 {
		t.Errorf("fn zavolána %dx, chci 0x — zrušený kontext se kontroluje před pokusem", got)
	}
}

func TestRetryCancelDuringBackoff(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var calls atomic.Int32
	start := time.Now()
	err := exercise.Retry(ctx, 10, time.Second, func(context.Context) error {
		if calls.Add(1) == 1 {
			cancel() // zrušíme ještě než začne čekání
		}
		return errors.New("dočasná chyba")
	})
	elapsed := time.Since(start)

	if !errors.Is(err, context.Canceled) {
		t.Fatalf("Retry = %v, chci chybu obalující context.Canceled", err)
	}
	if got := calls.Load(); got != 1 {
		t.Errorf("fn zavolána %dx, chci 1x", got)
	}
	if elapsed > 2*time.Second {
		t.Errorf("Retry čekal %v — čekání musí reagovat na zrušení kontextu", elapsed)
	}
}

func TestRetryInvalidAttempts(t *testing.T) {
	t.Parallel()

	var calls atomic.Int32
	err := exercise.Retry(context.Background(), 0, time.Millisecond, func(context.Context) error {
		calls.Add(1)
		return nil
	})
	if !errors.Is(err, exercise.ErrNoAttempts) {
		t.Errorf("Retry s attempts=0 = %v, chci ErrNoAttempts", err)
	}
	if got := calls.Load(); got != 0 {
		t.Errorf("fn zavolána %dx, chci 0x", got)
	}
}

func TestRunServerGracefulShutdown(t *testing.T) {
	t.Parallel()

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("net.Listen selhal: %v", err)
	}

	started := make(chan struct{})
	srv := &http.Server{
		Handler: http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			close(started)
			time.Sleep(100 * time.Millisecond) // rozpracovaná práce
			_, _ = io.WriteString(w, "hotovo")
		}),
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	runDone := make(chan error, 1)
	go func() {
		runDone <- exercise.RunServer(ctx, srv, ln)
	}()

	type result struct {
		body string
		err  error
	}
	reqDone := make(chan result, 1)
	go func() {
		resp, err := http.Get("http://" + ln.Addr().String() + "/")
		if err != nil {
			reqDone <- result{err: err}
			return
		}
		defer func() { _ = resp.Body.Close() }()
		body, err := io.ReadAll(resp.Body)
		reqDone <- result{body: string(body), err: err}
	}()

	select {
	case <-started:
	case <-time.After(2 * time.Second):
		t.Fatal("server se nespustil — RunServer musí obsluhovat listener")
	}
	cancel() // shutdown uprostřed rozpracovaného požadavku

	select {
	case res := <-reqDone:
		if res.err != nil {
			t.Fatalf("rozpracovaný požadavek selhal: %v", res.err)
		}
		if res.body != "hotovo" {
			t.Errorf("tělo odpovědi = %q, chci %q — shutdown nesmí utnout běžící požadavek", res.body, "hotovo")
		}
	case <-time.After(5 * time.Second):
		t.Fatal("rozpracovaný požadavek se nedokončil do 5 s")
	}

	select {
	case err := <-runDone:
		if err != nil {
			t.Errorf("RunServer = %v, chci nil při čistém ukončení", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("RunServer se nevrátil do 5 s po zrušení kontextu")
	}

	// Po ukončení už server nesmí přijímat spojení.
	if _, err := http.Get("http://" + ln.Addr().String() + "/"); err == nil {
		t.Error("server po ukončení pořád odpovídá")
	}
}

func TestRunServerListenerError(t *testing.T) {
	t.Parallel()

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("net.Listen selhal: %v", err)
	}
	_ = ln.Close() // Serve na zavřeném listeneru musí skončit chybou

	srv := &http.Server{Handler: http.NotFoundHandler()}
	done := make(chan error, 1)
	go func() {
		done <- exercise.RunServer(context.Background(), srv, ln)
	}()

	select {
	case err := <-done:
		if err == nil {
			t.Error("RunServer = nil, chci chybu z Serve")
		}
	case <-time.After(5 * time.Second):
		t.Fatal("RunServer se nevrátil do 5 s")
	}
}

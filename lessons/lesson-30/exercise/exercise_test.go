package exercise_test

import (
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"testing"
	"time"

	exercise "github.com/rdurica/go-deep/lessons/lesson-30/exercise"
)

func TestPermanentUnwrap(t *testing.T) {
	t.Parallel()

	sentinel := errors.New("400 bad request")
	wrapped := exercise.Permanent(sentinel)

	if !errors.Is(wrapped, sentinel) {
		t.Errorf("errors.Is na Permanent(%v) = false, chci true — Unwrap musí vracet obalenou chybu", sentinel)
	}
	var permanent *exercise.PermanentError
	if !errors.As(wrapped, &permanent) {
		t.Fatal("errors.As nevrátilo *PermanentError")
	}
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
			time.Sleep(100 * time.Millisecond)
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
	cancel()

	select {
	case res := <-reqDone:
		if res.err != nil {
			t.Fatalf("rozpracovaný požadavek selhal: %v", res.err)
		}
		if res.body != "hotovo" {
			t.Errorf("tělo odpovědi = %q, chci %q", res.body, "hotovo")
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
	_ = ln.Close()

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

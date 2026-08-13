package solutions_test

import (
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"

	exercise "github.com/rdurica/go-deep/lessons/lesson-20/solutions"
)

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func TestNewClientDefaultValues(t *testing.T) {
	c, err := exercise.NewClient("https://api.example.com/")
	if err != nil {
		t.Fatalf("NewClient vrátil chybu %v, chci nil", err)
	}
	if got := c.BaseURL(); got != "https://api.example.com" {
		t.Errorf("BaseURL() = %q, chci %q (koncové lomítko se ořezává)", got, "https://api.example.com")
	}
	if got := c.Timeout(); got != exercise.DefaultTimeout {
		t.Errorf("Timeout() = %v, chci %v", got, exercise.DefaultTimeout)
	}
	if got := c.Retries(); got != exercise.DefaultRetries {
		t.Errorf("Retries() = %d, chci %d", got, exercise.DefaultRetries)
	}
	if got := c.UserAgent(); got != exercise.DefaultUserAgent {
		t.Errorf("UserAgent() = %q, chci %q", got, exercise.DefaultUserAgent)
	}
}

func TestNewClientOptions(t *testing.T) {
	tests := []struct {
		name          string
		opts          []exercise.Option
		wantTimeout   time.Duration
		wantRetries   int
		wantUserAgent string
	}{
		{
			name:          "timeout only",
			opts:          []exercise.Option{exercise.WithTimeout(2 * time.Second)},
			wantTimeout:   2 * time.Second,
			wantRetries:   exercise.DefaultRetries,
			wantUserAgent: exercise.DefaultUserAgent,
		},
		{
			name: "all options",
			opts: []exercise.Option{
				exercise.WithTimeout(time.Minute),
				exercise.WithRetries(0),
				exercise.WithUserAgent("radek/2.0"),
			},
			wantTimeout:   time.Minute,
			wantRetries:   0,
			wantUserAgent: "radek/2.0",
		},
		{
			name: "last option wins",
			opts: []exercise.Option{
				exercise.WithRetries(7),
				exercise.WithRetries(9),
			},
			wantTimeout:   exercise.DefaultTimeout,
			wantRetries:   9,
			wantUserAgent: exercise.DefaultUserAgent,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, err := exercise.NewClient("http://localhost:8080", tt.opts...)
			if err != nil {
				t.Fatalf("NewClient vrátil chybu %v, chci nil", err)
			}
			if got := c.Timeout(); got != tt.wantTimeout {
				t.Errorf("Timeout() = %v, chci %v", got, tt.wantTimeout)
			}
			if got := c.Retries(); got != tt.wantRetries {
				t.Errorf("Retries() = %d, chci %d", got, tt.wantRetries)
			}
			if got := c.UserAgent(); got != tt.wantUserAgent {
				t.Errorf("UserAgent() = %q, chci %q", got, tt.wantUserAgent)
			}
		})
	}
}

func TestNewClientValidation(t *testing.T) {
	tests := []struct {
		name    string
		baseURL string
		opts    []exercise.Option
		wantErr error
	}{
		{"empty address", "", nil, exercise.ErrMissingBaseURL},
		{"missing scheme", "api.example.com", nil, exercise.ErrInvalidBaseURL},
		{"ftp scheme", "ftp://example.com", nil, exercise.ErrInvalidBaseURL},
		{
			"zero timeout",
			"https://example.com",
			[]exercise.Option{exercise.WithTimeout(0)},
			exercise.ErrInvalidTimeout,
		},
		{
			"negative timeout",
			"https://example.com",
			[]exercise.Option{exercise.WithTimeout(-time.Second)},
			exercise.ErrInvalidTimeout,
		},
		{
			"negative retries",
			"https://example.com",
			[]exercise.Option{exercise.WithRetries(-1)},
			exercise.ErrInvalidRetries,
		},
		{
			"empty user agent",
			"https://example.com",
			[]exercise.Option{exercise.WithUserAgent("")},
			exercise.ErrEmptyUserAgent,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, err := exercise.NewClient(tt.baseURL, tt.opts...)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("chyba = %v, chci %v", err, tt.wantErr)
			}
			if c != nil {
				t.Errorf("při chybě chci nil klienta, mám %+v", c)
			}
		})
	}
}

func TestNewServer(t *testing.T) {
	logger := testLogger()

	srv, err := exercise.NewServer(":8080", logger)
	if err != nil {
		t.Fatalf("NewServer vrátil chybu %v, chci nil", err)
	}
	if got := srv.Addr(); got != ":8080" {
		t.Errorf("Addr() = %q, chci %q", got, ":8080")
	}
	if srv.Logger() != logger {
		t.Error("Logger() nevrací logger předaný konstruktoru")
	}
}

func TestNewServerValidation(t *testing.T) {
	logger := testLogger()

	tests := []struct {
		name    string
		addr    string
		logger  *slog.Logger
		wantErr error
	}{
		{"empty address", "", logger, exercise.ErrMissingAddr},
		{"address without colon", "localhost", logger, exercise.ErrInvalidAddr},
		{"missing logger", ":8080", nil, exercise.ErrMissingLogger},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv, err := exercise.NewServer(tt.addr, tt.logger)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("chyba = %v, chci %v", err, tt.wantErr)
			}
			if srv != nil {
				t.Errorf("při chybě chci nil server, mám %+v", srv)
			}
		})
	}
}

func TestMustNewServer(t *testing.T) {
	t.Run("valid input does not panic", func(t *testing.T) {
		srv := exercise.MustNewServer("127.0.0.1:9000", testLogger())
		if got := srv.Addr(); got != "127.0.0.1:9000" {
			t.Errorf("Addr() = %q, chci %q", got, "127.0.0.1:9000")
		}
	})

	t.Run("invalid input panics", func(t *testing.T) {
		defer func() {
			r := recover()
			if r == nil {
				t.Fatal("MustNewServer nepanikoval, chci paniku")
			}
			err, ok := r.(error)
			if !ok || !errors.Is(err, exercise.ErrMissingLogger) {
				t.Errorf("panika nese %v, chci ErrMissingLogger", r)
			}
		}()
		exercise.MustNewServer(":8080", nil)
	})
}

func TestRegistryZeroValue(t *testing.T) {
	var reg exercise.Registry // žádný konstruktor — zero value musí fungovat

	if got := reg.Len(); got != 0 {
		t.Errorf("Len() = %d, chci 0", got)
	}
	if _, ok := reg.Lookup("chybí"); ok {
		t.Error("Lookup na prázdné Registry vrátil ok = true")
	}

	reg.Set("zeta", "3")
	reg.Set("alfa", "1")
	reg.Set("beta", "2")
	reg.Set("alfa", "42") // přepis

	if got := reg.Len(); got != 3 {
		t.Errorf("Len() = %d, chci 3", got)
	}
	if v, ok := reg.Lookup("alfa"); !ok || v != "42" {
		t.Errorf("Lookup(\"alfa\") = (%q, %v), chci (\"42\", true)", v, ok)
	}
}

func TestRegistryVeStructu(t *testing.T) {
	// Registry se dá vložit do jiného typu a pořád funguje bez inicializace.
	type app struct {
		flags exercise.Registry
	}
	var a app
	a.flags.Set("debug", "true")
	if v, ok := a.flags.Lookup("debug"); !ok || v != "true" {
		t.Errorf("vnořená Registry nefunguje: (%q, %v)", v, ok)
	}
}

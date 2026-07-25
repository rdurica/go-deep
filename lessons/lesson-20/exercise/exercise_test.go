package exercise_test

import (
	"errors"
	"io"
	"log/slog"
	"strings"
	"testing"
	"time"

	exercise "github.com/rdurica/go-deep/lessons/lesson-20/exercise"
)

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
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

func TestNewServerValidace(t *testing.T) {
	logger := testLogger()

	tests := []struct {
		name    string
		addr    string
		logger  *slog.Logger
		wantErr error
	}{
		{"prázdná adresa", "", logger, exercise.ErrMissingAddr},
		{"adresa bez dvojtečky", "localhost", logger, exercise.ErrInvalidAddr},
		{"chybějící logger", ":8080", nil, exercise.ErrMissingLogger},
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
	t.Run("platný vstup nepanikuje", func(t *testing.T) {
		srv := exercise.MustNewServer("127.0.0.1:9000", testLogger())
		if got := srv.Addr(); got != "127.0.0.1:9000" {
			t.Errorf("Addr() = %q, chci %q", got, "127.0.0.1:9000")
		}
	})

	t.Run("neplatný vstup panikuje", func(t *testing.T) {
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

func TestNewClientVychoziHodnoty(t *testing.T) {
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
			name:          "jen timeout",
			opts:          []exercise.Option{exercise.WithTimeout(2 * time.Second)},
			wantTimeout:   2 * time.Second,
			wantRetries:   exercise.DefaultRetries,
			wantUserAgent: exercise.DefaultUserAgent,
		},
		{
			name: "všechny options",
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
			name: "poslední option vyhrává",
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

func TestNewClientValidace(t *testing.T) {
	tests := []struct {
		name    string
		baseURL string
		opts    []exercise.Option
		wantErr error
	}{
		{"prázdná adresa", "", nil, exercise.ErrMissingBaseURL},
		{"chybí schéma", "api.example.com", nil, exercise.ErrInvalidBaseURL},
		{"ftp schéma", "ftp://example.com", nil, exercise.ErrInvalidBaseURL},
		{
			"nulový timeout",
			"https://example.com",
			[]exercise.Option{exercise.WithTimeout(0)},
			exercise.ErrInvalidTimeout,
		},
		{
			"záporný timeout",
			"https://example.com",
			[]exercise.Option{exercise.WithTimeout(-time.Second)},
			exercise.ErrInvalidTimeout,
		},
		{
			"záporné retries",
			"https://example.com",
			[]exercise.Option{exercise.WithRetries(-1)},
			exercise.ErrInvalidRetries,
		},
		{
			"prázdný user agent",
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

// memStore je in-memory fake implementující exercise.Store.
// Testovací dvojník patří k testu, ne do produkčního balíčku.
type memStore struct {
	records  []exercise.Record
	failWith error
}

func (m *memStore) Save(r exercise.Record) error {
	if m.failWith != nil {
		return m.failWith
	}
	m.records = append(m.records, r)
	return nil
}

func (m *memStore) All() []exercise.Record {
	return m.records
}

func TestServiceSeStorem(t *testing.T) {
	store := &memStore{}
	svc, err := exercise.NewService(store)
	if err != nil {
		t.Fatalf("NewService vrátil chybu %v, chci nil", err)
	}
	if got := svc.Count(); got != 0 {
		t.Errorf("Count() = %d, chci 0", got)
	}

	if err := svc.Add("a", "alfa"); err != nil {
		t.Fatalf("Add vrátil chybu %v, chci nil", err)
	}
	if err := svc.Add("b", "beta"); err != nil {
		t.Fatalf("Add vrátil chybu %v, chci nil", err)
	}

	if got := svc.Count(); got != 2 {
		t.Errorf("Count() = %d, chci 2", got)
	}
	want := []string{"alfa", "beta"}
	got := svc.Values()
	if len(got) != len(want) {
		t.Fatalf("Values() = %v, chci %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("Values() = %v, chci %v", got, want)
		}
	}
}

func TestServiceChyby(t *testing.T) {
	t.Run("nil store", func(t *testing.T) {
		svc, err := exercise.NewService(nil)
		if !errors.Is(err, exercise.ErrMissingStore) {
			t.Fatalf("chyba = %v, chci ErrMissingStore", err)
		}
		if svc != nil {
			t.Error("při chybě chci nil službu")
		}
	})

	t.Run("prázdné ID", func(t *testing.T) {
		svc, err := exercise.NewService(&memStore{})
		if err != nil {
			t.Fatalf("NewService vrátil chybu %v", err)
		}
		if err := svc.Add("", "hodnota"); !errors.Is(err, exercise.ErrEmptyRecordID) {
			t.Errorf("chyba = %v, chci ErrEmptyRecordID", err)
		}
		if got := svc.Count(); got != 0 {
			t.Errorf("Count() = %d, chci 0 — neplatný záznam se neukládá", got)
		}
	})

	t.Run("chyba ze storu je obalená", func(t *testing.T) {
		boom := errors.New("disk full")
		svc, err := exercise.NewService(&memStore{failWith: boom})
		if err != nil {
			t.Fatalf("NewService vrátil chybu %v", err)
		}
		err = svc.Add("a", "alfa")
		if !errors.Is(err, boom) {
			t.Fatalf("chyba = %v, chci obalenou %v", err, boom)
		}
		if !strings.Contains(err.Error(), `"a"`) {
			t.Errorf("chyba %q neobsahuje ID záznamu", err.Error())
		}
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
	if got := reg.Keys(); len(got) != 0 {
		t.Errorf("Keys() = %v, chci prázdný slice", got)
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
	want := []string{"alfa", "beta", "zeta"}
	got := reg.Keys()
	if len(got) != len(want) {
		t.Fatalf("Keys() = %v, chci %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("Keys() = %v, chci %v", got, want)
		}
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

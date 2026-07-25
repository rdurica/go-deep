package solutions_test

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	exercise "github.com/rdurica/go-deep/lessons/lesson-28/solutions"
)

// fakeEnv vyrobí getenv funkci nad mapou. Díky tomu testujeme bez globálního
// stavu procesu a testy mohou běžet paralelně.
func fakeEnv(m map[string]string) func(string) string {
	return func(key string) string { return m[key] }
}

// validEnv je minimální prostředí, ve kterém Load uspěje.
func validEnv() map[string]string {
	return map[string]string{
		"DATABASE_URL": "postgres://app:s3cr3t@db:5432/app",
		"API_KEY":      "sk-live-abcdef",
	}
}

func TestLookupString(t *testing.T) {
	t.Parallel()

	env := fakeEnv(map[string]string{"ADDR": "127.0.0.1", "EMPTY": ""})

	tests := []struct {
		name string
		key  string
		def  string
		want string
	}{
		{"nastavená hodnota vyhraje nad výchozí", "ADDR", "0.0.0.0", "127.0.0.1"},
		{"nenastavený klíč dá výchozí hodnotu", "MISSING", "0.0.0.0", "0.0.0.0"},
		{"prázdná hodnota se bere jako nenastavená", "EMPTY", "0.0.0.0", "0.0.0.0"},
		{"prázdná výchozí hodnota je v pořádku", "MISSING", "", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := exercise.LookupString(env, tt.key, tt.def); got != tt.want {
				t.Errorf("LookupString(env, %q, %q) = %q, chci %q", tt.key, tt.def, got, tt.want)
			}
		})
	}
}

func TestLookupInt(t *testing.T) {
	t.Parallel()

	env := fakeEnv(map[string]string{
		"PORT":     "9090",
		"NEGATIVE": "-3",
		"JUNK":     "osm tisíc",
		"EMPTY":    "",
	})

	tests := []struct {
		name    string
		key     string
		def     int
		want    int
		wantErr bool
	}{
		{"platné číslo", "PORT", 8080, 9090, false},
		{"záporné číslo je platný int", "NEGATIVE", 8080, -3, false},
		{"nenastavený klíč dá výchozí hodnotu", "MISSING", 8080, 8080, false},
		{"prázdná hodnota dá výchozí hodnotu", "EMPTY", 8080, 8080, false},
		{"nečíselná hodnota je chyba", "JUNK", 8080, 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := exercise.LookupInt(env, tt.key, tt.def)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("LookupInt(env, %q, %d) = (%d, nil), chci chybu", tt.key, tt.def, got)
				}
				if !errors.Is(err, exercise.ErrInvalid) {
					t.Errorf("chyba %v se nedá porovnat přes errors.Is s ErrInvalid", err)
				}
				if !strings.Contains(err.Error(), tt.key) {
					t.Errorf("chyba %q neobsahuje jméno klíče %q", err.Error(), tt.key)
				}
				return
			}
			if err != nil {
				t.Fatalf("LookupInt(env, %q, %d) vrátilo chybu %v", tt.key, tt.def, err)
			}
			if got != tt.want {
				t.Errorf("LookupInt(env, %q, %d) = %d, chci %d", tt.key, tt.def, got, tt.want)
			}
		})
	}
}

func TestLoadDefaults(t *testing.T) {
	t.Parallel()

	cfg, err := exercise.Load(fakeEnv(validEnv()))
	if err != nil {
		t.Fatalf("Load vrátilo chybu %v, chci nil", err)
	}

	want := exercise.Config{
		Addr:        "0.0.0.0",
		Port:        8080,
		ReadTimeout: 5 * time.Second,
		Debug:       false,
		DatabaseURL: "postgres://app:s3cr3t@db:5432/app",
		APIKey:      "sk-live-abcdef",
	}
	if cfg != want {
		t.Errorf("Load = %#v, chci %#v", cfg, want)
	}
}

func TestLoadOverrides(t *testing.T) {
	t.Parallel()

	env := validEnv()
	env["ADDR"] = "127.0.0.1"
	env["PORT"] = "9000"
	env["READ_TIMEOUT"] = "250ms"
	env["DEBUG"] = "true"

	cfg, err := exercise.Load(fakeEnv(env))
	if err != nil {
		t.Fatalf("Load vrátilo chybu %v, chci nil", err)
	}
	if cfg.Addr != "127.0.0.1" {
		t.Errorf("Addr = %q, chci %q", cfg.Addr, "127.0.0.1")
	}
	if cfg.Port != 9000 {
		t.Errorf("Port = %d, chci %d", cfg.Port, 9000)
	}
	if cfg.ReadTimeout != 250*time.Millisecond {
		t.Errorf("ReadTimeout = %v, chci %v", cfg.ReadTimeout, 250*time.Millisecond)
	}
	if !cfg.Debug {
		t.Error("Debug = false, chci true")
	}
}

func TestLoadMissingRequired(t *testing.T) {
	t.Parallel()

	_, err := exercise.Load(fakeEnv(map[string]string{}))
	if err == nil {
		t.Fatal("Load na prázdném prostředí vrátilo nil, chci chybu")
	}
	if !errors.Is(err, exercise.ErrMissing) {
		t.Errorf("chyba %v se nedá porovnat přes errors.Is s ErrMissing", err)
	}
	// Fail-fast musí ohlásit VŠECHNY chybějící hodnoty, ne jen tu první.
	for _, key := range []string{"DATABASE_URL", "API_KEY"} {
		if !strings.Contains(err.Error(), key) {
			t.Errorf("chyba %q neobsahuje %q — sbíráš chyby přes errors.Join?", err.Error(), key)
		}
	}
}

func TestLoadCollectsAllErrors(t *testing.T) {
	t.Parallel()

	env := map[string]string{
		"PORT":         "70000",
		"READ_TIMEOUT": "pět vteřin",
		"DEBUG":        "možná",
	}

	_, err := exercise.Load(fakeEnv(env))
	if err == nil {
		t.Fatal("Load na rozbitém prostředí vrátilo nil, chci chybu")
	}
	for _, key := range []string{"PORT", "READ_TIMEOUT", "DEBUG", "DATABASE_URL", "API_KEY"} {
		if !strings.Contains(err.Error(), key) {
			t.Errorf("chyba neobsahuje %q; celý text:\n%s", key, err.Error())
		}
	}
	if !errors.Is(err, exercise.ErrInvalid) {
		t.Error("chyba se nedá porovnat přes errors.Is s ErrInvalid")
	}
	if !errors.Is(err, exercise.ErrMissing) {
		t.Error("chyba se nedá porovnat přes errors.Is s ErrMissing")
	}
}

func TestLoadPortRange(t *testing.T) {
	t.Parallel()

	tests := []struct {
		port    string
		wantErr bool
	}{
		{"1", false},
		{"65535", false},
		{"0", true},
		{"-1", true},
		{"65536", true},
	}
	for _, tt := range tests {
		t.Run("PORT="+tt.port, func(t *testing.T) {
			env := validEnv()
			env["PORT"] = tt.port

			_, err := exercise.Load(fakeEnv(env))
			if tt.wantErr && err == nil {
				t.Errorf("Load s PORT=%s vrátilo nil, chci chybu", tt.port)
			}
			if !tt.wantErr && err != nil {
				t.Errorf("Load s PORT=%s vrátilo chybu %v", tt.port, err)
			}
		})
	}
}

// TestLoadWithOSGetenv ukazuje, že stejná funkce jde použít i s reálným
// prostředím procesu. t.Setenv se po testu automaticky uklidí.
func TestLoadWithOSGetenv(t *testing.T) {
	t.Setenv("ADDR", "10.0.0.1")
	t.Setenv("PORT", "3000")
	t.Setenv("READ_TIMEOUT", "1s")
	t.Setenv("DEBUG", "1")
	t.Setenv("DATABASE_URL", "postgres://app:s3cr3t@db:5432/app")
	t.Setenv("API_KEY", "sk-test")

	cfg, err := exercise.Load(os.Getenv)
	if err != nil {
		t.Fatalf("Load(os.Getenv) vrátilo chybu %v", err)
	}
	if cfg.Addr != "10.0.0.1" || cfg.Port != 3000 || cfg.ReadTimeout != time.Second || !cfg.Debug {
		t.Errorf("Load(os.Getenv) = %#v, nesedí s nastaveným prostředím", cfg)
	}
}

func TestConfigStringMasksSecrets(t *testing.T) {
	t.Parallel()

	const (
		password = "s3cr3t-heslo"
		apiKey   = "sk-live-9f8e7d6c"
	)
	cfg := exercise.Config{
		Addr:        "0.0.0.0",
		Port:        8080,
		ReadTimeout: 5 * time.Second,
		DatabaseURL: "postgres://app:" + password + "@db:5432/app",
		APIKey:      apiKey,
	}

	outputs := map[string]string{
		"String()": cfg.String(),
		"%v":       fmt.Sprintf("%v", cfg),
		"%+v":      fmt.Sprintf("%+v", cfg),
		"%s":       fmt.Sprintf("%s", cfg),
	}
	for verb, out := range outputs {
		if strings.Contains(out, password) {
			t.Errorf("%s vypsalo heslo: %s", verb, out)
		}
		if strings.Contains(out, apiKey) {
			t.Errorf("%s vypsalo API klíč: %s", verb, out)
		}
		// Netajné hodnoty naopak vidět musí, jinak by log byl k ničemu.
		if !strings.Contains(out, "8080") {
			t.Errorf("%s neobsahuje port: %s", verb, out)
		}
		if !strings.Contains(out, "db:5432") {
			t.Errorf("%s neobsahuje adresu databáze: %s", verb, out)
		}
	}
}

func TestConfigStringWithoutSecrets(t *testing.T) {
	t.Parallel()

	cfg := exercise.Config{Addr: "0.0.0.0", Port: 80, DatabaseURL: "postgres://db:5432/app"}
	out := cfg.String()
	if !strings.Contains(out, "postgres://db:5432/app") {
		t.Errorf("URL bez hesla se nemá měnit, dostal jsem %q", out)
	}
}

func TestLoadFromEnviron(t *testing.T) {
	t.Parallel()

	environ := []string{
		"PATH=/usr/bin",
		"ADDR=127.0.0.1",
		"PORT=9999",
		"DATABASE_URL=postgres://app:pwd@db:5432/app?sslmode=disable",
		"API_KEY=sk-abc",
		"DEBUG=true",
		"BROKEN_ENTRY_WITHOUT_EQUALS",
		"=hodnota-bez-klice",
	}

	cfg, err := exercise.LoadFromEnviron(environ)
	if err != nil {
		t.Fatalf("LoadFromEnviron vrátilo chybu %v", err)
	}
	if cfg.Addr != "127.0.0.1" {
		t.Errorf("Addr = %q, chci %q", cfg.Addr, "127.0.0.1")
	}
	if cfg.Port != 9999 {
		t.Errorf("Port = %d, chci %d", cfg.Port, 9999)
	}
	// Hodnota smí obsahovat '=' — dělí se jen na prvním výskytu.
	if want := "postgres://app:pwd@db:5432/app?sslmode=disable"; cfg.DatabaseURL != want {
		t.Errorf("DatabaseURL = %q, chci %q", cfg.DatabaseURL, want)
	}
}

func TestLoadFromEnvironMissing(t *testing.T) {
	t.Parallel()

	if _, err := exercise.LoadFromEnviron([]string{"PATH=/usr/bin"}); err == nil {
		t.Fatal("LoadFromEnviron bez povinných hodnot vrátilo nil, chci chybu")
	}
}

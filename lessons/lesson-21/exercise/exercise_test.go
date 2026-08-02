package exercise_test

import (
	"errors"
	"io"
	"strings"
	"testing"

	exercise "github.com/rdurica/go-deep/lessons/lesson-21/exercise"
)

func TestReadConfigValidInput(t *testing.T) {
	in := strings.Join([]string{
		"# konfigurace služby",
		"",
		"name = api",
		"port=8080",
		"   debug = true   ",
	}, "\n")

	got, err := exercise.ReadConfig(strings.NewReader(in))
	if err != nil {
		t.Fatalf("ReadConfig vrátil chybu %v, chci nil", err)
	}
	want := exercise.Config{Name: "api", Port: 8080, Debug: true}
	if got != want {
		t.Errorf("ReadConfig = %+v, chci %+v", got, want)
	}
}

func TestReadConfigOptionalDebug(t *testing.T) {
	got, err := exercise.ReadConfig(strings.NewReader("name=api\nport=1\n"))
	if err != nil {
		t.Fatalf("ReadConfig vrátil chybu %v, chci nil", err)
	}
	want := exercise.Config{Name: "api", Port: 1}
	if got != want {
		t.Errorf("ReadConfig = %+v, chci %+v", got, want)
	}
}

func TestReadConfigLastKeyWins(t *testing.T) {
	got, err := exercise.ReadConfig(strings.NewReader("name=a\nname=b\nport=80\nport=443\n"))
	if err != nil {
		t.Fatalf("ReadConfig vrátil chybu %v, chci nil", err)
	}
	want := exercise.Config{Name: "b", Port: 443}
	if got != want {
		t.Errorf("ReadConfig = %+v, chci %+v", got, want)
	}
}

func TestReadConfigErrors(t *testing.T) {
	tests := []struct {
		name    string
		in      string
		wantErr error
		wantMsg string
	}{
		{
			name:    "line without equals",
			in:      "name=api\nport=8080\nfoo\n",
			wantErr: exercise.ErrMalformedLine,
			wantMsg: `line 3: malformed line: "foo"`,
		},
		{
			name:    "unknown key",
			in:      "name=api\ncolour=red\nport=8080\n",
			wantErr: exercise.ErrUnknownKey,
			wantMsg: `line 2: unknown key: "colour"`,
		},
		{
			name:    "key is case sensitive",
			in:      "Name=api\n",
			wantErr: exercise.ErrUnknownKey,
			wantMsg: `line 1: unknown key: "Name"`,
		},
		{
			name:    "port is not a number",
			in:      "name=api\nport=abc\n",
			wantErr: exercise.ErrInvalidPort,
			wantMsg: `line 2: invalid port: "abc"`,
		},
		{
			name:    "port below range",
			in:      "name=api\nport=0\n",
			wantErr: exercise.ErrInvalidPort,
			wantMsg: `line 2: invalid port: "0"`,
		},
		{
			name:    "port above range",
			in:      "name=api\nport=70000\n",
			wantErr: exercise.ErrInvalidPort,
			wantMsg: `line 2: invalid port: "70000"`,
		},
		{
			name:    "debug is not bool",
			in:      "name=api\nport=80\ndebug=yes\n",
			wantErr: exercise.ErrInvalidBool,
			wantMsg: `line 3: invalid bool: "yes"`,
		},
		{
			name:    "missing name",
			in:      "port=8080\n",
			wantErr: exercise.ErrMissingKey,
			wantMsg: `missing key: "name"`,
		},
		{
			name:    "missing port",
			in:      "name=api\n",
			wantErr: exercise.ErrMissingKey,
			wantMsg: `missing key: "port"`,
		},
		{
			name:    "empty input reports name first",
			in:      "",
			wantErr: exercise.ErrMissingKey,
			wantMsg: `missing key: "name"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := exercise.ReadConfig(strings.NewReader(tt.in))
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("chyba = %v, chci obalenou %v", err, tt.wantErr)
			}
			if got := err.Error(); got != tt.wantMsg {
				t.Errorf("err.Error() = %q, chci %q", got, tt.wantMsg)
			}
			if cfg != (exercise.Config{}) {
				t.Errorf("při chybě chci nulový Config, mám %+v", cfg)
			}
		})
	}
}

// failingReader vždy selže — simuluje rozbitý soubor nebo síť.
type failingReader struct {
	err error
}

func (f failingReader) Read([]byte) (int, error) {
	return 0, f.err
}

func TestReadConfigReadError(t *testing.T) {
	boom := errors.New("disk on fire")

	_, err := exercise.ReadConfig(failingReader{err: boom})
	if !errors.Is(err, boom) {
		t.Fatalf("chyba = %v, chci obalenou %v", err, boom)
	}
	if got, want := err.Error(), "read config: disk on fire"; got != want {
		t.Errorf("err.Error() = %q, chci %q", got, want)
	}
}

func TestPipelineHappyPath(t *testing.T) {
	p := exercise.NewPipeline(
		exercise.Step{Name: "trim", Fn: func(s string) (string, error) {
			return strings.TrimSpace(s), nil
		}},
		exercise.Step{Name: "upper", Fn: func(s string) (string, error) {
			return strings.ToUpper(s), nil
		}},
		exercise.Step{Name: "suffix", Fn: func(s string) (string, error) {
			return s + "!", nil
		}},
	)

	got, err := p.Run("  ahoj  ")
	if err != nil {
		t.Fatalf("Run vrátil chybu %v, chci nil", err)
	}
	if want := "AHOJ!"; got != want {
		t.Errorf("Run = %q, chci %q", got, want)
	}
}

func TestPipelineEmpty(t *testing.T) {
	p := exercise.NewPipeline()

	got, err := p.Run("unchanged")
	if err != nil {
		t.Fatalf("Run vrátil chybu %v, chci nil", err)
	}
	if want := "unchanged"; got != want {
		t.Errorf("Run = %q, chci %q", got, want)
	}
}

func TestPipelineErrorCarriesStepName(t *testing.T) {
	errTooShort := errors.New("input too short")

	p := exercise.NewPipeline(
		exercise.Step{Name: "trim", Fn: func(s string) (string, error) {
			return strings.TrimSpace(s), nil
		}},
		exercise.Step{Name: "validate", Fn: func(s string) (string, error) {
			if len(s) < 5 {
				return "", errTooShort
			}
			return s, nil
		}},
		exercise.Step{Name: "nikdy", Fn: func(string) (string, error) {
			t.Error("krok za chybou se nesmí spustit")
			return "", nil
		}},
	)

	got, err := p.Run(" ok ")
	if !errors.Is(err, errTooShort) {
		t.Fatalf("chyba = %v, chci obalenou %v", err, errTooShort)
	}
	if want := `step "validate": input too short`; err.Error() != want {
		t.Errorf("err.Error() = %q, chci %q", err.Error(), want)
	}
	if got != "" {
		t.Errorf("při chybě chci prázdný výstup, mám %q", got)
	}
}

func TestPipelineContextChain(t *testing.T) {
	errBoom := errors.New("boom")

	inner := exercise.NewPipeline(
		exercise.Step{Name: "boom", Fn: func(string) (string, error) {
			return "", errBoom
		}},
	)
	outer := exercise.NewPipeline(
		exercise.Step{Name: "inner", Fn: inner.Run},
	)

	_, err := outer.Run("cokoli")
	if !errors.Is(err, errBoom) {
		t.Fatalf("chyba = %v, chci obalenou %v", err, errBoom)
	}
	want := `step "inner": step "boom": boom`
	if err.Error() != want {
		t.Errorf("err.Error() = %q, chci %q", err.Error(), want)
	}
}

func TestPipelineNilStep(t *testing.T) {
	p := exercise.NewPipeline(exercise.Step{Name: "broken"})

	_, err := p.Run("x")
	if !errors.Is(err, exercise.ErrNilStep) {
		t.Fatalf("chyba = %v, chci obalenou ErrNilStep", err)
	}
	if !strings.Contains(err.Error(), `step "broken"`) {
		t.Errorf("chyba %q neobsahuje název kroku", err.Error())
	}
}

// fakeCloser počítá zavření a umí selhat.
type fakeCloser struct {
	closed int
	err    error
}

func (c *fakeCloser) Close() error {
	c.closed++
	return c.err
}

func TestCloseAllSuccess(t *testing.T) {
	a, b := &fakeCloser{}, &fakeCloser{}

	if err := exercise.CloseAll([]io.Closer{a, b}); err != nil {
		t.Fatalf("CloseAll vrátil chybu %v, chci nil", err)
	}
	if a.closed != 1 || b.closed != 1 {
		t.Errorf("zavření = (%d, %d), chci (1, 1)", a.closed, b.closed)
	}
}

func TestCloseAllEmptyAndNil(t *testing.T) {
	if err := exercise.CloseAll(nil); err != nil {
		t.Errorf("CloseAll(nil) = %v, chci nil", err)
	}
	if err := exercise.CloseAll([]io.Closer{}); err != nil {
		t.Errorf("CloseAll([]) = %v, chci nil", err)
	}
	if err := exercise.CloseAll([]io.Closer{nil, nil}); err != nil {
		t.Errorf("CloseAll([nil, nil]) = %v, chci nil", err)
	}
}

func TestCloseAllJoinsErrors(t *testing.T) {
	errFirst := errors.New("first failed")
	errThird := errors.New("third failed")

	a := &fakeCloser{err: errFirst}
	b := &fakeCloser{}
	c := &fakeCloser{err: errThird}

	err := exercise.CloseAll([]io.Closer{a, nil, b, c})
	if err == nil {
		t.Fatal("CloseAll vrátil nil, chci spojenou chybu")
	}
	if !errors.Is(err, errFirst) {
		t.Errorf("chyba %v neobsahuje %v", err, errFirst)
	}
	if !errors.Is(err, errThird) {
		t.Errorf("chyba %v neobsahuje %v", err, errThird)
	}
	if a.closed != 1 || b.closed != 1 || c.closed != 1 {
		t.Errorf("zavření = (%d, %d, %d), chci (1, 1, 1) — zavírá se všechno", a.closed, b.closed, c.closed)
	}
}

func TestWithCleanup(t *testing.T) {
	errWork := errors.New("work failed")
	errClean := errors.New("cleanup failed")

	t.Run("both succeed", func(t *testing.T) {
		cleaned := 0
		err := exercise.WithCleanup(
			func() error { return nil },
			func() error { cleaned++; return nil },
		)
		if err != nil {
			t.Errorf("chyba = %v, chci nil", err)
		}
		if cleaned != 1 {
			t.Errorf("cleanup proběhl %dx, chci 1x", cleaned)
		}
	})

	t.Run("f fails, cleanup succeeds", func(t *testing.T) {
		cleaned := 0
		err := exercise.WithCleanup(
			func() error { return errWork },
			func() error { cleaned++; return nil },
		)
		if !errors.Is(err, errWork) {
			t.Fatalf("chyba = %v, chci %v", err, errWork)
		}
		if err.Error() != "work failed" {
			t.Errorf("err.Error() = %q, chci %q — chyba z f se nemá obalovat", err.Error(), "work failed")
		}
		if cleaned != 1 {
			t.Errorf("cleanup proběhl %dx, chci 1x i při chybě f", cleaned)
		}
	})

	t.Run("f succeeds, cleanup fails", func(t *testing.T) {
		err := exercise.WithCleanup(
			func() error { return nil },
			func() error { return errClean },
		)
		if !errors.Is(err, errClean) {
			t.Fatalf("chyba = %v, chci obalenou %v", err, errClean)
		}
		if want := "cleanup: cleanup failed"; err.Error() != want {
			t.Errorf("err.Error() = %q, chci %q", err.Error(), want)
		}
	})

	t.Run("both fail", func(t *testing.T) {
		err := exercise.WithCleanup(
			func() error { return errWork },
			func() error { return errClean },
		)
		if !errors.Is(err, errWork) {
			t.Errorf("chyba %v neobsahuje %v", err, errWork)
		}
		if !errors.Is(err, errClean) {
			t.Errorf("chyba %v neobsahuje %v", err, errClean)
		}
		if !strings.Contains(err.Error(), "cleanup: cleanup failed") {
			t.Errorf("chyba %q neobsahuje kontext cleanupu", err.Error())
		}
	})

	t.Run("cleanup is nil", func(t *testing.T) {
		if err := exercise.WithCleanup(func() error { return nil }, nil); err != nil {
			t.Errorf("chyba = %v, chci nil", err)
		}
		if err := exercise.WithCleanup(func() error { return errWork }, nil); !errors.Is(err, errWork) {
			t.Errorf("chyba = %v, chci %v", err, errWork)
		}
	})

	t.Run("f is nil", func(t *testing.T) {
		cleaned := 0
		err := exercise.WithCleanup(nil, func() error { cleaned++; return nil })
		if !errors.Is(err, exercise.ErrNilFunc) {
			t.Fatalf("chyba = %v, chci ErrNilFunc", err)
		}
		if cleaned != 0 {
			t.Errorf("cleanup proběhl %dx, chci 0x", cleaned)
		}
	})
}

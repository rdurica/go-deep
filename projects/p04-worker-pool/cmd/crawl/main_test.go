package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"runtime"
	"strings"
	"testing"
	"time"
)

func waitNoLeak(t *testing.T, before int) {
	t.Helper()
	for i := 0; i < 200; i++ {
		if runtime.NumGoroutine() <= before+1 {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Errorf("goroutine leak: před testem %d goroutin, po testu %d", before, runtime.NumGoroutine())
}

func testConfig() config {
	return config{workers: 4, queue: 8, rate: 100_000, burst: 1000}
}

func TestRunProcessesAllLinesInOrder(t *testing.T) {
	before := runtime.NumGoroutine()

	var in bytes.Buffer
	for i := 0; i < 30; i++ {
		fmt.Fprintf(&in, "uloha %02d\n", i)
	}
	in.WriteString("\n# komentar se preskoci\n   \n")

	var out, logOut bytes.Buffer
	if err := run(context.Background(), testConfig(), &in, &out, &logOut); err != nil {
		t.Fatalf("run = %v, chci nil", err)
	}

	lines := strings.Split(strings.TrimSpace(out.String()), "\n")
	if len(lines) != 30 {
		t.Fatalf("vypsáno %d řádků, chci 30:\n%s", len(lines), out.String())
	}
	for i, line := range lines {
		want := fmt.Sprintf("OK   uloha %02d ", i)
		if !strings.HasPrefix(line, want) {
			t.Errorf("řádek %d = %q, chci prefix %q", i, line, want)
		}
	}
	if !strings.Contains(logOut.String(), "prijato=30") {
		t.Errorf("report neobsahuje prijato=30:\n%s", logOut.String())
	}
	if !strings.Contains(logOut.String(), "zpracovano=30") {
		t.Errorf("report neobsahuje zpracovano=30:\n%s", logOut.String())
	}
	waitNoLeak(t, before)
}

func TestRunReportsFailedTasks(t *testing.T) {
	before := runtime.NumGoroutine()

	in := strings.NewReader("dobra uloha\n!spatna uloha\njeste jedna dobra\n")
	var out, logOut bytes.Buffer

	err := run(context.Background(), testConfig(), in, &out, &logOut)
	if err == nil {
		t.Fatal("run = nil, chci chybu kvůli neúspěšné úloze")
	}
	if !strings.Contains(err.Error(), "1 úloh") {
		t.Errorf("chyba = %q, chci zmínku o jedné neúspěšné úloze", err.Error())
	}
	if !strings.Contains(out.String(), "ERR  !spatna uloha") {
		t.Errorf("výstup neobsahuje chybný řádek:\n%s", out.String())
	}
	if !strings.Contains(logOut.String(), "chyb=1") {
		t.Errorf("report neobsahuje chyb=1:\n%s", logOut.String())
	}
	waitNoLeak(t, before)
}

func TestRunEmptyInput(t *testing.T) {
	before := runtime.NumGoroutine()

	var out, logOut bytes.Buffer
	if err := run(context.Background(), testConfig(), strings.NewReader("\n#\n  \n"), &out, &logOut); err != nil {
		t.Fatalf("run = %v, chci nil", err)
	}
	if out.Len() != 0 {
		t.Errorf("výstup = %q, chci prázdný", out.String())
	}
	if !strings.Contains(logOut.String(), "prázdný") {
		t.Errorf("report neoznámil prázdný vstup:\n%s", logOut.String())
	}
	waitNoLeak(t, before)
}

func TestRunRespectsCanceledContext(t *testing.T) {
	before := runtime.NumGoroutine()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	var out, logOut bytes.Buffer
	err := run(ctx, testConfig(), strings.NewReader("a\nb\nc\n"), &out, &logOut)
	if !errors.Is(err, context.Canceled) {
		t.Errorf("run se zrušeným kontextem = %v, chci Canceled", err)
	}
	waitNoLeak(t, before)
}

func TestRunRejectsBadConfig(t *testing.T) {
	var out, logOut bytes.Buffer
	cfg := testConfig()
	cfg.workers = 0

	err := run(context.Background(), cfg, strings.NewReader("a\n"), &out, &logOut)
	if err == nil {
		t.Fatal("run s nulou workerů = nil, chci chybu")
	}
}

func TestProcess(t *testing.T) {
	got, err := process(context.Background(), task{line: "jedna dve tri"})
	if err != nil {
		t.Fatalf("process = %v, chci nil", err)
	}
	if got.words != 3 {
		t.Errorf("words = %d, chci 3", got.words)
	}
	if got.sum == 0 {
		t.Error("hash = 0, chci spočítanou hodnotu")
	}
	if !strings.Contains(got.String(), "slov=3") {
		t.Errorf("String() = %q, chci zmínku o počtu slov", got.String())
	}

	if _, err := process(context.Background(), task{line: "!spatna"}); err == nil {
		t.Error("process(\"!spatna\") = nil, chci chybu")
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := process(ctx, task{line: "a"}); !errors.Is(err, context.Canceled) {
		t.Errorf("process se zrušeným kontextem = %v, chci Canceled", err)
	}
}

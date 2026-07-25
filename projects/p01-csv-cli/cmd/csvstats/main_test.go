package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const sampleCSV = "name,amount,category\nAda,120.50,food\nBob,80,transport\nGrace,200.25,food\n"

func writeTemp(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "spend.csv")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("příprava souboru selhala: %v", err)
	}
	return path
}

func TestRunStdin(t *testing.T) {
	var out, errBuf bytes.Buffer

	code := run(nil, strings.NewReader(sampleCSV), &out, &errBuf)

	if code != exitOK {
		t.Fatalf("run(...) = %d, chci %d (stderr: %s)", code, exitOK, errBuf.String())
	}
	for _, want := range []string{"KATEGORIE", "food", "transport", "CELKEM", "400.75"} {
		if !strings.Contains(out.String(), want) {
			t.Errorf("výstup neobsahuje %q:\n%s", want, out.String())
		}
	}
}

func TestRunFile(t *testing.T) {
	path := writeTemp(t, sampleCSV)
	var out, errBuf bytes.Buffer

	code := run([]string{"-file", path}, strings.NewReader(""), &out, &errBuf)

	if code != exitOK {
		t.Fatalf("run(-file) = %d, chci %d (stderr: %s)", code, exitOK, errBuf.String())
	}
	if !strings.Contains(out.String(), "food") {
		t.Errorf("výstup neobsahuje kategorii food:\n%s", out.String())
	}
}

func TestRunTop(t *testing.T) {
	var out, errBuf bytes.Buffer

	code := run([]string{"-top", "2"}, strings.NewReader(sampleCSV), &out, &errBuf)

	if code != exitOK {
		t.Fatalf("run(-top 2) = %d, chci %d (stderr: %s)", code, exitOK, errBuf.String())
	}
	got := out.String()
	_, topTable, found := strings.Cut(got, "\n\n")
	if !found || !strings.Contains(topTable, "JMÉNO") {
		t.Fatalf("chybí tabulka největších útrat oddělená prázdným řádkem:\n%s", got)
	}
	for _, want := range []string{"Grace", "Ada"} {
		if !strings.Contains(topTable, want) {
			t.Errorf("v top 2 chybí %q:\n%s", want, topTable)
		}
	}
	if strings.Contains(topTable, "Bob") {
		t.Errorf("Bob nemá být v top 2:\n%s", topTable)
	}
}

func TestRunChyby(t *testing.T) {
	tests := []struct {
		name  string
		args  []string
		stdin string
		want  int
	}{
		{"prázdný stdin", nil, "", exitFailure},
		{"rozbitý CSV", nil, "name,amount,category\nAda,neni-cislo,food\n", exitFailure},
		{"neexistující soubor", []string{"-file", "/neexistuje/nikde.csv"}, "", exitFailure},
		{"neznámý flag", []string{"-nesmysl"}, sampleCSV, exitUsage},
		{"argument navíc", []string{"navic.csv"}, sampleCSV, exitUsage},
		{"záporné top", []string{"-top", "-1"}, sampleCSV, exitUsage},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var out, errBuf bytes.Buffer
			code := run(tt.args, strings.NewReader(tt.stdin), &out, &errBuf)
			if code != tt.want {
				t.Errorf("run(%v) = %d, chci %d (stderr: %s)", tt.args, code, tt.want, errBuf.String())
			}
			if errBuf.Len() == 0 {
				t.Error("chybová cesta nic nenapsala na stderr")
			}
		})
	}
}

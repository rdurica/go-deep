// Command crawl načte seznam úloh ze standardního vstupu a zpracuje je
// s omezenou souběžností, omezenou rychlostí a s reportem na konci.
//
// Použití:
//
//	cat tasks.txt | crawl -workers=8 -queue=16 -rate=200
//
// Záměrně nic nestahuje ze sítě — „práce" je lokální výpočet nad řádkem.
// Cílem projektu je životní cyklus poolu, ne HTTP klient. Díky tomu je
// program plně testovatelný bez internetu.
//
// SIGINT nebo SIGTERM běh korektně ukončí: rozpracované úlohy dostanou
// zrušený kontext, hotové výsledky se dopíšou a program vypíše metriky.
package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"hash/fnv"
	"io"
	"os"
	"os/signal"
	"sort"
	"strings"
	"syscall"
	"time"

	"github.com/rdurica/go-deep/projects/p04-worker-pool/pool"
	"github.com/rdurica/go-deep/projects/p04-worker-pool/ratelimit"
)

type config struct {
	workers int
	queue   int
	rate    float64
	burst   int
}

// task je jeden řádek vstupu.
type task struct {
	line string
}

// report je výsledek zpracování jednoho řádku.
type report struct {
	words int
	sum   uint32
}

func (r report) String() string {
	return fmt.Sprintf("slov=%d hash=%08x", r.words, r.sum)
}

func main() {
	var cfg config
	flag.IntVar(&cfg.workers, "workers", 8, "počet souběžných workerů")
	flag.IntVar(&cfg.queue, "queue", 16, "kapacita vstupní fronty (backpressure)")
	flag.Float64Var(&cfg.rate, "rate", 200, "maximální počet úloh za sekundu")
	flag.IntVar(&cfg.burst, "burst", 32, "kolik úloh smí jít v nárazu")
	flag.Parse()

	// signal.NotifyContext zruší kontext při SIGINT/SIGTERM — to je celý
	// graceful shutdown, protože kontext teče až do jednotlivých úloh.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := run(ctx, cfg, os.Stdin, os.Stdout, os.Stderr); err != nil {
		fmt.Fprintln(os.Stderr, "crawl:", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, cfg config, in io.Reader, out, logOut io.Writer) error {
	lines, err := readLines(in)
	if err != nil {
		return fmt.Errorf("čtení vstupu: %w", err)
	}
	if len(lines) == 0 {
		fmt.Fprintln(logOut, "vstup je prázdný, není co dělat")
		return nil
	}

	limiter, err := ratelimit.New(cfg.rate, cfg.burst, nil)
	if err != nil {
		return fmt.Errorf("rate limiter: %w", err)
	}

	tasks := make([]task, len(lines))
	for i, line := range lines {
		tasks[i] = task{line: line}
	}

	results, stats, err := pool.Collect(ctx, pool.Config[task, report]{
		Workers:   cfg.workers,
		QueueSize: cfg.queue,
		Handler: func(ctx context.Context, t task) (report, error) {
			// Nejdřív rate limit, pak práce. Čekání na token je zrušitelné.
			if err := limiter.Wait(ctx); err != nil {
				return report{}, err
			}
			return process(ctx, t)
		},
	}, tasks)
	if err != nil {
		return fmt.Errorf("dávka: %w", err)
	}

	failed := writeResults(out, results)
	writeSummary(logOut, stats)

	if failed > 0 {
		return fmt.Errorf("%d úloh skončilo chybou", failed)
	}
	return nil
}

// process je „práce" nad jedním řádkem. Nedělá IO, jen počítá — a kontroluje
// kontext, aby byla zrušitelná i uprostřed dávky.
func process(ctx context.Context, t task) (report, error) {
	if err := ctx.Err(); err != nil {
		return report{}, err
	}
	if strings.HasPrefix(t.line, "!") {
		return report{}, fmt.Errorf("neplatná úloha %q", t.line)
	}

	h := fnv.New32a()
	if _, err := io.WriteString(h, t.line); err != nil {
		return report{}, fmt.Errorf("hash: %w", err)
	}
	return report{words: len(strings.Fields(t.line)), sum: h.Sum32()}, nil
}

// readLines načte neprázdné řádky bez komentářů. Vstup čteme celý dopředu,
// takže producent do poolu už jen sype hotový slice.
func readLines(in io.Reader) ([]string, error) {
	var lines []string
	sc := bufio.NewScanner(in)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		lines = append(lines, line)
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	return lines, nil
}

// writeResults vypíše výsledky v pořadí vstupu a vrátí počet chyb.
func writeResults(out io.Writer, results []pool.Result[task, report]) int {
	sorted := make([]pool.Result[task, report], len(results))
	copy(sorted, results)
	sort.SliceStable(sorted, func(i, j int) bool { return sorted[i].Index < sorted[j].Index })

	var failed int
	for _, res := range sorted {
		if res.Err != nil {
			failed++
			fmt.Fprintf(out, "ERR  %s: %v\n", res.Input.line, res.Err)
			continue
		}
		fmt.Fprintf(out, "OK   %s %s\n", res.Input.line, res.Value)
	}
	return failed
}

func writeSummary(logOut io.Writer, stats pool.Stats) {
	fmt.Fprintf(logOut,
		"hotovo: prijato=%d zpracovano=%d chyb=%d max_soubeznost=%d doba=%s\n",
		stats.Submitted, stats.Processed, stats.Failed, stats.MaxInFlight,
		stats.Elapsed.Round(time.Millisecond))
}

# Kanonické zdroje

Čti v tomto pořadí, kopíruje to fáze kurzu. Odkazy v jednotlivých lekcích jsou
konkrétnější — tohle je mapa.

## Povinné (oficiální)

1. [How to Write Go Code](https://go.dev/doc/code) — lekce 01
2. [A Tour of Go](https://go.dev/tour/) — lekce 03–15, průběžně
3. [Effective Go](https://go.dev/doc/effective_go) — lekce 19–20
4. [Go Code Review Comments](https://go.dev/wiki/CodeReviewComments) — lekce 21 a dál
   jako denní reference
5. [Google Go Style Guide](https://google.github.io/styleguide/go/) — průběžně
6. [Language Specification](https://go.dev/ref/spec) — výběrově
7. [The Go Memory Model](https://go.dev/ref/mem) — lekce 48
8. [Go Modules Reference](https://go.dev/ref/mod) — lekce 51

## Blog a přednášky (vysoká hodnota)

- [Go Slices: usage and internals](https://go.dev/blog/slices-intro) — lekce 07
- [Arrays, slices: the mechanics of append](https://go.dev/blog/slices) — lekce 07
- [Strings, bytes, runes and characters](https://go.dev/blog/strings) — lekce 09
- [Package names](https://go.dev/blog/package-names) — lekce 19
- [Error handling and Go](https://go.dev/blog/error-handling-and-go) — lekce 14
- [Working with Errors in Go 1.13](https://go.dev/blog/go1.13-errors) — lekce 14
- [Go Concurrency Patterns](https://go.dev/talks/2012/concurrency.slide) — lekce 41–42
- [Pipelines and cancellation](https://go.dev/blog/pipelines) — lekce 45
- [Go Concurrency Patterns: Context](https://go.dev/blog/context) — lekce 27
- [Share Memory By Communicating](https://go.dev/blog/codelab-share) — lekce 43
- [Introducing the Go Race Detector](https://go.dev/blog/race-detector) — lekce 44
- [Profiling Go Programs](https://go.dev/blog/pprof) — lekce 53
- [Fuzzing is Beta Ready](https://go.dev/blog/fuzz-beta) — lekce 52
- [Structured Logging with slog](https://go.dev/blog/slog) — lekce 29
- [Routing Enhancements for Go 1.22](https://go.dev/blog/routing-enhancements) — lekce 25

## Style guides

- [Uber Go Style Guide](https://github.com/uber-go/guide) — lekce 19–21
- [Go Code Review Comments](https://go.dev/wiki/CodeReviewComments) — reference při review

## Knihy (doporučené, text nekopírujeme)

- *100 Go Mistakes and How to Avoid Them*, Teiva Harsanyi — mapa pastí
- *Let's Go* a *Let's Go Further*, Alex Edwards — praktické HTTP bez frameworku
- *The Go Programming Language*, Donovan & Kernighan — hloubka jazyka
- *Concurrency in Go*, Katherine Cox-Buday — k modulu M5

## Stdlib jako učebnice

Čti zdrojáky, ne jen dokumentaci. `go doc -src` nebo skok do definice v IDE.

| Balíček | Co se z něj naučíš | Lekce |
|---------|--------------------|-------|
| `io` | minimální interfacy a kompozice | 13 |
| `errors` | jak funguje wrapping a Is/As | 14 |
| `net/http` | Handler, HandlerFunc, ServeMux | 22, 24 |
| `encoding/json` | dekodér a cena reflexe | 16, 54 |
| `context` | rušení a deadliny | 27 |
| `sync` | zámky, Once, WaitGroup | 43 |
| `log/slog` | handler jako rozšiřitelný bod | 29 |

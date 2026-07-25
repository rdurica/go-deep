# P01 — csvstats: CLI nad CSV

Projekt patří k [lekci 17](../../lessons/lesson-17/README.md) a uzavírá fázi 1.
Cílem je napsat **skutečný příkazový nástroj**: knihovna dělá práci, `main` řeší
jen argumenty, vstup, výstup a exit kódy.

## Zadání

Nástroj `csvstats` přečte CSV se sloupci `name,amount,category`, spočítá souhrn po
kategoriích a vypíše ho jako tabulku. Volitelně vypíše N největších útrat.

```bash
go run ./cmd/csvstats -file csvstats/testdata/sample.csv
cat csvstats/testdata/sample.csv | go run ./cmd/csvstats -top 3
```

Výstup:

```
KATEGORIE  POČET  SOUČET  PRŮMĚR
food       3      365.85  121.95
fun        2      120.00  60.00
transport  2      95.75   47.88
CELKEM     7      581.60
```

## Struktura

```
projects/p01-csv-cli/
  ACCEPTANCE.md
  csvstats/               # package csvstats — parsování, agregace, formátování
    csvstats.go
    format.go
    csvstats_test.go
    format_test.go
    testdata/
      sample.csv
      summary.golden
      top3.golden
  cmd/csvstats/
    main.go               # package main — flagy, stdin/soubor, exit kódy
    main_test.go
```

## Akceptační kritéria

- [ ] Balíček `csvstats` je importovatelný jako
      `github.com/rdurica/go-deep/projects/p01-csv-cli/csvstats` a **nevypisuje**
      nic na `os.Stdout` ani nečte `os.Args`.
- [ ] `ParseRecords(io.Reader) ([]Record, error)` vyžaduje hlavičku `name,amount,category`
      (case-insensitive), validuje jméno, kategorii i částku a chyba obsahuje **číslo řádku**.
      Prázdný vstup vrací `ErrEmptyInput`.
- [ ] `LoadFile(path)` obaluje chybu otevření souboru přes `%w`, takže funguje
      `errors.Is(err, os.ErrNotExist)`.
- [ ] `Summarize([]Record) Summary` vrací počet záznamů, celkový součet a statistiku
      po kategoriích (počet, součet, průměr) seřazenou sestupně podle součtu, při shodě
      abecedně.
- [ ] `TopN([]Record, n) []Record` řadí stabilně sestupně podle částky, nemění vstupní
      slice, pro `n <= 0` vrací prázdný výsledek a pro `n > len` všechny záznamy.
- [ ] `RenderSummary(io.Writer, Summary) error` a `RenderTop(io.Writer, []Record) error`
      píšou do `io.Writer` (ne na stdout) zarovnanou tabulku přes `text/tabwriter`.
- [ ] `cmd/csvstats` podporuje `-file` (prázdné = stdin) a `-top N`, chybové hlášky píše
      na stderr a vrací exit kódy: `0` úspěch, `1` chyba běhu (nečitelný nebo neplatný
      vstup), `2` chybné použití (neznámý flag, argument navíc, záporné `-top`).
- [ ] Jádro příkazu je funkce `run(args []string, stdin io.Reader, stdout, stderr io.Writer) int`,
      takže jde testovat bez spouštění procesu.
- [ ] Existuje **golden test**: výstup `RenderSummary` se porovnává s
      `csvstats/testdata/summary.golden`.
- [ ] `go test ./...` v adresáři projektu prochází a `go vet ./...` mlčí.

## Jak ověřit

```bash
cd projects/p01-csv-cli
go test ./...
go vet ./...
go run ./cmd/csvstats -file csvstats/testdata/sample.csv -top 3
```

Golden soubory se přegenerují (a je potřeba je ručně zkontrolovat v diffu):

```bash
go test ./csvstats -update
```

## Rozšíření pro odvážné

- Přidej `-category food` a filtruj záznamy před agregací.
- Přidej `-format json` a vypiš souhrn jako JSON (lekce 16).
- Přidej `-sort name|total` a nech uživatele zvolit řazení tabulky.

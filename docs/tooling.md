# Tooling

Kurz běží na Go 1.26+ a standardní knihovně. Níže je, co potřebuješ lokálně
a co je volitelné.

## Povinné

| Nástroj | Účel |
|---------|------|
| Go 1.26+ | `go test`, `go vet`, `gofmt` |
| `make` | cíle v kořenovém `Makefile` |
| Python 3 | `scripts/generate_index.py` (regenerace `BOOK.md` / `PROGRESS.md`) |

Ověření instalace:

```bash
go version
make help
python3 --version
```

## Doporučené

| Nástroj | Účel | Jak |
|---------|------|-----|
| [golangci-lint](https://golangci-lint.run/welcome/install/) | širší lint než `go vet` | `make lint` — bez instalace Makefile jen vypíše varování |
| editor s gopls | doplňování, jump to def | VS Code / GoLand / Neovim |

Konfigurace linteru: [`.golangci.yml`](../.golangci.yml). Lekce 21 a 23 na něj odkazují
v review cvičeních; CI ho nepouští — stačí `gofmt` + `go vet` + testy.

## Běžné příkazy

```bash
make lesson L=07          # celé cvičení
make lesson L=07 PART=1   # jen stupeň 1 (tiers.txt)
make race L=44            # s -race (volitelně PART=)
make solutions            # všechna referenční řešení
make project P=02         # testy projektu
make check                # to samé co CI (bez race na solutions)
make mirror               # zkopíruje testy z exercise/ do solutions/
make fmt                  # gofmt -w .
make lint                 # golangci-lint, pokud je nainstalovaný
```

`make mirror` / `scripts/mirror_tests.sh` přepíše `exercise_test` → `solutions_test`
(a import cesty) nebo whitebox `package exercise` → `package solutions`. Testy musí
izolovat stuby — viz [authoring.md](authoring.md#izolace-stubů).

## CI

Workflow [`.github/workflows/ci.yml`](../.github/workflows/ci.yml) na `main` a PR:
formátování, `go vet`, solutions (+ race), projekty (+ race), stuby cvičení musí padat,
index (`BOOK.md` / `PROGRESS.md`) musí sedět s generátorem.

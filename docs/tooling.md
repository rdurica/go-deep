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
make lesson L=07     # testy cvičení jedné lekce
make race L=44       # totéž s -race
make solutions       # všechna referenční řešení
make project P=02    # testy projektu
make check           # to samé co CI (bez race na solutions)
make mirror          # zkopíruje testy z exercise/ do solutions/
make fmt             # gofmt -w .
make lint            # golangci-lint, pokud je nainstalovaný
```

## CI

Workflow [`.github/workflows/ci.yml`](../.github/workflows/ci.yml) na `main` a PR:
formátování, `go vet`, solutions (+ race), projekty (+ race), stuby cvičení musí padat,
index (`BOOK.md` / `PROGRESS.md`) musí sedět s generátorem.

# Poznámky pro lektora

## Typické PHP → Go pasty

1. Panic místo `error` (reflex `throw`).
2. Pointer všude, protože „v PHP je objekt reference".
3. Obří `UserServiceInterface` 1:1 se Symfony service.
4. Balíčky `controller` / `service` / `repository` jako dogma.
5. Hledání DI kontejneru první týden.
6. Zápis do `nil` mapy, protože zero value vypadá použitelně.
7. Očekávání, že `append` mutuje původní slice.
8. Context uložený ve struct fieldu.
9. Ignorované chyby `_ = err` z AI snippetů.
10. `len()` na stringu s diakritikou a překvapení, že to nejsou znaky.

## Kde se studenti nejčastěji zaseknou

| Lekce | Zádrhel | Co pomůže |
|-------|---------|-----------|
| 07 | aliasing slice a chování `append` | nakreslit backing array na tabuli |
| 12 | nil interface vs typovaný nil pointer | pustit to v debuggeru |
| 14 | kdy `errors.Is` a kdy `errors.As` | tabulka sentinel vs typ |
| 26 | pořadí middlewaru | očíslovat logy před a po `next.ServeHTTP` |
| 33 | interface u konzumenta, ne u implementace | ukázat na jejich Symfony kódu |
| 40–45 | vše kolem souběžnosti | společné pouštění s `-race` |
| 48 | happens-before jako uspořádání, ne čas | příklad s přeuspořádáním zápisů |

## AI abuse patterns (sleduj u studentů)

- Copilot dopisuje celá cvičení v modulu M1.
- „Napiš řešení lekce N" místo vlastního pokusu.
- Accept-all na diff bez projetí checklistu.
- Generování concurrency kódu bez spuštění `-race`.
- Sáhnutí po Ginu nebo Echu dřív, než umí `net/http`.
- Přijetí odpovědi, která volá neexistující funkci stdlib.

## Jak vést živou lekci

1. Student ukáže padající test, pak vlastní pokus, teprve potom diskuse.
2. V modulu M5 pouštějte `-race` společně a čtěte výstup nahlas.
3. V modulu M7 vyžadujte vyplněný checklist z `docs/ai-playbook.md`.
4. Diff lab (lekce 57): dvě řešení vedle sebe, hlasování „co je Go-ish a proč".
5. Checkpointy nechte studenty vyplnit samostatně a teprve pak proberte rubriku.

## Časové bufferování

- Modul M1 trvá seniorům déle (ego a odvykání návyků). Neuspěchejte ho.
- Lekce 07 (slices) a 14 (errors) klidně roztáhněte na dvě sezení.
- Modul M5 je pro PHP vývojáře nejcizejší — počítejte s dvojnásobkem otázek.
- P03 a P05 jsou místa, kde studenti over-engineerují. Vracejte je k akceptačním kritériím.

## Kontrola kvality odevzdání

```bash
make check          # gofmt, vet, všechna referenční řešení, projekty
make race L=44      # u concurrency lekcí vždy
make project P=03
```

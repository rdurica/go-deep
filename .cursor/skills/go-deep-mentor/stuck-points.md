# Zádrhele a PHP → Go pasty

Zkráceno z `course/facilitator-notes.md`. Použij při mentoringu — pojmenuj reflex, pak nápověda.

## Typické PHP → Go pasty

1. Panic místo `error` (reflex `throw`)
2. Pointer všude, protože „v PHP je objekt reference“
3. Obří `UserServiceInterface` 1:1 se Symfony service
4. Balíčky `controller` / `service` / `repository` jako dogma
5. Hledání DI kontejneru první týden
6. Zápis do `nil` mapy (zero value vypadá použitelně)
7. Očekávání, že `append` mutuje původní slice
8. Context uložený ve struct fieldu
9. Ignorované chyby `_ = err` z AI snippetů
10. `len()` na stringu s diakritikou ≠ počet znaků

## Kde se studenti zaseknou

| Lekce | Zádrhel | Co pomůže (bez spoileru kódu) |
|-------|---------|-------------------------------|
| 07 | Aliasing slice, chování `append` | Nakresli / popiš backing array, len/cap, kdy vzniká nové pole |
| 12 | Nil interface vs typovaný nil pointer | Nech porovnat `(T)(nil)` vs `nil` jako interface value |
| 14 | Kdy `errors.Is` vs `errors.As` | Tabulka: sentinel hodnota vs typ chyby |
| 26 | Pořadí middlewaru | Očísluj logy před a po `next.ServeHTTP` |
| 33 | Interface u konzumenta, ne u implementace | Porovnej se Symfony service interface v jejich kódu |
| 40–45 | Panic kolem souběžnosti | Společně `make race L=NN`, čti výstup |
| 48 | Happens-before jako uspořádání, ne čas | Příklad s přeuspořádáním zápisů |

## Tempo a projekty

- M1 (01–18): senioři potřebují víc času — neodvykej návykům spěchem
- Lekce 07 a 14: klidně dvě sezení
- M5 (concurrency): dvojnásobek otázek; vždy `-race`
- P03 / P05: over-engineering → vrať k `ACCEPTANCE.md`

# Lekce 21 — Error handling v review

> **Čas:** ~85 min · **Fáze:** 2 — Idiomatický Go · **AI režim:** `JEN VYSVĚTLENÍ`

## Co budeš umět

- Napsat text chyby, který se dá bez úprav vložit do větší chyby.
- Postavit řetěz kontextu při probublávání a rozhodnout, kdy wrapovat a kdy chybu nahradit.
- Vysvětlit, proč `_ = err` a `defer f.Close()` bez kontroly nejsou totéž.
- Zkombinovat chybu z hlavní práce a z úklidu přes named return, `defer` a `errors.Join`.
- Poznat v code review špatný error na první pohled a říct proč.

## Teorie

### Text chyby je API

Chybové hlášky vidí uživatel, hledá je v logu a grepuje je v Sentry. Jsou součástí
kontraktu, i když je nekontroluje kompilátor. Konvence Go (a stdlib je dodržuje do puntíku):

- **malé počáteční písmeno** — `"invalid port"`, ne `"Invalid port"`;
- **žádná tečka ani vykřičník na konci**;
- **žádný `\n`**;
- **žádné „failed to" ani „error"** na začátku každé úrovně — že jde o chybu, už čtenář ví;
- **kontext, ne opakování** — každá úroveň přidá informaci, kterou nižší vrstva neměla.

```go
// ŠPATNĚ — každá úroveň opakuje totéž a přidává balast
return fmt.Errorf("Error: failed to process the user request: %w", err)
// -> Error: failed to process the user request: Error: failed to read: failed to open file.

// SPRÁVNĚ — každá úroveň přidá právě jeden nový fakt
return fmt.Errorf("process user %d: %w", id, err)
// -> process user 42: read profile: open /data/42.json: no such file or directory
```

Dobrý test: **vezmi výslednou zprávu a přečti ji nahlas.** Pokud dává smysl jako cesta
`co → kde → proč`, je to v pořádku. Pokud tam je třikrát „failed", není.

### Kdy wrapovat, kdy nahradit

`%w` zachová původní chybu pro `errors.Is` a `errors.As`. To je silné, ale zároveň to
znamená, že **typ chyby se stává součástí tvého veřejného API**. Volající se na něj může
navázat.

Wrapuj (`%w`), když:

- volající může na konkrétní chybu smysluplně reagovat (`os.ErrNotExist`, `sql.ErrNoRows`);
- jsi uvnitř jednoho modulu a chceš zachovat diagnostiku.

Nahraď (`%v` nebo vlastní sentinel), když:

- chyba pochází z implementačního detailu, který nechceš vystavit (dnes SQL, zítra HTTP);
- chceš klientovi nabídnout stabilní sentinel: `if errors.Is(err, sql.ErrNoRows) { return ErrNotFound }`.

Sentinely deklaruj jako balíčkové proměnné s prefixem `Err`:

```go
var ErrNotFound = errors.New("not found")
```

### Nikdy neignoruj `err`

`_ = err` je v review červená vlajka. Existují tři legitimní reakce na chybu: vrátit ji
(s kontextem), zpracovat ji (fallback, retry, log), nebo — vzácně — vědomě ji zahodit
s komentářem, **proč**.

Speciální případ, který AI kazí nejčastěji, je `defer f.Close()`:

```go
// U čtení je ignorovaný Close v pořádku — nic se neztratí.
f, err := os.Open(path)
if err != nil {
	return err
}
defer f.Close() //nolint: chyba při zavírání čtení nemá následek

// U ZÁPISU je ignorovaný Close ztráta dat: Close flushuje buffer.
func writeAll(path string, data []byte) (err error) {
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create %q: %w", path, err)
	}
	defer func() {
		if cerr := f.Close(); cerr != nil && err == nil {
			err = fmt.Errorf("close %q: %w", path, cerr)
		}
	}()

	if _, err := f.Write(data); err != nil {
		return fmt.Errorf("write %q: %w", path, err)
	}
	return nil
}
```

Klíčem je **named return** (`(err error)`). Bez něj `defer` nemá do čeho zapsat a chyba
z `Close` zmizí. Tenhle vzor si zapamatuj, protože se ti bude vracet u transakcí,
souborů i síťových spojení.

### `errors.Join` pro nezávislé chyby

Když má selhat víc nezávislých operací a ty chceš ohlásit všechny, ne jen první, použij
`errors.Join` (Go 1.20+):

```go
var errs []error
for i, c := range closers {
	if err := c.Close(); err != nil {
		errs = append(errs, fmt.Errorf("close %d: %w", i, err))
	}
}
return errors.Join(errs...) // nil, když je slice prázdný nebo obsahuje jen nil
```

`errors.Is` prochází i spojené chyby, takže volající pořád najde konkrétní příčinu.
`Join` používej na **paralelní** chyby (zavři všechno, zvaliduj všechna pole).
Na sekvenční řetěz „selhalo A, protože selhalo B" je pořád správný `%w`.

### Chyba vs panika

Panika je pro **chybu programátora**, ne pro chybu vstupu. Nedostupná databáze, neplatný
JSON od klienta ani chybějící soubor nejsou důvod k panice — to jsou očekávané stavy,
které patří do návratové hodnoty. Panikuj jen tam, kde další běh nemá smysl: porušený
invariant, nemožná větev `switch`, `Must` konstruktor nad konstantou v kódu.

A ještě jedno pravidlo pro dobu, kdy budeš psát souběžný kód: **chyba z goroutiny se
nesmí ztratit.** Buď má goroutina kanál, kudy ji pošle, nebo běží pod `errgroup`.
`go doWork()` bez odchytu chyby je leak informace. K tomu se vrátíme ve fázi 5.

### Jak vypadá review

| Kód v PR | Verdikt |
|---|---|
| `if err != nil { return err }` v knihovní funkci | často OK, ale zvaž kontext |
| `if err != nil { return fmt.Errorf("failed to do thing: %v", err) }` | chybí `%w`, přebývá „failed to" |
| `_ = json.Unmarshal(b, &v)` | blok |
| `log.Println(err); return nil` | blok — chybu jsi „vyřešil" logem |
| `panic(err)` v HTTP handleru | blok |
| `defer tx.Rollback()` bez kontroly | OK, rollback po commitu vrací chybu záměrně |

Nástroj, který tohle hlídá strojově, je `errcheck` (součást `golangci-lint`). Než ho
pustíš, nauč se to vidět sám — jinak jen umlčíš linter komentářem.

## Rozdíly proti PHP

V Symfony je výjimka objekt s vlastní třídou a zprávou pro člověka. Kontext přidáváš tím,
že výjimku zabalíš do jiné třídy:

```php
try {
    $this->loadConfig($path);
} catch (IOException $e) {
    throw new ConfigurationException(
        sprintf('Failed to load configuration from "%s".', $path), 0, $e
    );
}
```

V Go je chyba **hodnota** a text chyby je stavební kámen, ne celá věta. Skládá se zleva
doprava jako cesta:

```go
cfg, err := loadConfig(path)
if err != nil {
	return fmt.Errorf("load config %q: %w", path, err)
}
// výsledek na nejvyšší úrovni:
// serve: load config "/etc/app.yaml": open /etc/app.yaml: no such file or directory
```

Návyk k opuštění: **přestaň psát celé věty s velkým písmenem a tečkou.** Tvoje chyba
skoro nikdy není konec řetězu — někdo ji obalí a tvoje tečka skončí uprostřed.

## Časté chyby

| Chyba | Proč vzniká | Jak to udělat správně |
|-------|-------------|------------------------|
| `"Failed to open file."` | výjimka v PHP je věta pro člověka | `"open %q: %w"` — malé písmeno, bez tečky |
| `%v` místo `%w` | vypadá to stejně | `%w`, když má volající šanci reagovat |
| `_ = err` | „tady to nemůže selhat" | vrať, zpracuj, nebo zdůvodni komentářem |
| `defer f.Close()` u zápisu | čtení a zápis vypadají stejně | named return + `defer` s kontrolou |
| `log.Error(err); return nil` | zvyk na globální error handler | logování není zpracování chyby |
| Panika při chybě vstupu | reflex `throw` | `error` v návratové hodnotě |
| Vlastní typ chyby pro každý případ | „výjimka je třída" | sentinel `var ErrX = errors.New(...)` |

## AI kvíz

Po přečtení teorie spusť v Cursoru **`/go-deep-quiz 21`**. AI tě ~5 minut prověří mentální model (ne hotové cvičení). Slabiny si uloží do [`GAPS.md`](../../GAPS.md).

## Úkol

Pracuj v `exercise/`. Po doplnění spouštěj testy:

Stupně jdou od jednodušších ke složitějším — po každém stupni spusť review, než jdeš dál.

### Jednoduchý

Funkce: `ReadConfig`

```bash
make lesson L=21 PART=1
```

Pak **`/go-deep-review 21 easy`**.

### Střední

Funkce: `NewPipeline`, `Run`

```bash
make lesson L=21 PART=2
```

Pak **`/go-deep-review 21 medium`**.

### Obtížný

Funkce: `CloseAll`, `WithCleanup`

```bash
make lesson L=21 PART=3
```

Pak **`/go-deep-review 21 hard`**.

Až budou stupně hotové, porovnej se `solutions/` (spoiler).

## Závěrečné otázky

Spusť **`/go-deep-review 21 final`**. AI projde body níže, doptá se a ověří pochopení. Celé cvičení ověří `make lesson L=21` (+ `make race L=21`, pokud to lekce vyžaduje).

- [ ] V žádné své chybě nemáš velké počáteční písmeno, tečku ani „failed to"
- [ ] Nikde v tvém řešení není `_ = err`
- [ ] Umíš vysvětlit, proč `WithCleanup` potřebuje named return
- [ ] Umíš vysvětlit rozdíl mezi `%w` a `%v` z pohledu zpětné kompatibility
- [ ] Umíš říct, kdy `defer f.Close()` bez kontroly chyby je a kdy není v pořádku
- [ ] Umíš vysvětlit, kdy použít `errors.Join` a kdy `%w`

## AI režim

`JEN VYSVĚTLENÍ` — viz [docs/ai-playbook.md](../../docs/ai-playbook.md).

Legitimní prompt: *„Projdi tenhle diff a vypiš každé místo, kde je chyba ignorovaná nebo
obalená bez užitečného kontextu. Body, ne kód."* Přepis napiš sám — error handling je
přesně to, co AI generuje syntakticky správně a sémanticky špatně.

## Další čtení

1. [Go blog — Working with Errors in Go 1.13](https://go.dev/blog/go1.13-errors)
2. [Go Code Review Comments — Error Strings](https://go.dev/wiki/CodeReviewComments#error-strings)
3. [pkg.go.dev — errors](https://pkg.go.dev/errors) — přečti si zdrojáky `Join` a `Is`
4. [Dave Cheney — Don't just check errors, handle them gracefully](https://dave.cheney.net/2016/04/27/dont-just-check-errors-handle-them-gracefully)

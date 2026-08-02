# Lekce 17 — Testování a projekt P01 (CSV CLI)

> **Čas:** ~90 min · **Fáze:** 1 — Jazyk a paměťový model · **AI režim:** `ZAKÁZÁNO`

## Co budeš umět

- Napsat table-driven test s podtesty a poznat, kdy patří `t.Fatal` a kdy `t.Error`.
- Rozhodnout mezi interním (`package foo`) a externím (`package foo_test`) testovacím balíčkem.
- Použít `t.Helper`, `t.Cleanup`, `t.TempDir` a adresář `testdata/` místo ručního úklidu.
- Otestovat kód pracující se soubory bez toho, abys po sobě nechal smetí v repozitáři.
- Vysvětlit, proč se v Go místo mocků píšou fake implementace, a napsat jednu.

## Teorie

### Konvence, které toolchain vynucuje

- Soubor musí končit na `_test.go`. Do buildu se nedostane, `go build` ho ignoruje.
- Funkce se jmenuje `TestXxx(t *testing.T)`, `BenchmarkXxx(b *testing.B)`,
  `FuzzXxx(f *testing.F)`, `ExampleXxx()`.
- Test běží s pracovním adresářem nastaveným na adresář balíčku, takže relativní cesta
  `testdata/sample.csv` funguje vždy stejně.
- Adresář `testdata/` je pro nástroje Go posvátný: ignoruje ho `go build` i `go vet`.

Nejpoužívanější přepínače:

| Příkaz | K čemu |
|--------|--------|
| `go test ./...` | celý strom balíčků |
| `go test -run TestMedian/sudý` | jen jeden podtest (regulární výraz) |
| `go test -v` | jméno a výsledek každého testu |
| `go test -count=1` | vypne cache výsledků, „opravdu to spusť" |
| `go test -cover` | procento pokrytí |
| `go test -race` | detektor souběhů (uvidíš ve fázi 5) |

`-count=1` používej v CI i v Makefile. Go si totiž úspěšné testy kešuje a při nezměněném
kódu je nepustí znovu — což je matoucí ve chvíli, kdy testu záleží na okolí.

### `package foo` vs `package foo_test`

V jednom adresáři smí výjimečně žít dva balíčky: `foo` a `foo_test`.

```go
package csvstats      // interní test: vidí i neexportované funkce
package csvstats_test // externí test: vidí jen veřejné API
```

Externí test je **výchozí volba**. Testuješ balíček tak, jak ho uvidí jeho uživatel,
takže tě test začne bolet ve chvíli, kdy je API nepohodlné — a to je zpětná vazba, o kterou
stojíš. Interní test si nech na případy, kdy je nutné ověřit vnitřní pomocnou funkci
s netriviální logikou.

V tomhle kurzu jsou testy vždycky externí, navíc s pojmenovaným importem, aby šel stejný
soubor použít nad `exercise/` i `solutions/`:

```go
package exercise_test

import exercise "github.com/rdurica/go-deep/lessons/lesson-17/exercise"
```

### `t.Error` vs `t.Fatal`, `t.Helper`, `t.Cleanup`

`t.Error` selhání zaznamená a **pokračuje**. `t.Fatal` test ukončí. Pravidlo: `Fatal`
tehdy, kdy by další kód stejně spadl (chyba, kterou jsi nečekal, nil hodnota, krátký
slice), `Error` pro každé jednotlivé porovnání, protože chceš vidět všechna selhání
naráz, ne je odkrývat po jednom.

```go
got, err := ParseRecords(r)
if err != nil {
	t.Fatalf("ParseRecords(...) = _, %v, chci nil", err) // dál nemá smysl pokračovat
}
if len(got) != 2 {
	t.Fatalf("dostal jsem %d záznamů, chci 2", len(got))  // jinak by index spadl
}
if got[0].Name != "Ada" {
	t.Errorf("got[0].Name = %q, chci %q", got[0].Name, "Ada") // jen jedno tvrzení
}
```

Pozor: `t.Fatal` smí volat **jen goroutina testu**. Z jiné goroutiny použij `t.Error`
a vrať se.

Vlastní asserty piš jako obyčejné funkce s `t.Helper()` na prvním řádku. Díky tomu se
selhání nahlásí na řádku volajícího, ne uvnitř pomocníka:

```go
func equalFloat(t *testing.T, got, want float64) {
	t.Helper()
	if math.Abs(got-want) > 1e-9 {
		t.Errorf("dostal jsem %v, chci %v", got, want)
	}
}
```

`t.Cleanup(fn)` je Go verze `tearDown`, jen lokální — registruje se tam, kde vzniká
zdroj, a běží i při `t.Fatal`. Je to stejný nápad jako `defer`, ale funguje i uvnitř
pomocných funkcí, které skončí dřív než test.

### `t.TempDir` a `testdata/`

Pro práci se soubory máš dvě možnosti a obě jsou lepší než „vytvořím soubor vedle testu":

- **`testdata/`** — vstupy, které se nemění a chceš je vidět v gitu (ukázkové CSV, golden
  soubory, fuzz korpus).
- **`t.TempDir()`** — soubory, které si test vyrábí sám. Vrací unikátní adresář a smaže
  ho po skončení testu včetně obsahu. Nemusíš volat `os.RemoveAll`, nemusíš řešit kolize
  mezi paralelními testy.

```go
dir := t.TempDir()
path := filepath.Join(dir, "spend.csv")
os.WriteFile(path, []byte("name,amount,category\nZoe,42.5,fun\n"), 0o600)

recs, err := LoadFile(path)
```

### Podtesty a `t.Parallel`

`t.Run(jméno, func(t *testing.T){...})` dává každému případu vlastní jméno
(`TestMedian/sudý_počet`), vlastní výsledek a možnost pustit ho samostatně přes `-run`.
Uvnitř podtestu můžeš zavolat `t.Parallel()`; podtest se pozastaví, dokud nedoběhne tělo
rodiče, a pak běží souběžně s ostatními paralelními sourozenci.

Paralelizuj jen tam, kde to má smysl (pomalé I/O), a nikdy nesdílej mezi podtesty
zapisovatelný stav. V Go 1.22 už se `tt` v `for` cyklu chová bezpečně — proměnná cyklu je
nová v každé iteraci, takže klasická past `tt := tt` odpadá.

### Fake místo mocku

V PHP je mock reflex: `$this->createMock(RepositoryInterface::class)` s očekáváními na
volání. V Go se mockovací frameworky používají minimálně a je to důsledek návrhu jazyka:
rozhraní se implementuje implicitně a bývají malá, takže **napsat fake je otázka deseti
řádků**.

```go
type Store interface {
	Save(rec Record) error
}

type fakeStore struct {
	saved []Record
	err   error
}

func (f *fakeStore) Save(rec Record) error {
	if f.err != nil {
		return f.err
	}
	f.saved = append(f.saved, rec)
	return nil
}
```

Rozdíl proti mocku je v tom, co ověřuješ. Mock kontroluje **interakce** („Save byl zavolán
dvakrát s těmito argumenty"), fake ti dovolí ověřit **výsledek** („po zpracování jsou
uložené tyhle dva záznamy"). Druhé je odolnější vůči refaktoringu: přepíšeš vnitřek
a test drží.

## Rozdíly proti PHP

PHPUnit je framework: dědíš z `TestCase`, máš `setUp`, `tearDown`, data providery,
`assertSame` a k tomu Mockery nebo Prophecy.

```php
final class MedianTest extends TestCase
{
    #[DataProvider('cases')]
    public function testMedian(array $in, float $expected): void
    {
        self::assertSame($expected, median($in));
    }
}
```

V Go je testování součást toolchainu, ne knihovna. Žádná bázová třída, žádné anotace,
žádné asserty:

```go
func TestMedian(t *testing.T) {
	tests := []struct {
		name string
		in   []float64
		want float64
	}{
		{"lichý počet", []float64{3, 1, 2}, 2},
		{"sudý počet", []float64{4, 1, 3, 2}, 2.5},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got, _ := Median(tt.in); got != tt.want {
				t.Errorf("Median(%v) = %v, chci %v", tt.in, got, tt.want)
			}
		})
	}
}
```

Co se mění v uvažování: **data provider je obyčejný slice a assert je obyčejný `if`.**
Chybí ti `assertEquals`? Ve skutečnosti ti chybí jen jeho hláška — a tu si píšeš sám,
takže je konkrétnější než cokoli, co by vygeneroval framework. Standardní tvar hlášky je
`funkce(vstup) = dostal, chci chtěl`; drž se ho a nikdy nebudeš u červeného testu tápat.

## Časté chyby

| Chyba | Proč vzniká | Jak to udělat správně |
|-------|-------------|------------------------|
| Assertion knihovna hned v prvním testu | zvyk na `assertSame` | nejdřív si zvykni na `if` a formát `got/want` |
| `t.Error` tam, kde další řádek spadne na nil | v PHPUnit assert test ukončí | `t.Fatal`, pokud nemá smysl pokračovat |
| Test si vytváří soubory v repozitáři | není `setUp`/`tearDown` po ruce | `t.TempDir()` a `testdata/` |
| Interní testovací balíček ze zvyku | v PHPUnit vidíš `private` přes reflexi | `package foo_test` a testuj veřejné API |
| Mock frameworky pro každé rozhraní | reflex z Mockery/Prophecy | malé rozhraní + ruční fake |
| Zapomenuté `-count=1` v CI | keš testů vypadá jako „už to prošlo" | `go test -count=1 ./...` |

## AI kvíz

Po přečtení teorie spusť v Cursoru **`/go-deep-quiz 17`**. AI tě ~5 minut prověří mentální model (ne hotové cvičení). Slabiny si uloží do [`GAPS.md`](../../GAPS.md).

## Úkol

Pracuj v `exercise/`. Po doplnění spouštěj testy:

Stupně jdou od jednodušších ke složitějším — po každém stupni spusť review, než jdeš dál.

### Jednoduchý

Funkce: `Median`

```bash
make lesson L=17 PART=1
```

Pak **`/go-deep-review 17 easy`**.

### Střední

Funkce: `ParseRecords`, `SumByCategory`

```bash
make lesson L=17 PART=2
```

Pak **`/go-deep-review 17 medium`**.

### Obtížný

Funkce: `TopN`, `LoadFile`

```bash
make lesson L=17 PART=3
```

Pak **`/go-deep-review 17 hard`**.

Až budou stupně hotové, porovnej se `solutions/` (spoiler).

## Projekt P01 — csvstats

Zadání a akceptační kritéria jsou v
[projects/p01-csv-cli/ACCEPTANCE.md](../../projects/p01-csv-cli/ACCEPTANCE.md).
Z funkcí, které jsi napsal, uděláš skutečný CLI nástroj: balíček `csvstats` počítá a
formátuje, `cmd/csvstats/main.go` řeší flagy, stdin vs. soubor a exit kódy. Součástí je
i **golden test** — výstup tabulky se porovnává se souborem v `testdata/`.

```bash
cd projects/p01-csv-cli && go test ./...
```

## Závěrečné otázky

Spusť **`/go-deep-review 17 final`**. AI projde body níže, doptá se a ověří pochopení. Celé cvičení ověří `make lesson L=17` (+ `make race L=17`, pokud to lekce vyžaduje).

- [ ] Umíš vysvětlit rozdíl mezi `package foo` a `package foo_test` a kdy který zvolit
- [ ] Umíš vysvětlit, proč `t.Helper()` mění místo, kde se hlásí chyba
- [ ] Umíš vysvětlit, kdy patří `t.Fatal` a kdy `t.Error`
- [ ] Umíš vysvětlit, co dělá `-count=1` a proč ho chceš v CI
- [ ] Umíš vysvětlit, proč se v Go místo mocku píše fake

## AI režim

`ZAKÁZÁNO` — viz [docs/ai-playbook.md](../../docs/ai-playbook.md).

Mentor, kvíz i review (dialog) jsou vždy OK; v tomto režimu AI nesmí psát kód cvičení.
## Další čtení

1. [pkg.go.dev — testing](https://pkg.go.dev/testing)
2. [Go Wiki — TableDrivenTests](https://go.dev/wiki/TableDrivenTests)
3. [Go blog — Using Subtests and Sub-benchmarks](https://go.dev/blog/subtests)
4. [go.dev — go test command](https://pkg.go.dev/cmd/go#hdr-Testing_flags)

# Lekce 15 — Generics

> **Čas:** ~35 min · **Fáze:** 1 — Jazyk a paměťový model · **AI režim:** `ZAKÁZÁNO`

## Co budeš umět

- Napsat generickou funkci i generický typ a nechat kompilátor odvodit type parametry.
- Vysvětlit, co znamená vlnovka v constraintu, a kdy bez ní kód neprojde.
- Rozhodnout, kdy generika **ne**použít, a obhájit to pravidlem tří výskytů.
- Nahradit vlastní pomocníky funkcemi z balíčků `slices`, `maps` a `cmp`.

## Teorie

### Jak se dřív žilo bez generik

Do Go 1.18 existovaly dvě možnosti a obě byly nepříjemné. První bylo `interface{}`
s type assertion — funkční, ale typová kontrola se odsunula na runtime:

```go
func Sum(values []interface{}) interface{} {
	var total int
	for _, v := range values {
		n, ok := v.(int) // a co když tam je float64?
		if !ok {
			panic("nečekaný typ")
		}
		total += n
	}
	return total
}
```

Druhá byla copy-paste: `SumInt`, `SumInt64`, `SumFloat64`. Přesně to dodnes dělá
`sort.Ints`, `sort.Strings`, `sort.Float64s`. Generika obojí ruší.

### Type parametry a instanciace

Type parametry jsou v hranatých závorkách mezi jménem a seznamem argumentů. Každý má
jméno a **constraint** — interface, který říká, co smí typ být:

```go
func Map[T, U any](s []T, f func(T) U) []U { … }
```

`T` a `U` jsou dva nezávislé parametry, oba s constraintem `any`. Zápis `[T, U any]`
je zkratka za `[T any, U any]`, stejně jako u parametrů funkce.

Volat se dá explicitně, nebo s odvozením:

```go
Map[int, string]([]int{1, 2}, strconv.Itoa) // explicitní instanciace
Map([]int{1, 2}, strconv.Itoa)              // type inference — obvyklý zápis
```

Kompilátor `T` odvodí z prvního argumentu a `U` z návratového typu funkce. Explicitně to
píšeš jen tam, kde inference nestačí — typicky když se type parametr v argumentech
vůbec neobjeví (`New[int]()`).

### Constraints: `any`, `comparable`, vlastní

Constraint je interface v rozšířeném významu. Kromě metod smí obsahovat i **množinu typů**:

```go
type Number interface {
	~int | ~int64 | ~float64
}
```

Tři vestavěné, které potkáš nejčastěji:

| Constraint | Co dovoluje | Kde je |
|------------|-------------|--------|
| `any` | cokoli, ale nejde s tím nic dělat | vestavěný (alias `interface{}`) |
| `comparable` | `==` a `!=`, tedy i klíč mapy | vestavěný |
| `cmp.Ordered` | `<`, `<=`, `>`, `>=` | balíček `cmp` (Go 1.21+) |

`comparable` je přesně to, co potřebuješ pro `map[K]V`, proto má `Keys[K comparable, V any]`
takový podpis. Pozor, `comparable` **nesplňují** slice, mapa ani funkce — u nich `==`
neexistuje.

Constraint s metodami i s typy najednou je taky legální:

```go
type StringerNumber interface {
	~int | ~float64
	String() string
}
```

Constraint smí být použit **jen** jako constraint. `var x Number` nezkompiluješ, pokud
interface obsahuje množinu typů.

### Vlnovka `~` — nejdůležitější detail lekce

`int` v constraintu znamená „přesně typ `int`". `~int` znamená „libovolný typ, jehož
**podkladovým typem** je `int`". Rozdíl je vidět okamžitě, jakmile do hry vstoupí
pojmenovaný typ z lekce 3:

```go
type Celsius float64
type UserID int

type Strict interface{ int | float64 }
type Loose interface{ ~int | ~float64 }

func SumStrict[T Strict](s []T) T { … }
func SumLoose[T Loose](s []T) T   { … }

SumStrict([]Celsius{1, 2}) // chyba: Celsius does not satisfy Strict
SumLoose([]Celsius{1, 2})  // OK
```

Protože doménové typy typu `Celsius`, `UserID` nebo `Level` chceš v Go zavádět často
(je to nejlevnější typová bezpečnost, jakou máš), je **vlnovka v číselném constraintu
skoro vždy správně**. Constraint bez vlnovky napíšeš, jen když ti na přesné identitě
typu záleží.

Návratový typ si přitom pojmenovaný typ zachová: `SumLoose([]Celsius{...})` vrátí
`Celsius`, ne `float64`.

### Generické typy a metody

Generický může být i typ. Type parametry pak patří k typu a metody je jen přebírají:

```go
type Stack[T any] struct {
	items []T
}

func (s *Stack[T]) Push(v T) { s.items = append(s.items, v) }

func (s *Stack[T]) Pop() (T, bool) {
	if len(s.items) == 0 {
		var zero T // jediný způsob, jak získat zero value neznámého typu
		return zero, false
	}
	last := len(s.items) - 1
	v := s.items[last]
	s.items = s.items[:last]
	return v, true
}
```

Dvě věci k zapamatování. Za prvé, `var zero T` je idiom pro zero value type parametru —
nemůžeš napsat `return nil` ani `return 0`, protože nevíš, co `T` je.

Za druhé — a je to častý zdroj zklamání — **metoda nemůže mít vlastní type parametry**:

```go
// nezkompiluje se: method must have no type parameters
func (s *Stack[T]) Map[U any](f func(T) U) *Stack[U] { … }
```

Řešením je volná funkce: `func MapStack[T, U any](s *Stack[T], f func(T) U) *Stack[U]`.
Právě proto je `slices.Map` funkce a ne metoda. Omezení je záměrné — bez něj by nešlo
implementovat method sety pro interfacy.

Instanciovaný generický typ je normální typ: `Stack[int]` a `Stack[string]` jsou dva
různé typy a nedají se navzájem přiřadit.

### `slices`, `maps` a `cmp` ze stdlib

Od Go 1.21 je většina generických pomocníků přímo ve standardní knihovně. Než si napíšeš
vlastní, podívej se tam:

```go
slices.Contains(s, v)         // je prvek v slice?
slices.Index(s, v)            // pozice, nebo -1
slices.Sort(s)                // řazení pro cmp.Ordered
slices.SortFunc(s, cmpFunc)   // řazení s vlastním porovnáním
slices.Equal(a, b)            // porovnání po prvcích
slices.Clone(s)               // mělká kopie
slices.Delete(s, i, j)        // odstranění rozsahu
slices.Max(s), slices.Min(s)  // extrémy (panikují na prázdném vstupu)

maps.Keys(m), maps.Values(m)  // iterátory přes klíče/hodnoty (od Go 1.23)

cmp.Compare(a, b)             // -1, 0, 1
cmp.Or(a, b, c)               // první nenulová hodnota
```

Všimni si, že `slices.Max` na prázdném slice **panikuje**. Tvoje `Max` vrací
`(T, bool)` — to je bezpečnější varianta, kterou v produkčním kódu často chceš.

### Kdy generika nepoužívat

Generika svádějí. Pravidlo, které funguje:

> Napiš to nejdřív konkrétně. Zobecni až při **třetím** výskytu.

Dvě kopie kódu jsou levnější než špatná abstrakce. Konkrétní důvody, kdy generika nechat být:

- **Stačí interface.** Potřebuješ-li jen chování, ne konkrétní typ, je `io.Writer`
  jednodušší než `[T Writer]`. Generika jsou pro případy, kdy chceš zachovat *identitu*
  typu (vstup `[]Celsius` → výstup `Celsius`).
- **Jediný type parametr, jediné použití.** Pak jsi jen přidal hranaté závorky.
- **Čitelnost.** `func Process[T any, K comparable, V Number](m map[K]V, f func(T) V) []T`
  nikdo nepřečte. Rozděl to.
- **Výkon není důvod.** Go generika se kompilují po tzv. GC shapes, takže občas přidají
  nepřímé volání. Nezrychlují.

Pozitivní znamení naopak jsou: kontejnery (`Stack`, `Cache`, `Set`), operace nad slice
a mapami, funkce, které zachovávají typ vstupu na výstupu.

## Rozdíly proti PHP

PHP typový systém generika nemá. Řeší se to docblockem pro statickou analýzu a doufáním:

```php
/**
 * @template T
 * @param list<T> $items
 * @param callable(T): bool $keep
 * @return list<T>
 */
function filter(array $items, callable $keep): array
{
    return array_values(array_filter($items, $keep));
}
```

PHPStan to pochopí, ale runtime ne — v `$items` může být cokoli a chyba spadne až
za běhu. Go od verze 1.18 má generika **v jazyce**, takže je kontroluje kompilátor:

```go
func Filter[T any](s []T, keep func(T) bool) []T {
	out := make([]T, 0, len(s))
	for _, v := range s {
		if keep(v) {
			out = append(out, v)
		}
	}
	return out
}

nums := Filter([]int{1, 2, 3}, func(n int) bool { return n%2 == 1 }) // []int
```

Co se mění v uvažování: v PHP je generický typ **anotace pro nástroj**, v Go je to
**součást typu**. Zároveň platí varování v opačném směru: Go generika jsou schválně
skromná. Nejsou tu od toho, abys s nimi postavil framework — jsou tu, abys nemusel psát
`MapStringToInt`, `MapStringToString` a `MapIntToString` třikrát.

## Časté chyby

| Chyba | Proč vzniká | Jak to udělat správně |
|-------|-------------|------------------------|
| Constraint bez `~` | vlnovka vypadá jako ozdoba | `~int \| ~float64`, jinak neprojdou pojmenované typy |
| `return nil` / `return 0` u type parametru | zvyk na konkrétní typ | `var zero T; return zero` |
| Type parametr na metodě | reflex z jiných jazyků | volná funkce s parametry navíc |
| Generika hned napoprvé | snaha o znovupoužitelnost dopředu | konkrétně, zobecni při třetím výskytu |
| `any` místo `comparable` u klíče mapy | `any` se zdá univerzálnější | `comparable` — jinak mapa nezkompiluje |
| Vlastní `Contains`, `Sort`, `Clone` | neznalost `slices` a `maps` | koukni nejdřív do stdlib |
| `var x Number` | constraint vypadá jako interface | constraint s množinou typů jde použít jen jako constraint |
| Generický „BaseRepository" | přenos vrstveného návrhu z PHP | konkrétní typ za rozumnou cenu duplikace |

## AI kvíz

Po přečtení teorie spusť v Cursoru **`/go-deep-quiz 15`**. AI tě ~5 minut prověří mentální model (ne hotové cvičení). Slabiny si uloží do [`GAPS.md`](../../GAPS.md).

## Úkol

Pracuj v `exercise/`. Kontrakt je v komentáři nad funkcí. Stupně jdou od jednodušších
ke složitějším — po každém stupni spusť review, než jdeš dál.

### Jednoduchý

Oprav: `Max` (kód je záměrně vadný — panika na prázdném vstupu)

```bash
make lesson L=15 PART=1
```

Pak **`/go-deep-review 15 easy`**.

### Střední

Implementuj: `Filter`, `Sum`

```bash
make lesson L=15 PART=2
```

Pak **`/go-deep-review 15 medium`**.

### Obtížný

Doplň: `Push`, `Pop` (generický typ — zero value při odebrání)

```bash
make lesson L=15 PART=3
```

Pak **`/go-deep-review 15 hard`**.

Až budou stupně hotové, porovnej se `solutions/` (spoiler).

## Závěrečné otázky

Spusť **`/go-deep-review 15 final`**. AI projde body níže, doptá se a ověří pochopení. Celé cvičení ověří `make lesson L=15` (+ `make race L=15`, pokud to lekce vyžaduje).

- [ ] Umíš vysvětlit rozdíl mezi `int` a `~int` v constraintu
- [ ] Umíš vysvětlit, proč metoda nemůže mít vlastní type parametry
- [ ] Umíš získat zero value type parametru
- [ ] Umíš uvést dva případy, kdy je interface lepší volba než generika
- [ ] Umíš z hlavy jmenovat pět funkcí z balíčku `slices`

## AI režim

`ZAKÁZÁNO` — viz [docs/ai-playbook.md](../../docs/ai-playbook.md).

Mentor, kvíz i review (dialog) jsou vždy OK; v tomto režimu AI nesmí psát kód cvičení.
## Další čtení

1. [Go blog — An Introduction To Generics](https://go.dev/blog/intro-generics)
2. [Go blog — When To Use Generics](https://go.dev/blog/when-generics)
3. [pkg.go.dev — slices](https://pkg.go.dev/slices)
4. [Type Parameters Proposal](https://go.googlesource.com/proposal/+/refs/heads/master/design/43651-type-parameters.md)

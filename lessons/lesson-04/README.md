# Lekce 04 — Funkce, multiple returns, closures

> **Čas:** ~90 min · **Fáze:** 1 — Jazyk a paměťový model · **AI režim:** `ZAKÁZÁNO`

## Co budeš umět

- Navrhnout signaturu funkce s víc návratovými hodnotami a rozhodnout, kdy použít
  `(T, error)`, kdy `(T, bool)` a kdy pojmenované návratové hodnoty.
- Vysvětlit, co closure zachytává a jak dlouho to žije.
- Popsat, jak se od Go 1.22 chová proměnná cyklu a proč byla dřív klasickou pastí.
- Napsat funkci, která vrací funkci, a použít funkci jako parametr bez interface.

## PHP → Go most

V PHP se z funkce vrací jedna hodnota. Když jich potřebuješ víc, vrací se pole nebo DTO,
a když něco selže, letí výjimka.

```php
function minMax(array $nums): array
{
    if ($nums === []) {
        throw new InvalidArgumentException('empty');
    }
    return ['min' => min($nums), 'max' => max($nums)];
}

['min' => $min, 'max' => $max] = minMax($nums);
```

V Go se prostě vrátí víc hodnot a chybějící výsledek se signalizuje `bool`:

```go
func MinMax(nums []int) (min, max int, ok bool) {
	if len(nums) == 0 {
		return 0, 0, false
	}
	// ...
}

min, max, ok := MinMax(nums)
if !ok {
	// prázdný vstup
}
```

Co se mění v uvažování: přestaneš vymýšlet obalové struktury pro „dvě věci najednou".
Zároveň platí, že návratových hodnot mají být nejvýš tři a poslední je `error` nebo `bool` —
`(int, string, bool, error, time.Time)` je signál, že funkce dělá moc věcí.

## Teorie

### Multiple return values

Návratových hodnot může být libovolně mnoho a nemají žádný obal — nejsou to tuple ani pole,
kompilátor je předává přímo.

```go
func divmod(a, b int) (int, int) {
	return a / b, a % b
}

q, r := divmod(17, 5) // 3, 2
q, _ = divmod(9, 2)   // druhou hodnotu zahodíš podtržítkem
```

Podtržítko `_` je *blank identifier*. Není to proměnná — nejde ho číst a nezabírá paměť.
Používá se přesně tam, kde tě druhá hodnota nezajímá.

Konvence pro pořadí je pevná a stdlib ji dodržuje beze zbytku:

| Vzor | Kdy |
|------|-----|
| `(T, error)` | operace může selhat z důvodu, který stojí za popsání |
| `(T, bool)` | hodnota buď je, nebo není; není co vysvětlovat |
| `(T1, T2)` | dvě rovnocenné části jednoho výsledku (`divmod`) |

Comma-ok idiom (`v, ok := ...`) je v jazyce zabudovaný na třech místech: čtení z mapy,
příjem z kanálu a type assertion. Když ho použiješ i ve vlastních funkcích, chová se tvoje
API stejně jako jazyk.

### Pojmenované návratové hodnoty

Návratové hodnoty můžeš pojmenovat. Tím vzniknou proměnné inicializované na zero value a
`return` bez argumentů („naked return") vrátí jejich aktuální obsah.

```go
func MinMax(nums []int) (min, max int, ok bool) {
	if len(nums) == 0 {
		return // vrátí 0, 0, false
	}
	min, max, ok = nums[0], nums[0], true
	// ...
	return
}
```

Používej to **střídmě**. Naked return v delší funkci se čte hrozně: čtenář musí odrolovat na
signaturu, aby zjistil, co se vlastně vrací. Dvě situace, kde se pojmenované návratové
hodnoty vyplatí:

1. **Dokumentace signatury.** `func Split(path string) (dir, file string)` řekne mnohem víc
   než `(string, string)` — a tady vůbec nemusíš naked return použít.
2. **Změna výsledku v `defer`** — typicky obalení chyby. Bez pojmenovaného výsledku to nejde,
   protože `defer` běží až po vyhodnocení `return`.

Praktické pravidlo: pojmenuj je pro čitelnost, ale v těle piš explicitní
`return min, max, true`. Nejlepší z obou světů.

### Variadické funkce

Poslední parametr může mít `...`, čímž funkce přijme libovolný počet argumentů. Uvnitř je
z toho obyčejný slice.

```go
func Sum(nums ...int) int {
	total := 0
	for _, n := range nums { // nums je []int
		total += n
	}
	return total
}

Sum()           // 0, nums je nil slice
Sum(1, 2, 3)    // 6
values := []int{1, 2, 3}
Sum(values...)  // rozbalení existujícího slice
```

Pozor na dvě věci. Zaprvé, `Sum(values...)` **nekopíruje** — funkce dostane slice ukazující
na stejné pole, takže když ho uvnitř změní, projeví se to venku. Zadruhé, `Sum()` dostane
`nil` slice, ne prázdný; pro `len`, `range` a `append` je to jedno, takže se tím většinou
nemusíš zabývat.

### Funkce jsou hodnoty

Funkce je v Go plnohodnotný typ. Můžeš ji přiřadit do proměnné, poslat jako parametr, vrátit
z jiné funkce nebo uložit do slice.

```go
type transform func(int) int // pojmenovaný typ pro čitelnost, není povinný

func Apply(nums []int, f func(int) int) []int {
	out := make([]int, len(nums))
	for i, n := range nums {
		out[i] = f(n)
	}
	return out
}

doubled := Apply([]int{1, 2, 3}, func(x int) int { return x * 2 })
```

Tohle je v Go náhrada za drobné strategy interfacy, na které jsi zvyklý ze Symfony. Pokud má
„strategie" jednu metodu, nedělej z ní interface — udělej z ní typ funkce. Ostatně
`http.HandlerFunc` ve standardní knihovně dělá přesně to.

Zero value funkčního typu je `nil` a volání `nil` funkce panikuje. Když je funkce v parametru
volitelná, musíš to ošetřit: `if f == nil { f = identity }`.

### Closures a co zachytávají

Anonymní funkce vidí na proměnné z okolního rozsahu a **zachytává je odkazem**, ne hodnotou.
Zachycená proměnná žije tak dlouho, jak dlouho žije closure — kompilátor ji pro tenhle případ
přesune z zásobníku na haldu.

```go
func Counter() func() int {
	n := 0 // přežije návrat z Counter
	return func() int {
		n++
		return n
	}
}

next := Counter()
next() // 1
next() // 2

other := Counter()
other() // 1 — vlastní n, na sobě nezávislé
```

Tohle je nejlevnější způsob, jak v Go získat „objekt se stavem" tam, kde by struct s jedním
polem byl zbytečný.

Protože se zachytává odkazem, sdílí několik closure nad stejnou proměnnou **jeden** stav:

```go
func Pair() (func(), func() int) {
	n := 0
	return func() { n++ }, func() int { return n }
}
```

Přesně o tohle se opírá úkol C. A pozor — sdílený stav mezi closures není chráněný proti
souběhu. Jakmile budeš closure volat z víc goroutin, potřebuješ mutex (fáze 5).

### Proměnná cyklu: past, která od Go 1.22 zmizela

Historicky měl `for` cyklus **jednu** proměnnou pro všechny iterace. Closure vytvořené v těle
proto po skončení cyklu všechny viděly stejnou, poslední hodnotu:

```go
// Chování do Go 1.21 včetně:
var fs []func() int
for _, v := range []int{1, 2, 3} {
	fs = append(fs, func() int { return v })
}
// fs[0]() == fs[1]() == fs[2]() == 3
```

Standardní obcházení bylo `v := v` na začátku těla — tenhle řádek uvidíš ve spoustě staršího
kódu a teď už víš, k čemu byl.

**Od Go 1.22 (a jen pokud má modul v `go.mod` napsáno `go 1.22` nebo vyšší — kurz má `go 1.26`) má každá iterace
vlastní proměnnou.** Předchozí příklad vrátí `1`, `2`, `3`. Sémantika se řídí verzí modulu,
ne verzí kompilátoru, takže starý modul přeložený novým Go se pořád chová postaru.

Dva důsledky, na které si dej pozor:

- Nekopíruj bezmyšlenkovitě staré Stack Overflow odpovědi — `v := v` je dnes zbytečné
  (`go vet` na to v novějších verzích upozorní jako na nadbytečné).
- Když čteš cizí kód, podívej se do `go.mod`, než začneš usuzovat, co closure zachytí.

Změna se týká proměnných deklarovaných v hlavičce cyklu. Proměnná deklarovaná **před** cyklem
je pořád jedna sdílená:

```go
sum := 0
for _, v := range nums {
	sum += v // jeden sum, žádná změna chování
}
```

### `defer` a rekurze v jedné větě

`defer` odloží volání funkce na konec té funkce, ve které stojí — hodí se na úklid
(`defer f.Close()`). Argumenty se vyhodnotí hned, tělo až při návratu. Celá lekce 10 je o něm
i o `panic`/`recover`.

Rekurze funguje, jak čekáš, včetně anonymních funkcí (které si ale musíš nejdřív deklarovat
přes `var f func(int) int`, aby na sebe mohly odkazovat). Go nemá garantovanou optimalizaci
koncové rekurze, takže hluboká rekurze roste na zásobníku — ten sice roste dynamicky, ale má
strop. U průchodu stromem je rekurze v pořádku, u iterace přes milion prvků použij cyklus.

## Časté chyby

| Chyba | Proč vzniká | Jak to udělat správně |
|-------|-------------|------------------------|
| Vracení structu pro dvě hodnoty | v PHP jiná možnost není | vrať dvě hodnoty, obal až od tří souvisejících |
| Naked return v dlouhé funkci | vypadá to úsporně | pojmenuj výsledky, ale `return` piš explicitně |
| `v := v` v každém cyklu | zvyk z Go do 1.21 | od Go 1.22 zbytečné, iterace má vlastní proměnnou |
| Interface s jednou metodou jako strategie | reflex ze Symfony | typ funkce, viz `http.HandlerFunc` |
| Volání `nil` funkce z parametru | zero value funkce je `nil` | ošetři `if f == nil` nebo dokumentuj povinnost |
| Sdílený stav closure z víc goroutin | closure vypadá nevinně | zachycená proměnná je sdílená paměť, chraň ji |

## Úkol

Pracuj v `exercise/`. Postupuj A → B → C, po každé části spusť test.

### A — rozcvička (~10 min)

1. `Sum(nums ...int) int` — variadický součet. Bez argumentů vrací `0`.
2. `MinMax(nums []int) (min, max int, ok bool)` — nejmenší a největší prvek. Pro prázdný
   nebo `nil` vstup vrať `0, 0, false`. Signatura má pojmenované návratové hodnoty; v těle
   piš explicitní `return`. Funkce nesmí měnit vstupní slice (tedy žádné třídění na místě).

### B — jádro (~35 min)

1. `Counter() func() int` — vrátí funkci, která při každém volání vrátí o jedna víc. První
   volání vrací `1`. Dva čítače získané dvěma voláními `Counter()` na sobě musí být nezávislé.
2. `Apply(nums []int, f func(int) int) []int` — vrátí **nový** slice se stejnou délkou, kde
   je na každý prvek použita `f`. Vstup se nemění a vrácený slice nesmí sdílet podkladové
   pole se vstupem. Pro `nil` a prázdný vstup vrať slice délky `0`.
3. `Compose(fs ...func(int) int) func(int) int` — složí funkce **zleva doprava**, tedy
   `Compose(f, g)(x)` je `g(f(x))`. Bez argumentů vrať identitu (`func(x int) int { return x }`).

### C — rozšíření (~25 min)

`Memoize(f func(int) int) (func(int) int, func() int)` vrátí dvojici:

- memoizovanou variantu `f` — pro stejný argument vrátí uloženou hodnotu a `f` znovu nevolá,
- funkci, která vrátí počet **skutečných** volání `f`.

Obě vrácené funkce se dělí o stejný stav (cache a čítač), a přesně to je pointa úkolu:
closure nad sdílenou proměnnou. Cache si udělej mapou `map[int]int` — mapy jsou detailně až
v lekci 08, pro teď stačí `make(map[int]int)`, zápis `m[k] = v` a comma-ok čtení
`v, ok := m[k]`.

Každé volání `Memoize` musí vytvořit nezávislou instanci s vlastní cache.

```bash
make lesson L=04
```

Až budeš hotový, porovnej se `solutions/` (spoiler).

## Ověření

- [ ] `make lesson L=04` prochází
- [ ] Umíš vysvětlit, kdy vrátit `(T, error)` a kdy `(T, bool)`
- [ ] Umíš vysvětlit, proč je naked return v dlouhé funkci problém
- [ ] Umíš popsat, jak se chová proměnná cyklu v Go 1.22 a jak se chovala dřív
- [ ] Umíš vysvětlit, proč dva čítače z `Counter()` mají oddělený stav
- [ ] Umíš najít ve stdlib aspoň jeden typ funkce použitý místo interface

## AI režim

`ZAKÁZÁNO` — viz [docs/ai-playbook.md](../../docs/ai-playbook.md).

## Další čtení

1. [Tour of Go — Function values a Closures](https://go.dev/tour/moretypes/24)
2. [Go 1.22 Release Notes — for loop variable scoping](https://go.dev/doc/go1.22#language)
3. [Go blog — Fixing for loops in Go 1.22](https://go.dev/blog/loopvar-preview)
4. [Effective Go — Multiple return values, Named result parameters](https://go.dev/doc/effective_go#multiple-returns)

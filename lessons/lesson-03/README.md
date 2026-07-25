# Lekce 03 — Typy, zero values a konstanty

> **Čas:** ~90 min · **Fáze:** 1 — Jazyk a paměťový model · **AI režim:** `ZAKÁZÁNO`

## Co budeš umět

- Vyjmenovat zero value každého základního i složeného typu a odvodit ji u vlastního typu.
- Rozhodnout mezi `var`, `:=` a `const` a vědět, proč to není jen otázka vkusu.
- Vysvětlit, proč Go nemá implicitní konverze a co to znamená pro `int` × `float64`.
- Vytvořit vlastní pojmenovaný typ, dát mu konstanty přes `iota` a metodu `String()`.

## PHP → Go most

PHP proměnná, která „není nastavená", je `null`. Skoro každá knihovna proto řeší
`?? ''`, `isset()`, `?->` a nullable typy.

```php
$name = null;
echo strlen($name ?? '');   // musíš ošetřit null
$count = "5" + 3;            // 8 — PHP typ převede za tebe
```

Go tenhle problém z velké části odstraňuje: **každá proměnná má vždy platnou hodnotu**.
Není nutné ji inicializovat, protože zero value je definovaná a použitelná.

```go
var name string   // "" — ne null, rovnou použitelný string
var count int     // 0
fmt.Println(len(name)) // 0, žádná kontrola nutná

count := 3
sum := "5" + count // chyba kompilace — Go nekonvertuje nic samo
```

Přenos návyku: přestaň psát obranné kontroly na „nenastaveno". Pokud potřebuješ rozlišit
„nula" od „nevyplněno", musíš to modelovat explicitně (pointer nebo dvojice hodnota+bool).
Právě proto, že to Go nutí přiznat, je to čitelnější než všudypřítomný `null`.

## Teorie

### Zero values

Deklarace bez inicializace nastaví typově specifickou nulu:

| Typ | Zero value | Použitelné bez inicializace? |
|-----|------------|------------------------------|
| `int`, `int64`, `uint`, … | `0` | ano |
| `float64` | `0` | ano |
| `bool` | `false` | ano |
| `string` | `""` | ano |
| `[]T` (slice) | `nil` | ano pro čtení, `len`, `append`, `range` |
| `map[K]V` | `nil` | jen pro čtení a `len` — **zápis panikuje** |
| `chan T` | `nil` | ne, blokuje navždy |
| `*T` (pointer) | `nil` | dereference panikuje |
| `interface` | `nil` | volání metody panikuje |
| `func` | `nil` | volání panikuje |
| `struct` | struct se všemi poli na zero value | ano |
| `[N]T` (array) | array N zero values | ano |

Dvě z nich stojí za zapamatování hned, protože jsou nejčastějším zdrojem paniky u
začátečníka a zároveň nejužitečnější vlastností jazyka:

```go
var s []string
s = append(s, "a")     // funguje, nil slice je platný začátek
fmt.Println(len(s))    // 1

var m map[string]int
fmt.Println(m["x"])    // 0 — čtení z nil mapy je OK
fmt.Println(len(m))    // 0
m["x"] = 1             // panic: assignment to entry in nil map
```

Nil mapu musíš vytvořit: `m := make(map[string]int)` nebo `m := map[string]int{}`.

### Užitečná zero value je designový cíl

Tohle je jeden z mála skutečných „návrhových vzorů" v Go standardní knihovně: typ se
navrhuje tak, aby jeho zero value byla rovnou funkční.

```go
var buf bytes.Buffer   // připravený k použití, žádný NewBuffer()
buf.WriteString("ahoj")

var mu sync.Mutex      // odemčený mutex, připravený
mu.Lock()

var wg sync.WaitGroup  // připravená
```

Když navrhuješ vlastní typ, zeptej se: *dává jeho zero value smysl?* Pokud ano, ušetříš
uživatelům konstruktor. V PHP je tenhle luxus nemyslitelný, protože objekt bez `new`
neexistuje.

### `var` vs `:=` vs `const`

```go
var timeout time.Duration          // chci zero value a explicitní typ
var name = "radek"                 // typ odvozen, ale mimo funkci nutné var
name := "radek"                    // jen uvnitř funkce, nejběžnější
const MaxRetries = 3               // vyhodnoceno při kompilaci, nelze měnit
```

Praktické pravidlo:

- Uvnitř funkce používej `:=`, dokud nechceš zero value nebo jiný typ, než by se odvodil.
- `var x T` použij, když je zero value přesně to, co chceš (`var sb strings.Builder`).
- Mimo funkci `:=` nefunguje vůbec — na úrovni balíčku musí být `var` nebo `const`.
- Nikdy nepiš `var x int = 0`. Buď `var x int`, nebo `x := 0`.

Konstanty v Go jsou zvláštní v tom, že mohou být **netypované**. Netypovaná konstanta se
přizpůsobí kontextu:

```go
const factor = 3          // netypovaná
var i int = factor        // OK
var f float64 = factor    // taky OK
const typed int = 3
var g float64 = typed     // chyba: cannot use typed (int) as float64
```

Proto se konstanty většinou píšou bez typu — jsou pak flexibilnější.

### Žádné implicitní konverze

Go nepřevede `int` na `float64` za tebe, ani `int` na `int64`. Nikdy.

```go
var cents int = 1999
var price float64 = cents          // chyba kompilace
var price float64 = float64(cents) // OK
price = float64(cents) / 100       // 19.99
```

Klasická past — celočíselné dělení proběhne dřív, než ho převedeš:

```go
fmt.Println(float64(1999 / 100)) // 19  (dělí se jako int, pak konvertuje)
fmt.Println(float64(1999) / 100) // 19.99
```

Druhá past je konverze do menšího typu, která tiše ořízne:

```go
var big int = 300
var small int8 = int8(big) // 44, žádné varování
```

Kompilátor tohle nezachytí, protože jsi konverzi napsal explicitně. Když na velikosti
záleží, musíš rozsah zkontrolovat sám.

### Vlastní typy a `iota`

Pojmenovaný typ nad základním typem není alias — je to nový typ s vlastní identitou.
To je nejlevnější způsob, jak v Go získat typovou bezpečnost tam, kde bys v PHP použil
enum nebo konstanty na třídě.

```go
type Level int

const (
	LevelDebug Level = iota // 0
	LevelInfo               // 1
	LevelWarn               // 2
	LevelError              // 3
)

func (l Level) String() string {
	switch l {
	case LevelDebug:
		return "DEBUG"
	case LevelInfo:
		return "INFO"
	case LevelWarn:
		return "WARN"
	case LevelError:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}
```

`iota` je čítač, který se v každém `const` bloku resetuje na 0 a s každým řádkem se
zvýší. Řádky bez výrazu opakují ten předchozí, proto stačí `iota` napsat jednou.

Metoda `String() string` je součástí interface `fmt.Stringer`. Jakmile ji typ má,
`fmt.Println(LevelWarn)` vypíše `WARN` místo `2`. Tuhle metodu chceš skoro u každého
enum-like typu.

Pozor na jednu nepříjemnost: zero value takového typu je `0`, tedy `LevelDebug`. Když
chceš rozlišit „nenastaveno", nech nulu prázdnou:

```go
const (
	LevelUnknown Level = iota // 0 = nenastaveno
	LevelDebug
	LevelInfo
)
```

## Časté chyby

| Chyba | Proč vzniká | Jak to udělat správně |
|-------|-------------|------------------------|
| Zápis do `nil` mapy | zero value mapy vypadá použitelně | `make(map[K]V)` před prvním zápisem |
| `var x int = 0` | zvyk explicitně inicializovat | `var x int` |
| Celočíselné dělení místo desetinného | konverze až po dělení | konvertuj operand, ne výsledek |
| Nula jako platná hodnota enumu | `iota` začíná na 0 | první konstanta = `Unknown` |
| Pointer jen kvůli „nenastaveno" | reflex z nullable PHP | zvaž `(value, bool)` nebo prázdnou hodnotu |
| Konstanta s typem bez důvodu | zvyk na strict types | nech ji netypovanou |

## Úkol

Pracuj v `exercise/`. Postupuj A → B → C, po každé části spusť test.

### A — rozcvička (~10 min)

Implementuj `Classify(n int) string`, která vrací `"negative"`, `"zero"` nebo `"positive"`.
Použij `switch` bez výrazu (`switch { case n < 0: ... }`), ne řetězec `if/else`.

### B — jádro (~35 min)

1. `ZeroValueOf(kind string) string` — pro názvy typů `"int"`, `"float64"`, `"string"`,
   `"bool"`, `"slice"`, `"map"`, `"pointer"`, `"chan"`, `"interface"` vrať textovou podobu
   zero value: `"0"`, `"0"`, `""` (prázdný řetězec), `"false"`, `"nil"`, `"nil"`, `"nil"`,
   `"nil"`, `"nil"`. Cokoli jiného → `"unknown"`.
2. `CentsToPrice(cents int) float64` — převede celé centy na desetinnou cenu
   (1999 → 19.99). Pozor na pořadí konverze a dělení.
3. `ToInt8(n int) (int8, bool)` — vrátí hodnotu jako `int8` a `true`, pokud se do rozsahu
   `int8` vejde; jinak `0, false`. Rozsah si odvoď, nebo použij `math.MinInt8` a
   `math.MaxInt8`.

### C — rozšíření (~25 min)

Vytvoř typ `Level int` s konstantami `LevelUnknown`, `LevelDebug`, `LevelInfo`,
`LevelWarn`, `LevelError` v tomto pořadí přes `iota` a implementuj:

- `func (l Level) String() string` → `"UNKNOWN"`, `"DEBUG"`, `"INFO"`, `"WARN"`, `"ERROR"`;
  hodnota mimo rozsah také `"UNKNOWN"`.
- `func ParseLevel(s string) Level` — case-insensitive převod ze jména na `Level`,
  neznámý vstup dá `LevelUnknown`.
- `func (l Level) Enabled(min Level) bool` — vrací `true`, pokud je úroveň alespoň `min`.

Až budeš hotový, ověř si v `go doc fmt.Stringer`, co jsi právě implementoval.

```bash
make lesson L=03
```

Až budeš hotový, porovnej se `solutions/` (spoiler).

## Ověření

Po dokončení úkolů spusť v Cursoru **`/go-deep-review`** a zadej třeba jen `03`. AI tě postupně projde body níže, doptá se a ověří pochopení — nestačí jen zelené testy.

- [ ] `make lesson L=03` prochází
- [ ] Umíš zpaměti zero value pro slice, mapu, pointer a struct
- [ ] Umíš vysvětlit, proč `float64(1999/100)` není `19.99`
- [ ] Umíš vysvětlit rozdíl mezi typovanou a netypovanou konstantou
- [ ] Umíš vysvětlit, proč je dobré, aby `iota` enum začínal hodnotou „unknown"
- [ ] Umíš uvést dva typy ze stdlib, jejichž zero value je rovnou použitelná

## AI režim

`ZAKÁZÁNO` — viz [docs/ai-playbook.md](../../docs/ai-playbook.md).

## Další čtení

1. [Tour of Go — Basic types, Zero values](https://go.dev/tour/basics/11)
2. [Effective Go — Constants](https://go.dev/doc/effective_go#constants)
3. [Go blog — Constants](https://go.dev/blog/constants) — proč jsou netypované konstanty chytré

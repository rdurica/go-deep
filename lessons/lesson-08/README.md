# Lekce 08 — Mapy

> **Čas:** ~35 min · **Fáze:** 1 — Jazyk a paměťový model · **AI režim:** `ZAKÁZÁNO`

## Co budeš umět

- Vysvětlit, proč zápis do `nil` mapy panikuje, i když čtení z ní projde.
- Rozlišit „klíč chybí" od „klíč má zero value" pomocí comma-ok a vědět, kdy na tom záleží.
- Vysvětlit, proč je pořadí iterace mapy náhodné, a napsat deterministický výpis.
- Rozhodnout mezi `map[K]V` a `map[K]*V` podle toho, jestli potřebuješ mutovat prvek.
- Použít `map[string]struct{}` jako množinu a vědět, proč ne `map[string]bool`.

## Teorie

### Mapa je reference type

Proměnná typu `map[K]V` je ukazatel na hashovací tabulku v runtime. Přiřazení a předání
do funkce kopíruje **jen ten ukazatel**, ne data.

```go
m := map[string]int{"a": 1}
n := m
n["b"] = 2
fmt.Println(len(m)) // 2 — m i n jsou tatáž tabulka

func addKey(m map[string]int) { m["c"] = 3 }
addKey(m)
fmt.Println(len(m)) // 3 — funkce mutovala mapu volajícího
```

To je zásadní rozdíl proti slice. U slice funkce nemůže změnit délku volajícímu; u mapy
může přidávat i mazat, protože délku drží sdílená tabulka. Funkce, která bere mapu jako
parametr, ji tedy může libovolně měnit — a musíš to čekat, i když se to v signatuře
nijak neprojeví.

Mapa **není** porovnatelná: `m1 == m2` se nezkompiluje (kromě porovnání s `nil`).
Na porovnání obsahu použij `reflect.DeepEqual` v testu, nebo si napiš cyklus.

### nil mapa: čtení ano, zápis panic

Zero value mapy je `nil`. Vypadá použitelně a z velké části i je:

```go
var m map[string]int

fmt.Println(m == nil)  // true
fmt.Println(len(m))    // 0
fmt.Println(m["x"])    // 0 — čtení z nil mapy je v pořádku
for k := range m {     // iterace nic neudělá, neselže
	_ = k
}
delete(m, "x")         // no-op, taky OK

m["x"] = 1             // panic: assignment to entry in nil map
```

Tenhle asymetrický design je záměr: čtení má definovanou odpověď („nic tam není"),
zápis nemá kam zapsat. Vytvoření je jedním ze dvou způsobů:

```go
m := make(map[string]int)          // prázdná, připravená
m := make(map[string]int, 1000)    // s předalokací na ~1000 klíčů
m := map[string]int{"a": 1}        // literál
m := map[string]int{}              // prázdný literál
```

Druhý argument `make` je jen nápověda pro předalokaci, ne limit. Když víš, kolik klíčů
bude, ušetříš tím rehashování.

Praktický důsledek pro API: **struktura s mapou uvnitř potřebuje konstruktor**, protože
její zero value nefunguje pro zápis. Buď dej typu `New…()`, nebo mapu líně vytvářej
v metodě, která zapisuje.

### comma-ok: chybí, nebo je to nula?

Čtení chybějícího klíče vrátí zero value hodnoty. To je pohodlné, ale ničí informaci.

```go
scores := map[string]int{"radek": 0}

fmt.Println(scores["radek"]) // 0
fmt.Println(scores["nikdo"]) // 0 — nerozeznatelné!

v, ok := scores["radek"]     // 0, true
v, ok = scores["nikdo"]      // 0, false
```

Druhá návratová hodnota je „klíč existuje". Používej ji vždy, když může být zero value
platná hodnota — u `int`, `bool`, `string` a pointerů prakticky vždy.

Nejčastější idiom je zkrácený `if`:

```go
if v, ok := scores[name]; ok {
	fmt.Println("mám skóre", v)
}
```

Naopak, když ti zero value vyhovuje, comma-ok nepiš. Počítadlo je krásný příklad:

```go
counts := make(map[string]int)
for _, w := range words {
	counts[w]++ // chybějící klíč se čte jako 0, pak zapíše jako 1
}
```

`delete(m, k)` klíč smaže a je bezpečný i pro klíč, který neexistuje, i pro `nil` mapu.
Vrací nic — když chceš vědět, jestli tam klíč byl, zeptej se comma-ok předtím.

### Iterace je záměrně náhodná

Go pořadí iterace **randomizuje při každém běhu**:

```go
m := map[string]int{"a": 1, "b": 2, "c": 3}
for k := range m {
	fmt.Print(k, " ") // třeba "c a b", příště "b c a"
}
```

Není to vedlejší efekt implementace, ale rozhodnutí návrhářů jazyka: kdyby pořadí bylo
náhodou stabilní, začal by na něj kód spoléhat a rozbil by se při jakékoli změně
implementace hashovací tabulky. Radši to selže hned a všem.

Deterministický výstup si vyrobíš přes seřazené klíče:

```go
keys := make([]string, 0, len(m))
for k := range m {
	keys = append(keys, k)
}
sort.Strings(keys)
for _, k := range keys {
	fmt.Println(k, m[k])
}
```

Tenhle blok napíšeš v Go tolikrát, že se ti dostane do prstů. Platí pro každý výstup,
který někdo uvidí nebo porovná: log, JSON, HTTP odpověď, testovací assert.

Jedna související záludnost: **úprava mapy během `range`** má definované chování jen
částečně. Mazat je bezpečné — smazaný klíč se prostě už nenavštíví. Přidávat bezpečné
není: nový klíč se během téhle iterace objevit může a nemusí, podle toho, kam v tabulce
padne.

```go
for k := range m {
	delete(m, k)     // OK, tohle je legální způsob, jak mapu vyprázdnit
}

for k := range m {
	m[k+"!"] = 1     // nedefinované: možná se navštíví, možná ne
}
```

Když potřebuješ během procházení přidávat, posbírej si klíče předem do slice a iteruj
přes něj.

### Nelze vzít adresu prvku

Prvek mapy **není adresovatelný**. Hashovací tabulka si data při růstu přesouvá, takže
pointer na prvek by se mohl kdykoli stát neplatným.

```go
type Item struct{ Qty int }

m := map[string]Item{"sku": {Qty: 1}}
p := &m["sku"]     // chyba kompilace: cannot take address of m["sku"]
m["sku"].Qty = 5   // chyba kompilace: cannot assign to struct field
```

Dvě cesty ven. První: přečti, změň, zapiš zpátky celou hodnotu.

```go
it := m["sku"]
it.Qty = 5
m["sku"] = it
```

Druhá, obvyklejší u větších struktur a u dat, která se často mění: ulož **pointer**.

```go
m := map[string]*Item{"sku": {Qty: 1}}
m["sku"].Qty = 5 // OK — mapa vrací pointer, ten je adresovatelný
```

U mapy pointerů ale pozor na chybějící klíč: `m["neexistuje"]` vrátí `nil` a
`m["neexistuje"].Qty` je nil dereference. Vždycky nejdřív comma-ok:

```go
if it, ok := m[sku]; ok {
	it.Qty += n
} else {
	m[sku] = &Item{Qty: n}
}
```

Volba mezi `map[K]V` a `map[K]*V` je tedy věcná: hodnoty pro malé neměnné struktury
(čitelnější, žádný nil), pointery tam, kde prvky mutuješ nebo je sdílíš jinam.

### Klíče musí být porovnatelné

Typ klíče musí podporovat `==`. To vylučuje slice, mapy a funkce:

```go
map[string]int      // OK
map[int]string      // OK
map[[2]int]bool     // OK — pole je porovnatelné
map[Point]string    // OK, pokud Point je struct z porovnatelných polí
map[[]byte]int      // chyba kompilace: invalid map key type
```

Pozor na `interface{}` jako klíč: zkompiluje se, ale když do něj vložíš slice, panikuje
to až za běhu. A pozor na struct s polem typu float — `NaN != NaN`, takže takový klíč
nikdy nenajdeš.

### Množina přes `map[string]struct{}`

Go nemá typ set. Idiom je mapa s prázdnou strukturou jako hodnotou:

```go
type Set map[string]struct{}

s := Set{}
s["a"] = struct{}{}
_, ok := s["a"]  // true
delete(s, "a")
```

`struct{}` je typ o velikosti **nula bajtů**, takže množina nespotřebuje nic navíc
na hodnoty. Alternativa `map[string]bool` je čitelnější na zápis (`s["a"] = true`), ale
zavádí nesmyslný stav „klíč je v mapě s hodnotou false" — a čtení `if s["a"]` pak splývá
dvě různé věci. U veřejného API volíme `struct{}`.

Protože je `Set` pojmenovaný typ nad mapou, můžeš na něj věšet metody. Stačí **value
receiver** — mapa je reference, takže `func (s Set) Add(...)` mění tu samou tabulku.

### Mapa není thread-safe

Souběžný zápis do mapy z více goroutin runtime detekuje a shodí program s
`fatal error: concurrent map writes`. Není to panic, kterou bys odchytil — je to
tvrdý pád. Souběžné čtení bez zápisu je bezpečné.

Řešení jsou `sync.Mutex` kolem mapy nebo `sync.Map` pro specifické případy; detailně
v lekci 43. Zatím si zapamatuj jen to, že běžná mapa se mezi goroutinami nesdílí.

## Rozdíly proti PHP

PHP má jeden kontejner na všechno. Pole je zároveň seznam i slovník, a `foreach` po něm
jde v pořadí vkládání.

```php
$counts = [];
$counts['ahoj']++;          // funguje (s warningem), klíč se vytvoří
foreach ($counts as $k => $v) { /* pořadí vkládání */ }
if (isset($counts['x'])) { /* … */ }
```

Go má mapu jako samostatný typ s pevným typem klíče i hodnoty, s náhodným pořadím iterace
a bez „automatického vytvoření" celé mapy.

```go
counts := make(map[string]int)
counts["ahoj"]++            // funguje: chybějící klíč se čte jako 0
for k, v := range counts {  // POŘADÍ JE NÁHODNÉ
	fmt.Println(k, v)
}
if _, ok := counts["x"]; ok { /* … */ }
```

Přenos návyku: přestaň předpokládat pořadí. V PHP se `foreach` chová deterministicky
a spousta kódu na to nevědomky spoléhá — třeba když se výstup posílá do šablony nebo
porovnává v testu. V Go musíš pořadí vždy vyrobit sám, přes seřazený seznam klíčů.
Druhá změna: mapu musíš explicitně vytvořit, jinak dostaneš `nil` a panic při zápisu.

## Časté chyby

| Chyba | Proč vzniká | Jak to udělat správně |
|-------|-------------|------------------------|
| Zápis do `nil` mapy | `var m map[string]int` vypadá jako prázdná mapa | `make(map[string]int)` nebo literál |
| Spoléhání na pořadí iterace | v PHP `foreach` jede v pořadí vkládání | posbírej klíče a `sort.Strings` |
| `if m[k] != 0` místo comma-ok | zvyk na `isset()`, který v Go nemá přímý protějšek | `if v, ok := m[k]; ok` |
| `m[k].Field = x` u mapy hodnot | reflex z PHP, kde je všechno objekt | `map[K]*V`, nebo přečíst → změnit → zapsat |
| Nil dereference u `map[K]*V` | chybějící klíč vrátí `nil` pointer | comma-ok a založení prvku, když chybí |
| `map[string]bool` jako množina | vypadá čitelněji | `map[string]struct{}` — nula bajtů, jeden stav |
| Sdílení mapy mezi goroutinami | v PHP-FPM žádný souběh není | mutex kolem mapy (lekce 43) |

## AI kvíz

Po přečtení teorie spusť v Cursoru **`/go-deep-quiz 08`**. AI tě ~5 minut prověří mentální model (ne hotové cvičení). Slabiny si uloží do [`GAPS.md`](../../GAPS.md).

## Úkol

Pracuj v `exercise/`. Kontrakt je v komentáři nad funkcí. Stupně jdou od jednodušších
ke složitějším — po každém stupni spusť review, než jdeš dál.

### Jednoduchý

Oprav: `CloneMap` (kód je záměrně vadný — sdílení mapy)

```bash
make lesson L=08 PART=1
```

Pak **`/go-deep-review 08 easy`**.

### Střední

Implementuj: `WordCount`, `NewSet`, `Add`, `Has`

```bash
make lesson L=08 PART=2
```

Pak **`/go-deep-review 08 medium`**.

### Obtížný

Doplň: `AddStock` (`map[K]*V` — mutace přes pointer)

```bash
make lesson L=08 PART=3
```

Pak **`/go-deep-review 08 hard`**.

Až budou stupně hotové, porovnej se `solutions/` (spoiler).

## Závěrečné otázky

Spusť **`/go-deep-review 08 final`**. AI projde body níže, doptá se a ověří pochopení. Celé cvičení ověří `make lesson L=08` (+ `make race L=08`, pokud to lekce vyžaduje).

- [ ] Umíš vysvětlit, proč je čtení z `nil` mapy OK a zápis ne
- [ ] Umíš z hlavy napsat deterministický výpis mapy
- [ ] Umíš vysvětlit, proč `&m[k]` nejde a jaké jsou dvě obcházky
- [ ] Umíš říct, kdy volit `map[K]V` a kdy `map[K]*V`
- [ ] Umíš vysvětlit, proč je `map[string]struct{}` lepší množina než `map[string]bool`

## AI režim

`ZAKÁZÁNO` — viz [docs/ai-playbook.md](../../docs/ai-playbook.md).

Mentor, kvíz i review (dialog) jsou vždy OK; v tomto režimu AI nesmí psát kód cvičení.
## Další čtení

1. [Go blog — Go maps in action](https://go.dev/blog/maps)
2. [Effective Go — Maps](https://go.dev/doc/effective_go#maps)
3. [Go spec — Map types](https://go.dev/ref/spec#Map_types)

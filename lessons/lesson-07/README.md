# Lekce 07 — Slices: pole, append, internals, aliasing

> **Čas:** ~90 min · **Fáze:** 1 — Jazyk a paměťový model · **AI režim:** `ZAKÁZÁNO`

## Co budeš umět

- Nakreslit z hlavy paměťový model slice (`ptr`, `len`, `cap`) a odvodit z něj chování `append`.
- Vysvětlit, proč `append` někdy mutuje volajícího a někdy ne, a proč se návratová hodnota nikdy nesmí zahodit.
- Rozpoznat aliasing dvou slice nad stejným backing polem a zabránit mu přes `s[a:b:c]` nebo `copy`.
- Rozhodnout mezi `nil` slice a prázdným slice a vědět, kdy na tom rozdílu záleží.
- Napsat mazání prvku se zachováním i bez zachování pořadí a vědět, co každá varianta stojí.

## PHP → Go most

PHP pole je jeden univerzální kontejner: uspořádaná hash mapa, která umí být seznam
i slovník. Roste sama, kopíruje se při přiřazení (copy-on-write) a nemusíš o ní
přemýšlet.

```php
$a = [1, 2, 3];
$b = $a;        // logická kopie, $b je nezávislé
$b[] = 4;
count($a);      // 3 — $a se nezměnilo
```

Go rozděluje tenhle kontejner na dvě věci. **Pole** (`[3]int`) má délku v typu, je to
hodnota a při přiřazení se opravdu zkopíruje. **Slice** (`[]int`) je malá struktura
ukazující do cizí paměti a při přiřazení se zkopíruje jen ta struktura, ne data.

```go
a := []int{1, 2, 3}
b := a          // zkopíruje se jen header, data jsou sdílená
b[0] = 99
fmt.Println(a[0]) // 99 — a se změnilo!

b = append(b, 4)  // tady už se možná alokovalo nové pole
```

Přenos návyku: v PHP se ptáš „co je v poli". V Go se musíš ptát **„kdo vlastní backing
array a kdo do něj ještě kouká"**. Slice je pohled na paměť, ne vlastník dat. Tohle je
jediný koncept z celé fáze 1, na kterém opravdu zakopne každý, kdo přichází z PHP —
protože vypadá jako pole, ale chová se jako ukazatel s délkou.

## Teorie

### Pole je hodnota, slice je pohled

Pole má délku součástí typu. `[3]int` a `[4]int` jsou dva různé, nekompatibilní typy.
Přiřazení i předání do funkce pole **celé zkopíruje**.

```go
arr := [3]int{1, 2, 3}
cp := arr
cp[0] = 99
fmt.Println(arr[0], cp[0]) // 1 99 — nezávislé kopie

func mutate(a [3]int) { a[0] = 7 }
mutate(arr)
fmt.Println(arr[0]) // 1 — funkce dostala kopii
```

Pole se v praxi používá zřídka: jako pevný buffer (`[32]byte` pro hash), jako klíč mapy
(pole je porovnatelné, slice ne) nebo tam, kde chceš zaručit, že se data zkopírují.
Ve zbytku kódu potkáš slice.

Slice je **header**, tři slova (na 64bitové platformě 24 bajtů):

```
s := make([]int, 3, 5)

slice header                       backing array (5 × int)
+---------+---------+---------+    +-----+-----+-----+-----+-----+
|   ptr   |  len=3  |  cap=5  |--->|  0  |  0  |  0  |  -  |  -  |
+---------+---------+---------+    +-----+-----+-----+-----+-----+
                                      0     1     2     3     4
                                   <----- len=3 ---->
                                   <---------- cap=5 ----------->
```

- `len` je počet prvků, které smíš indexovat. `s[3]` panikuje, i když tam paměť je.
- `cap` je počet prvků, které se do backing pole vejdou od `ptr` dál.
- Prvky mezi `len` a `cap` existují, jsou vynulované, ale nejsou „tvoje" — dokud je
  nezpřístupníš přes `append` nebo reslicing.

`make([]int, 3)` je zkratka pro `make([]int, 3, 3)`. Literál `[]int{1, 2, 3}` alokuje pole
o třech prvcích a udělá nad ním slice s `len == cap == 3`.

### append a realokace

`append` se pokusí zapsat za konec `len`. Pokud tam je místo (`cap > len`), zapíše **do
existujícího backing pole** a vrátí header s větším `len`. Pokud místo není, alokuje nové,
větší pole, zkopíruje do něj data a vrátí header ukazující jinam.

```go
s := make([]int, 0, 2)
fmt.Println(len(s), cap(s)) // 0 2

s = append(s, 1)            // vejde se, stejné pole
s = append(s, 2)            // vejde se, stejné pole
fmt.Println(len(s), cap(s)) // 2 2

s = append(s, 3)            // nevejde se → nové pole
fmt.Println(len(s), cap(s)) // 3 4
```

Růst je **amortizovaný**: runtime kapacitu zhruba zdvojnásobuje (u velkých slice pomaleji),
takže `n` volání `append` stojí `O(n)`, ne `O(n²)`. Konkrétní čísla ale nejsou součástí
specifikace jazyka a mezi verzemi Go se mění. **Nikdy nespoléhej na konkrétní `cap` po
appendu.**

Z toho plyne pravidlo číslo jedna: **návratovou hodnotu `append` musíš vždy přiřadit.**

```go
append(s, 1)        // chyba kompilace: výsledek se nepoužívá
s = append(s, 1)    // správně
```

A pravidlo číslo dvě: `append` uvnitř funkce **nemůže** spolehlivě rozšířit slice
volajícího, protože funkce dostala kopii headeru.

```go
func addBroken(s []int) {
	s = append(s, 42) // změní jen lokální kopii headeru
}

func addOK(s []int) []int {
	return append(s, 42) // volající si výsledek uloží
}

nums := []int{1, 2, 3}
addBroken(nums)
fmt.Println(len(nums)) // 3 — nic se nestalo
nums = addOK(nums)
fmt.Println(len(nums)) // 4
```

Zákeřné je, že `addBroken` občas „funguje". Když má slice volná místa v `cap`, zapíše se
hodnota do sdíleného pole a volající ji uvidí, jakmile si `len` sám zvětší. Chování tedy
závisí na kapacitě — což je přesně ten druh bugu, který v testech neuvidíš a v produkci ano.

Když počet prvků znáš dopředu, předalokuj. Ušetříš všechna mezialokace i kopírování:

```go
// dobré: jedna alokace, len 0 a cap n
out := make([]string, 0, len(rows))
for _, r := range rows {
	out = append(out, r.Name)
}

// častá chyba: make([]T, n) dá n prázdných prvků a append přidává ZA ně
out := make([]string, len(rows))
for _, r := range rows {
	out = append(out, r.Name) // výsledek má 2×len(rows) prvků, půlka prázdných
}
```

Rozdíl mezi druhým a třetím argumentem `make` je tedy rozdíl mezi „kolik prvků tam už je"
a „pro kolik prvků je místo". S `append` chceš skoro vždycky `make([]T, 0, n)`.

### Aliasing: dva slice, jedno pole

Slicing `s[a:b]` nekopíruje. Vytvoří nový header ukazující do stejného pole.

```go
a := []int{10, 20, 30, 40, 50}
b := a[1:3] // len=2, cap=4 (od indexu 1 do konce pole)

           +----+----+----+----+----+
backing:   | 10 | 20 | 30 | 40 | 50 |
           +----+----+----+----+----+
             ^    ^
             |    |
           a.ptr  b.ptr
                  <-- b.len=2 -->
                  <--------- b.cap=4 --------->

b[0] = 99
fmt.Println(a) // [10 99 30 40 50] — zápis přes b je vidět v a
```

Nová `cap` je `cap(a) - a`, tedy sahá **až na konec původního pole**. A právě tady vzniká
klasická past:

```go
a := []int{10, 20, 30, 40, 50}
b := a[1:3]              // [20 30], cap 4
b = append(b, 999)       // vejde se do cap → přepíše a[3]!
fmt.Println(a)           // [10 20 30 999 50]
```

`append` do subslice tiše přepsal data v originálu. Žádná chyba, žádné varování. Pokud
`b` předáš cizí funkci, ta ti může rozbít data, o kterých nic neví.

Obrana je tříindexový slicing `s[low:high:max]`, který navíc omezí kapacitu:

```go
b := a[1:3:3]            // len = 3-1 = 2, cap = 3-1 = 2
b = append(b, 999)       // cap je plná → alokuje nové pole
fmt.Println(a)           // [10 20 30 40 50] — originál nedotčen
```

Vzorec: `s[low:high:max]` dá `len = high-low` a `cap = max-low`. Když `max == high`, je
slice „plný" a první `append` vždy realokuje. Tomuhle se říká *full slice expression* a
používá se pokaždé, když vracíš výřez cizích dat ven z balíčku.

### copy, mazání a nezávislé kopie

`copy(dst, src)` zkopíruje `min(len(dst), len(src))` prvků a vrátí jejich počet. Nikdy
nealokuje a nikdy neroste — když je `dst` kratší, zbytek se prostě nezkopíruje. Tohle je
nejčastější zdroj „proč mám prázdný slice":

```go
src := []int{1, 2, 3}
var dst []int
fmt.Println(copy(dst, src)) // 0 — dst má len 0, nic se nezkopíruje

dst = make([]int, len(src))
copy(dst, src)              // 3, teď je dst nezávislá kopie
```

Mazání prvku se zachováním pořadí je `append` se spread operátorem:

```go
// smaže s[i], zachová pořadí, O(n)
s = append(s[:i], s[i+1:]...)
```

Když na pořadí nezáleží, je varianta „swap s posledním" v `O(1)`:

```go
s[i] = s[len(s)-1]
s = s[:len(s)-1]
```

Obě varianty **mutují původní backing pole** — po smazání jsou data volajícího posunutá.
A obě nechávají poslední prvek v paměti; u slice pointerů nebo struktur s pointery
to brání GC uvolnit paměť, takže se tam osiřelé místo nuluje (`s[len(s)-1] = nil`).

### nil slice vs prázdný slice

```go
var a []int          // nil: ptr=nil, len=0, cap=0
b := []int{}         // ne-nil: ptr ukazuje na prázdné pole, len=0, cap=0

fmt.Println(a == nil, b == nil) // true false
fmt.Println(len(a), len(b))     // 0 0
a = append(a, 1)                // funguje, nil slice je platný začátek
for range a {}                  // funguje
```

Pro 95 % kódu je rozdíl nepodstatný — `len`, `range`, `append` i `copy` fungují na obou.
Idiomatické je deklarovat `var s []T` a nechat `append`, ať alokuje. Rozdíl je vidět jen
tam, kde se identita `nil` zkoumá: `encoding/json` serializuje nil slice jako `null`
a prázdný jako `[]`, a `reflect.DeepEqual(a, b)` je pro tuhle dvojici `false`.

Praktický důsledek pro API: pokud funkce vrací „žádné výsledky", vracej klidně `nil`.
Volající to nemusí řešit. Ale pokud tvoje funkce má vrátit **kopii**, musí `nil` vstup
vracet jako `nil` výstup a ne-nil vstup jako ne-nil výstup — jinak se ti rozjede JSON.

### Slice jako parametr funkce

Funkce dostane **kopii headeru**, ne kopii dat. Z toho plynou tři pravidla:

1. Změna prvku (`s[i] = x`) je vidět venku.
2. Změna délky (`s = append(...)`, `s = s[:2]`) venku vidět **není**.
3. Chceš-li měnit délku, vracej nový slice: `func Grow(s []int) []int`.

To je také důvod, proč skoro žádná funkce ve stdlib nebere `*[]T`. Pointer na slice
potřebuješ jen výjimečně (typicky `json.Unmarshal`). Idiomatické Go místo toho vrací.

Pokud funkce nesmí vstup rozbít, musí si udělat vlastní kopii — buď přes `make` + `copy`,
nebo vstup předat jako `s[:len(s):len(s)]`, aby první `append` vždy alokoval.

## Časté chyby

| Chyba | Proč vzniká | Jak to udělat správně |
|-------|-------------|------------------------|
| `append(s, x)` bez přiřazení | v PHP `$a[] = $x` mutuje na místě | vždy `s = append(s, x)` |
| Funkce „přidá" prvek přes parametr | reflex z PHP, kde objekty jdou po referenci | vracej nový slice: `s = Add(s)` |
| `b := a[1:3]; b = append(b, x)` přepíše `a` | `cap` subslice sahá na konec pole | `a[1:3:3]` nebo `copy` do nového slice |
| `copy(dst, src)` do prázdného `dst` | očekává se, že `copy` alokuje | `dst := make([]T, len(src))` |
| Spoléhání na konkrétní `cap` po `append` | domněnka, že růst je definovaný | testuj `len`, ne `cap` |
| Vracení výřezu interního bufferu | v PHP by šlo o kopii | vrať `Clone(s)` nebo `s[a:b:b]` |
| Porovnávání slice přes `==` | zvyk na `$a == $b` v PHP | prvek po prvku, nebo `reflect.DeepEqual` v testu |

## Úkol

Pracuj v `exercise/`. Postupuj A → B → C, po každé části spusť test.

### A — rozcvička (~10 min)

1. `Sum(nums []int) int` — součet prvků. Prázdný i `nil` vstup dá `0`.
2. `Reverse(nums []int)` — otočí pořadí **in-place**, nic nevrací. Musí fungovat pro
   sudou i lichou délku, pro prázdný i `nil` vstup. Použij dva indexy zleva a zprava,
   nealokuj pomocný slice.

Rozmysli si, proč `Reverse` nemusí nic vracet, i když `Grow` z části B musí.

### B — jádro (~35 min)

1. `Grow(s []int, n int) []int` — zajistí, že výsledek má `cap >= n`, a **zachová `len`
   i obsah**. Pokud už `cap(s) >= n`, vrať `s` beze změny (tedy stejný header nad stejným
   backing polem — test to kontroluje porovnáním adresy prvního prvku). Jinak alokuj nové
   pole a data zkopíruj.
2. `RemoveAt(s []int, i int) []int` — smaže prvek na indexu `i` a **zachová pořadí**
   zbytku. Index mimo rozsah (`i < 0` nebo `i >= len(s)`) znamená „nedělej nic": vrať `s`
   beze změny. Funkce smí (a bude) mutovat backing pole volajícího — to je záměr.
3. `RemoveFast(s []int, i int) []int` — totéž, ale v `O(1)` a **bez zachování pořadí**:
   na uvolněné místo přesuň poslední prvek a slice zkrať. Index mimo rozsah opět vrací
   `s` beze změny.
4. `Clone(s []int) []int` — vrátí **nezávislou** kopii: zápis do výsledku nesmí být vidět
   ve vstupu a naopak. `nil` vstup vrací `nil`. Prázdný ne-nil slice vrací prázdný ne-nil
   slice.

### C — rozšíření (~25 min)

1. `Chunk(s []int, size int) [][]int` — rozdělí `s` na kusy délky `size`; poslední kus
   může být kratší. `size <= 0` nebo prázdný vstup vrací výsledek s nulovou délkou.
   **Každý chunk musí být nezávislá kopie** — test zapíše do jednoho chunku a ověří, že
   se to neprojevilo v ostatních ani ve vstupu. Naivní `s[i:j]` tímhle testem neprojde.
2. `Filter(s []int, keep func(int) bool) []int` — vrátí prvky, pro které `keep` vrátí
   `true`, **implementovaný trikem `s[:0]`**: výsledek se skládá do stejného backing pole
   jako vstup, bez jediné alokace. Zachovej pořadí.

   Test tenhle trik cíleně odhaluje: po zavolání `Filter` zkontroluje, že se **vstupní
   slice přepsal**. To je cena za nulovou alokaci a musíš o ní vědět — proto se tenhle
   vzor nikdy nepoužívá na data, která ti nepatří.

```bash
make lesson L=07
```

Až budeš hotový, porovnej se `solutions/` (spoiler).

## Ověření

- [ ] `make lesson L=07` prochází
- [ ] Umíš nakreslit slice header a vysvětlit rozdíl mezi `len` a `cap`
- [ ] Umíš vysvětlit, proč `append` uvnitř funkce nezvětší slice volajícího
- [ ] Umíš vysvětlit, kdy `a[1:3]` a `a[1:3:3]` dají různý výsledek
- [ ] Umíš říct, proč `copy(dst, src)` do nil `dst` zkopíruje nula prvků
- [ ] Umíš uvést situaci, kdy na rozdílu mezi `nil` a prázdným slice opravdu záleží

## AI režim

`ZAKÁZÁNO` — viz [docs/ai-playbook.md](../../docs/ai-playbook.md).

## Další čtení

1. [Go blog — Go Slices: usage and internals](https://go.dev/blog/slices-intro)
2. [Go blog — Arrays, slices (and strings): The mechanics of 'append'](https://go.dev/blog/slices)
3. [Go wiki — SliceTricks](https://go.dev/wiki/SliceTricks)
4. [Go spec — Slice expressions](https://go.dev/ref/spec#Slice_expressions)

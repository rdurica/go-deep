# Lekce 06 — Pointery: hodnota vs reference

> **Čas:** ~90 min · **Fáze:** 1 — Jazyk a paměťový model · **AI režim:** `ZAKÁZÁNO`

## Co budeš umět

- Přečíst `&` a `*` v cizím kódu a říct, co se v paměti děje.
- Zdůvodnit, kdy použít pointer a kdy je hodnota lepší volba.
- Rozpoznat situace, kdy je pointer ve struct fieldu správně (a kdy je to jen zbytečná past).
- Vysvětlit, proč `slice` mění data volajícího i bez pointeru a proč `append` musíš vracet.
- Bezpečně pracovat s `nil` pointerem, včetně `nil` receiveru.

## PHP → Go most

V PHP je proměnná s objektem handle. Předáš ji dál a příjemce mění tvoje data — a ty to
na signatuře nepoznáš.

```php
function applyDefaults(Config $c): void
{
    $c->host ??= 'localhost';  // mění objekt volajícího
}

function rename(string $s): void
{
    $s = 'jiny';               // nemění nic, string je hodnota
}
```

Go dělá totéž, ale explicitně: co se má měnit, se předává pointerem, a je to vidět v typu.

```go
func ApplyDefaults(c *Config) { // hvězdička = "budu ti do toho sahat"
	if c.Host == "" {
		c.Host = "localhost"
	}
}

func rename(c Config) { // bez hvězdičky = dostávám kopii
	c.Host = "jiny" // volajícího to nezmění
}
```

Co se mění v uvažování: přestaneš hádat, jestli funkce mutuje vstup. Signatura to říká.
Zároveň se otočí výchozí volba — v PHP jsou reference zadarmo a všude, v Go začínáš
hodnotou a pointer si musíš zdůvodnit.

## Teorie

### Co je pointer a co s ním jde

Pointer je hodnota, která obsahuje adresu jiné hodnoty. Dva operátory:

```go
x := 42
p := &x     // p je *int, adresa x
fmt.Println(*p) // 42 — dereference, čtení skrz adresu
*p = 7          // zápis skrz adresu
fmt.Println(x)  // 7
```

Zero value pointeru je `nil`. Dereference `nil` pointeru je runtime panika
`invalid memory address or nil pointer dereference` — ne tichá nula jako `$x?->y` v PHP.

Co v Go **nejde**, na rozdíl od C: pointerová aritmetika. `p+1` je chyba kompilace. Důvod je
garbage collector — ten musí přesně vědět, které paměťové bloky jsou ještě dosažitelné.
Kdyby si program mohl adresy vyrábět počítáním, nešlo by to. Bezpečnou práci „s posunem"
zajišťují slices, které kombinují pointer, délku a kapacitu do jedné hodnoty.

### Všechno se předává hodnotou

Tohle je jedna věta, ze které plyne skoro celý zbytek lekce: **Go předává argumenty vždycky
kopií.** Když předáš `int`, zkopíruje se číslo. Když předáš `Config`, zkopíruje se celý
struct. Když předáš `*Config`, zkopíruje se adresa — a protože kopie adresy ukazuje na totéž
místo, může přes ni funkce originál změnit.

Z toho plyne past, na kterou se naráží u pointerů na pointery:

```go
func replace(c *Config) {
	c = &Config{Host: "novy"} // mění jen lokální kopii pointeru!
}

func fill(c *Config) {
	c.Host = "novy" // tohle funguje — píšeme skrz adresu
}
```

Když funkce potřebuje volajícímu vyměnit celý objekt, musí buď nový vrátit, nebo dostat
`**Config`. To druhé v Go skoro nikdy nechceš; vracení je čitelnější.

### `new()` vs `&T{}`

Dvě cesty, jak vyrobit pointer na novou hodnotu:

```go
c1 := new(Config)  // *Config, všechna pole zero value
c2 := &Config{}    // totéž
c3 := &Config{Host: "localhost", Port: 8080} // navíc rovnou inicializuje
```

`new(T)` alokuje zero value a vrátí `*T`. `&T{}` umí to samé a ještě dovolí pole vyplnit,
takže v praxi uvidíš skoro výhradně tu druhou formu. `new` má smysl u typů, které nemají
literál — `p := new(int)` je jediný způsob, jak dostat `*int` bez pomocné proměnné.

Poznámka pro klid duše: `&Config{}` uvnitř funkce není chyba, i když se ta lokální proměnná
„vrátí ven". Kompilátor si přes **escape analysis** spočítá, že hodnota přežije návrat, a
alokuje ji na haldě. Nemusíš (a nemůžeš) to řídit ručně. Praktický důsledek je jen výkonnostní:
zbytečné pointery znamenají víc práce pro GC. Detaily jsou téma pro lekci o pprof.

### Kdy pointer receiver

Pravidlo z lekce 05 platí i tady, jen se na něj podíváme z paměťové strany:

```go
func (c *Counter) Inc() { c.n++ }  // musí měnit stav → pointer
func (p Point) Add(q Point) Point  // jen počítá → hodnota
```

Value receiver navíc znamená, že metodu můžeš bezpečně volat souběžně, protože pracuje
s vlastní kopií. To je silný argument pro hodnotové typy v doméně (`Money`, `Point`, `Date`).

Zvláštnost, která překvapí: metoda s pointer receiverem se dá zavolat i na `nil` pointeru,
pokud tělo na receiver nesáhne. To se využívá u stromových struktur:

```go
func (n *Node) Len() int {
	if n == nil {
		return 0 // legální, žádná panika
	}
	return 1 + n.Next.Len()
}
```

Panika nastane až při skutečné dereferenci (`n.Val`).

### Kdy pointer ve struct fieldu

Pointer v poli má tři legitimní důvody:

1. **Optional hodnota.** Když je zero value platná hodnota a potřebuješ rozlišit
   „nenastaveno". Klasika je `Debug *bool`: `nil` znamená „uživatel neřekl", `&false` znamená
   „uživatel řekl vypnout". Se `bool` tyhle dva stavy nerozlišíš.
2. **Sdílená identita.** Dva kusy kódu mají mít tentýž objekt, ne dvě kopie.
3. **Rekurzivní struktura.** `Next *Node` musí být pointer — struct nemůže obsahovat sám
   sebe hodnotou, protože by měl nekonečnou velikost.

Ve všech ostatních případech je pointer daň navíc: kontrola na `nil` u každého přístupu, horší
čitelnost a práce pro GC. Pro optional hodnoty ve vlastním API často zvládneš i dvojici
`(value, bool)` nebo `sql.NullString`-like typ.

### Slice, mapa a „reference bez pointeru"

Slice je struct o třech polích: pointer na podkladové pole, délka, kapacita. Předává se
hodnotou jako všechno ostatní, ale ten vnitřní pointer se kopíruje taky — takže funkce sahá
na stejná data:

```go
func IncrementAll(nums []int) {
	for i := range nums {
		nums[i]++ // vidět i u volajícího
	}
}

func broken(nums []int) {
	nums = append(nums, 1) // volající nic neuvidí: měníme kopii headeru
}
```

`append` může (a nemusí) alokovat nové pole a v každém případě mění délku, která je součástí
**kopie** headeru. Proto se výsledek `append` vždycky vrací nebo přiřazuje.

Ošidnější je opačný případ: když má slice volný prostor v kapacitě, `append` zapíše do
původního pole a přepíše data, na která se ještě někdo dívá. Proto úkol B žádá funkci, která
si data nejdřív okopíruje. Celé to rozebereme v lekci 07.

Mapa je oproti tomu ukazatel na hashtable, takže funkce vidí i vložené klíče. Má ale
omezení, které se přímo týká pointerů: **adresu prvku mapy nelze vzít**.

```go
m := map[string]Counter{"a": {}}
p := &m["a"]  // chyba kompilace: cannot take address of m["a"]
m["a"].Inc()  // chyba: prvek mapy není adresovatelný
```

Důvod je, že mapa si prvky při růstu přesouvá, takže by adresa přestala platit. Řešení jsou
dvě: buď ukládej pointery (`map[string]*Counter`), nebo hodnotu vytáhni, změň a vlož zpátky
(`c := m["a"]; c.Inc(); m["a"] = c`).

Stejné omezení má návratová hodnota funkce a konstanta. Prvek slice adresovatelný **je**
(`&s[0]` je v pořádku), protože slice si pole sám nepřesouvá — ale pozor, `append` ho
přesunout může a starý pointer pak ukazuje do neaktuálních dat.

## Časté chyby

| Chyba | Proč vzniká | Jak to udělat správně |
|-------|-------------|------------------------|
| Pointer u každého parametru | objekty v PHP jsou reference | výchozí je hodnota, pointer když měníš nebo sdílíš |
| `c = &Config{}` uvnitř funkce | vypadá to jako mutace | vrať novou hodnotu, měň skrz `c.Pole` |
| Zapomenutý výsledek `append` | slice vypadá jako reference | `s = append(s, v)` vždy |
| `&m["klic"]` | prvek mapy vypadá jako proměnná | `map[K]*V`, nebo hodnotu vlož zpátky |
| Dereference bez kontroly `nil` | zvyk na `?->` a tiché null | ošetři `nil` nebo garantuj neexistenci v API |
| `*bool` „pro jistotu" | pointer jako univerzální optional | pointer jen tam, kde `false` a „nenastaveno" nejsou totéž |

## Úkol

Pracuj v `exercise/`. Postupuj A → B → C, po každé části spusť test.

### A — rozcvička (~10 min)

`Swap(a, b *int)` prohodí hodnoty, na které pointery ukazují. Když je kterýkoli z nich `nil`,
funkce **nesmí panikovat** a nemá dělat nic. Volání `Swap(&x, &x)` musí `x` nechat beze změny.

### B — jádro (~35 min)

1. `ApplyDefaults(c *Config)` doplní do konfigurace výchozí hodnoty:
   - prázdný `Host` → `"localhost"`,
   - `Port` rovný nule → `8080`,
   - `Debug` rovný `nil` → pointer na `false`.

   Už nastavené hodnoty se nesmí přepsat — a to včetně `Debug`, který ukazuje na `false`.
   Přesně kvůli tomu je pointer: `nil` je „nenastaveno", `&false` je „nastaveno na false".
   Pokud dostane `nil` místo konfigurace, neudělá nic a nepanikuje.

2. `IncrementAll(nums []int)` zvýší každý prvek o jedna **na místě**, bez pointeru a bez
   návratové hodnoty. `nil` i prázdný slice projdou bez paniky. Test ověří i to, že změna
   přes podslice (`nums[1:]`) je vidět v původním slice.

3. `AppendSafe(nums []int, v int) []int` vrátí nový slice s přidaným prvkem a **nesmí sáhnout
   na podkladové pole volajícího** — ani když má vstupní slice volnou kapacitu. To znamená,
   že si musíš data zkopírovat, ne jen zavolat `append`. Pro `nil` vstup vrať slice s jedním
   prvkem.

### C — rozšíření (~25 min)

Jednosměrně zřetězený seznam. Typ `Node{Val int; Next *Node}` je předpřipravený, prázdný
seznam reprezentuje `nil` hlava. Všechny tři funkce musí `nil` hlavu zvládnout.

- `Push(head *Node, v int) *Node` — vloží prvek na **začátek** a vrátí novou hlavu.
  `Push(nil, 42)` vytvoří jednoprvkový seznam.
- `Len(head *Node) int` — počet prvků, `nil` hlava dá `0`. Napiš to cyklem, ne rekurzí.
- `Reverse(head *Node) *Node` — otočí pořadí a vrátí novou hlavu. Zvládni to jedním průchodem
  s pomocnými pointery (`prev`, `next`), bez pomocného slice. `Reverse(nil)` vrací `nil`.

Až budeš hotový, projdi si `Reverse` po řádcích s tužkou nad papírem a nakresli, kam ukazují
`prev`, `node` a `next` v každé iteraci. Tenhle jeden obrázek je důvod, proč je v kurzu
zřetězený seznam, i když bys ho v produkci nikdy nepsal.

```bash
make lesson L=06
```

Až budeš hotový, porovnej se `solutions/` (spoiler).

## Ověření

- [ ] `make lesson L=06` prochází
- [ ] Umíš vysvětlit, proč `func f(c *Config) { c = &Config{} }` nic neudělá
- [ ] Umíš vysvětlit, proč `IncrementAll` mění data i bez pointeru, ale `append` uvnitř funkce ne
- [ ] Umíš uvést tři důvody pro pointer ve struct fieldu
- [ ] Umíš vysvětlit, proč nejde vzít adresu prvku mapy
- [ ] Umíš říct, kdy volání metody na `nil` pointeru nepanikuje

## AI režim

`ZAKÁZÁNO` — viz [docs/ai-playbook.md](../../docs/ai-playbook.md).

## Další čtení

1. [Tour of Go — Pointers](https://go.dev/tour/moretypes/1)
2. [Go FAQ — When are function parameters passed by value?](https://go.dev/doc/faq#pass_by_value)
3. [Go blog — Go Slices: usage and internals](https://go.dev/blog/slices-intro)
4. [Go Code Review Comments — Pointers to Interfaces](https://go.dev/wiki/CodeReviewComments#pointers-to-interfaces)

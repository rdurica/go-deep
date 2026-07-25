# Lekce 59 — Capstone P05: spec a implementace

> **Čas:** ~90 min · **Fáze:** 8 — Capstone · **AI režim:** `TECH LEAD`

## Co budeš umět

- Napsat specifikaci capstone služby tak, aby z ní šlo přímo odvodit akceptační testy.
- Rozdělit práci mezi sebe a agenta podle protokolu z lekce 58 a udržet si vlastnictví domény.
- Postavit doménový model se skutečnou normalizací a validací na hranici, ne s anemickým DTO.
- Napsat in-memory store s indexem, který je bezpečný pro souběžné použití a nesdílí vnitřní data.
- Navrhnout stránkované hledání se stabilním řazením a cursorem, který se nerozbije při zápisu.

## PHP → Go most

Capstone v Symfony bys složil z hotových dílů: entity s Doctrine anotacemi, repository
s `findBy`, validaci přes atributy, paginátor z bundle.

```php
#[Assert\Url]
private string $url;

$paginator = $this->repo->findBy(['tag' => $tag], ['createdAt' => 'DESC'], 20, $offset);
```

V Go žádný z těch dílů nedostaneš — a to je smysl capstone. Napíšeš si validaci jako
funkci, index jako mapu, stránkování jako slice s cursorem, a poznáš, kolik z toho, co ti
framework dělal, jsou vlastně tři řádky:

```go
func (b Bookmark) Validate() error {
	if strings.TrimSpace(b.ID) == "" {
		return ErrEmptyID
	}
	// ...
}
```

Změna v uvažování: **offset paginace je v Go stejně špatný nápad jako v PHP, jen ti to
tady nikdo nezakryje.** Když mezi dvěma stránkami někdo přidá záložku, offset ti položku
zopakuje nebo přeskočí. Cursor tenhle problém nemá, protože se neptá „kolikátý", ale
„po kterém".

## Teorie

### Spec, ze které jde odvodit test

Capstone P05 je služba záložek: doména, in-memory store, HTTP API, konfigurace, provoz.
Kompletní zadání je v [projects/p05-capstone/SPEC.md](../../projects/p05-capstone/SPEC.md)
a odškrtávací kritéria v `ACCEPTANCE.md`. Tahle lekce staví **jádro té spec v malém**,
abys ho mohl otestovat dřív, než k němu přilepíš HTTP.

Dobrá spec pro doménu vypadá takhle konkrétně:

> `NormalizeURL` sjednotí schéma a host na malá písmena, odstraní výchozí port
> (`:80` pro http, `:443` pro https), zahodí fragment, odstraní všechny query parametry
> začínající `utm_` (case-insensitive), zbytek seřadí podle klíče a ořízne koncové
> lomítko. Vstup bez schématu `http`/`https` je chyba `ErrInvalidURL`.

Z tohohle odstavce napíšeš tabulkový test za pět minut a nikdy se nebudeš ptát „a co
když". Přesně tohle je vstup, který dáváš agentovi. Kdyby tam chybělo slovo
„case-insensitive", dostaneš implementaci, která `UTM_SOURCE` nechá — a nebude to jeho
chyba.

### Pořadí stavby

Stavěj **odspodu podle závislostí**, ne odshora podle vrstev:

1. **Doména bez závislostí** — typy, normalizace, validace, doménové chyby. Testuješ
   čistě, bez `httptest`, bez store.
2. **Port a in-memory adaptér** — `Add`, `Get`, `Delete`, `ByTag`. Testuješ s `-race`.
3. **Dotazy** — filtr, řazení, stránkování. Nejvíc hraničních případů celého projektu.
4. **HTTP** — až když všechno pod tím prochází. Handler je pak tenký překlad.
5. **Provoz** — konfigurace, logy, shutdown, hardening (lekce 60).

Rozdělení práce s agentem podle protokolu z lekce 58: **spec a acceptance testy píšeš ty**
(kroky 1–3 jsou tvoje doména), agent smí psát implementaci proti testům, HTTP wiring
a testovací tabulky. Domain model si nikdy nenech vygenerovat jako první krok — je to
jediná část, kterou budeš měnit nejdráž.

### Normalizace patří na hranici

Klíčové rozhodnutí, které v capstone uděláš: **kde se data čistí**. Dvě možnosti:

```go
// A) normalizace uvnitř Validate — validace mění data, což nikdo nečeká
func (b *Bookmark) Validate() error { b.URL = normalize(b.URL); /* ... */ }

// B) normalizace v konstruktoru, Validate jen kontroluje — jasné a testovatelné
func New(id, rawURL, title string, tags []string, at time.Time) (Bookmark, error)
func (b Bookmark) Validate() error   // hodnotový receiver, nic nemění
```

Vyhrává B. `Validate` s hodnotovým receiverem nemůže nic změnit, takže se dá volat
kdekoli, i ve store jako pojistka. `New` je jediné místo, kde vzniká platná záložka —
a protože `Validate` navíc kontroluje, že URL **je** v normalizovaném tvaru, nedostaneš do
store nic, co by neprošlo konstruktorem.

Tenhle vzor („konstruktor normalizuje, validace jen ověřuje, invariant je kontrolovatelný
zvenčí") je v Go běžný a dobře se vysvětluje v review.

### Index a defenzivní kopie

In-memory store není `map[string]Bookmark` a hotovo. Vyhledávání podle tagu přes průchod
všemi položkami je O(n) a rozpadne se na prvním tisíci záložek, takže potřebuješ index:

```go
type Store struct {
	mu    sync.RWMutex
	items map[string]Bookmark
	byTag map[string]map[string]struct{} // tag -> množina ID
}
```

Index musí být **udržovaný při každé změně** — `Add` ho plní, `Delete` z něj maže a když
zůstane prázdná množina, maže i klíč. Osiřelý index je klasická chyba, kterou test
odhalí jen tehdy, když po `Delete` sáhneš na `ByTag`.

Druhá past je **aliasing slice**. `Bookmark.Tags` je slice, takže když ji uložíš tak, jak
přišla, volající do ní může sáhnout i potom:

```go
b := Bookmark{Tags: []string{"go"}}
store.Add(b)
b.Tags[0] = "podvrh"   // změnil obsah store, protože backing array je sdílené
```

Proto se při vstupu i výstupu dělá **hluboká kopie** — vzpomeň si na lekci 07. `RWMutex`
tě před tímhle neochrání: zámek chrání mapu, ne pole, na které se z ní ukazuje.

### Stabilní řazení a cursor

Stránkování stojí na tom, že **pořadí je totální**. `CreatedAt` sestupně nestačí; dvě
záložky mohou vzniknout ve stejnou nanosekundu (v testu skoro jistě). Bez tiebreaku ti
`sort.Slice` může mezi dvěma voláními vrátit jiné pořadí a cursor začne přeskakovat.

```go
sort.SliceStable(items, func(i, j int) bool {
	if !items[i].CreatedAt.Equal(items[j].CreatedAt) {
		return items[i].CreatedAt.After(items[j].CreatedAt)
	}
	return items[i].ID < items[j].ID   // tiebreak dělá pořadí totálním
})
```

Cursor pak je prostě **ID poslední položky předchozí stránky**: najdi ho ve výsledku
a pokračuj za ním. Neznámý cursor je chyba (`ErrInvalidCursor`), ne tichý začátek od
začátku — jinak si klient nikdy nevšimne, že mu vypršel.

A validace dotazu: záporný limit, limit nad strop, neznámé řazení a neplatný tag jsou
`ErrInvalidQuery`. Limit `0` znamená „nezadáno" a nahradí se výchozím. Tohle je hranice
systému, i když zrovna nemá HTTP — pravidlo z lekce 36 platí i mezi balíčky.

## Časté chyby

| Chyba | Proč vzniká | Jak to udělat správně |
|-------|-------------|------------------------|
| Normalizace uvnitř `Validate` | v Symfony to dělal validátor i normalizér naráz | konstruktor normalizuje, `Validate` jen ověřuje |
| Uložení slice tak, jak přišla | reflex „objekty jsou stejně reference" | hluboká kopie na vstupu i výstupu |
| Index se aktualizuje jen v `Add` | `Delete` se dopisuje později | test, který po `Delete` volá `ByTag` |
| Řazení bez tiebreaku | `CreatedAt` vypadá jako unikátní | druhotné řazení podle ID |
| Offset paginace | zvyk z `findBy($c, $o, $limit, $offset)` | cursor = ID poslední položky |
| `RWMutex` jako záruka bezpečí | zámek vypadá jako řešení všeho | zámek chrání mapu, ne data za pointery |

## Úkol

Pracuj v `exercise/`. Postupuj A → B → C, po každé části spusť test.

Píšeš jádro capstone v malém: doménu, store a hledání. Stejné rozhraní pak najdeš
v `projects/p05-capstone/`.

### A — rozcvička (~15 min)

1. `NormalizeURL(raw string) (string, error)` — pravidla přesně podle teorie výše:
   trim, povolené jen `http`/`https` (jinak `ErrInvalidURL`), malý host, odstranění
   výchozího portu, prázdný host je chyba, pryč s fragmentem, pryč s parametry `utm_*`
   (case-insensitive), zbytek query seřazený podle klíče, useknuté koncové lomítko.
   Funkce musí být **idempotentní**: `NormalizeURL(NormalizeURL(x)) == NormalizeURL(x)`.
2. `NormalizeTags(tags []string) ([]string, error)` — trim, malá písmena, zahození
   prázdných, deduplikace, abecední řazení. Tag smí obsahovat jen `a–z`, `0–9` a `-`,
   jinak `ErrInvalidTag`. Víc než `MaxTags` (po deduplikaci) je `ErrTooManyTags`.
3. `(Bookmark).Validate() error` — hodnotový receiver, nic nemění. Kontroluje: neprázdné
   `ID` (`ErrEmptyID`), URL **v normalizovaném tvaru** (`ErrInvalidURL`), neprázdný
   titulek (`ErrEmptyTitle`) do `MaxTitleLen` **run** (`ErrTitleTooLong`), tagy platné
   (`ErrInvalidTag`), bez duplicit (`ErrDuplicateTag`) a nanejvýš `MaxTags`
   (`ErrTooManyTags`).
4. `New(...) (Bookmark, error)` — normalizuje ID (trim), URL, titulek (trim) a tagy,
   pak zavolá `Validate`.

např. `NormalizeURL("HTTPS://Example.com/a/?utm_source=x")` → `"https://example.com/a"`

### B — jádro (~35 min)

In-memory store bezpečný pro souběžné použití, s indexem podle tagu.

- `NewStore() *Store` — připravené mapy.
- `Add(b Bookmark) error` — nejdřív `Validate` (neplatná záložka se **neuloží**), pak
  `ErrDuplicateID` při kolizi ID, jinak uložení a aktualizace indexu.
- `Get(id string) (Bookmark, error)` — kopie záložky, `ErrNotFound` pro neznámé ID.
- `Delete(id string) error` — smaže položku i všechny její záznamy v indexu (prázdný klíč
  z indexu zmizí), `ErrNotFound` pro neznámé ID.
- `Len() int`.
- `ByTag(tag string) []Bookmark` — tag se normalizuje stejně jako při ukládání
  (trim, malá písmena), výsledek je seřazený od nejnovější, při shodě času podle ID.
  Neznámý tag vrací prázdný výsledek, ne `nil` dereference.

Store nesmí sdílet `Tags` s volajícím ani na vstupu, ani na výstupu. Test to zkouší
a běží s `-race`.

např. `ByTag("go")` po Add b1…b3 → `[b2, b1]` (od nejnovější)

### C — rozšíření (~25 min)

`Search(q Query) (Page, error)`:

- **Validace:** `Limit < 0` nebo `> MaxLimit` → `ErrInvalidQuery`; neznámé `Sort` →
  `ErrInvalidQuery`; tag, který neprojde `NormalizeTags`, → `ErrInvalidQuery`;
  `Limit == 0` znamená `DefaultLimit`.
- **Filtr:** prázdné `Tags` znamená „bez omezení". `MatchAll == false` je OR,
  `MatchAll == true` je AND. `Text` je case-insensitive podřetězec v titulku.
- **Řazení:** `SortNewest` = `CreatedAt` sestupně, tiebreak `ID` vzestupně;
  `SortTitle` = titulek vzestupně (case-insensitive), tiebreak `ID`.
- **Stránkování:** `Cursor` je ID poslední položky předchozí stránky; když ve výsledku
  není, vrať `ErrInvalidCursor`. `NextCursor` je ID poslední vrácené položky, pokud ještě
  něco zbývá, jinak prázdný řetězec. `Total` je počet položek **před** stránkováním.

např. `Search({Tags:["http"]})` → IDs `[b4, b2]`

```bash
make lesson L=59
```

Až budeš hotový, porovnej se `solutions/` (spoiler). Pak si projdi
`projects/p05-capstone/SPEC.md` a `ACCEPTANCE.md` — capstone je tenhle model plus HTTP
vrstva, konfigurace a provoz.

## Ověření

Po dokončení úkolů spusť v Cursoru **`/go-deep-review`** a zadej třeba jen `59`. AI tě postupně projde body níže, doptá se a ověří pochopení — nestačí jen zelené testy.

- [ ] `make lesson L=59` prochází, včetně `make race L=59`
- [ ] Umíš vysvětlit, proč `Validate` nemá měnit data
- [ ] Umíš ukázat, kde by bez hluboké kopie vzniklo sdílení dat se store
- [ ] Umíš vysvětlit, proč řazení potřebuje tiebreak a co se stane bez něj
- [ ] Umíš vysvětlit, čím je cursor lepší než offset, a co dělat s neznámým cursorem
- [ ] Umíš říct, které části capstone píšeš ty a které smí psát agent

## AI režim

`TECH LEAD` — viz [docs/ai-playbook.md](../../docs/ai-playbook.md).

Doménu a acceptance testy vlastníš ty. Agent smí psát store a HTTP vrstvu proti tvým
testům. Každé předání zdůvodni — protokol z lekce 58 platí pro celý capstone.

## Další čtení

1. [net/url — dokumentace balíčku](https://pkg.go.dev/net/url)
2. [sort — dokumentace balíčku](https://pkg.go.dev/sort)
3. [Go blog — Go Slices: usage and internals](https://go.dev/blog/slices-intro)
4. [Effective Go — Concurrency](https://go.dev/doc/effective_go#concurrency)

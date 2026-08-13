# Lekce 25 — Routing: metody, wildcardy, PathValue

> **Čas:** ~70 min · **Fáze:** 3 — net/http a tooling · **AI režim:** `BOILERPLATE OK`

## Co budeš umět

- Zapsat routu včetně metody a parametru jako `"GET /items/{id}"` a přečíst hodnotu přes `r.PathValue`.
- Vysvětlit pravidla precedence vzorů a poznat dvojici vzorů, která shodí aplikaci při startu.
- Rozhodnout, kdy `http.ServeMux` stačí a kdy má smysl sáhnout po externím routeru.
- Načíst a zvalidovat query parametry tak, aby špatný vstup skončil na 400 a ne na 500.
- Postavit endpoint servírující soubory, který neumožní útok path traversal.

## Teorie

### Vzory ServeMux (od Go 1.22)

Do Go 1.21 uměl `ServeMux` jen prefixy a přesné shody. Od 1.22 má vzor tvar
`[METODA ][HOST]/cesta` a v cestě smí být wildcardy:

```go
mux := http.NewServeMux()

mux.HandleFunc("GET /items", listItems)          // přesná cesta, jen GET
mux.HandleFunc("POST /items", createItem)        // stejná cesta, jiná metoda
mux.HandleFunc("GET /items/{id}", getItem)       // {id} = právě jeden segment
mux.HandleFunc("DELETE /items/{id}", deleteItem)
mux.HandleFunc("GET /files/{path...}", serveFile) // {path...} = zbytek cesty včetně /
mux.HandleFunc("GET /{$}", index)                 // právě kořen, nic jiného
```

Čtyři pravidla, která si zapamatuj:

- `{name}` odpovídá **právě jednomu** segmentu a nesmí být prázdný.
- `{name...}` na konci vzoru odpovídá **zbytku cesty**, tedy i více segmentům.
- `{$}` znamená „tady cesta končí". Bez něj je `/` prefixový vzor nad úplně vším,
  takže by ti kořen chytal i `/tohle-neexistuje`.
- Vzor bez metody odpovídá všem metodám. Vzor s `GET` navíc automaticky odpovídá i `HEAD`.

Hodnoty wildcardů čteš přes `r.PathValue("id")`. Pro segment, který ve vzoru není,
vrátí prázdný řetězec — takže překlep ve jméně nezpůsobí paniku, ale tichou prázdnou
hodnotu. V testu si můžeš hodnotu nastavit ručně přes `r.SetPathValue("id", "42")`
a zavolat handler bez routeru.

### Precedence a konflikty

Pravidlo je jediné: **vyhrává specifičtější vzor.** Vzor A je specifičtější než B, pokud
A odpovídá striktní podmnožině požadavků, kterým odpovídá B. Pořadí registrace nehraje
roli — na rozdíl od většiny PHP routerů, kde vyhrává první zapsaná routa.

```go
mux.HandleFunc("/items/{id}", generic)
mux.HandleFunc("GET /items/latest", latest) // vyhraje pro GET /items/latest
```

Když ani jeden vzor není specifičtější, jde o **konflikt** a `ServeMux` při registraci
panikuje:

```go
mux.HandleFunc("GET /items/{id}", a)
mux.HandleFunc("/items/latest", b)
// panic: pattern "/items/latest" conflicts with "GET /items/{id}"
```

To je dobrá zpráva: nejednoznačný routing spadne při startu aplikace, ne v produkci na
podivně směrovaném požadavku.

Druhá vestavěná vymoženost: pokud cesta odpovídá nějakému vzoru, ale metoda ne, `ServeMux`
sám vrátí **405 Method Not Allowed** s hlavičkou `Allow`. `PUT /items/1` tedy dostane
`405, Allow: DELETE, GET` bez jediného řádku tvého kódu. V lekci 24 sis tuhle logiku psal
ručně — teď ji můžeš smazat.

Pozor na jednu věc, kterou lidé objevují až v produkci: pokud registruješ catch-all vzor
`/` (třeba kvůli JSON 404, jako v minulé lekci), sebere si všechny neobsloužené požadavky
včetně těch, které by jinak vedly na 405. Buď catch-all, nebo automatické 405 — obojí
naráz nedostaneš.

### Potřebuješ chi, gorilla nebo gin?

Pro drtivou většinu služeb ne. Externí routery historicky řešily přesně ty tři věci,
které `ServeMux` od 1.22 umí: metody, parametry v cestě a rozumnou precedenci. Nulová
závislost navíc znamená nulové bezpečnostní záplaty, nulovou migraci mezi major verzemi
a handlery, které jsou pořád jen `http.Handler`.

Externí router zvaž, když opravdu potřebuješ:

- **regulární výrazy v cestě** (`{id:[0-9]+}`) — `ServeMux` je nemá,
- **skupiny rout se sdíleným middlewarem** a doslova stovky endpointů,
- **hromadné mountování subrouterů** třetích stran.

Ani jedno není důvod sáhnout po celém frameworku. Reflex „nejdřív framework, pak teprve
zjistím, co potřebuju" je nejdražší návyk, který si z PHP můžeš přinést.

### Query parametry

Query string není součástí routingu. `r.URL.Query()` vrací `url.Values`, což je
`map[string][]string` s pomocnou metodou `Get`, která vrátí první hodnotu nebo `""`.

```go
values := r.URL.Query()
raw := values.Get("limit") // "" když parametr chybí

limit := 0
if raw != "" {
	n, err := strconv.Atoi(raw)
	if err != nil || n < 1 {
		writeError(w, http.StatusBadRequest, "limit must be a positive integer")
		return
	}
	limit = n
}
```

Tři chyby, které v tomhle kódu potkáš nejčastěji:

```go
limit, _ := strconv.Atoi(values.Get("limit")) // ignorovaná chyba → "abc" tiše dá 0
if limit > 100 { limit = 100 }                // horní mez bez dolní → -1 projde
n, _ := strconv.Atoi(values.Get("page"))      // a pak page*size v SQL → hezký den
```

Parsování dej do samostatné funkce nad `url.Values`. Otestuješ ji bez HTTP, tabulkou
vstupů, a handler zůstane čitelný.

### Path traversal

Tohle je jediné místo v lekci, kde chyba není „vrátí to špatný status", ale „vydá to
`/etc/passwd`".

```go
// ŠPATNĚ — klasická díra
full := filepath.Join(root, r.PathValue("path"))
http.ServeFile(w, r, full)
```

`filepath.Join` **cestu vyčistí**, takže `root + "../../etc/passwd"` se poslušně vyhodnotí
na `/etc/passwd` a odešle se. Nikdy neskládej cestu k souboru přímo z uživatelského vstupu.

Bezpečná verze kontroluje vstup dvakrát — nejdřív segmenty, pak výsledek:

```go
func SafeJoin(root, rel string) (string, error) {
	if rel == "" || strings.HasPrefix(rel, "/") || filepath.IsAbs(rel) {
		return "", ErrInvalidPath
	}
	for _, segment := range strings.Split(rel, "/") {
		if segment == "" || segment == "." || segment == ".." {
			return "", ErrInvalidPath
		}
	}
	full := filepath.Join(root, filepath.FromSlash(rel))

	absRoot, _ := filepath.Abs(root)
	absFull, _ := filepath.Abs(full)
	if !strings.HasPrefix(absFull, absRoot+string(filepath.Separator)) {
		return "", ErrInvalidPath
	}
	return full, nil
}
```

Sama o sobě by stačila i druhá kontrola, ale první dává jasnou chybovou hlášku a nemusí
sahat na souborový systém. U bezpečnostních kontrol se redundance vyplácí.

Poznámka k testování: `ServeMux` normalizuje cestu a na neuklizenou cestu odpoví
přesměrováním (`/files/../secret.txt` → `301` na `/secret.txt`). Přes HTTP klienta se ti
tedy „hloupý" traversal do handleru vůbec nedostane. To ale **není** tvoje obrana —
proxy před tebou může poslat cestu jinak zakódovanou. V testech proto handler voláme přímo
a hodnotu wildcardu nastavujeme přes `r.SetPathValue`, abychom otestovali handler, ne mux.

Pokud jen servíruješ statický adresář a nepotřebuješ vlastní logiku, použij
`http.FileServer(http.Dir(root))` nebo `http.FileServerFS` — mají tuhle ochranu uvnitř.

## Rozdíly proti PHP

V Symfony je routa atribut nad metodou controlleru a parametry ti přijdou jako argumenty,
už přetypované a případně i převedené na entitu.

```php
#[Route('/items/{id}', name: 'item_show', methods: ['GET'], requirements: ['id' => '\d+'])]
public function show(int $id): JsonResponse { /* ... */ }
```

Go 1.22 dostal do `http.ServeMux` metody a wildcardy, takže totéž vypadá takhle:

```go
mux.HandleFunc("GET /items/{id}", func(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id") // vždy string, žádná konverze ani validace
	// ...
})
```

Co se mění v uvažování: **router ti nic nekonvertuje a nic nevaliduje.** Neexistuje
`requirements`, neexistuje ParamConverter, neexistuje `int $id` v signatuře. Dostaneš
řetězec ze segmentu cesty a jsi za něj zodpovědný ty. Znělo by to jako krok zpět, ale
je to poctivější: validace je vidět v kódu handleru, ne schovaná v atributu, a chybová
odpověď má tvar, který jsi navrhl ty.

## Časté chyby

| Chyba | Proč vzniká | Jak to udělat správně |
|-------|-------------|------------------------|
| `mux.HandleFunc("/items/{id}", ...)` bez metody | zvyk, že routu řeší atribut a metody se filtrují jinde | uveď metodu ve vzoru, 405 pak řeší mux |
| Očekávání, že vyhraje první zaregistrovaná routa | tak to dělá Symfony i většina PHP routerů | vyhrává specifičtější vzor, konflikt = panika při startu |
| `strconv.Atoi(...)` s ignorovanou chybou | v PHP by `(int)"abc"` dalo 0 a jelo se dál | vždy zkontroluj chybu a vrať 400 |
| `filepath.Join(root, userInput)` | vypadá to bezpečně, protože Join cestu „uklidí" | validuj segmenty a ověř prefix výsledku |
| Catch-all `/` vedle vzorů s metodami | snaha mít JSON 404 | rozhodni se: buď catch-all, nebo vestavěné 405 |
| Sáhnutí po chi/gin na první routě | framework-first reflex | `ServeMux` a přidej závislost, až narazíš |

## AI kvíz

Po přečtení teorie spusť v Cursoru **`/go-deep-quiz 25`**. AI tě ~5 minut prověří mentální model (ne hotové cvičení). Slabiny si uloží do [`GAPS.md`](../../GAPS.md).

## Úkol

Pracuj v `exercise/`. `Store`, `WriteJSON`, `FilesHandler` a `FilesRouter` jsou hotové —
soustřeď se na body níže. Stupně jdou od jednodušších ke složitějším; po každém stupni
spusť review, než jdeš dál.

### Jednoduchý

Doplň: `ParseListQuery`

```bash
make lesson L=25 PART=1
```

Pak **`/go-deep-review 25 easy`**.

### Střední

Implementuj: `ItemsRouter`

```bash
make lesson L=25 PART=2
```

Pak **`/go-deep-review 25 medium`**.

### Obtížný

Oprav: `SafeJoin` (kód je záměrně vadný — filepath.Join bez validace segmentů)

```bash
make lesson L=25 PART=3
```

Pak **`/go-deep-review 25 hard`**.

Až budou stupně hotové, porovnej se `solutions/` (spoiler).

## Závěrečné otázky

Spusť **`/go-deep-review 25 final`**. AI projde body níže, doptá se a ověří pochopení. Celé cvičení ověří `make lesson L=25` (+ `make race L=25`, pokud to lekce vyžaduje).

- [ ] Umíš vysvětlit rozdíl mezi `{id}`, `{path...}` a `{$}`
- [ ] Umíš vysvětlit, proč pořadí registrace vzorů nehraje roli a kdy mux panikuje
- [ ] Umíš vysvětlit, kdy `ServeMux` sám vrátí 405 a proč to catch-all `/` vypne
- [ ] Umíš vyjmenovat dva důvody, kdy má externí router pořád smysl
- [ ] Umíš vysvětlit, proč `filepath.Join(root, vstup)` není bezpečný

## AI režim

`BOILERPLATE OK` — viz [docs/ai-playbook.md](../../docs/ai-playbook.md).

AI ti může vygenerovat tabulku testovacích případů pro `ParseListQuery`. Kontrakt rout,
status kódy a bezpečnostní kontrolu v `SafeJoin` napiš a projdi sám — tohle je přesně ta
kategorie kódu, kde vygenerovaná „skoro správná" verze vypadá nejlíp a bolí nejvíc.

## Další čtení

1. [Go blog — Routing Enhancements for Go 1.22](https://go.dev/blog/routing-enhancements)
2. [pkg.go.dev — http.ServeMux](https://pkg.go.dev/net/http#ServeMux)
3. [pkg.go.dev — Request.PathValue](https://pkg.go.dev/net/http#Request.PathValue)
4. [pkg.go.dev — net/url.Values](https://pkg.go.dev/net/url#Values)

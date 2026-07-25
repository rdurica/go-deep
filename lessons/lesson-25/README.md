# Lekce 25 — Routing: metody, wildcardy, PathValue

> **Čas:** ~90 min · **Fáze:** 3 — net/http a tooling · **AI režim:** `BOILERPLATE OK`

## Co budeš umět

- Zapsat routu včetně metody a parametru jako `"GET /items/{id}"` a přečíst hodnotu přes `r.PathValue`.
- Vysvětlit pravidla precedence vzorů a poznat dvojici vzorů, která shodí aplikaci při startu.
- Rozhodnout, kdy `http.ServeMux` stačí a kdy má smysl sáhnout po externím routeru.
- Načíst a zvalidovat query parametry tak, aby špatný vstup skončil na 400 a ne na 500.
- Postavit endpoint servírující soubory, který neumožní útok path traversal.

## PHP → Go most

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

## Časté chyby

| Chyba | Proč vzniká | Jak to udělat správně |
|-------|-------------|------------------------|
| `mux.HandleFunc("/items/{id}", ...)` bez metody | zvyk, že routu řeší atribut a metody se filtrují jinde | uveď metodu ve vzoru, 405 pak řeší mux |
| Očekávání, že vyhraje první zaregistrovaná routa | tak to dělá Symfony i většina PHP routerů | vyhrává specifičtější vzor, konflikt = panika při startu |
| `strconv.Atoi(...)` s ignorovanou chybou | v PHP by `(int)"abc"` dalo 0 a jelo se dál | vždy zkontroluj chybu a vrať 400 |
| `filepath.Join(root, userInput)` | vypadá to bezpečně, protože Join cestu „uklidí" | validuj segmenty a ověř prefix výsledku |
| Catch-all `/` vedle vzorů s metodami | snaha mít JSON 404 | rozhodni se: buď catch-all, nebo vestavěné 405 |
| Sáhnutí po chi/gin na první routě | framework-first reflex | `ServeMux` a přidej závislost, až narazíš |

## Úkol

Pracuj v `exercise/`. `Store` i `WriteJSON` už jsou hotové — soustřeď se na routing.
Chybové odpovědi mají tvar `ErrorResponse`.

### A — rozcvička (~10 min)

`ItemsRouter(store *Store) http.Handler` postavený na `http.ServeMux`:

- `GET /{$}` → **200** a `ServiceInfo{Service: "items"}`. Cokoli jiného než přesný kořen
  musí skončit na 404, takže `{$}` je povinné.
- `GET /items/{id}` → **200** a položku ze `Store.Get`, nebo **404** s `ErrorResponse`,
  pokud neexistuje.

### B — jádro (~35 min)

1. `ParseListQuery(values url.Values) (limit int, q string, err error)`:
   chybějící nebo prázdný `limit` → `0` (bez omezení); jinak celé číslo ≥ 1.
   Nečíselná, nulová i záporná hodnota → chyba (`ErrInvalidQuery`). `q` se ořízne
   od bílých znaků. Při chybě vracej nulové hodnoty.
2. Doplň do `ItemsRouter`:
   - `GET /items` → **200** a JSON **pole** položek v pořadí vložení. Parametry přes
     `ParseListQuery`; chyba → **400**. `q` filtruje podle podřetězce ve jménu
     case-insensitive, `limit` ořízne počet výsledků. Prázdný výsledek je `[]`, ne `null`.
   - `POST /items` → tělo `CreateItemRequest`. Rozbitý JSON i prázdné jméno (po ořezání)
     → **400**. Úspěch → **201**, hlavička `Location: /items/{id}` a tělo s vytvořenou
     položkou.
   - `DELETE /items/{id}` → **204** bez těla, neexistující ID → **404**.

Ověř si v testu, že `PUT /items/1` vrací 405 s hlavičkou `Allow` — a že jsi pro to
nenapsal ani řádek.

### C — rozšíření (~25 min)

1. `SafeJoin(root, rel string) (string, error)` — vrátí cestu k souboru pod `root`,
   nebo `ErrInvalidPath`. Odmítni prázdný `rel`, absolutní cestu a jakýkoli segment
   `""`, `"."` nebo `".."`. Nakonec ověř, že absolutní výsledek leží pod absolutním `root`.
2. `FilesHandler(root string) http.Handler` — přečte `r.PathValue("path")`, zavolá
   `SafeJoin` (chyba → **400**), ověří přes `os.Stat`, že cíl existuje a není adresář
   (jinak **404**), a soubor pošle přes `http.ServeFile`.
3. `FilesRouter(root string) http.Handler` — `ServeMux` se vzorem
   `GET /files/{path...}` napojeným na `FilesHandler`.

```bash
make lesson L=25
```

Až budeš hotový, porovnej se `solutions/` (spoiler).

## Ověření

- [ ] `make lesson L=25` prochází
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

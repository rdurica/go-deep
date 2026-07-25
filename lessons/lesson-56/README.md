# Lekce 56 — Spec-first, ADR a prompting pro Go

> **Čas:** ~90 min · **Fáze:** 7 — Inženýrství v době AI · **AI režim:** `TECH LEAD`

## Co budeš umět

- Napsat specifikaci, kterou model nemůže splnit „skoro" — akceptační kritéria, příklady a hraniční případy místo přání.
- Napsat acceptance testy **dřív** než prompt a používat je jako jediný zdroj pravdy při iteraci.
- Rozhodnout, kdy je rozhodnutí hodné ADR, a napsat ho tak, aby ho za rok někdo pochopil bez tebe.
- Sestavit promptovací kontext pro Go (verze, závislosti, error model, hranice balíčků) a poznat halucinované API dřív, než ho pustíš do repozitáře.

## PHP → Go most

V Symfony projektu bývá „specifikace" ticket a pár testů, které dopíšeš po implementaci.
Framework za tebe drží spoustu rozhodnutí: kde je controller, jak vypadá validace, co je
service. Když pak řekneš agentovi „přidej endpoint na záložky", má z konvencí bundle
dost kontextu, aby trefil něco použitelného.

```php
// Symfony: konvence nese půlku specifikace
#[Route('/bookmarks', methods: ['POST'])]
public function create(#[MapRequestPayload] CreateBookmark $dto): JsonResponse
```

V Go žádné konvence tohohle typu nejsou. `net/http` ti nenapoví, jestli chceš
`map[string]any`, nebo typovaný request; jestli chyby vracíš jako problem details, nebo
jako holý text; jestli je store interface u konzumenta, nebo v balíčku domény.

```go
// Go: rozhodnutí, která nikdo neudělá za tebe
type CreateRequest struct {
	URL   string   `json:"url"`
	Title string   `json:"title"`
	Tags  []string `json:"tags"`
}

func (s *Server) createBookmark(w http.ResponseWriter, r *http.Request) { /* ... */ }
```

Změna v uvažování: **prázdné místo po frameworku musí vyplnit spec, ne model.** Když ho
nevyplníš ty, vyplní ho agent tím, co viděl nejčastěji na GitHubu — a to bývá Go psané
jako Java. Spec je v Go dražší než v Symfony, protože nese víc informace. Zároveň se ti
mnohonásobně vrátí, protože přesně tahle informace je to, co model nemá.

## Teorie

### Spec jako akceptační kritéria, ne jako přání

Zadání typu „udělej robustní service pro záložky" je nesplnitelné, protože nemá kritérium
selhání. Použitelná spec má čtyři části:

1. **Akceptační kritéria** — pozorovatelné chování: „`POST /bookmarks` s URL bez schématu
   vrátí 400 a tělo `{"type":"…/invalid-url"}`".
2. **Příklady** — konkrétní dvojice vstup → výstup. Jeden dobrý příklad ušetří odstavec
   prózy a model ho použije jako testovací vektor.
3. **Hraniční případy** — prázdný vstup, `nil` slice, duplicitní ID, souběžné volání,
   příliš dlouhý titulek, URL s `utm_` parametry.
4. **Omezení** — verze Go, zákaz závislostí, error model, cílová struktura balíčků.

Zkouška kvality: **dokázal bys spec předat dvěma lidem, kteří spolu nemluví, a dostat
kompatibilní implementace?** Pokud ne, chybí v ní kritéria, ne slova.

### Acceptance testy před promptem

Tohle je nejlevnější trik celé fáze. Než napíšeš prompt, napiš test:

```go
func TestNormalizeURL(t *testing.T) {
	tests := map[string]string{
		"HTTPS://Example.com/a/?utm_source=x": "https://example.com/a",
		"http://example.com:80/":              "http://example.com",
		"https://example.com/a?b=2&a=1":       "https://example.com/a?a=1&b=2",
	}
	for in, want := range tests {
		got, err := bookmark.NormalizeURL(in)
		if err != nil || got != want {
			t.Errorf("NormalizeURL(%q) = (%q, %v), chci (%q, nil)", in, got, err, want)
		}
	}
}
```

Test dělá tři věci najednou: přesně definuje chování (i to, které bys prózou zapomněl
popsat), dává ti neúplatné kritérium hotovo a zároveň je to nejlepší prompt, jaký můžeš
modelu dát. Iterace pak vypadá „test padá takhle, oprav to", ne „ještě to není ono".

Pravidlo: **iteruj přes testy, ne přes popis.** Každé kolo, ve kterém upřesňuješ slovy,
co jsi mohl vyjádřit assertem, je promarněné kolo.

### ADR — architecture decision record

ADR je jednostránkový záznam jednoho rozhodnutí. Má čtyři části: **kontext** (co nás
tlačí), **rozhodnutí** (co děláme), **důsledky** (co tím platíme) a **alternativy** (co
jsme zvážili a proč ne). Formát je záměrně nudný a strojově čitelný:

```markdown
# 7. Použij stdlib router

- Status: Accepted
- Date: 2024-05-01

## Context

Potřebujeme routing s metodami a wildcardy.

## Decision

Použijeme net/http ServeMux se vzory (metody, wildcardy).

## Consequences

Méně závislostí. Musíme si napsat vlastní middleware chain.
```

ADR piš, když rozhodnutí splňuje aspoň jedno: je **drahé vrátit** (persistence, formát
API, hranice modulu), **omezuje ostatní** (error model, styl logů), nebo **vypadá divně
bez kontextu** (proč tady máme vlastní pool místo `database/sql`). Nepiš ADR na
„pojmenujeme balíček `bookmark`" — to je code review, ne architektura.

V éře agentů má ADR ještě jednu funkci: je to **kontext, který přiložíš k promptu**.
Model, který dostane tvá tři ADR, přestane navrhovat ORM, které jsi před měsícem odmítl.
Proto se vyplatí `Status` držet v definované množině a mít index se sekvenčními čísly —
je to malá databáze rozhodnutí, ne sbírka poznámek.

### Promptovací vzory pro Go

Do promptu patří to, co model **nemůže odvodit** ze zadání:

```
Piš idiomatický Go 1.26, stdlib first, bez web frameworku a bez nových závislostí.
Accept interfaces, return structs. Interface definuj u konzumenta.
Chyby: sentinel hodnoty + wrapping přes %w, žádná panika pro business chyby.
Balíčky podle domény, ne podle vrstvy. Žádné utils/common.
Testy table-driven, souběžný kód ověřený -race.
Nejdřív vysvětli hranice balíčků třemi body, teprve pak piš kód.
```

Poslední řádek je nejcennější. Když necháš model **napsat plán před kódem**, dostaneš
levnou příležitost ho zastavit dřív, než vygeneruje 300 řádků špatné struktury. Tři body
přečteš za deset sekund, diff za deset minut.

Do promptu naopak **nepatří** popis toho, co je v kódu vidět („funkce má vracet error"),
ani prosby o kvalitu („piš čistý kód", „buď důkladný"). Kvalitu neurčuje adjektivum,
určuje ji kritérium.

### Halucinované API

Model hladce vymyslí funkci, která ve stdlib není, protože „by tam logicky být měla":

```go
// tohle ve stdlib neexistuje, i když to vypadá naprosto normálně
u, _ := url.ParseStrict(raw)
s := strings.RemoveAll(text, "utm_")
d := time.ParseDurationString("5s")
```

Ověření je mechanické a rychlé — nikdy ho nedělej okem:

```bash
go build ./...            # nejlevnější detektor: neexistující funkce se nezkompiluje
go doc strings            # existuje v balíčku vůbec?
go doc net/url.URL.Query  # a má ten podpis, který model použil?
go vet ./...              # a nepoužil ho blbě?
```

Nebezpečnější varianta halucinace je API, které **existuje, ale znamená něco jiného** —
`context.WithValue` použité jako DI kontejner, `sync.Map` tam, kde stačí mutex,
`io.ReadAll` na tělo požadavku bez limitu. Tohle kompilátor nechytí; chytí to jen tvoje
znalost stdlib z fází 1–6 a review z lekce 57.

## Časté chyby

| Chyba | Proč vzniká | Jak to udělat správně |
|-------|-------------|------------------------|
| Spec bez kritéria selhání | zvyk psát tickety pro lidi, kteří doptají | každé kritérium formuluj jako pozorovatelné chování |
| Testy až po generování | v Symfony byl framework kostrou, tady není | acceptance test je součást zadání, ne úklid |
| Iterace přes prózu | „ještě to není ono" je rychlejší napsat | přidej assert a nech padat test |
| ADR na každou drobnost | reflex dokumentovat vše | ADR jen na drahá, omezující nebo nečekaná rozhodnutí |
| Prompt bez verze a zákazu závislostí | předpoklad, že model zná náš kontext | verze Go, „bez závislostí" a error model do každého promptu |
| Důvěra v hezky vypadající volání stdlib | kód vypadá idiomaticky | `go build`, `go doc`, `go vet` na každý vygenerovaný soubor |

## Úkol

Pracuj v `exercise/`. Postupuj A → B → C, po každé části spusť test.

Stavíš nástroje, které spec-first workflow drží při životě: generátor a parser ADR
a analyzátor kvality zadání.

### A — rozcvička (~10 min)

1. `Status` — implementuj `String()` (`"Unknown"`, `"Proposed"`, `"Accepted"`,
   `"Rejected"`, `"Superseded"`; hodnota mimo rozsah `"Unknown"`) a
   `ParseStatus(s string) (Status, error)` — case-insensitive, s oříznutím mezer, neznámý
   vstup vrací chybu obalující `ErrInvalidStatus`.
2. `Fold(s string) string` — malá písmena bez diakritiky (`"Akceptační Kritéria"` →
   `"akceptacni kriteria"`). Znaky bez mapování nech být.
3. `Slug(title string) string` — z `Fold` nech jen `[a-z0-9]`, ostatní skupiny znaků
   nahraď jedinou pomlčkou, bez pomlčky na začátku a na konci (`"Použij stdlib router"` →
   `"pouzij-stdlib-router"`, `"---"` → `""`).
4. `(ADR).Filename() string` → `"0007-pouzij-stdlib-router.md"`. Číslo je na čtyři místa
   doplněné nulami, delší číslo se nezkracuje. Prázdný slug nahraď `"adr"`.

### B — jádro (~35 min)

1. `(ADR).Render() string` — přesně tenhle markdown (bez odsazení, `\n` na konci):

```
# 7. Použij stdlib router

- Status: Accepted
- Date: 2024-05-01

## Context

<Context>

## Decision

<Decision>

## Consequences

<Consequences>
```

   Datum formátuj `2006-01-02`, texty sekcí ořízni o okolní bílé znaky.

2. `ParseADR(s string) (ADR, error)` — round-trip parser k `Render`. Musí zvládnout
   víceřádkové sekce a `\r\n`. Chyby (vždy obalené, aby fungovalo `errors.Is`):
   - hlavička nemá tvar `# <číslo>. <titulek>` → `ErrInvalidHeader` (i prázdný vstup),
   - neznámý status → `ErrInvalidStatus`,
   - datum v jiném formátu → `ErrInvalidDate`,
   - chybí `- Status:`, `- Date:` nebo některá ze tří sekcí (nebo je prázdná) →
     `ErrMissingSection`.

3. `Index(adrs []ADR) string` — markdown tabulka se sloupci `Číslo | Titulek | Status |
   Datum`, seřazená podle čísla, při shodě podle titulku. Prázdný vstup vrací
   `"_Žádné ADR._\n"`. Pokud se nějaké číslo opakuje, přidej za tabulku prázdný řádek
   a pro každé duplicitní číslo vzestupně řádek
   `> pozor: duplicitní číslo 7 (2×)`.

### C — rozšíření (~25 min)

Analyzátor zadání. `SpecCheck` drží pravidla, `Rule` má `ID`, klíčová slova, závažnost
a zprávu.

1. `Severity.String()` → `"INFO"`, `"WARN"`, `"ERROR"`.
2. `DefaultSpecCheck() SpecCheck` — pravidla v tomto pořadí:

| ID | Klíčová slova (stačí jedno) | Závažnost | Zpráva |
|----|------------------------------|-----------|--------|
| `acceptance` | `akceptační kritéria`, `acceptance criteria`, `kritéria přijetí` | ERROR | `chybí akceptační kritéria` |
| `edge-cases` | `hraniční případ`, `edge case` | ERROR | `chybí hraniční případy` |
| `errors` | `chybový stav`, `chybové stavy`, `error handling` | ERROR | `chybí popis chybových stavů` |
| `go-version` | `go 1.` | WARN | `chybí cílová verze Go` |
| `deps` | `bez závislostí`, `žádné závislosti`, `pouze stdlib`, `stdlib only` | WARN | `chybí pravidlo o závislostech` |

3. `(SpecCheck).Check(spec string) []Finding` — pro každé pravidlo, jehož žádné klíčové
   slovo se v textu nevyskytuje, vrať nález. Porovnávej přes `Fold`, takže na velikosti
   písmen ani diakritice nezáleží. Pořadí nálezů = pořadí pravidel. Bez pravidel žádné
   nálezy.
4. `CheckSpec(spec string) []Finding` — zkratka nad `DefaultSpecCheck()`.

```bash
make lesson L=56
```

Až budeš hotový, porovnej se `solutions/` (spoiler).

## Ověření

- [ ] `make lesson L=56` prochází
- [ ] Umíš vysvětlit, proč se acceptance test píše před promptem, a ne po něm
- [ ] Umíš rozhodnout, které ze tří konkrétních rozhodnutí zaslouží ADR
- [ ] Umíš vyjmenovat pět věcí, které patří do promptu pro Go, a dvě, které tam nepatří
- [ ] Umíš popsat dva způsoby, jak ověřit, že model nevymyslel neexistující API
- [ ] Umíš říct, čím se liší halucinované API od špatně použitého existujícího API

## AI režim

`TECH LEAD` — viz [docs/ai-playbook.md](../../docs/ai-playbook.md).

Ty vlastníš spec, acceptance testy a ADR. Agent smí navrhovat implementaci a alternativy
k rozhodnutí; ADR ale podepisuješ ty, protože důsledky ponese tvůj tým.

## Další čtení

1. [Go Code Review Comments](https://go.dev/wiki/CodeReviewComments)
2. [Effective Go](https://go.dev/doc/effective_go)
3. [go doc — dokumentace nástroje](https://pkg.go.dev/cmd/doc)
4. [Go blog — Error handling and Go](https://go.dev/blog/error-handling-and-go)

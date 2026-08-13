# Lekce 38 — Projekt P03: hexagonální služba

> **Čas:** ~70 min · **Fáze:** 4 — Architektura v Go · **AI režim:** `BOILERPLATE OK`

## Co budeš umět

- Postavit službu v pořadí doména → porty → fake adaptér → testy → HTTP → wiring a vysvětlit, proč právě v tomhle.
- Rozhodnout, kde má bydlet interface, a obhájit to pravidlem „interface u konzumenta".
- Ověřit hranici balíčků strojově, ne slibem v README.
- Napsat akceptační test proti HTTP API, který nezná vnitřek služby.
- Poznat, kdy je abstrakce přepálená, a smazat ji dřív, než se rozroste.

## Teorie

### Začni od domény, protože nemá závislosti

Doména je jediná část služby, kterou nemůžeš odvodit z ničeho jiného. HTTP API se dá
odvodit z domény, schéma databáze se dá odvodit z domény, ale obráceně to nefunguje —
z tabulky nikdo nevyčte, že odeslat lze jen zaplacenou objednávku.

Praktický důsledek: doménový balíček píšeš první a jeho test nepotřebuje **nic**.
Žádný server, žádnou databázi, žádný fixture soubor. Když zjistíš, že doménový test
potřebuje `httptest`, máš chybu v návrhu, ne v testu.

Agregát drží svá pravidla tak, že nedovolí špatný stav vzniknout:

```go
type Order struct {
	ID     string
	Lines  []Line
	Status Status
}

// transition je JEDINÉ místo, kde se stav mění. Celý stavový automat je pak
// vidět na pěti řádcích, ne rozeseto po aplikaci.
func (o Order) transition(to Status, from ...Status) (Order, error) {
	for _, allowed := range from {
		if o.Status == allowed {
			o.Status = to    // o je kopie, originál voláním netrpí
			return o, nil
		}
	}
	return o, fmt.Errorf("%w: ze stavu %s nelze do %s", ErrInvalidTransition, o.Status, to)
}
```

Klíčové rozhodnutí je **hodnotový receiver**. `o` je kopie, takže přechod nemůže
původní objednávku zmutovat: `o.Pay()` vrací novou hodnotu a když se výsledek nikam
nepřiřadí, nestalo se nic. Polovičně provedený přechod je horší než žádný a s hodnotovou
semantikou nemůže vzniknout ani při chybě uprostřed metody.

Pole jsou exportovaná, což se v jiných jazycích považuje za hřích. Tady je to únosné
právě proto, že hodnota je nesdílená kopie: kdo si `Status` přepíše, poškodí jen svůj
výtisk, ne stav ve úložišti. Jedinou branou do systému zůstává konstruktor `New`, který
invarianty ověří a slice položek si **zkopíruje** — bez té kopie by změna vstupního
slice u volajícího prolezla dovnitř a validaci obešla.

Čas se do domény nikdy nebere z `time.Now()`, ale předává jako parametr nebo port.
Jinak dostaneš test, který jednou za rok spadne na přechodu letního času.

### Nakresli si porty

Port je díra ve zdi hexagonu: pojmenovaná potřeba, ne technologie. Nakreslit si je
znamená napsat na papír dvě otázky:

1. Co potřebuje doména od okolí? (výstupní porty: úložiště, odesílání e-mailu, hodiny)
2. Co okolí potřebuje od domény? (vstupní porty: případy užití, které volá HTTP)

Výstupní port se v Go píše jako interface **u konzumenta**:

```go
// v balíčku app, vedle Service, která ho volá
type Repository interface {
	Save(ctx context.Context, o order.Order) error
	Find(ctx context.Context, id string) (order.Order, error)
}

type IDGen interface {
	NewID() string
}
```

Proč ne v balíčku `memstore` u implementace? Protože pak by aplikační vrstva musela
importovat persistenci, aby uměla pojmenovat svou vlastní potřebu — a šipka závislosti
by mířila ven. Adaptér port nikde nezmiňuje; splňuje ho implicitně a hlídá si to jedním
řádkem:

```go
var _ app.Repository = (*memstore.Repository)(nil)
```

Dvě metody stačí. `Create` a `Update` zvlášť by byl závazek pro každý budoucí adaptér
a zatím ho nikdo nepotřebuje; kdo bude chtít rozlišit vložení od přepsání, přidá si
metodu, až bude mít volajícího.

`IDGen` je port ze stejného důvodu jako hodiny: generování identifikátoru je efekt
z okolního světa. S fake generátorem vydávajícím `ord-1`, `ord-2` je test čitelný,
s UUID by se v něm nedalo nic tvrdit.

Vstupní port v malé službě obvykle nepotřebuješ jako interface vůbec. `*app.Service`
je konkrétní typ a HTTP handler ho bere přímo. Interface by tu nic nekoupil, jen přidal
soubor.

### Pořadí prací

Pořadí není libovolné. Každý krok stojí na hotovém předchozím, takže se nikdy
nedostaneš do stavu „mám tři rozdělané vrstvy a nevím, kde to nefunguje":

| # | Krok | Kdy je hotový |
|---|---|---|
| 1 | Doména | testy pravidel a přechodů jsou zelené, balíček nic neimportuje |
| 2 | Port | interface má jen metody, které někdo skutečně volá |
| 3 | Fake adaptér | in-memory implementace portu, pár desítek řádků |
| 4 | Testy služby | případy užití proti fake adaptéru, pořád bez HTTP |
| 5 | HTTP adaptér | mapování požadavků a chyb na statusy |
| 6 | Wiring | `main.go`: konfigurace, výběr adaptérů, shutdown |

Kroky 1–4 nemají žádnou vazbu na okolní svět, takže jejich testy běží v milisekundách
a nikdy nejsou flaky. To je hlavní praktický zisk hexagonální architektury — ne diagram,
ale rychlost zpětné vazby.

### Jak se pozná, že hranice drží

Slovní dohoda vydrží do prvního spěchu. Napiš si test:

```go
func TestDomainDoesNotImportAdapters(t *testing.T) {
	fset := token.NewFileSet()
	bezTestu := func(fi fs.FileInfo) bool {
		return !strings.HasSuffix(fi.Name(), "_test.go")
	}
	pkgs, err := parser.ParseDir(fset, "order", bezTestu, parser.ImportsOnly)
	if err != nil {
		t.Fatalf("parsování balíčku selhalo: %v", err)
	}
	souboru := 0
	for _, pkg := range pkgs {
		for name, file := range pkg.Files {
			souboru++
			for _, spec := range file.Imports {
				path, _ := strconv.Unquote(spec.Path.Value)
				if path == "net/http" || path == "encoding/json" {
					t.Errorf("%s importuje %q — doména nesmí znát transport", name, path)
				}
			}
		}
	}
	if souboru == 0 {
		t.Error("neprošel se žádný soubor")
	}
}
```

`parser.ImportsOnly` přečte jen hlavičku souboru, takže je to rychlé. Filtr vynechává
`_test.go`, protože testy balíčku klidně `net/http` importovat mohou — hranici porušuje
až produkční kód.

Pozor na past falešně zeleného testu: kdyby `ParseDir` nenašel žádný soubor, test
projde a nic nezkontroluje. Proto se počítá, kolik souborů se skutečně prošlo.

### Acceptance testy proti HTTP API

Doménový test ověřuje pravidlo. Acceptance test ověřuje **slib daný klientovi**: že
`POST /orders` vrátí 201 s hlavičkou `Location` a že `ship` na nezaplacené objednávce
vrátí 409. Nesmí sahat do vnitřku:

```go
h := httpapi.NewHandler(app.NewService(memstore.New(), stubIDs{}))

rec := httptest.NewRecorder()
h.ServeHTTP(rec, httptest.NewRequest("POST", "/orders", body))
```

Pro handler stačí `httptest.NewRecorder` — žádný socket, žádný port, žádné čekání.
Když potřebuješ opravdového klienta (redirecty, timeouty, TLS), použij
`httptest.NewServer`, který si vezme volný port od operačního systému. Pevný port
v testu je záruka náhodného selhání na CI.

Generátor ID i hodiny se injektují i tady, takže odpověď má předvídatelné `id`
a `placed_at`.

### Jak nepřepálit abstrakci

Hexagonální architektura má jednu spolehlivou nemoc: člověk si oblíbí vrstvy a začne
je přidávat i tam, kde nic neřeší. Test na přepálenou abstrakci má tři otázky:

1. **Má ten interface víc než jednu implementaci** (fake se počítá)? Pokud ne a ani
   se nechystá, smaž ho a používej konkrétní typ.
2. **Přidala ta vrstva pravidlo, nebo jen přeposílá volání?** `func (s *Service) Get(id) { return s.repo.Get(id) }` není vrstva, je to daň.
3. **Dokázal bys tu vrstvu vysvětlit juniorovi za třicet vteřin?** Když ne, je to
   ceremoniál.

Konkrétně v Go: nedělej balíčky `service/`, `repository/`, `dto/`. To je Symfony
adresářová struktura přenesená do jazyka, kde balíček znamená hranici viditelnosti,
ne složku. Balíček se jmenuje podle domény (`order`), ne podle role.

## Rozdíly proti PHP

Symfony projekt začínáš tím, že vygeneruješ entitu. Doctrine anotace, repository třída,
controller, `services.yaml`. Doména vzniká jako vedlejší produkt persistence:

```php
#[ORM\Entity(repositoryClass: OrderRepository::class)]
class Order
{
    #[ORM\Column(type: 'string')]
    private string $status = 'draft';   // stav je sloupec, ne pravidlo
}
```

V Go se začíná z opačného konce — od pravidel, bez jediné závislosti:

```go
type Order struct {
	Status Status
}

func (o Order) Ship() (Order, error) {
	return o.transition(StatusShipped, StatusPaid) // pravidlo, ne sloupec
}
```

Co se mění v uvažování: přestaneš se ptát „jak to uložím" a začneš se ptát „co je
pravda o objednávce". Uložení je detail, který přijde na řadu jako třetí a dá se
vyměnit. Když někdo v code review řekne „tohle je jen anemický DTO se settery",
mluví přesně o tomhle rozdílu.

## Časté chyby

| Chyba | Proč vzniká | Jak to udělat správně |
|-------|-------------|------------------------|
| Interface v balíčku implementace | zvyk na `OrderRepositoryInterface` | interface patří ke konzumentovi |
| `time.Now()` uvnitř agregátu | je to nejkratší cesta | čas jako parametr nebo `func() time.Time` |
| Fake adaptér ukládá ukazatel | mapa ukazatelů je pohodlná | ukládej i vracej kopii, jako to dělá databáze |
| Mutující přechod na `*Order` | „vždyť to mění stav" | hodnotový receiver vracející novou hodnotu |
| Konstruktor si nezkopíruje slice | jeden řádek navíc | `copy` do vlastního slice |
| JSON tagy na doménovém typu | ať se to hned serializuje | DTO v adaptéru, doména bez tagů |
| Balíčky `service/`, `dto/` | Symfony struktura | balíček podle domény |
| Doménový test přes HTTP | „testuju to celé" | doména proti fake portu, HTTP zvlášť |

## AI kvíz

Po přečtení teorie spusť v Cursoru **`/go-deep-quiz 38`**. AI tě ~5 minut prověří mentální model (ne hotové cvičení). Slabiny si uloží do [`GAPS.md`](../../GAPS.md).

## Úkol

Pracuj v `exercise/httpapi/`. Balíčky `order/`, `app/` a `memstore/` jsou hotové —
ukazují hexagon bez opakování projektu P03. Ty píšeš **jen HTTP adaptér**.

### Jednoduchý

Oprav: `writeJSON` (tělo před `WriteHeader`)

```bash
make lesson L=38 PART=1
```

Pak **`/go-deep-review 38 easy`**.

### Střední

Doplň: `writeError` (mapování chyb na statusy na jednom místě)

```bash
make lesson L=38 PART=2
```

Pak **`/go-deep-review 38 medium`**.

### Obtížný

Implementuj: `NewHandler` (router + handlery use-casů)

```bash
make lesson L=38 PART=3
```

Pak **`/go-deep-review 38 hard`**.

Až budou stupně hotové, porovnej se `solutions/` (spoiler).

## Projekt P03

Cvičení je rozcvička. Vlastní projekt je v
[`projects/p03-hex-service/`](../../projects/p03-hex-service/ACCEPTANCE.md) — stejná
architektura, ale s `Money`, `internal/`, porty `Clock` a `IDGen`, strukturovaným
logováním přes `log/slog`, konfigurací z prostředí, endpointy `/healthz` a `/readyz`,
graceful shutdownem a testem hranice balíčků přes `go/parser`.
Akceptační kritéria jsou v `ACCEPTANCE.md`; projdi je bod po bodu, než projekt prohlásíš
za hotový.

## Závěrečné otázky

Spusť **`/go-deep-review 38 final`**. AI projde body níže, doptá se a ověří pochopení. Celé cvičení ověří `make lesson L=38` (+ `make race L=38`, pokud to lekce vyžaduje).

- [ ] Umíš vysvětlit, proč port patří k `Service` a ne k adaptéru
- [ ] Umíš vysvětlit, proč fake adaptér ukládá kopii
- [ ] Umíš vysvětlit, proč se čas injektuje místo `time.Now()` v doméně
- [ ] Umíš na svém kódu ukázat abstrakci, kterou bys smazal, a říct proč

## AI režim

`BOILERPLATE OK` — viz [docs/ai-playbook.md](../../docs/ai-playbook.md).
DTO, JSON tagy a scaffolding tabulkových testů si nech vygenerovat. Hranice balíčků,
error model a mapování statusů vlastníš ty.

## Další čtení

1. [Go blog — Organizing a Go module](https://go.dev/doc/modules/layout)
2. [Go Code Review Comments — Interfaces](https://go.dev/wiki/CodeReviewComments#interfaces)
3. [pkg.go.dev — go/parser](https://pkg.go.dev/go/parser)
4. [pkg.go.dev — net/http ServeMux (vzory metod a wildcardů)](https://pkg.go.dev/net/http#ServeMux)

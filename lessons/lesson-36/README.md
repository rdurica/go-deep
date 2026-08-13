# Lekce 36 — Validace na hranici

> **Čas:** ~70 min · **Fáze:** 4 — Architektura v Go · **AI režim:** `BOILERPLATE OK`

## Co budeš umět

- Rozhodnout, jestli je konkrétní pravidlo validace vstupu, nebo doménový invariant, a podle toho ho umístit.
- Napsat konstruktor ve stylu *parse, don't validate*, po kterém už nevalidní hodnota nemůže existovat.
- Vybrat mezi fail-fast a sběrem všech chyb podle toho, kdo je konzumentem API.
- Navrhnout chybovou odpověď podle RFC 7807 a namapovat doménové chyby na HTTP statusy.
- Vysvětlit, proč v Go není anotační validátor a proč to není chudoba, ale volba.

## Teorie

### Kde validovat

Máš tři kandidáty a každý dělá něco jiného.

**Na hranici systému** (HTTP handler, CLI parser, konzument fronty) přichází neznámý
vstup z vnějšího světa. Tady se rozhoduje, jestli data vůbec pustíš dovnitř.

**V doméně.** Doména nevaliduje — doména *neumožňuje* vzniknout nevalidnímu stavu.
Když má `Order.Ship()` fungovat jen na zaplacené objednávce, není to validace vstupu,
ale invariant: `Ship()` na nezaplacené objednávce vrátí chybu vždy, i když ho zavoláš
z testu nebo z migračního skriptu.

**V databázi.** `NOT NULL` a `UNIQUE` jsou poslední záchranná síť proti závodu dvou
požadavků, ne validace. Kdo se na ně spolehne jako na primární kontrolu, ukáže
uživateli `pq: duplicate key value violates unique constraint`.

Praktické pravidlo: **formát a rozsah na hranici, invariant v konstruktoru typu,
integritu v databázi.** „E-mail musí obsahovat zavináč" je hranice. „Objednávka nesmí
mít nulovou cenu" je invariant. „Dva uživatelé nesmí mít stejný e-mail" je databáze
(a handler ten konflikt přeloží na 409).

### Parse, don't validate

Nejčastější tvar validace, který si člověk přinese z PHP:

```go
// ŠPATNĚ — po zavolání pořád držíš obyčejný string
func IsValidEmail(s string) bool { /* … */ }

if !IsValidEmail(input) {
	return errors.New("neplatný e-mail")
}
sendWelcome(input) // sendWelcome bere string, takže musí věřit volajícímu
```

Problém: `IsValidEmail` vrátila `true` a ta informace se **okamžitě zahodila**.
`sendWelcome` dostane `string`, o kterém nic neví, a jestli je někdo o tři vrstvy
níž zvaliduje znovu, je otázka disciplíny.

Správný tvar vrátí typ, který nese důkaz:

```go
type Email struct {
	value string // neexportované: mimo balíček nejde naplnit
}

func ParseEmail(s string) (Email, error) {
	s = strings.ToLower(strings.TrimSpace(s))
	if s == "" {
		return Email{}, ErrEmptyEmail
	}
	local, domain, found := strings.Cut(s, "@")
	if !found || local == "" || !strings.Contains(domain, ".") {
		return Email{}, fmt.Errorf("%w: %q", ErrInvalidEmail, s)
	}
	return Email{value: s}, nil
}

func (e Email) String() string { return e.value }
func (e Email) IsZero() bool   { return e.value == "" }

func sendWelcome(to Email) { /* nic nekontroluje, nemá co */ }
```

Čtyři věci, které jsi dostal zdarma:

1. **Normalizace na jednom místě.** `Radek@Example.COM` a `radek@example.com` jsou po
   parsování stejná hodnota. Struktura s jediným porovnatelným polem zůstává
   porovnatelná přes `==`, takže i klíč v mapě sedí.
2. **Kompilátor kontroluje volání.** `sendWelcome(nejakyString)` se nepřeloží.
3. **Nejde to obejít konverzí.** Kdyby byl typ `type Email string`, přeložil by se
   `Email("evidentní nesmysl")`. S neexportovaným polem vede jediná cesta
   k neprázdnému `Email` přes `ParseEmail`.
4. **Nulová hodnota je poznatelná.** `Email{}` vzniknout může, ale `IsZero` to řekne.

### Fail-fast versus sběr chyb

`ParseEmail` je fail-fast: první problém, konec. To je správně pro *jednu hodnotu*.

Formulář je něco jiného. Když uživatel vyplní špatně tři pole a ty mu vrátíš jen první
chybu, čekají ho tři kola. Konzument API chce **všechny chyby najednou**, ve tvaru,
který se dá připnout k polím formuláře:

```go
// mapa pole → důvod zamítnutí
type ValidationErrors map[string]string

func (v ValidationErrors) Error() string { /* … */ }
```

Mapa je tady lepší volba než slice: pole je přirozený klíč, klient si podle něj chybu
připne k inputu a duplicitní hlášení k jednomu poli nemůže vzniknout. Cenou je, že
mapa nemá pořadí — a proto musí `Error()` klíče **seřadit**, jinak dostaneš pokaždé
jiný text a testy i logy začnou blikat.

Pozor na dvě věci. Za prvé, `ValidationErrors` implementuje `error`, takže ho můžeš
vrátit jako `error` a o vrstvu výš ho vytáhnout přes `errors.As` — i po obalení:

```go
var problems ValidationErrors
if errors.As(err, &problems) {
	WriteProblem(w, http.StatusUnprocessableEntity, "Neplatná data",
		"požadavek obsahuje neplatná pole", problems)
}
```

Za druhé, klasická past s typovaným nil:

```go
// ŠPATNĚ — problems je prázdná mapa, ale návratová hodnota je nenulový interface
func (r Req) Validate() error {
	problems := ValidationErrors{}
	// …nic se nepřidá…
	return problems        // err != nil !!
}
```

Interface je nil jen tehdy, když je nil typ i hodnota. Vždycky se explicitně zeptej na
`len(problems) == 0` a vrať doslova `nil`.

Zpráva u pole je věta pro člověka a smí se měnit, včetně překladu. Kdo potřebuje
i strojově čitelný kód (`required`, `format`, `range`), rozšíří hodnotu mapy na
strukturu — princip zůstane stejný. Lokalizace není práce validátoru: ten vrátí
identifikátor a překlad se dělá až při serializaci odpovědi podle `Accept-Language`.
V Symfony to za tebe dělá translator; v Go je to prostě mapa nebo
`golang.org/x/text/message`.

### Dekódování a validace jako jeden krok

Dekódovat tělo a zvalidovat ho jsou dvě věci, které se v handlerech opisují pořád
stejně. V Go 1.22 na to stačí generická funkce a jednometodový interface:

```go
type Validator interface {
	Validate() error
}

func DecodeAndValidate[T Validator](r *http.Request) (T, error) {
	var zero, v T
	dec := json.NewDecoder(io.LimitReader(r.Body, maxRequestBody))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&v); err != nil {
		return zero, fmt.Errorf("%w: %v", ErrMalformedJSON, err)
	}
	if err := v.Validate(); err != nil {
		return zero, err
	}
	return v, nil
}
```

Tři detaily, které tam nejsou náhodou. `io.LimitReader` je strop pro tělo — bez něj
pošle útočník nekonečný stream a dekodér ho poslušně nasype do paměti.
`DisallowUnknownFields` odmítne pole, která ve structu nejsou, takže překlep v názvu
neprojde tiše jako výchozí hodnota. A při chybě se vrací **nulová hodnota**, ne
částečně naplněná struktura, aby volající neměl co zneužít.

### Tvar chybové odpovědi

RFC 7807 (*problem details*) je jednoduchý standard: chybová odpověď je JSON objekt
s `type`, `title`, `status`, `detail` a Content-Type `application/problem+json`.

```go
type ProblemDetails struct {
	Type   string            `json:"type"`
	Title  string            `json:"title"`
	Status int               `json:"status"`
	Detail string            `json:"detail,omitempty"`
	Errors map[string]string `json:"errors,omitempty"`
}
```

`type` je URI dokumentace daného druhu chyby; když žádnou nemáš, patří tam
`about:blank`. `title` je krátký lidský titulek stejný pro celý druh chyby, `detail`
popisuje tenhle konkrétní výskyt. Pole `errors` standard nedefinuje — RFC 7807
rozšíření vysloveně povoluje a mapa pole → důvod je jeho nejběžnější podoba.
`omitempty` u `detail` a `errors` zajistí, že se do těla nedostanou prázdné hodnoty.

Dvě pravidla, na kterých se dá spálit:

- **Nikdy neposílej interní chybu do odpovědi.** `fmt.Fprintf(w, "%v", err)` ti jednou
  vypíše připojovací řetězec k databázi včetně uživatele. Interní detail patří do logu,
  ven jde neutrální věta a případně korelační ID.
- **`w.WriteHeader` se volá jednou** a až po `w.Header().Set(...)`. Jakmile jednou
  zapíšeš status, hlavičky se ignorují.

### Mapování chyb na statusy

Handler je jediné místo, které o HTTP vůbec ví. Doména vrací sentinel chyby, handler
z nich dělá čísla:

| Situace | Status | Proč |
|---|---|---|
| Tělo nejde dekódovat | 400 | klient poslal nesmysl na úrovni syntaxe |
| Data jsou syntakticky OK, ale nevalidní | 422 | rozumí se JSONu, ale ne obsahu |
| Konflikt se současným stavem | 409 | duplicitní e-mail, neplatný přechod stavu |
| Zdroj neexistuje | 404 | |
| Cokoli neočekávaného | 500 | a do logu celou chybu |

Rozhodování se dělá přes `errors.Is` / `errors.As`, nikdy přes porovnání textu. A vyplatí
se ho mít na **jednom místě**, ne rozeseté po handlerech:

```go
func ErrorHandler(err error) (int, ProblemDetails) {
	var problems ValidationErrors
	switch {
	case errors.As(err, &problems):
		return problem(422, "Neplatná data", "požadavek obsahuje neplatná pole", problems)
	case errors.Is(err, ErrMalformedJSON):
		return problem(400, "Neplatný požadavek", "tělo není platný JSON", nil)
	case errors.Is(err, ErrNotFound):
		return problem(404, "Nenalezeno", "zdroj neexistuje", nil)
	default:
		return problem(500, "Vnitřní chyba serveru", "požadavek se nepodařilo zpracovat", nil)
	}
}
```

Všimni si asymetrie: na vstupu je chyba se vším kontextem (`uložení objednávky: pq:
connection refused`), na výstupu jen to, co smí vidět klient. `default` větev je proto
holá pětistovka bez detailu — neznámá chyba je z definice ta, u které nevíš, co v ní je.

Handler se pak scvrkne na tři řádky a mapování se testuje samostatně, bez HTTP:

```go
req, err := DecodeAndValidate[CreateUserRequest](r)
if err != nil {
	WriteError(w, err) // uvnitř: ErrorHandler + WriteProblem
	return
}
```

## Rozdíly proti PHP

V Symfony je validace deklarativní. Přilepíš atributy na DTO a framework je za tebe
projde:

```php
final class CreateUserRequest
{
    #[Assert\NotBlank, Assert\Email]
    public string $email = '';

    #[Assert\Range(min: 13, max: 150)]
    public int $age = 0;
}

$violations = $validator->validate($request);   // magie přes reflexi
```

Go protějšek je obyčejná metoda. Žádná reflexe, žádný kontejner:

```go
type CreateUserRequest struct {
	Email string `json:"email"`
	Age   int    `json:"age"`
}

func (r CreateUserRequest) Validate() error {
	problems := ValidationErrors{}
	if _, err := ParseEmail(r.Email); err != nil {
		problems["email"] = "e-mail nemá platný tvar"
	}
	// …
	if len(problems) == 0 {
		return nil
	}
	return problems
}
```

Co se mění v uvažování: v Symfony je validace **vlastnost DTO**, kterou kdosi jinde
spustí. V Go je validace **krok v handleru**, který je vidět v kódu. Nikdo ti ji
nespustí za zády — ale taky ti ji nikdo nezapomene spustit v novém kontextu, protože
bez ní se ti do domény nedostane hodnota správného typu. A to je podstatnější rozdíl,
než se zdá.

## Časté chyby

| Chyba | Proč vzniká | Jak to udělat správně |
|-------|-------------|------------------------|
| `IsValid(s) bool` místo `Parse(s) (T, error)` | reflex z PHP validátoru | vrať silný typ, aby se důkaz nezahodil |
| `return problems` u prázdné mapy | typovaný nil v interface | `if len(problems) == 0 { return nil }` |
| Nedeterministický výpis chyb | iterace mapy má náhodné pořadí | klíče seřaď před složením textu |
| Dekódování bez stropu na tělo | `json.NewDecoder(r.Body)` je nejkratší | `io.LimitReader(r.Body, maxRequestBody)` |
| Validace uvnitř doménové metody | zvyk na fat service | invariant vynuť konstruktorem, vstup validuj na hranici |
| Text chyby posílaný klientovi | `fmt.Fprint(w, err)` je nejkratší cesta | detail do logu, ven neutrální zpráva |
| Porovnávání `err.Error() == "..."` | v PHP se chytá typ výjimky | `errors.Is` / `errors.As` nad sentinely |
| 400 na nevalidní data | nezná se rozdíl 400 vs 422 | 400 = nejde dekódovat, 422 = nedává smysl |

## AI kvíz

Po přečtení teorie spusť v Cursoru **`/go-deep-quiz 36`**. AI tě ~5 minut prověří mentální model (ne hotové cvičení). Slabiny si uloží do [`GAPS.md`](../../GAPS.md).

## Úkol

Pracuj v `exercise/`. Kontrakt je v komentáři nad funkcí.

### Jednoduchý

Oprav: `ValidationErrors.Error` (chybí řazení klíčů mapy)

```bash
make lesson L=36 PART=1
```

Pak **`/go-deep-review 36 easy`**.

### Střední

Implementuj: `ParseEmail`, `CreateUserRequest.Validate`

```bash
make lesson L=36 PART=2
```

Pak **`/go-deep-review 36 medium`**.

### Obtížný

Doplň: `WriteProblem`, `ErrorHandler`, `WriteError`

```bash
make lesson L=36 PART=3
```

Pak **`/go-deep-review 36 hard`**.

Až budou stupně hotové, porovnej se `solutions/` (spoiler).

## Závěrečné otázky

Spusť **`/go-deep-review 36 final`**. AI projde body níže, doptá se a ověří pochopení. Celé cvičení ověří `make lesson L=36` (+ `make race L=36`, pokud to lekce vyžaduje).

- [ ] Umíš vysvětlit rozdíl mezi validací vstupu a doménovým invariantem na vlastním příkladu
- [ ] Umíš vysvětlit, proč `IsValid(s) bool` zahazuje informaci a `Parse` ne
- [ ] Umíš vysvětlit, kdy vrátit 400 a kdy 422
- [ ] Umíš vysvětlit past s typovaným nil u `return problems`
- [ ] Víš, proč musí `Error()` nad mapou řadit klíče
- [ ] Umíš vysvětlit, co chrání `io.LimitReader` a co `DisallowUnknownFields`

## AI režim

`BOILERPLATE OK` — viz [docs/ai-playbook.md](../../docs/ai-playbook.md).

## Další čtení

1. [RFC 7807 — Problem Details for HTTP APIs](https://datatracker.ietf.org/doc/html/rfc7807)
2. [pkg.go.dev — encoding/json Decoder.DisallowUnknownFields](https://pkg.go.dev/encoding/json#Decoder.DisallowUnknownFields)
3. [Go blog — Working with Errors in Go 1.13](https://go.dev/blog/go1.13-errors)
4. [Go blog — Error handling and Go](https://go.dev/blog/error-handling-and-go)

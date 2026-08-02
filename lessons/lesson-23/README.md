# Lekce 23 — Checkpoint fáze 2: PHP zápachy a AI garbage

> **Čas:** ~90 min · **Fáze:** 2 — Idiomatický Go · **AI režim:** `JEN VYSVĚTLENÍ`

## Co budeš umět

- Projít cizí Go balíček a pojmenovat každý zápach jménem, ne pocitem „tohle se mi nelíbí".
- Přepsat vygenerovaný balíček do idiomatické podoby beze změny chování.
- Rozhodnout, kdy je pointer, interface nebo konstruktor zbytečný.
- Odhadnout podle stylu, který kus kódu psal model bez kontextu Go konvencí.
- Zhodnotit vlastní úroveň v tématech fáze 2 a vědět, co zopakovat.

## Recap

### Otázky a odpovědi

**Proč se getter nejmenuje `GetName`?** `Get` nese informaci „tohle něco načítá odjinud".
U prostého přístupu k poli je to šum. `user.Name()` a `user.SetName(x)`.

**Kdy je `i` lepší jméno než `index`?** Vždycky, když žije míň než pár řádků. Délka jména
roste s velikostí scope, ne s důležitostí.

**Proč je `utils` zápach?** Nemá hranici. Balíček musí jít popsat jednou větou bez slova
„a"; `utils` roste, dokud na něm nezávisí všechno.

**Co znamená „accept interfaces, return structs"?** Na vstupu ber nejobecnější typ, který
ti stačí; na výstupu vracej konkrétní. Vracený interface znamená, že každá nová metoda je
breaking change a volající nevidí zbytek API. Definuje ho **konzument** — balíček
s implementací o svém rozhraní obvykle vůbec neví.

**Kdy nepotřebuješ konstruktor?** Když je zero value použitelná. `var buf bytes.Buffer`,
`var mux http.ServeMux`. Mapy inicializuj líně v mutující metodě.

**Kdy použít functional options?** Od tří volitelných parametrů výš nebo když potřebuješ
rozlišit „nenastaveno" od zero value. Na dvě volby stačí `Config` struct.

**Jak vypadá dobrý text chyby?** Malé písmeno, bez tečky, bez „failed to", jeden nový
fakt na úroveň: `process user 42: read profile: open /data/42.json: no such file`.
`%w` je na sekvenční řetěz příčin, `errors.Join` na nezávislé chyby naráz.

**Proč potřebuje úklidová funkce named return?** Protože `defer` musí mít kam zapsat
chybu z `Close`. Bez `(err error)` ji nemáš jak vrátit.

**Kde hledat pravdu o stdlib?** `go doc -src` a `_test.go`. Testy popisují chování
přesněji než jakýkoli článek.

### Co si musíš pamatovat

| Téma | Pravidlo | Lekce |
|---|---|---|
| Zkratky | `URL`, `ID`, `HTTP` celé stejným případem | 19 |
| Konstanty | `MaxRetries`, ne `MAX_RETRIES` | 19 |
| Balíčky | krátké, jednoslovné, doménové; žádné `utils` | 19 |
| Koktání | `http.Server`, ne `http.HTTPServer` | 19 |
| Struktura | early return, happy path vlevo | 19 |
| Konstruktory | `New`, validace, `(T, error)`, `Must` navíc | 20 |
| Zero value | použitelná = žádný konstruktor | 20 |
| Interfacy | malé, u konzumenta, přijímat ne vracet | 20 |
| Chyby | malé písmeno, bez tečky, kontext přes `%w` | 21 |
| Ignorování | `_ = err` nikdy; `Close` u zápisu kontroluj | 21 |
| Panika | jen chyba programátora, ne chyba vstupu | 21 |
| Čtení kódu | `go doc -src`, testy, `go.mod`, `internal/` | 22 |

### Katalog zápachů

**1. Obří interface 1:1 se službou.** Deset metod, jedna implementace, existuje jen kvůli
mockování.

```go
// zápach
type OrderServiceInterface interface { /* 12 metod */ }
// oprava: port u konzumenta, jedna až tři metody
type OrderSaver interface{ Save(Order) error }
```

**2. Layer-cake balíčky.** `service/`, `repository/`, `entity/`, `dto/`. Vzniká import
cyklus a nikdo neví, kam patří nový soubor. Oprava: balíčky podle domény (`billing`,
`inventory`, `auth`), vrstva je soubor uvnitř balíčku.

**3. Pointer všude pro jistotu.** `func Process(o *Order) *Result`, `[]*User`, `*int`
pro nepovinné číslo. V PHP je objekt vždy reference, tak se to přenese. Oprava: hodnota
je výchozí volba; pointer jen když potřebuješ mutovat, sdílet velkou strukturu nebo
skutečně rozlišit „nenastaveno".

**4. Balíček `utils`.** Oprava: rozřež podle domény, nebo funkci přesuň tam, kde se používá.

**5. Getry s `Get`.** `GetName()`, `GetID()`. Oprava: `Name()`, `ID()`.

**6. Ignorované chyby.** `_ = json.Unmarshal(...)`, `defer f.Close()` u zápisu,
`log.Println(err); return nil`. Oprava: vrať s kontextem, nebo zpracuj a napiš proč.

**7. Konstruktor pro typ s užitečnou zero value.** `NewRegistry()`, který jen udělá mapu.
Oprava: líná inicializace v `Set`.

**8. DI framework.** Kontejner s reflexí, protože „ruční wiring je neudržitelný".
Oprava: `main` nebo `buildApp`, kompilátor jako kontrola.

**9. Panika místo erroru** v knihovní funkci. Oprava: `(T, error)`.

**10. Konstruktor, který vrací interface.** `func New() Servicer` nad neexportovaným
typem. Oprava: vracej `*Service`.

### Jak tyhle zápachy poznat v kódu od AI

Model generuje průměr svého trénovacího korpusu, a ten korpus je plný Javy a PHP. Znaky,
podle kterých to poznáš skoro okamžitě:

- **Symetrie, kterou nikdo nepotřebuje.** Ke každému typu interface, ke každému poli
  getter i setter.
- **Jména jsou dokonalá věta.** `NewUserServiceImplementation`, `currentIterationIndex`.
- **Chyby znějí jako log.** `"Failed to process the request: %v"` — velké písmeno, tečka,
  `%v` místo `%w`.
- **Komentáře popisují kód, ne důvod.** `// increment counter` nad `counter++`.
- **Struktura kopíruje framework.** `handlers/`, `services/`, `models/` bez doménového jména.
- **Nic není zero value.** Všechno má `New`, i prázdný registr.

Nejužitečnější věta při review AI kódu zní: **„Co by se stalo, kdybych to smazal?"**
Většina vygenerovaného kódu jsou vrstvy, které nic nedělají.

## Rozdíly proti PHP

Celá fáze 2 se dá shrnout jedním obrázkem. Takhle vypadá Symfony služba přepsaná do Go
doslova:

```go
// package service
type UserServiceInterface interface {
	GetUser(id *int) (*User, error)
	GetAllUsers() ([]*User, error)
	CreateUser(u *User) error
	UpdateUser(u *User) error
	DeleteUser(id *int) error
	ValidateUser(u *User) (bool, error)
}

type userService struct{ repo repository.UserRepositoryInterface }

func NewUserService(repo repository.UserRepositoryInterface) UserServiceInterface {
	return &userService{repo: repo}
}
```

A takhle vypadá totéž napsané v Go:

```go
// package user

// Store je port. Definuje ho ten, kdo ho volá, a má jen to, co skutečně potřebuje.
type Store interface {
	Save(User) error
}

type Service struct{ store Store }

func New(store Store) (*Service, error) { /* ... */ }
```

Rozdíl není v počtu řádků, ale v tom, kdo koho vlastní. V první verzi vlastní rozhraní
implementace a konzument bere, co dostane. Ve druhé ho vlastní konzument a implementace
se mu přizpůsobí. To je celá fáze 2 v jedné větě.

## Časté chyby

| Chyba | Proč vzniká | Jak to udělat správně |
|-------|-------------|------------------------|
| Interface „pro testovatelnost" u všeho | PHPUnit mockuje třídy přes interface | fake implementace portu v testu, 5 řádků |
| `[]*Record` bez důvodu | v PHP je pole objektů pole referencí | `[]Record`, dokud nepotřebuješ mutovat |
| Balíček `models` s celou doménou | Doctrine entity žijí pohromadě | balíček podle bounded contextu |
| Chyba jen zalogovaná a spolknutá | globální exception handler v Symfony | vrať ji volajícímu s kontextem |
| `New()` u typu s prázdnou mapou | „objekt bez konstruktoru neexistuje" | líná inicializace |
| Refaktor, který mimochodem mění chování | „stejně to bylo špatně" | testy první, chování zamčené |

## AI kvíz

Po přečtení teorie spusť v Cursoru **`/go-deep-quiz 23`**. AI tě ~5 minut prověří mentální model (ne hotové cvičení). Slabiny si uloží do [`GAPS.md`](../../GAPS.md).

## Úkol

Pracuj v `exercise/`. Po doplnění spouštěj testy:

Stupně jdou od jednodušších ke složitějším — po každém stupni spusť review, než jdeš dál.

### Jednoduchý

Funkce: `Describe`, `LoadRecords`

```bash
make lesson L=23 PART=1
```

Pak **`/go-deep-review 23 easy`**.

### Střední

Funkce: `Put`, `PutAll`, `Load`

```bash
make lesson L=23 PART=2
```

Pak **`/go-deep-review 23 medium`**.

### Obtížný

Funkce: `Remove`, `List`, `TotalQty`

```bash
make lesson L=23 PART=3
```

Pak **`/go-deep-review 23 hard`**.

Až budou stupně hotové, porovnej se `solutions/` (spoiler).

## Sebehodnocení

Za každou položku si dej body podle toho, jak jsi na tom **bez nápovědy**:
2 = zvládnu a umím vysvětlit proč, 1 = zvládnu s dohledáním, 0 = nezvládnu.

| # | Dovednost | Body |
|---|---|---|
| 1 | Pojmenuju typ, funkci i balíček podle konvencí bez přemýšlení | 0–2 |
| 2 | Přepíšu zanořený kód na early return beze změny chování | 0–2 |
| 3 | Rozhodnu, jestli typ potřebuje konstruktor | 0–2 |
| 4 | Napíšu functional options a vím, kdy je nepoužít | 0–2 |
| 5 | Navrhnu port u konzumenta a napíšu k němu fake | 0–2 |
| 6 | Napíšu text chyby, který obstojí v řetězu | 0–2 |
| 7 | Správně zkombinuju chybu z práce a z úklidu | 0–2 |
| 8 | Použiju `errors.Is`, `%w` i `errors.Join` na správném místě | 0–2 |
| 9 | Najdu odpověď ve zdrojáku stdlib místo na webu | 0–2 |
| 10 | Poznám v cizím diffu zápach a pojmenuju ho | 0–2 |

Vyhodnocení:

| Skóre | Co s tím |
|---|---|
| 18–20 | Jdi na fázi 3. Idiomatický Go máš v ruce. |
| 14–17 | Přepiš úkol C těch lekcí, kde jsi měl 0–1, znovu z prázdné složky. |
| 10–13 | Zopakuj celé lekce podle nejslabších položek: 1–2 → 19, 3–5 → 20, 6–8 → 21, 9–10 → 22. |
| pod 10 | Vrať se na lekci 19 a projdi fázi znovu. Bez idiomů budeš ve fázi 3 psát Symfony v Go. |

Zvlášť sleduj položky 5 a 7. Návrh portů a kombinování chyb ti fáze 3 a 4 budou vracet
každý den, a zároveň je AI generuje nejhůř.

## Závěrečné otázky

Spusť **`/go-deep-review 23 final`**. AI projde body níže, doptá se a ověří pochopení. Celé cvičení ověří `make lesson L=23` (+ `make race L=23`, pokud to lekce vyžaduje).

- [ ] `grep -c '\*Record' exercise/exercise.go` vrací 0
- [ ] Umíš vyjmenovat aspoň sedm zápachů z katalogu a ke každému opravu
- [ ] Umíš vysvětlit, proč `Store` nepotřebuje konstruktor
- [ ] Umíš vysvětlit, proč `Describe` bere `Loader` a ne `*Store`
- [ ] Umíš vysvětlit, kdy `errors.Join` a kdy `%w`
- [ ] Umíš popsat tři znaky, podle kterých poznáš Go vygenerované AI

## AI režim

`JEN VYSVĚTLENÍ` — viz [docs/ai-playbook.md](../../docs/ai-playbook.md).

Tenhle checkpoint uzavírá režim `JEN VYSVĚTLENÍ`. Od lekce 24 platí `BOILERPLATE OK`,
což znamená, že AI smí psát repetitivní kód — ale jen proto, že už umíš poznat, kdy ho
napsala špatně. Vyzkoušej si to hned: nech si vygenerovat řešení úkolu C, projdi ho
podle katalogu zápachů výše a vypiš, co bys v review komentoval.

## Další čtení

1. [Go Code Review Comments](https://go.dev/wiki/CodeReviewComments)
2. [Go blog — Package names](https://go.dev/blog/package-names)
3. [Effective Go](https://go.dev/doc/effective_go) — po fázi 2 ho přečteš celý za večer
4. [Go blog — Errors are values](https://go.dev/blog/errors-are-values)

# PHP/Symfony → Go: mentální mosty

Nejde o 1:1 mapování. Jde o to, **který problém řešíš** a jaký nástroj na něj Go nabízí.
Sloupec „Lekce" ukazuje, kde se to probírá.

| Symfony / PHP | Go ekvivalent | Poznámka | Lekce |
|---------------|---------------|----------|-------|
| Composer + `composer.json` | Go modules + `go.mod` | cesta importu je adresář, ne namespace v konfiguraci | 01 |
| Class | `struct` + metody | žádná hierarchie tříd | 05 |
| Dědičnost | embedding a kompozice | promotion metod není `extends` | 05 |
| Interface (PHP) | `interface` s implicitní implementací | žádné `implements`, žádný import implementace | 12 |
| `public` / `private` | velké / malé písmeno | hranicí je balíček, ne třída | 11 |
| `null` | zero value, případně `(hodnota, bool)` | proměnná je vždy inicializovaná | 03 |
| Objekty jako reference | hodnoty jako default | pointer jen když sdílíš nebo mutuješ | 06 |
| Exception | `error` jako návratová hodnota | součást signatury, tedy kontraktu | 14 |
| try/catch | `if err != nil` | explicitní tok řízení | 14 |
| finally | `defer` | LIFO, běží i při panice | 10 |
| Exception hierarchie | sentinel chyby a `errors.As` | typ chyby místo dědičnosti | 14 |
| Bundle / modul | Go module a balíčky | `internal/` je privátní pro modul | 11, 32 |
| Service container | ruční wiring ve `main` | žádná runtime magie | 20, 33 |
| Autowiring | konstruktory se závislostmi | závislosti jsou vidět v signatuře | 20 |
| `services.yaml` | funkce `Wire()` v `main` | konfigurace je kód, který se kompiluje | 33 |
| Doctrine entity | struct + repozitářový port | persistence není doména | 34, 35 |
| Doctrine repository | interface u konzumenta | často jen 1–2 metody | 33, 35 |
| ORM a lazy loading | explicitní dotazy přes `database/sql` | žádný unit of work, žádné překvapení | 35 |
| Migrace (DoctrineMigrations) | číslované SQL soubory, forward-only | pořadí a drift řešíš sám | 35 |
| Validator s atributy | explicitní `Validate()` a parse-don't-validate | typ nese invariant | 36 |
| Controller | `http.Handler` | píšeš do streamu, není Response objekt | 24 |
| Routing anotace | vzory `ServeMux` od Go 1.22 | `"GET /items/{id}"` a `r.PathValue` | 25 |
| EventListener / kernel.request | middleware `func(http.Handler) http.Handler` | explicitní kompozice, žádné priority tagy | 26 |
| RequestStack | `context.Context` jako první parametr | nikdy ve struct fieldu | 27 |
| `.env` + `%env(int:PORT)%` | `os.LookupEnv` + validace při startu | fail-fast, ne pád za běhu | 28 |
| Monolog | `log/slog` | strukturované atributy, ne formátovaný text | 29 |
| Guzzle | `http.Client` s timeoutem | `DefaultClient` timeout nemá | 30 |
| PHPUnit | `testing` + table-driven | žádné anotace, žádné setUp | 17 |
| Mockery / Prophecy | ruční fake struct | obvykle kratší než mock setup | 33 |
| Atributy / anotace | struct tagy, jen pro serializaci | logika v kódu, ne v metadatech | 16, 54 |
| Symfony Serializer | `encoding/json` + tagy | méně automatiky, víc kontroly | 16 |
| Žádná souběžnost (request = proces) | goroutiny a kanály | sdílený stav je najednou tvůj problém | 40–50 |

## Co si přenést

- Bounded context myšlení a doménový jazyk
- Explicitní hranice aplikace
- Testovatelnost přes porty
- Observabilita a provozní disciplína
- Zvyk psát akceptační kritéria

## Co nechat za dveřmi

- Abstraktní bázové třídy „pro budoucnost"
- Obří interfacy „pro mockování všeho"
- Anemické entity a fat services jako jediný styl
- Framework-first reflex
- Balíčky pojmenované podle vrstvy

## Typický zápach „Symfony v Go"

```go
// špatně: layer balíčky a obří interface u implementace
package service

type UserServiceInterface interface {
	Create(u *User) error
	Update(u *User) error
	Delete(id string) error
	Find(id string) (*User, error)
	FindAll() ([]*User, error)
	FindByEmail(e string) (*User, error)
	// ...
}
```

```go
// lépe: malý port definovaný tam, kde se používá
package billing

type UserStore interface {
	Save(ctx context.Context, u User) error
}
```

Rozdíl není v počtu řádků. Je v tom, že druhá varianta jde implementovat testovacím
fakem na pěti řádcích a že balíček `billing` říká, co dělá, zatímco `service` neříká nic.

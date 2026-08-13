# Lekce 35 — Persistence: repozitář, in-memory fake, SQL mindset

> **Čas:** ~75 min · **Fáze:** 4 — Architektura v Go · **AI režim:** `BOILERPLATE OK`

## Co budeš umět

- Držet doménový typ čistý: bez `db:` tagů, bez znalosti tabulky a bez SQL.
- Popsat mentální model `database/sql` — pool, `QueryRow` vs `Query` vs `Exec`,
  povinné `rows.Close()` a `rows.Err()`.
- Převést `sql.ErrNoRows` na doménovou chybu a vysvětlit, proč to musíš udělat.
- Sestavit dotaz tak, aby byl odolný proti SQL injection, včetně případu, kdy
  placeholder použít nelze.
- Naplánovat běh migrací a poznat drift mezi tím, co je aplikované, a tím, co je
  v repozitáři.

## Teorie

### Persistence není doména

Doménový typ nesmí obsahovat `db:` tagy, jméno tabulky, ani SQL. Důvod není
estetický: jakmile doména ví o schématu, každá změna sloupce je změna domény
a doménu nejde otestovat bez databáze.

Hranice se drží dvěma typy. Doménový `User` a — pokud se schéma liší od domény —
neexportovaný `userRow` uvnitř adaptéru:

```go
package postgres

type userRow struct {
	id        string
	email     string
	createdAt time.Time
}

func (r userRow) toDomain() app.User { /* mapování */ }
```

Devětkrát z deseti se struktury shodují a mapuješ přímo do doménového typu. Ten
desátý případ (denormalizace, historická jména sloupců) je přesně ten, kvůli
kterému se ta hranice vyplatí mít.

### Port u konzumenta, in-memory adaptér první

Repozitář je port. Definuje si ho ten, kdo ho používá, a obsahuje jen metody,
které opravdu potřebuje:

```go
type UserRepo interface {
	Get(ctx context.Context, id string) (User, error)
	Save(ctx context.Context, u User) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context) ([]User, error)
}
```

`context.Context` je první parametr každé metody. Nejde o formalitu: bez něj
nedokážeš dotaz zrušit, když klient zavře spojení, a bez něj nefunguje timeout.

První implementaci napiš in-memory. Není to „jen na testy" — je to plnohodnotný
adaptér, který ti dá pracující aplikaci dřív, než vůbec rozhodneš o databázi.
Musí se ale chovat stejně jako ta budoucí SQL verze: stejné sentinelové chyby,
stejné řazení, stejná reakce na zrušený context. Jinak testy lžou.

### `database/sql` mentální model

Pět věcí, které musíš mít v hlavě, než napíšeš první dotaz:

1. **`*sql.DB` je pool, ne spojení.** Otevřeš ho jednou při startu a předáváš
   všude. `db.Close()` patří do `main`, ne do funkce, která dělá dotaz.
   `sql.Open` nic nepřipojuje — první spojení vznikne až u prvního dotazu, ověřit
   ho můžeš přes `db.PingContext(ctx)`.
2. **Tři metody podle tvaru výsledku.** `QueryRowContext` pro jeden řádek,
   `QueryContext` pro víc řádků, `ExecContext` pro INSERT/UPDATE/DELETE.
3. **`rows` se musí zavřít a zkontrolovat.**

```go
rows, err := db.QueryContext(ctx, "SELECT id, email FROM users WHERE active = $1", true)
if err != nil {
	return nil, fmt.Errorf("select users: %w", err)
}
defer rows.Close()

var out []User
for rows.Next() {
	var u User
	if err := rows.Scan(&u.ID, &u.Email); err != nil {
		return nil, fmt.Errorf("scan user: %w", err)
	}
	out = append(out, u)
}
return out, rows.Err()   // chyba z iterace, kterou Next() zamlčel
```

Vynechané `rows.Err()` je nejčastější tichá chyba v Go kódu nad SQL: síť spadne
uprostřed čtení, `Next()` vrátí `false`, cyklus skončí a ty vrátíš neúplný výsledek
s `nil` chybou.

4. **`sql.ErrNoRows` musíš přeložit.** `QueryRow(...).Scan(...)` vrátí u prázdného
výsledku `sql.ErrNoRows`. Pokud ho pustíš ven, celá aplikace bude znát
`database/sql`:

```go
if errors.Is(err, sql.ErrNoRows) {
	return User{}, fmt.Errorf("user %q: %w", id, ErrNotFound)
}
```

5. **Transakce má rozsah jedné obchodní operace.** `tx, err := db.BeginTx(ctx, nil)`,
hned `defer tx.Rollback()` (po úspěšném `Commit` je to no-op) a `tx.Commit()`
na konci. Transakce nesmí přežít HTTP request a nesmí obalovat volání cizího API —
držet zámky v databázi, dokud odpoví platební brána, je spolehlivý recept na výpadek.

### Placeholdery, injection a co placeholderem nejde

Hodnoty vždycky přes placeholder. Nikdy konkatenací.

```go
// KATASTROFA — nikdy
q := "SELECT id FROM users WHERE email = '" + email + "'"

// správně
db.QueryRowContext(ctx, "SELECT id FROM users WHERE email = $1", email)
```

Placeholder je `$1` v PostgreSQL a `?` v MySQL a SQLite. Driver hodnotu neposílá
jako text dotazu, takže ji server nikdy nemůže interpretovat jako SQL.

Háček, o kterém se mlčí: **placeholder funguje jen pro hodnoty.** Jméno tabulky,
jméno sloupce, směr řazení ani `LIMIT` u některých driverů takhle předat nejde.
A právě tam vzniká injection, protože programátor sáhne po `fmt.Sprintf`. Jediná
obrana je whitelist — seznam povolených jmen na straně serveru:

```go
var allowedSort = map[string]string{
	"email":  "email",
	"newest": "created_at DESC",
}

order, ok := allowedSort[req.Sort]
if !ok {
	return ErrBadSort
}
q := "SELECT id FROM users ORDER BY " + order   // bezpečné: order je z whitelistu
```

Rozdíl proti Doctrine QueryBuilderu je v tom, že tady tu obranu vidíš. Tam je
schovaná a lidé ji obcházejí přes `->andWhere("u.email = '$email'")`.

### N+1, migrace a co ti Doctrine nebude chybět

**N+1** v Go nepoznáš podle profileru, ale podle tvaru kódu: dotaz uvnitř cyklu
nad výsledkem jiného dotazu. Protože žádný lazy loading neexistuje, musíš ho
napsat vědomě — a to je právě ta výhoda. Řešení je jeden dotaz s `IN` nebo `JOIN`
a mapa v paměti.

**Migrace** jsou očíslované SQL soubory, typicky `0001_create_users.sql`. Verze
určuje pořadí, nikdy se nepřepisují a nikdy se nemažou — jsou forward-only.
Nástroj si drží tabulku aplikovaných verzí a při startu spočítá rozdíl. Dvě
situace musí zastavit deploy: dvě migrace se stejnou verzí (dva lidé, dvě větve)
a *drift*, tedy verze zapsaná jako aplikovaná, která v repozitáři chybí — někdo
smazal migraci, která už na produkci proběhla, a stav schématu je neznámý.

Co ti z Doctrine chybět nebude: unit of work, který uloží něco, co jsi neměl
v úmyslu; lazy loading, který udělá dotaz při čtení atributu; a proxy třídy
v stack trace. Cena je, že napíšeš víc řádků. Odměna je, že každý dotaz, který
proběhne, je vidět v kódu.

## Rozdíly proti PHP

Doctrine ti dá persistenci skoro zdarma — za cenu toho, že entita ví o databázi:

```php
#[ORM\Entity(repositoryClass: UserRepository::class)]
#[ORM\Table(name: 'users')]
class User
{
    #[ORM\Id, ORM\Column(type: 'string')]
    private string $id;

    #[ORM\Column(type: 'string', unique: true)]
    private string $email;
}

$user = $repo->find($id);   // null když nic
$user->setName('Alice');    // změna se uloží sama při flush()
$em->flush();
```

Doménový objekt je poseshovaný anotacemi, unit of work sleduje změny za tebe
a lazy loading dotahuje vazby, když na ně sáhneš. Go protějšek je nudnější
a to je jeho hlavní přednost:

```go
type User struct {   // žádné tagy, doména neví o databázi
	ID     string
	Email  string
	Name   string
	Active bool
}

u, err := repo.Get(ctx, id)          // err, ne null
if errors.Is(err, ErrNotFound) { /* ... */ }
u.Name = "Alice"
if err := repo.Save(ctx, u); err != nil { /* ... */ }   // explicitní zápis
```

Návyk k opuštění: **přestaň čekat, že se změna uloží sama.** Ve stdlib není ORM,
unit of work ani lazy loading. Každý zápis napíšeš. Zní to jako krok zpátky, dokud
si nevzpomeneš, kolikrát jsi v Symfony ladil, proč se něco uložilo, proč se to
neuložilo, nebo proč jeden `find()` v cyklu vygeneroval čtyři sta dotazů.

## Časté chyby

| Chyba | Proč vzniká | Jak to udělat správně |
|-------|-------------|------------------------|
| `db:` tagy na doménovém typu | reflex z Doctrine anotací | doménový typ čistý, mapování v adaptéru |
| `sql.ErrNoRows` uniká z repozitáře | „chyba je chyba" | přelož ji na vlastní `ErrNotFound` |
| Chybí `rows.Err()` | `Next()` vypadá, že stačí | vždy `defer rows.Close()` a `return rows.Err()` |
| `sql.Open` v každém handleru | zvyk na `new PDO` | jeden `*sql.DB` na aplikaci, je to pool |
| Jméno sloupce z requestu do `Sprintf` | placeholder tam nefunguje | whitelist povolených jmen |
| Transakce kolem HTTP volání | „ať je to atomické" | transakce = jedna obchodní operace, krátká |
| Dotaz v cyklu | chybí lazy loading, tak to řeším ručně | jeden dotaz s `IN`, pak mapa |

## AI kvíz

Po přečtení teorie spusť v Cursoru **`/go-deep-quiz 35`**. AI tě ~5 minut prověří mentální model (ne hotové cvičení). Slabiny si uloží do [`GAPS.md`](../../GAPS.md).

## Úkol

Pracuj v `exercise/`. Většina typů a `List`/`Plan` je hotová — doplňuješ kritické
mezery v adaptéru a staviteli dotazů. Kontrakt je v komentáři nad metodou.

### Jednoduchý

Oprav: `Get` (u chybějícího ID vrací nulu místo `ErrNotFound`)

```bash
make lesson L=35 PART=1
```

Pak **`/go-deep-review 35 easy`**.

### Střední

Doplň: `Save`, `Delete`

```bash
make lesson L=35 PART=2
```

Pak **`/go-deep-review 35 medium`**.

### Obtížný

Implementuj: `BuildSelect` (whitelist tabulek a sloupců)

```bash
make lesson L=35 PART=3
```

Pak **`/go-deep-review 35 hard`**.

Až budou stupně hotové, porovnej se `solutions/` (spoiler).

## Závěrečné otázky

Spusť **`/go-deep-review 35 final`**. AI projde body níže, doptá se a ověří pochopení. Celé cvičení ověří `make lesson L=35` (+ `make race L=35`, pokud to lekce vyžaduje).

- [ ] Umíš vysvětlit, proč doménový typ nesmí mít `db:` tagy
- [ ] Umíš vyjmenovat, co udělá `rows.Err()` navíc oproti `rows.Next()`
- [ ] Umíš popsat, proč `sql.ErrNoRows` nesmí opustit repozitář
- [ ] Umíš říct, kdy placeholder použít nejde a co dělat místo něj
- [ ] Umíš popsat drift v migracích a proč kvůli němu zastavit deploy

## AI režim

`BOILERPLATE OK` — viz [docs/ai-playbook.md](../../docs/ai-playbook.md). Scan
řádků, mapování struktur a tabulkové testy si nech vygenerovat. Hranici mezi
doménou a adaptérem, převod chyb a whitelist si ohlídej sám. AI tady s oblibou
vyrobí doménový typ s `db:` tagy, zapomene `rows.Err()` a jméno sloupce vlepí
do dotazu přes `fmt.Sprintf`.

## Další čtení

1. [pkg.go.dev — database/sql](https://pkg.go.dev/database/sql)
2. [Go wiki — SQL Interface](https://go.dev/wiki/SQLInterface)
3. [Go blog — Contexts and structs](https://go.dev/blog/context-and-structs) — proč `ctx` jako parametr, ne jako pole
4. [Go blog — Working with Errors in Go 1.13](https://go.dev/blog/go1.13-errors) — `errors.Is` a překlad chyb na hranici

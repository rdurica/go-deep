# Lekce 54 — Generics v API, reflexe a build tagy

> **Čas:** ~90 min · **Fáze:** 6 — Production Go · **AI režim:** `JUNIOR POD REVIEW`

## Co budeš umět

- Rozhodnout, kdy generika veřejné API zlepší a kdy ho jen zkomplikují.
- Vysvětlit, proč metoda nesmí mít vlastní typový parametr, a obejít to funkcí.
- Napsat kód nad `reflect`, a zároveň zdůvodnit, proč ho v aplikační vrstvě nechceš.
- Rozdělit implementaci podle `//go:build` tagů a spustit test nad konkrétní variantou.
- Poznat rozdíl mezi tagem, příponou souboru (`_linux.go`) a `//go:generate`.

## PHP → Go most

V PHP jsou generika jen v docblocku a reflexe je běžný nástroj. Doctrine, Symfony DI i
Serializer stojí na tom, že runtime dokáže přečíst atributy a typy tříd:

```php
#[ORM\Column(name: 'user_name')]
private string $name;
```

Kontejner si na startu projde třídy, přečte atributy a poskládá služby. Je to pomalé, ale
výsledek se zkompiluje do cache a nikdo to neřeší.

V Go je tohle pořád možné, ale je to výslovně nástroj **knihoven**, ne aplikací:

```go
type User struct {
	Name string `map:"user_name"`
}
```

Co se mění v uvažování: **v Go platíš za reflexi při každém volání, ne jednou při warm-upu.**
Neexistuje kompilovaná cache kontejneru. Když si v aplikaci napíšeš vlastní mapper přes
reflexi, zaplatíš ho v každém requestu — a navíc přijdeš o kontrolu kompilátoru.
Reflexe patří do `encoding/json`, ne do tvého service layeru.

## Teorie

### Kdy generika v API pomáhají

Pravidlo, které se osvědčuje: **generika pro kontejnery a algoritmy, ne pro obecnost do
budoucna.**

```go
// Dobré použití: algoritmus, který se liší jen typem prvku.
func Filter[T any](in []T, keep func(T) bool) []T

// Špatné použití: "kdyby to někdy někdo chtěl jinak".
func NewService[T Repository](repo T) *Service[T]   // interface stačí
```

Rozhodovací otázka je jednoduchá: *potřebuju znát konkrétní typ, nebo jen jeho chování?*
Chování → rozhraní. Konkrétní typ (protože ho vracím, ukládám do slice nebo porovnávám)
→ typový parametr.

Rozhraní má jednu cenu, o které je dobré vědět: hodnota se do něj **zabalí** a při každém
volání metody se skáče přes tabulku. Generická funkce se pro každý typový argument
kompiluje do sdílené instance podle tvaru typu, takže bývá rychlejší — ale rozdíl je
obvykle v jednotkách nanosekund a není to důvod pro volbu.

### Typová inference a její meze

```go
r := Ok(42)                 // Result[int] — odvozeno z argumentu
e := Err[string](errTest)   // typ nejde odvodit, argument je jen error
```

Inference funguje z **argumentů**, ne z návratového typu. Jakmile se typový parametr
v parametrech nevyskytuje, musíš ho napsat. Proto má `Err` explicitní `[string]`.

### Proč `Map` není metoda

```go
// nejde přeložit: metody nesmí mít vlastní typové parametry
func (r Result[T]) Map[U any](f func(T) U) Result[U]

// jde: volná funkce se dvěma typovými parametry
func Map[T, U any](r Result[T], f func(T) U) Result[U]
```

Důvod je implementační: sada metod typu musí být známá při kompilaci typu, aby šlo
sestavit tabulku metod pro rozhraní. Metoda s vlastním parametrem by znamenala nekonečně
mnoho metod. Praktický dopad: transformace mezi typy se v Go píšou jako funkce, takže
místo `r.Map(f).Map(g)` píšeš `Map(Map(r, f), g)`. Není to hezké — a je to jeden z důvodů,
proč se `Result` v idiomatickém Go moc nepoužívá a vrací se prostě `(T, error)`.

### Reflexe

```go
rv := reflect.ValueOf(v)
if rv.Kind() == reflect.Pointer { rv = rv.Elem() }
rt := rv.Type()
for i := 0; i < rt.NumField(); i++ {
	f := rt.Field(i)
	if !f.IsExported() { continue }        // k neexportovaným se nedostaneš
	tag, ok := f.Tag.Lookup("map")          // Lookup rozliší "" od chybějícího
	_ = tag
	_ = ok
	fmt.Println(f.Name, rv.Field(i).Interface())
}
```

Tři věci, které se pletou:

- `Kind` je *druh* (struct, pointer, int), `Type` je *typ* (`main.User`). Přepínej podle
  `Kind`, porovnávej podle `Type`.
- K neexportovaným polím se přes reflexi nedostaneš — `Interface()` na nich panikuje.
  Proto se `json:"..."` na malém písmenu tiše ignoruje.
- Zápis vyžaduje **adresovatelnou** hodnotu, tedy pointer: `reflect.ValueOf(&c).Elem()`.
  Bez toho `CanSet()` vrátí `false`.

Legitimní použití reflexe jsou serializace, ORM a DI kontejnery — tedy knihovny, které
z principu nemůžou znát typ volajícího. V aplikačním kódu je skoro vždycky lepší napsat
mapování ručně: je typově kontrolované, čitelné a rychlejší.

Na měření tohohle cvičení vychází reflexivní `StructToMap` zhruba **3× pomalejší** než
ruční mapování (401 ns/op proti 141 ns/op). Rozdíl je tady tlumený tím, že obě verze
alokují mapu; u větších structů a při zápisu polí bývá poměr desetinásobný.

### Build tagy

```go
//go:build fancy

package shop
```

Řádek musí být **před** `package` a oddělený prázdným řádkem. Starší tvar
`// +build fancy` se od Go 1.17 negeneruje, ale pořád ho v cizím kódu potkáš — `gofmt` ho
umí převést. Výrazy podporují `&&`, `||`, `!` a závorky: `//go:build linux && !arm64`.

Pro OS a architekturu tagy většinou psát nemusíš, stačí přípona souboru:

| Soubor | Kdy se přeloží |
|--------|----------------|
| `store_linux.go` | jen na Linuxu |
| `store_windows.go` | jen na Windows |
| `store_amd64.go` | jen na amd64 |
| `store_test.go` | jen při `go test` |

Vlastní tag zapneš při buildu:

```bash
go build -tags fancy ./...
go test -tags fancy ./...
```

Typická použití jsou tři: platformní implementace, integrační testy, které se v běžném
CI nespouští (`//go:build integration`), a feature flag vyhodnocený při buildu — což je
levnější než runtime podmínka, ale nedá se přepnout bez nasazení.

Pozor na past: soubor, který kvůli tagu nikdy neprojde překladem, také **neprojde `vet`em**
a nikdo si nevšimne, že se rozbil. Proto musí mít každá varianta test, který ji v CI
skutečně sestaví.

`//go:generate` je něco jiného — není to build tag, jen komentář s příkazem, který spustí
`go generate ./...`. Nic se nespouští při buildu; vygenerovaný kód se commituje.

## Časté chyby

| Chyba | Proč vzniká | Jak to udělat správně |
|-------|-------------|------------------------|
| Generika tam, kde stačí rozhraní | zvyk na `<T>` z jiných jazyků | typový parametr jen když typ opravdu potřebuješ znát |
| Snaha napsat `func (r Result[T]) Map[U any](...)` | analogie s fluent API | volná funkce se dvěma parametry |
| Vlastní mapper přes reflexi v service vrstvě | reflex ze Symfony Serializeru | napiš mapování ručně, je rychlejší i čitelnější |
| Tag na neexportovaném poli | zvyk na `private` s atributem | reflexe k němu nemá přístup, exportuj ho |
| `reflect.ValueOf(c)` a pak `SetString` | zapomenutá adresovatelnost | `reflect.ValueOf(&c).Elem()` |
| Prázdný řádek chybí mezi `//go:build` a `package` | vypadá to jako obyčejný komentář | bez něj se řádek ignoruje a tag neplatí |

## Úkol

Pracuj v `exercise/`. Typy `User`, `Config` a `BadConfig` včetně tagů jsou předvyplněné.

### A — rozcvička (~10 min)

Generický `Result[T]` — hodnota, nebo chyba.

- `Ok[T any](v T) Result[T]` a `Err[T any](err error) Result[T]`.
  `Err(nil)` je platný Ok s nulovou hodnotou — tedy `IsOk()` závisí jen na tom, jestli je
  chyba nil. Nulová hodnota `Result[T]` musí být použitelná (je to Ok s nulovým `T`).
- `(r Result[T]) IsOk() bool` a `(r Result[T]) Unwrap() (T, error)`. Při chybě vrací
  `Unwrap` **nulovou** hodnotu, ne to, co bylo uvnitř.
- `Map[T, U any](r Result[T], f func(T) U) Result[U]` — nad chybou se `f` **nesmí zavolat**.
- `Must[T any](v T, err error) T` — při chybě panikuje hodnotou té chyby
  (test ji vytahuje přes `recover()` a `errors.Is`).

### B — jádro (~35 min)

1. `StructToMap(v any) (map[string]any, error)` — přes reflexi převede struct nebo pointer
   na struct na mapu. Neexportovaná pole přeskoč. Tag `map:"jmeno"` určí klíč, `map:"-"`
   pole vynechá, chybějící nebo prázdný tag znamená jméno pole beze změny. Cokoli, co není
   struct (včetně `nil` a nil pointeru), je chyba obalující `ErrNotStruct`.
2. `UserToMap(u User) map[string]any` — totéž pro `User`, ale ručně a bez reflexe. Test
   ověřuje, že obě funkce vrátí identickou mapu; benchmark je porovná.
3. `SetDefaults(v any) error` — vezme **nenilový pointer na struct** a každé exportované
   pole, které má tag `default` a je na nulové hodnotě, nastaví podle tagu. Podporuj
   `string`, celočíselné typy (`ParseInt` na správný počet bitů) a `bool`. Vyplněné pole
   se nepřepisuje, neexportované se ignoruje. Špatný vstup je `ErrNotPointer`,
   nepřevoditelná hodnota `ErrBadDefault`, nepodporovaný typ pole `ErrUnsupportedKind`.

### C — rozšíření (~20 min)

V `exercise/` už jsou dva soubory se stejnými funkcemi a opačnými build tagy:

- `feature_off.go` s `//go:build !fancy` — `FeatureName()` vrací `"basic"`, `Discount`
  vždy `0`,
- `feature_on.go` s `//go:build fancy` — `FeatureName()` vrací `"fancy"`, `Discount` vrací
  20 % z částky (celočíselně dolů), pro nekladný vstup `0`.

Doplň obě těla. Výchozí variantu ověřuje `feature_default_test.go` (má sám tag `!fancy`),
zapnutou `feature_fancy_test.go`:

```bash
make lesson L=54                                  # basic
cd exercise && go test -tags fancy -count=1 .     # fancy
```

Všimni si, že testy musí být otagované také — jinak by test na `"basic"` spadl při buildu
s `-tags fancy`. Obě varianty musí projít `gofmt` i `go vet -tags fancy`.

Až budeš hotový, porovnej se `solutions/` (spoiler).

## Ověření

Po dokončení úkolů spusť v Cursoru **`/go-deep-review`** a zadej třeba jen `54`. AI tě postupně projde body níže, doptá se a ověří pochopení — nestačí jen zelené testy.

- [ ] `make lesson L=54` prochází i `go test -tags fancy` v `exercise/`
- [ ] Umíš vysvětlit, proč metoda nesmí mít vlastní typový parametr
- [ ] Umíš uvést jeden případ, kdy je generika lepší než rozhraní, a jeden opačný
- [ ] Umíš vysvětlit, proč reflexe nevidí neexportovaná pole
- [ ] Umíš vysvětlit rozdíl mezi `Kind` a `Type`
- [ ] Umíš říct, proč otagovaný soubor bez testu je riziko

## AI režim

`JUNIOR POD REVIEW` — viz [docs/ai-playbook.md](../../docs/ai-playbook.md).

Reflexe je téma, kde agenti generují kód, který „funguje" a přitom tiše ignoruje polovinu
hraničních případů (nil pointer, embedded struct, neadresovatelná hodnota). Napiš testy
na hraniční případy dřív, než agenta pustíš k implementaci.

## Další čtení

1. [Go blog — An Introduction To Generics](https://go.dev/blog/intro-generics)
2. [Go blog — When To Use Generics](https://go.dev/blog/when-generics)
3. [Go blog — The Laws of Reflection](https://go.dev/blog/laws-of-reflection)
4. [go.dev — Build constraints](https://pkg.go.dev/cmd/go#hdr-Build_constraints)

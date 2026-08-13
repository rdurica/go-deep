# Lekce 44 — Race lab: detektor závodů

> **Čas:** ~35 min · **Fáze:** 5 — Concurrency do hloubky · **AI režim:** `JUNIOR POD REVIEW`

## Co budeš umět

- Definovat datový závod přesně a poznat ho v cizím kódu podle tří podmínek.
- Vysvětlit, proč závod není „občas špatné číslo", ale nedefinované chování.
- Pustit `go test -race`, přečíst jeho výstup a vědět, co detektor **nenajde**.
- Napsat stresový test, který dá závodu šanci se projevit.
- Opravit pět typických závodů: čítač, mapu, línou inicializaci, `append` a hot reload.

## Teorie

### Co je datový závod

Formálně jsou potřeba tři věci najednou:

1. dvě různé goroutiny přistupují ke **stejnému paměťovému místu**,
2. **aspoň jeden** z přístupů je zápis,
3. mezi přístupy **není uspořádání** (happens-before) vynucené synchronizací.

Když kterákoli chybí, závod to není. Dvě goroutiny čtoucí tutéž proměnnou jsou v pohodě.
Zápisy na různé indexy jednoho slice taky (lekce 40). A dva zápisy oddělené odesláním
hodnoty kanálem, mutexem nebo `WaitGroup` jsou uspořádané, takže také v pořádku.

### Závod je nedefinované chování

Nejnebezpečnější mýtus zní „no jo, občas mi vyjde o něco míň". Go memory model říká
něco jiného: program s datovým závodem **nemá definované chování**. Kompilátor smí
předpokládat, že závod neexistuje, a podle toho optimalizovat — čtení vytáhnout ze
smyčky, zápisy přeuspořádat, hodnotu držet v registru. To, co uvidíš, tedy nemusí
odpovídat žádnému prokládání zdrojového kódu.

V praxi se to projevuje takhle:

```go
// goroutina A
for !done {          // kompilátor smí načíst done jen jednou → nekonečná smyčka
    work()
}

// goroutina B
done = true
```

A u větších hodnot než jedno slovo (interface, string, slice, struct) hrozí **roztržená
hodnota**: interface má pointer na typ a pointer na data, string má pointer a délku. Když
někdo mění interface a někdo ho zároveň čte, může vzniknout kombinace typu z jedné
hodnoty a dat z druhé. Dereference takové hodnoty pak není špatný výsledek, ale pád nebo
tichá katastrofa. Přesně proto se konfigurace vyměňuje přes `atomic.Value`, a ne
poličkováním jednotlivých polí.

### Jak funguje `-race`

Race detektor (ThreadSanitizer) instrumentuje **každý** přístup do paměti a udržuje pro
každé paměťové místo historii, kdo na něj naposledy sahal, plus vektorové hodiny
goroutin. Když najde dvojici přístupů bez happens-before vztahu, vypíše hlášení se
stackem obou.

```
$ go test -race ./...
WARNING: DATA RACE
Read at 0x00c0000183a8 by goroutine 12:
  ...SafeIncrement.func1()
Previous write at 0x00c0000183a8 by goroutine 9:
  ...SafeIncrement.func1()
```

Cena je citelná: 5–10× pomalejší běh a 5–10× víc paměti. Proto `-race` běží v CI a při
vývoji, ne v produkci.

Zásadní je vědět, co detektor **neumí**:

- **Najde jen to, co se skutečně provedlo.** Když se závodná větev v testu nikdy
  nespustí, detektor mlčí. Není to statická analýza.
- **Nenajde závod, který se v daném běhu neprovedl souběžně.** Proto potřebuješ zátěž a
  opakování, ne jedno volání.
- **Nenajde logické závody.** `if !c.Has(k) { c.Set(k, v) }` pod korektními zámky je
  bez datového závodu a přesto špatně (TOCTOU).
- **Neřekne ti, jestli to opravdu vadí.** Řekne ti, že to je závod. To stačí — opravuje
  se každý.

Užitečné přepínače:

```bash
go test -race ./...
go test -race -count=20 ./...             # dvacet běhů, závod dostane šanci
go test -race -run TestRegistry -count=50 .
GORACE="halt_on_error=1" go test -race ./...   # zastavit hned u prvního nálezu
GORACE="history_size=4" go test -race ./...    # delší historie přístupů (víc paměti)
```

`-count` je tvůj nejlepší kamarád: vypíná cache výsledků a pouští test znovu a znovu.

### Kde závody vznikají

Pět zdrojů, které pokrývají většinu reálných případů — a všech pět je v dnešním cvičení:

1. **Čítač.** `counter++` z více goroutin. Oprava: `atomic.Int64` nebo mutex.
2. **Sdílená mapa.** Souběžný zápis do mapy navíc runtime aktivně detekuje a shodí
   program hláškou `fatal error: concurrent map writes`, kterou nejde odchytit
   `recover`em. Oprava: mutex, nebo mapa vlastněná jedinou goroutinou.
3. **Líná inicializace.** `if x == nil { x = new(...) }` ze dvou goroutin vyrobí dva
   objekty a jeden zápis se ztratí. Oprava: `sync.Once`.
4. **`append` do sdíleného slice.** `append` čte a přepisuje hlavičku (pointer, délka,
   kapacita). Oprava: mutex, nebo předalokovaný slice a zápis na vlastní index.
5. **Struktura měněná po polích.** Hot reload konfigurace, kde reloader přepisuje pole
   za polem a čtenář si přečte půlku staré a půlku nové. Oprava: vyměň celou hodnotu
   přes `atomic.Value` / `atomic.Pointer[T]`, nebo ji celou čti i piš pod `RWMutex`.

Za zmínku stojí ještě dva jevy, které se s závody pletou. **Uzávěr nad proměnnou cyklu**
byl do Go 1.21 klasickým zdrojem závodu; od 1.22 má každá iterace vlastní kopii, takže
past platí jen pro proměnné deklarované mimo cyklus (lekce 40). A **falešné sdílení** —
dvě nezávislé proměnné na jedné cache line, které si procesory přehazují — není závod,
program je korektní, jen pomalý; řeší se paddingem a je to téma pro profilování.

### Jak psát testy, které závod odhalí

Test, který zavolá funkci jednou, neodhalí nic. Potřebuješ tři věci: **souběh**
(dostatek goroutin), **opakování** (dost iterací a `-count`) a **společný start**, aby se
goroutiny potkaly ve stejném okamžiku:

```go
func StressTest(t *testing.T, f func()) {
    t.Helper()
    start := make(chan struct{})
    var wg sync.WaitGroup
    wg.Add(goroutines)
    for i := 0; i < goroutines; i++ {
        go func() {
            defer wg.Done()
            <-start // startovní výstřel: všichni vyrazí naráz
            for j := 0; j < iterations; j++ {
                f()
            }
        }()
    }
    close(start)
    wg.Wait()
}
```

Vedle toho piš tvrzení, která selžou i **bez** `-race`: součet, který nesedí, délka
slice, která je menší, invariant, který neplatí. Takový test chytí i závod, který
detektor v daném běhu minul, a hlavně dává smysl i tomu, kdo `-race` zapomene zapnout.

## Rozdíly proti PHP

V PHP se datový závod uvnitř procesu nemůže stát. Jeden request, jedno vlákno, žádná
sdílená paměť. Souběžnost řešíš na úrovni **dat mimo proces** a nástroje na to jsou
databázové:

```php
// dva requesty, klasický lost update — a řešení je transakce nebo zámek v DB
$row = $db->fetch('SELECT views FROM posts WHERE id = ?', [$id]);
$db->exec('UPDATE posts SET views = ? WHERE id = ?', [$row['views'] + 1, $id]);
```

Důležité je, že v PHP tenhle problém **vidíš** — je v SQL, je o transakcích, mluví se o
něm na code review. V Go je stejná chyba jedním znakem a nikdo si jí nevšimne:

```go
p.views++ // dvě goroutiny, a je to lost update se stejnými následky
```

Co si přenést: instinkt „tohle čte a zapisuje sdílenou věc, kdo to hlídá?". Co opustit:
představu, že když kód vypadá jako jedna operace, tak jedna operace je. `views++` je ve
skutečnosti načtení, přičtení a uložení — přesně ten `SELECT` a `UPDATE` jako výše, jen
bez transakce.

## Časté chyby

| Chyba | Proč vzniká | Jak to udělat správně |
|-------|-------------|------------------------|
| „Závod je jen občas špatné číslo" | zkušenost s lost update v DB | je to nedefinované chování; opravuje se každý nález |
| Test bez zátěže a bez `-count` | jedno volání „přece stačí" | stresový test + `go test -race -count=20` |
| `-race` jen lokálně | běh je pomalý | povinně v CI; lokálně aspoň před PR |
| Líná inicializace bez `Once` | v PHP je objekt vždy jen jeden na request | `sync.Once` |
| Přepis konfigurace po polích | „vždyť je to jen pár přiřazení" | vyměň celou hodnotu (`atomic.Value`) |
| `recover` na `concurrent map writes` | vypadá jako panika | není panika, je to fatal error; oprav mapu |

## AI kvíz

Po přečtení teorie spusť v Cursoru **`/go-deep-quiz 44`**. AI tě ~5 minut prověří mentální model (ne hotové cvičení). Slabiny si uloží do [`GAPS.md`](../../GAPS.md).

## Úkol

Pracuj v `exercise/`. Kontrakt je v komentáři nad funkcí. Stupně jdou od jednodušších
ke složitějším — po každém stupni spusť review, než jdeš dál.

### Jednoduchý

Oprav: `SafeIncrement`, `Set`, `Get`, `Len` na `Registry`

```bash
make lesson L=44 PART=1
```

Pak **`/go-deep-review 44 easy`**.

### Střední

Implementuj: `StressTest`

```bash
make lesson L=44 PART=2
```

Pak **`/go-deep-review 44 medium`**.

### Obtížný

Oprav: `Store`, `Load` na `Config` (atomická výměna snapshotu; `StartReloader` je hotový)

```bash
make lesson L=44 PART=3
```

Pak **`/go-deep-review 44 hard`**.

Až budou stupně hotové, porovnej se `solutions/` (spoiler).

## Závěrečné otázky

Spusť **`/go-deep-review 44 final`**. AI projde body níže, doptá se a ověří pochopení. Celé cvičení ověří `make lesson L=44` (+ `make race L=44`, pokud to lekce vyžaduje).

- [ ] Umíš vyjmenovat tři podmínky datového závodu
- [ ] Umíš vysvětlit, proč je závod nedefinované chování, a ne „nepřesný výsledek"
- [ ] Umíš říct, co race detektor nikdy nenajde
- [ ] Umíš vysvětlit, proč `sync.Once` řeší línou inicializaci a `if x == nil` ne
- [ ] Umíš popsat, proč se konfigurace vyměňuje celá a ne po polích

## AI režim

`JUNIOR POD REVIEW` — viz [docs/ai-playbook.md](../../docs/ai-playbook.md).

Zkus si experiment: dej agentovi původní závodní `Config` a požádej ho o opravu. Velmi
pravděpodobně přidá mutex kolem každého jednotlivého pole — kód bude bez `WARNING`, ale
invariant mezi poli bude pořád rozbitý. To je typická vlastnost AI oprav: umlčí nástroj,
neopraví návrh. Reviewerská otázka zní vždycky „co je tady ta věc, která musí platit jako
celek?".

## Další čtení

1. [Go blog — Introducing the Go Race Detector](https://go.dev/blog/race-detector)
2. [Data Race Detector](https://go.dev/doc/articles/race_detector) — včetně proměnné `GORACE`
3. [The Go Memory Model](https://go.dev/ref/mem)
4. [pkg.go.dev — sync/atomic](https://pkg.go.dev/sync/atomic) — `Value`, `Pointer[T]`

# Lekce 41 — Kanály a ownership

> **Čas:** ~70 min · **Fáze:** 5 — Concurrency do hloubky · **AI režim:** `JUNIOR POD REVIEW`

## Co budeš umět

- Rozhodnout mezi nebufferovaným a bufferovaným kanálem a říct, co buffer skutečně řeší.
- Popsat v každém kusu souběžného kódu, kdo kanál **vlastní** — tedy kdo do něj zapisuje
  a kdo ho zavírá.
- Použít směrové typy `chan<- T` a `<-chan T` jako dokumentaci kontraktu.
- Napsat generátor, fan-in a publish/subscribe, které po sobě neuklidí nikdo jiný než ony.
- Vysvětlit, proč `nil` kanál blokuje navždy a k čemu se to hodí.

## Teorie

### Kanál je typované potrubí s rendez-vous

```go
ch := make(chan string)      // nebufferovaný
buf := make(chan string, 10) // buffer na 10 hodnot
```

Nebufferovaný kanál synchronizuje dvě goroutiny: zápis se dokončí až v okamžiku, kdy
někdo čte. To je vlastnost, ne omezení — dostáváš zadarmo happens-before vztah a
přirozený backpressure. Když producent běží rychleji než konzument, sám se zpomalí.

Buffer řeší **jednu jedinou věc**: vyrovnává krátkodobé výkyvy rychlosti. Neřeší
pomalého konzumenta (buffer se zaplní a jsi tam, kde jsi byl), neřeší chybějící
konzumenty (zaplní se a producent uvízne) a rozhodně neřeší chyby v návrhu. Když buffer
zvětšuješ, protože „to občas zatuhne", opravuješ symptom.

Praktické pravidlo: začni nebufferovaným kanálem. Buffer přidej, až když víš, jaké číslo
tam patří a proč (typicky „počet workerů", ať odesílatel nečeká na převzetí).

### Směrové typy jsou kontrakt

Uvnitř funkce máš `chan T`, ale v signatuře skoro nikdy:

```go
func Generate(nums ...int) <-chan int   // volající smí jen číst
func Consume(in <-chan int)             // konzument nemůže zapsat ani zavřít
func Emit(out chan<- int)               // producent nemůže číst
```

Konverze `chan T` → `<-chan T` je automatická, zpátky to nejde. Tím kompilátor vynutí
to, co jinak bývá jen v komentáři: konzument fyzicky **nemůže** zavolat `close`, protože
`close` na `<-chan T` je chyba překladu. Signatura tak dokumentuje směr toku dat a
zároveň ho hlídá.

### Kdo zavírá

Jediné pravidlo, které si musíš odnést: **kanál zavírá vždy odesílatel, nikdy příjemce.**

Vyplývá to z toho, co `close` znamená: „už nikdy nic nepošlu". To může tvrdit jen ten,
kdo posílá. Zápis do zavřeného kanálu je **panika**, ne chyba, takže když ho zavře
příjemce, producent zabije celý program.

```go
close(ch)
ch <- 1 // panic: send on closed channel
close(ch) // panic: close of closed channel
```

Čtení ze zavřeného kanálu naopak funguje pořád a vrací zero value:

```go
v, ok := <-ch // ok == false, když je kanál zavřený a prázdný
for v := range ch {
    // skončí sám, jakmile je kanál zavřený a vyčerpaný
}
```

Právě proto je `close` tak užitečný jako **broadcast**: zavřený kanál obslouží libovolný
počet čekajících čtenářů okamžitě, zatímco `done <- struct{}{}` obslouží právě jednoho a
bez čtenáře sám uvízne. Odtud `done := make(chan struct{})` a `close(done)` jako
univerzální signál „konec".

Když do jednoho kanálu posílá víc goroutin, nesmí ho zavřít žádná z nich. Zavření
koordinuje ten, kdo je spustil:

```go
var wg sync.WaitGroup
wg.Add(len(inputs))
for _, in := range inputs {
    go func() {
        defer wg.Done()
        for v := range in {
            out <- v
        }
    }()
}
go func() {
    wg.Wait()
    close(out) // právě jednou, až když už nikdo neposílá
}()
```

Zavírat nemusíš vždycky. `close` není `free()` — kanál uklidí GC, jakmile na něj nikdo
neukazuje. Zavírá se jen tehdy, když to čtenáři potřebují vědět (typicky kvůli `range`).

### `nil` kanál blokuje navždy

Zero value kanálu je `nil` a operace na něm nikdy neskončí — ani zápis, ani čtení.
Zní to jako recept na deadlock a většinou to tak i je (`var ch chan int` bez `make`).
V `select` se z toho ale stává užitečný nástroj: větev s `nil` kanálem se nikdy nestane
připravenou, takže si větve můžeš „vypínat" tím, že proměnnou nastavíš na `nil`. Detailně
v lekci 42.

### Generátorový vzor

Nejběžnější tvar souběžného kódu v Go: funkce vrátí kanál, o jehož naplnění i zavření se
sama postará.

```go
func Generate(nums ...int) <-chan int {
    out := make(chan int)
    go func() {
        defer close(out) // vlastník kanálu ho i zavírá
        for _, n := range nums {
            out <- n
        }
    }()
    return out
}

for v := range Generate(1, 2, 3) {
    fmt.Println(v)
}
```

Všimni si tří věcí: kanál se vytváří **uvnitř** funkce (nikdo jiný ho nevlastní), vrací
se jako `<-chan int` (volající nemůže zapsat ani zavřít) a `close` je v `defer` hned na
prvním řádku goroutiny, aby přežil i případný `return` uprostřed.

Vzor má jedno známé riziko: když volající přestane číst dřív, než generátor doběhne,
zůstane goroutina viset na `out <- n`. Řešení je `done`/`ctx` větev v `select`, jak jsi
ji viděl v lekci 40 a jak ji dotáhneme v lekci 42.

Typ kanálu může být cokoli, tedy i kanál — `chan chan Result` se používá tam, kde s
požadavkem posíláš i adresu, kam poslat odpověď. Je to elegantní, čitelné to bývá zřídka.

### Share memory by communicating

Slogan z Go blogu neznamená „kanály jsou lepší než mutexy". Znamená, že místo sdílení
jednoho místa v paměti a hlídání přístupu k němu si goroutiny **předávají vlastnictví**
dat. Hodnota, kterou jsi poslal do kanálu, už není tvoje — nesaháš na ni.

Typický přešlap je poslat do kanálu pointer nebo slice a pak s ním dál pracovat:

```go
buf := make([]byte, 1024)
n, _ := r.Read(buf)
ch <- buf[:n]      // předáno dál
buf = buf[:0]      // ...a hned zase přepisuji stejnou paměť. Závod.
```

Kdy volit kanál a kdy mutex, rozebírá lekce 43. Zjednodušeně: kanál na **předání a
orchestraci**, mutex na **ochranu stavu**, který někde musí zůstat.

## Rozdíly proti PHP

V Symfony je předání práce jinam vždycky přes prostředníka: Messenger bus, Redis fronta,
databázový outbox. Odesílatel a příjemce se nikdy nepotkají a nikdo neřeší, kdo frontu
„zavírá" — ta existuje nezávisle na obou.

```php
// odesílatel neví nic o příjemci a nikdy nečeká
$this->bus->dispatch(new SendInvoice($id));

// příjemce běží v jiném procesu, fronta žije dál, i když oba spadnou
```

Go kanál vypadá podobně, ale chová se přesně naopak. Nebufferovaný kanál je
**rendez-vous**: odesílatel čeká, dokud si hodnotu někdo nepřevezme.

```go
ch := make(chan Invoice)   // žádný buffer
ch <- inv                  // blokuje, dokud jiná goroutina nezavolá <-ch
```

A hlavně: kanál je objekt v tvé paměti, ne infrastruktura. Někdo ho musí zavřít, jinak
čtenáři visí; a když ho zavře nesprávná strana, program panikuje. Návyk „pošlu to do
fronty a je to vyřízené" tady nefunguje — u každého kanálu musíš vědět, **kdo je jeho
vlastník**.

## Časté chyby

| Chyba | Proč vzniká | Jak to udělat správně |
|-------|-------------|------------------------|
| Kanál zavírá příjemce | „už to nepotřebuju, tak to zavřu" | zavírá vždy odesílatel; příjemce dostane `<-chan T` |
| `close` volaný z více goroutin | fan-in bez koordinace | jedna goroutina: `wg.Wait(); close(out)` |
| `done <- struct{}{}` místo `close(done)` | zápis zní přirozeněji než zavření | `close` funguje bez čtenáře a pro všechny |
| Buffer jako lék na zatuhnutí | reflex z fronty, která „pobere všechno" | najdi chybějícího konzumenta; buffer jen s odůvodněným číslem |
| Zapomenutý `make` u kanálu | zero value vypadá použitelně jako u slice | `nil` kanál blokuje navždy — vždy `make` |
| Sdílení slice po odeslání do kanálu | v PHP se serializuje kopie do fronty | po odeslání je hodnota cizí, nesahej na ni |

## AI kvíz

Po přečtení teorie spusť v Cursoru **`/go-deep-quiz 41`**. AI tě ~5 minut prověří mentální model (ne hotové cvičení). Slabiny si uloží do [`GAPS.md`](../../GAPS.md).

## Úkol

Pracuj v `exercise/`. Kontrakt je v komentáři nad funkcí. Stupně jdou od jednodušších
ke složitějším — po každém stupni spusť review, než jdeš dál.

### Jednoduchý

Oprav: `ForgetClose` (kanál se po odeslání nezavírá)

```bash
make lesson L=41 PART=1
```

Pak **`/go-deep-review 41 easy`**.

### Střední

Implementuj: `Generate`, `Collect`

```bash
make lesson L=41 PART=2
```

Pak **`/go-deep-review 41 medium`**.

### Obtížný

Doplň: `Merge` (fan-in, zavření výstupu po všech vstupech)

```bash
make lesson L=41 PART=3
```

Pak **`/go-deep-review 41 hard`**.

Až budou stupně hotové, porovnej se `solutions/` (spoiler).

## Závěrečné otázky

Spusť **`/go-deep-review 41 final`**. AI projde body níže, doptá se a ověří pochopení. Celé cvičení ověří `make lesson L=41` (+ `make race L=41`, pokud to lekce vyžaduje).

- [ ] Umíš vysvětlit, proč kanál nikdy nezavírá příjemce
- [ ] Umíš popsat, co přesně řeší buffer a co neřeší
- [ ] Umíš říct, proč se konec signalizuje přes `close(done)` a ne zápisem
- [ ] Umíš vysvětlit, proč `Merge` potřebuje `WaitGroup` a samostatnou goroutinu na `close`
- [ ] Umíš u každého kanálu ve svém kódu ukázat prstem na jeho vlastníka

## AI režim

`JUNIOR POD REVIEW` — viz [docs/ai-playbook.md](../../docs/ai-playbook.md).

Nech si vygenerovat `Broker` a projdi ho jako reviewer: kdo zavírá kanály odběratelů?
Co se stane, když odběratel nečte? Je `Close` idempotentní? Zavírá se `subs` slice, nebo
tam zůstanou zavřené kanály, do kterých někdo pošle? Tohle jsou přesně chyby, které
agenti dělají, protože vypadají jako detail.

## Další čtení

1. [Go blog — Share Memory By Communicating](https://go.dev/blog/codelab-share)
2. [Effective Go — Channels](https://go.dev/doc/effective_go#channels)
3. [Go blog — Go Concurrency Patterns: Pipelines and cancellation](https://go.dev/blog/pipelines)
4. [The Go Programming Language Specification — Channel types](https://go.dev/ref/spec#Channel_types)

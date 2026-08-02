# Lekce 53 — pprof a profilování

> **Čas:** ~90 min · **Fáze:** 6 — Production Go · **AI režim:** `JUNIOR POD REVIEW`

## Co budeš umět

- Rozhodnout, kdy má smysl profilovat, a kdy si tím jen kazíš kód.
- Pořídit CPU a heap profil z testu i z běžícího procesu a otevřít ho v `go tool pprof`.
- Přečíst z profilu rozdíl mezi flat a cum a najít podle něj skutečné hrdlo.
- Odstranit typické alokace v horké cestě a doložit zlepšení číslem.
- Vystavit `net/http/pprof` na vlastním muxu a vysvětlit, proč nikdy ne na veřejném portu.

## Teorie

### Kdy profilovat

Profilování začíná tam, kde už máš dvě věci: **měřitelný problém** (pomalý endpoint,
rostoucí RSS) a **benchmark, který ho reprodukuje**. Bez benchmarku nemáš jak ověřit, že
tvoje změna pomohla, a skončíš u „mně to přijde rychlejší".

Pořadí, které funguje:

1. Reprodukuj problém benchmarkem nebo zátěžovým testem.
2. Pořiď profil.
3. Najdi jedno místo, které stojí za víc než 20 % nákladů.
4. Oprav ho, změř znovu, porovnej `benchstat`em.
5. Vrať se na krok 2, dokud se to vyplácí.

### Druhy profilů

| Profil | Co ukazuje | Kdy sáhnout |
|--------|-----------|-------------|
| `cpu` | kde procesor tráví čas (vzorkuje 100×/s) | pomalý výpočet |
| `heap` | živé objekty v paměti při snímku | rostoucí RSS, podezření na leak |
| `allocs` | všechny alokace od startu | tlak na GC, hodně krátkých objektů |
| `goroutine` | zásobníky všech goroutin | leak goroutin, deadlock |
| `block` | čekání na synchronizaci | goroutiny stojí na kanálu |
| `mutex` | kontence na zámcích | zámek je hrdlo |

`block` a `mutex` se musí zapnout (`runtime.SetBlockProfileRate`,
`runtime.SetMutexProfileFraction`), zbytek běží pořád.

### Profil z testu

```bash
go test -run xxx -bench BenchmarkCountWords -cpuprofile cpu.out -memprofile mem.out .
go tool pprof cpu.out
```

V interaktivním pprof jsou čtyři příkazy, se kterými vystačíš:

| Příkaz | Co udělá |
|--------|----------|
| `top` | deset nejdražších funkcí |
| `top -cum` | totéž seřazené podle kumulativního času |
| `list CountWords` | zdrojový kód funkce s náklady po řádcích |
| `peek regexp` | kdo funkci volá a kdo z ní |
| `web` | graf v prohlížeči (potřebuje Graphviz) |

**Flat** je čas strávený přímo v těle funkce. **Cum** je čas včetně všeho, co zavolala.
Funkce s vysokým cum a nízkým flat je jen průchoďák — hledej níž. Funkce s vysokým flat
je ta, kterou chceš opravit.

### Profil z běžícího procesu

```go
mux := http.NewServeMux()
mux.HandleFunc("/debug/pprof/", pprof.Index)
mux.HandleFunc("/debug/pprof/profile", pprof.Profile)

go func() {
	// jen na loopback, nikdy na 0.0.0.0
	log.Print(http.ListenAndServe("127.0.0.1:6060", mux))
}()
```

Import `_ "net/http/pprof"` má vedlejší efekt, o kterém musíš vědět: v `init()` si
zaregistruje endpointy na `http.DefaultServeMux`. Pokud tvůj veřejný server běží na
`DefaultServeMux`, právě jsi na internet vystavil zásobníky všech goroutin, argumenty
příkazové řádky a možnost spustit třicetivteřinový CPU profil komukoli. Proto vlastní mux
a vlastní port, dostupný jen přes loopback nebo interní síť.

Programově se profil pořídí přes `runtime/pprof`:

```go
f, _ := os.Create("cpu.out")
pprof.StartCPUProfile(f)
defer pprof.StopCPUProfile()
```

```go
runtime.GC()                              // ať v profilu nejsou mrtvoly
pprof.Lookup("heap").WriteTo(f, 0)        // 0 = protobuf, 1 = čitelný text
```

Výsledek je gzipovaný protobuf — začíná bajty `1f 8b`, takže se dá snadno ověřit i testem.

### Typické nálezy

Většina prvních profilů vypadá stejně. Tohle je pořadí podle četnosti:

```go
// 1. Konverze v cyklu — každý string(r) je alokace
for _, r := range s {
	n, _ := strconv.Atoi(string(r))   // 444 alokací na jeden text
}
for i := 0; i < len(s); i++ {
	if c := s[i]; c >= '0' && c <= '9' { sum += int(c - '0') }  // 0 alokací
}

// 2. Regulární výraz v horké cestě
wordRE.FindAllString(strings.ToLower(text), -1)   // ToLower alokuje celý text

// 3. Chybějící předalokace
m := map[string]int{}                 // roste a rehashuje
m := make(map[string]int, odhad)      // jednou a dost

// 4. string ↔ []byte tam a zpět
b := []byte(s); ...; s2 := string(b)  // dvě kopie; často stačí slice původního stringu
```

Bonus: u malých kolekcí (do zhruba deseti prvků) je lineární průchod slice rychlejší než
mapa, protože se vejde do cache a nemusí se počítat hash.

### Paměť a GC

```go
var ms runtime.MemStats
runtime.ReadMemStats(&ms)
fmt.Println(ms.HeapAlloc, ms.NumGC)   // pozor, ReadMemStats zastaví svět
```

Go má generační-less mark-and-sweep GC laděný dvěma knoflíky:

- `GOGC` (výchozí 100) — GC se spustí, když heap naroste o 100 % proti stavu po minulém
  úklidu. Vyšší hodnota = méně GC, víc paměti.
- `GOMEMLIMIT` — měkký strop celkové paměti. V kontejneru ho nastav zhruba na 90 % limitu
  podu; GC pak bude pracovat víc, místo aby proces zabil OOM killer.

Snížit počet alokací je skoro vždy lepší než ladit GC. Méně objektů znamená kratší značkovací
fázi i menší RSS zároveň.

## Rozdíly proti PHP

V Symfony si zapneš Blackfire nebo profiler z web debug toolbaru, klikneš na požadavek
a dostaneš strom volání. Je to externí nástroj a v produkci ho zapínáš výjimečně.

```php
// obvyklá "optimalizace" v PHP: cache navrch a hotovo
$result = $this->cache->get($key, fn () => $this->slowThing());
```

V Go je profiler součástí runtime a jeho zapnutí stojí jednotky procent výkonu:

```go
import _ "net/http/pprof"      // registruje /debug/pprof/* na DefaultServeMux
```

```bash
go tool pprof http://127.0.0.1:6060/debug/pprof/profile?seconds=30
```

Co se mění v uvažování: **nejdřív změř, potom mysli, teprve pak optimalizuj.** V PHP je
první reflex přidat cache, protože měřit je drahé. V Go je měření tak levné, že přidat
cache bez profilu je čirá lenost — a obvykle to řeší jiný problém, než jaký máš.

## Časté chyby

| Chyba | Proč vzniká | Jak to udělat správně |
|-------|-------------|------------------------|
| Optimalizace bez profilu | intuice z PHP, kde se neměří | nejdřív benchmark, pak profil, pak změna |
| Cache jako první řešení | reflex ze Symfony | cache je poslední krok, ne první |
| `net/http/pprof` na veřejném portu | stačí `import _` a je to tam | vlastní mux, loopback nebo interní port |
| Čtení jen `flat` sloupce | vypadá jako jediný důležitý | `top -cum` odhalí, kdo tu práci zadal |
| Heap profil bez `runtime.GC()` | „snímek je snímek" | bez GC vidíš i objekty, které už nikdo nedrží |
| `ReadMemStats` v horké cestě | vypadá jako levné čtení | zastavuje svět, patří do metrik s periodou |

## AI kvíz

Po přečtení teorie spusť v Cursoru **`/go-deep-quiz 53`**. AI tě ~5 minut prověří mentální model (ne hotové cvičení). Slabiny si uloží do [`GAPS.md`](../../GAPS.md).

## Úkol

Pracuj v `exercise/`. Po doplnění spouštěj testy:

Stupně jdou od jednodušších ke složitějším — po každém stupni spusť review, než jdeš dál.

### Jednoduchý

Funkce: `SumDigitsSlow`, `CountWordsSlow`, `IsWordRune`

```bash
make lesson L=53 PART=1
```

Pak **`/go-deep-review 53 easy`**.

### Střední

Funkce: `SumDigits`, `CountWords`, `JoinIDs`

```bash
make lesson L=53 PART=2
```

Pak **`/go-deep-review 53 medium`**.

### Obtížný

Funkce: `CaptureCPUProfile`, `CaptureHeapProfile`, `PprofHandler`

```bash
make lesson L=53 PART=3
```

Pak **`/go-deep-review 53 hard`**.

Až budou stupně hotové, porovnej se `solutions/` (spoiler).

## Závěrečné otázky

Spusť **`/go-deep-review 53 final`**. AI projde body níže, doptá se a ověří pochopení. Celé cvičení ověří `make lesson L=53` (+ `make race L=53`, pokud to lekce vyžaduje).

- [ ] Umíš vysvětlit rozdíl mezi flat a cum ve výstupu `top`
- [ ] Umíš říct, kdy sáhnout po heap a kdy po allocs profilu
- [ ] Umíš vysvětlit, proč `import _ "net/http/pprof"` může být bezpečnostní chyba
- [ ] Umíš pojmenovat tři typické nálezy z prvního profilu
- [ ] Umíš vysvětlit, co dělá `GOMEMLIMIT` a proč se hodí v kontejneru

## AI režim

`JUNIOR POD REVIEW` — viz [docs/ai-playbook.md](../../docs/ai-playbook.md).

Agent ti ochotně „zoptimalizuje" kód bez jediného měření a výsledek bude nečitelný a
stejně rychlý. Dej mu profil a benchmark jako vstup a vyžaduj, aby změnu obhájil čísly.

## Další čtení

1. [Go blog — Profiling Go Programs](https://go.dev/blog/pprof)
2. [pkg.go.dev — runtime/pprof](https://pkg.go.dev/runtime/pprof)
3. [pkg.go.dev — net/http/pprof](https://pkg.go.dev/net/http/pprof)
4. [Go — A Guide to the Go Garbage Collector](https://go.dev/doc/gc-guide)

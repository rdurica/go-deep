# Lekce 10 — defer, panic, recover

> **Čas:** ~85 min · **Fáze:** 1 — Jazyk a paměťový model · **AI režim:** `ZAKÁZÁNO`

## Co budeš umět

- Předpovědět pořadí a hodnoty argumentů u několika `defer` ve stejné funkci.
- Použít `defer` k úpravě pojmenované návratové hodnoty a vědět, kdy to je a kdy není vhodné.
- Rozhodnout, jestli je daná situace `panic`, nebo `error`.
- Napsat `recover` na správném místě a vysvětlit, proč nefunguje přes hranici goroutiny.
- Vysvětlit, proč `defer f.Close()` u zapisovaného souboru není dost.

## PHP → Go most

V PHP máš `finally`, které se provede vždycky, a bloky se čtou shora dolů.

```php
$fh = fopen($path, 'w');
try {
    fwrite($fh, $data);
} catch (\Throwable $e) {
    $logger->error($e->getMessage());
    throw $e;
} finally {
    fclose($fh);   // úklid je dole, daleko od otevření
}
```

Go nemá `finally` jako blok. Má `defer`, což je **odložené volání registrované u zdroje**.
Úklid píšeš hned vedle získání zdroje, ne o třicet řádků níž.

```go
f, err := os.Create(path)
if err != nil {
	return err
}
defer f.Close()   // úklid hned u otevření, provede se při návratu

_, err = f.Write(data)
return err
```

Přenos návyku: `try/catch` v PHP je běžný nástroj řízení toku. V Go je `panic/recover`
**výjimečný** nástroj — chyby se vracejí jako hodnoty (lekce 14). `defer` naopak budeš
psát pořád, ale ne k chytání chyb, nýbrž k úklidu.

## Teorie

### Sémantika `defer`

`defer` zaregistruje volání, které proběhne, **až se funkce vrací** — ať už `return`,
pádem na konec těla, nebo panikou. Platí tři pravidla:

**1. LIFO.** Odložená volání se provádějí v opačném pořadí registrace.

```go
func main() {
	defer fmt.Println("první")
	defer fmt.Println("druhý")
	defer fmt.Println("třetí")
}
// třetí
// druhý
// první
```

Dává to smysl: zdroje se uklízejí v opačném pořadí, než se otevíraly.

**2. Argumenty se vyhodnotí při registraci, volání proběhne až na konci.**

```go
func main() {
	i := 0
	defer fmt.Println("hodnota při registraci:", i) // 0
	i = 42
	fmt.Println("na konci funkce:", i)              // 42
}
// na konci funkce: 42
// hodnota při registraci: 0
```

Tohle plete každého. Když chceš vidět hodnotu **v okamžiku provedení**, musíš to
zabalit do uzávěru — ten čte proměnnou, až když běží:

```go
i := 0
defer func() { fmt.Println(i) }() // vypíše 42
i = 42
```

Rozdíl mezi `defer f(x)` a `defer func(){ f(x) }()` je jedna z mála věcí, kterou musíš
znát nazpaměť.

**3. `defer` je vázaný na funkci, ne na blok.** Tohle je past v cyklech:

```go
// ŠPATNĚ — všech 10 000 souborů zůstane otevřených až do konce funkce
for _, path := range paths {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	process(f)
}
```

Odložená volání se hromadí a spustí se až na konci celé funkce — do té doby drží
file descriptory a paměť. Řešení je vytáhnout tělo cyklu do funkce:

```go
for _, path := range paths {
	if err := processFile(path); err != nil {
		return err
	}
}

func processFile(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close() // zavře se na konci každé iterace
	return process(f)
}
```

Anonymní funkce volaná na místě funguje taky, ale pojmenovaná funkce je čitelnější.

### `defer` a pojmenované návratové hodnoty

Pokud má funkce **pojmenované** návratové hodnoty, jsou to skutečné proměnné a `defer`
je může po `return` ještě změnit:

```go
func withNamed() (result int) {
	defer func() { result *= 2 }()
	return 5 // nastaví result = 5, pak běží defer, vrací se 10
}

func withoutNamed() int {
	result := 5
	defer func() { result *= 2 }() // změní lokální proměnnou, na návratu se to neprojeví
	return result                  // 5
}
```

`return 5` v Go není jedna operace: nejdřív se přiřadí do návratové proměnné, pak
proběhnou defery, pak funkce skutečně skončí.

Hlavní legitimní použití je doplnění kontextu k chybě nebo úklid, který sám může selhat:

```go
func save(path string, data []byte) (err error) {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer func() {
		if cerr := f.Close(); cerr != nil && err == nil {
			err = cerr // chybu ze zavření nezahodíme
		}
	}()
	_, err = f.Write(data)
	return err
}
```

A tady je odpověď na otázku, proč `defer f.Close()` u zápisu nestačí: zápis do souboru
je bufferovaný a data se dopisují a synchronizují právě při `Close()`. Když `Close`
selže (plný disk, síťový svazek), zápis se **nepovedl**, ale ty se to nedozvíš, protože
`defer f.Close()` návratovou hodnotu zahodí. U čtení je `defer f.Close()` v pořádku, tam
se nemá co pokazit.

Pozor na druhou stranu mince: `defer` s uzávěrem, který mění výsledek, je pro čtenáře
neviditelná změna toku. Používej to na úklid a chyby, ne na počítání.

### `panic` — kdy ano a kdy ne

`panic` odvine zásobník, cestou spustí všechny odložené funkce a program spadne
s výpisem stacku. Není to výjimka, kterou se má běžně řídit tok.

Panika je na místě, když jde o **chybu programátora nebo nemožný stav**:

```go
func mustCompile(pattern string) *regexp.Regexp {
	re, err := regexp.Compile(pattern)
	if err != nil {
		panic("neplatný regulární výraz v konstantě: " + pattern)
	}
	return re
}
```

Proto stdlib nabízí dvojice `regexp.Compile`/`regexp.MustCompile` a
`template.New(...).Parse`/`Must`. Konvence `MustXxx` znamená „paniká místo vracení chyby,
používej jen při inicializaci".

Panika **není** na místě u čehokoli, co může nastat za normálního provozu: neplatný
uživatelský vstup, chybějící soubor, nedostupná databáze, HTTP 404. To všechno je
`error`, protože volající s tím může něco udělat.

Runtime paniká sám v několika případech, které bys měl poznat z výpisu:

| Panika | Kdy |
|--------|-----|
| `index out of range` | `s[5]` na slice délky 3 |
| `nil pointer dereference` | metoda nebo pole na `nil` pointeru |
| `assignment to entry in nil map` | zápis do nevytvořené mapy |
| `integer divide by zero` | `a / b`, kde `b == 0` |
| `interface conversion` | špatný type assert bez comma-ok |
| `close of closed channel` | dvojité zavření kanálu |

### `recover` — jen v `defer`, jen ve své goroutině

`recover()` zastaví paniku a vrátí hodnotu, se kterou se panikovalo. Funguje **jen když
je volaná přímo z odložené funkce**:

```go
func safe() (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("zotaveno z paniky: %v", r)
		}
	}()
	return risky()
}
```

Mimo `defer` vrátí `recover()` vždycky `nil` a nic neudělá.

Dvě věci, které lidi překvapí:

**Recover neplatí přes hranici goroutiny.** Každá goroutina má vlastní zásobník. Panika
v goroutině, kterou tam nikdo neodchytí, shodí **celý program** — i když volající má
`recover`:

```go
func nefunguje() {
	defer func() { recover() }() // tohle goroutinu nezachrání
	go func() {
		panic("bum") // shodí celý proces
	}()
	time.Sleep(time.Second)
}
```

Každá goroutina, která může panikovat, si musí `recover` udělat sama. Ve své první
funkci, hned na začátku.

**Recover po zotavení pokračuje za volající funkcí, ne v místě paniky.** Zbytek těla se
neprovede. Stav, který funkce rozdělala, zůstane rozdělaný — proto se s panikou nedá
pracovat jako s `catch` a pokračovat.

Kde `recover` legitimně patří:

- na hranici, kde jedna vadná operace nesmí shodit celý proces (`net/http` má
  `recover` v každém handleru — proto ti server nespadne kvůli nil pointeru v jednom
  endpointu),
- v knihovně, která uvnitř používá paniku jako rychlý únik z hluboké rekurze a na
  své veřejné hranici ji převede na `error` (tak to dělá `encoding/json`),
- ve workeru, který zpracovává nezávislé úlohy.

Kam **nepatří**: do každé funkce „pro jistotu". Odchycená a zalogovaná panika, po které
program běží dál v nekonzistentním stavu, je horší než pád.

Ještě dvě praktické poznámky. `panic` bere **libovolnou hodnotu**, ne jen string, takže
`recover()` vrací `any` a musíš ho typovat:

```go
if r := recover(); r != nil {
	switch v := r.(type) {
	case error:
		err = fmt.Errorf("zotaveno: %w", v)
	default:
		err = fmt.Errorf("zotaveno: %v", v)
	}
}
```

A když paniku zachytíš, ztratíš s ní i stack trace. Pokud tě zajímá, kde vznikla —
a u loggeru na hranici serveru vždycky zajímá — musíš si ho vyžádat sám přes
`debug.Stack()` z balíčku `runtime/debug`, ještě uvnitř toho odloženého uzávěru.

### Srovnání s PHP

| PHP | Go | Rozdíl |
|-----|----|--------|
| `throw new DomainException` | `return err` | očekávané chyby jsou hodnoty, ne skoky |
| `catch (\Throwable $e)` | `if err != nil` | řeší se na místě, ne o pět rámců výš |
| `finally { … }` | `defer …` | píše se u zdroje, ne pod ním |
| `throw` u chyby programátora | `panic(...)` | jediný skutečný protějšek |
| catch kdekoli v zásobníku | `recover` jen v `defer` své funkce | nelze chytat na dálku |

Nejdůležitější rozdíl je kulturní: v PHP je normální řídit tok výjimkami. V Go je
`panic` v cestě běžného požadavku code smell.

## Časté chyby

| Chyba | Proč vzniká | Jak to udělat správně |
|-------|-------------|------------------------|
| `defer` v cyklu | čte se jako `finally` u bloku | vytáhni tělo do funkce |
| `defer fmt.Println(i)` vypíše starou hodnotu | argumenty se vyhodnotí hned | zabal do `defer func(){ … }()` |
| `defer f.Close()` u zápisu | v PHP `fclose` taky nikdo nekontroluje | uzávěr, který `Close` chybu přiřadí do `err` |
| `panic` na neplatný vstup | reflex z `throw new InvalidArgumentException` | vrať `error` |
| `recover()` mimo `defer` | vypadá jako běžná funkce | volej ho **přímo** z odložené funkce |
| `recover` v goroutině, která ji nespustila | očekávání „catch chytí všechno" | `recover` v každé goroutině zvlášť |
| Zotavení a pokračování v poloviční práci | zvyk pokračovat po `catch` | po `recover` stav zahoď nebo vrať chybu |

## Úkol

Pracuj v `exercise/`. Postupuj A → B → C, po každé části spusť test.

### A — rozcvička (~10 min)

`DeferOrder() []string` — funkce s **pojmenovanou** návratovou hodnotou, která
zaregistruje tři odložené uzávěry. Každý přidá do výsledku svůj řetězec, v pořadí
registrace `"first"`, `"second"`, `"third"`. Návratová hodnota tedy bude
`["third", "second", "first"]`, protože se defery provádějí v LIFO pořadí.

Poznámka: bez pojmenované návratové hodnoty by tohle nešlo napsat, protože `return`
by zafixoval prázdný slice dřív, než defery doběhnou. Zkus si obě varianty.

### B — jádro (~35 min)

1. `SumWithLog(nums []int) (total int, steps []string)` — sečte čísla a vede u toho
   protokol:
   - pro každé číslo přičti a přidej krok ve tvaru `"+3=6"` (přičtená hodnota,
     rovnítko, mezisoučet po přičtení),
   - **před cyklem** zaregistruj `defer`, který na konec protokolu přidá `"total=6"`
     s konečným součtem. Musí to být uzávěr, aby viděl finální hodnotu, a musí měnit
     pojmenovanou návratovou hodnotu `steps`.

   Pro `[]int{1, 2, 3}` tedy dostaneš `6` a `["+1=1", "+2=3", "+3=6", "total=6"]`.
   Pro prázdný vstup `0` a `["total=0"]`.

2. `SafeDivide(a, b int) (result int, err error)` — vydělí `a / b`. Dělení nulou
   v Go panikuje; tvým úkolem je paniku odchytit přes `recover` v odloženém uzávěru
   a převést ji na `error` (stačí `fmt.Errorf`, error model je až lekce 14). Při chybě
   musí být `result` nula. Nesmíš dělitele testovat přes `if b == 0` — smysl cvičení je
   napsat `recover`, ne se panice vyhnout.

3. `CloseAll(closers []func() error) error` — zavolá **všechny** funkce ze slice, i když
   některá vrátí chybu, a vrátí **první** vzniklou chybu (nebo `nil`). `nil` položky
   přeskoč. Prázdný i `nil` vstup vrací `nil`.

### C — rozšíření (~25 min)

Doplň `type Stack struct` (obsah si navrhni sám, hodí se `items []int`) s metodami:

- `func (s *Stack) Push(v int)` — vloží prvek navrch.
- `func (s *Stack) Pop() int` — odebere a vrátí vrchní prvek. Nad prázdným zásobníkem
  **paniká** s hodnotou `"pop from empty stack"`. Je to legitimní panika: volat `Pop`
  na prázdném zásobníku je chyba programátora, ne provozní stav.
- `func (s *Stack) Len() int` — počet prvků. Musí fungovat i na `nil` pointeru.
- `func TryPop(s *Stack) (v int, ok bool)` — bezpečná varianta: zavolá `Pop` a paniku
  odchytí přes `recover`. Při panice vrací `0, false`.

Test ověřuje, že zásobník je **po zotavení dál použitelný** — po neúspěšném `TryPop`
musí `Push` a `Pop` fungovat normálně. To je právě ta hranice, kde je `recover`
v pořádku: nezotavuješ se z poloviční mutace, jen ze zjištění „nic tam nebylo".

```bash
make lesson L=10
```

Až budeš hotový, porovnej se `solutions/` (spoiler).

## Ověření

Po dokončení úkolů spusť v Cursoru **`/go-deep-review`** a zadej třeba jen `10`. AI tě postupně projde body níže, doptá se a ověří pochopení — nestačí jen zelené testy.

- [ ] `make lesson L=10` prochází
- [ ] Umíš z hlavy říct, co vypíše `i := 0; defer fmt.Println(i); i = 42`
- [ ] Umíš vysvětlit, proč je `defer` v cyklu problém a jak ho vyřešit
- [ ] Umíš vysvětlit, jak `defer` mění pojmenovanou návratovou hodnotu
- [ ] Umíš uvést dvě situace, kdy je `panic` správně, a dvě, kdy ne
- [ ] Umíš vysvětlit, proč `recover` nezachrání paniku z jiné goroutiny
- [ ] Umíš vysvětlit, proč `defer f.Close()` u zápisu může tiše ztratit data

## AI režim

`ZAKÁZÁNO` — viz [docs/ai-playbook.md](../../docs/ai-playbook.md).

## Další čtení

1. [Go blog — Defer, Panic, and Recover](https://go.dev/blog/defer-panic-and-recover)
2. [Effective Go — Defer](https://go.dev/doc/effective_go#defer)
3. [Go spec — Handling panics](https://go.dev/ref/spec#Handling_panics)
4. [Go Wiki — Code Review Comments: Don't panic](https://go.dev/wiki/CodeReviewComments#dont-panic)

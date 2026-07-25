# Lekce 22 — Čtení stdlib a cizího kódu

> **Čas:** ~85 min · **Fáze:** 2 — Idiomatický Go · **AI režim:** `JEN VYSVĚTLENÍ`

## Co budeš umět

- Najít odpověď ve zdrojáku stdlib rychleji než ve vyhledávači.
- Vysvětlit trik `http.HandlerFunc` a napsat stejný adaptér pro vlastní interface.
- Popsat pravidlo výběru vzoru v `http.ServeMux` a implementovat ho.
- Odhadnout, co dělá `encoding/json` uvnitř, a proč k tomu potřebuje reflexi.
- Zorientovat se v neznámém repozitáři za deset minut a rozhodnout, jestli si ho pustíš
  do `go.mod`.

## PHP → Go most

V PHP se do zdrojáků knihoven chodí nerado. Symfony komponenta je vrstva abstrakcí,
`vendor/` má sto tisíc souborů a odpověď stejně najdeš v dokumentaci na webu.

```php
// Co přesně dělá tenhle řádek? Odpověď je pět tříd a dva compiler passy hluboko.
$response = $this->httpClient->request('GET', $url);
```

V Go je čtení zdrojáku **první, ne poslední** volba. Stdlib je psaná tak, aby se četla:
malé soubory, žádná dědičnost, doc comment nad každou exportovanou věcí a testy vedle.

```go
// go doc net/http Handler
// type Handler interface {
//     ServeHTTP(ResponseWriter, *Request)
// }
```

Návyk k opuštění: **přestaň hledat tutoriál a otevři zdroják.** V devíti z deseti případů
je kratší než blogpost o něm.

## Teorie

### Nástroje: `go doc`, pkg.go.dev, skok do zdrojáku

Tři vstupní body, každý na jinou otázku:

```bash
go doc net/http                 # přehled balíčku, seznam typů a funkcí
go doc net/http.ServeMux        # typ a jeho metody
go doc net/http.ServeMux.Handle # jedna metoda včetně komentáře
go doc -src net/http.HandlerFunc # rovnou implementace
go doc -all encoding/json | less # úplně všechno
```

`go doc` funguje offline a na verzi, kterou máš skutečně nainstalovanou — to je rozdíl
proti webu, kde snadno čteš dokumentaci k jiné verzi. [pkg.go.dev](https://pkg.go.dev)
použij, když chceš proklikat odkazy a podívat se na příklady (`Example` funkce v testech
se renderují jako spustitelné ukázky). V IDE si nastav skok do definice a používej ho
i na stdlib — zdrojáky máš lokálně v `$(go env GOROOT)/src`.

Doc comment je v Go zdroj pravdy, ne dekorace. Když komentář a implementace nesouhlasí,
je to bug hlášení hodný. A druhý zdroj pravdy jsou **testy**: `net/http/serve_test.go`
ti řekne o chování `ServeMux` víc než jakýkoli článek, protože popisuje i případy,
o kterých by ses sám nezamyslel.

### Případová studie: `net/http` Handler a HandlerFunc

Celý `net/http` stojí na jednom dvouřádkovém interface a jednom triku.

```go
type Handler interface {
	ServeHTTP(ResponseWriter, *Request)
}

type HandlerFunc func(ResponseWriter, *Request)

// ServeHTTP calls f(w, r).
func (f HandlerFunc) ServeHTTP(w ResponseWriter, r *Request) {
	f(w, r)
}
```

Přečti si to pomalu. `HandlerFunc` je **pojmenovaný funkční typ** a metoda na něm volá
sama sebe jako funkci. Důsledek: libovolnou funkci se správnou signaturou proměníš
v `Handler` jednou konverzí — `http.HandlerFunc(myFunc)`. Žádná třída, žádné `implements`,
žádný adapter objekt.

Tenhle vzor uvidíš v Go pořád: `sort.Reverse`, `http.HandlerFunc`, `context.CancelFunc`,
`io/fs.WalkDirFunc`. Metoda na funkčním typu je nejlevnější adaptér, jaký jazyk nabízí,
a v úkolu A si ho napíšeš sám.

`ServeMux` je pak jen `Handler`, který drží mapu vzorů a deleguje. Pravidlo výběru
(v klasické podobě, tedy bez metod a wildcardů z Go 1.22 — ty přijdou v lekci 25):

- vzor **bez** koncového lomítka odpovídá jen přesné cestě;
- vzor **s** koncovým lomítkem odpovídá celému podstromu;
- vyhrává **nejdelší** vyhovující vzor, ne první nebo poslední registrovaný;
- `ServeMux` sám implementuje `Handler`, takže muxy jdou vnořovat.

Poslední bod je elegantní: router není zvláštní kategorie objektu, je to obyčejný handler.

### Případová studie: `encoding/json`

`json.Unmarshal(data, &v)` dostane `any` a ukazatel. V době kompilace neví nic o typu,
který dekóduje — takže musí za běhu **zjistit strukturu hodnoty**. Na to slouží reflexe
(`reflect.Type`, `reflect.Value`): projde pole struktury, přečte tagy
(`json:"name,omitempty"`), spáruje je s klíči a nastaví hodnoty přes `reflect.Value.Set`.

Z toho plyne všechno, co tě na `encoding/json` kdy zaskočí:

- **musíš předat ukazatel** — jinak by dekodér měnil kopii;
- **neexportovaná pole se ignorují** — reflexe je nesmí nastavit;
- **je to pomalejší** než ručně psaný kodek, proto existují generátory;
- **tag je jediný způsob**, jak dekodéru sdělit jméno, protože jinak zná jen jméno pole.

Naproti tomu **kódování** reflexi nutně nepotřebuje. Kdyby šlo o pár konkrétních typů,
stačí `switch v := v.(type)` a rekurze — a přesně to si napíšeš v úkolu C. Až uvidíš,
kolik kódu je potřeba na čtyři typy, pochopíš, proč `encoding/json` sáhl po reflexi.

### Jak číst cizí repozitář

Pořadí, které funguje skoro vždy:

1. **README** — co to je, ne jak to funguje. Pět minut.
2. **`go.mod`** — verze Go a hlavně **počet závislostí**. Balíček s třiceti závislostmi
   na sobě prozradí víc než README.
3. **`cmd/`** — pokud existuje, `main` ti ukáže, jak se to skutečně skládá dohromady.
4. **kořenové `.go` soubory** — v Go knihovnách je veřejné API typicky v kořeni balíčku,
   často v souboru se jménem podle balíčku (`chi.go`, `zap.go`).
5. **`internal/`** — implementace, kterou autor **záměrně** nevystavil. Když je toho tam
   hodně, je to dobré znamení o disciplíně autora.
6. **testy** — `_test.go` u veřejného API ti dá skutečné použití, ne marketingové ukázky.

Před přidáním do `go.mod` se zeptej na pět věcí: kdy byl poslední commit, kolik má
závislostí, má tagovanou verzi `v1`+, kolik má otevřených issues bez odpovědi, a existuje
řešení ve stdlib. Poslední otázka je nejdůležitější — v Go je odpověď překvapivě často ano.

## Časté chyby

| Chyba | Proč vzniká | Jak to udělat správně |
|-------|-------------|------------------------|
| Hledání odpovědi na webu místo `go doc` | zvyk na Symfony docs | `go doc` je offline a na správné verzi |
| Ignorování `_test.go` v cizím repu | „testy jsou pro autora" | testy jsou nejpřesnější dokumentace |
| Přidání knihovny bez pohledu do `go.mod` | Composer reflex | zkontroluj strom závislostí |
| Framework místo `net/http` | zvyk, že bez frameworku to nejde | `net/http` je router i server |
| Struktura s malými písmeny do JSON | „vždyť je to pole" | reflexe neexportovaná pole nesmí nastavit |
| `json.Unmarshal(data, v)` bez `&` | v PHP je objekt reference | vždycky ukazatel |

## Úkol

Pracuj v `exercise/`. Postupuj A → B → C, po každé části spusť test.

Celá tahle sada je „přepiš si stdlib nanečisto". Než začneš, otevři si
`go doc -src net/http.HandlerFunc` a `go doc net/http.ServeMux.Handler` — je to legální
nápověda, a přesně o tom lekce je.

### A — rozcvička (~10 min)

Interface `Handler` a typ `HandlerFunc` už v `exercise.go` jsou. Doplň jedinou metodu:

```go
func (f HandlerFunc) Handle(path string) string
```

Musí platit, že `var h Handler = HandlerFunc(nejakaFunkce)` se zkompiluje a `h.Handle("/x")`
zavolá tu funkci. Je to jeden řádek — ale musíš pochopit, proč funguje.

### B — jádro (~35 min)

`Mux` je zjednodušený `ServeMux`. Jeho **zero value je použitelná**, konstruktor nepiš
(`var mux Mux` musí fungovat, mapu inicializuj líně).

`Register(pattern string, h Handler)`:

- panikuje `"mux: empty pattern"` u prázdného vzoru;
- panikuje `"mux: nil handler"` u `nil` handleru;
- panikuje `"mux: multiple registrations for " + pattern` při druhé registraci téhož vzoru.

Panika je tady správně, protože registrace rout se dělá při startu z konstant v kódu —
je to chyba programátora, ne vstupu. `net/http` to dělá úplně stejně.

`SetNotFound(h Handler)` uloží fallback; `nil` panikuje `"mux: nil handler"`.

`Handle(path string) string`:

- vzor bez koncového `/` odpovídá jen přesně stejné cestě;
- vzor s koncovým `/` odpovídá každé cestě, která jím začíná;
- vyhrává **nejdelší** vyhovující vzor;
- handler dostane **celou** cestu, ne zbytek za vzorem;
- bez shody se volá notFound handler, a když není nastavený, vrátí se
  `"404 not found: " + path`.

Ověř si na testech tuhle pointu: při vzorech `/`, `/api/` a `/api/users` jde `/api`
na `/` (protože `/api/` není prefix `/api`), zatímco `/api/users/42` jde na `/api/`.

Protože `Handle` má stejnou signaturu jako metoda v `Handler`, tvůj `*Mux` sám splňuje
`Handler` a půjde vnořit do jiného muxu. Nic pro to dělat nemusíš — a to je ta lekce.

### C — rozšíření (~25 min)

`Marshal(v any) (string, error)` — mini `encoding/json` **bez reflexe**, jen přes
type switch.

| Vstup | Výstup |
|---|---|
| `nil` | `null` |
| `"ahoj"` | `"ahoj"` |
| `42`, `-42`, `0` | `42`, `-42`, `0` |
| `[]string{"a","b"}` | `["a","b"]` |
| `[]string{}` | `[]` |
| `[]string(nil)` | `null` |
| `map[string]string{"b":"2","a":"1"}` | `{"a":"1","b":"2"}` |
| `map[string]string{}` | `{}` |
| `map[string]string(nil)` | `null` |
| cokoli jiného | `""` a chyba `fmt.Errorf("%w: %T", ErrUnsupportedType, v)` |

Detaily, na kterých to stojí:

- Klíče mapy se **řadí vzestupně**. Iterace přes mapu je v Go náhodná, takže bez řazení
  bys měl nedeterministický výstup — `encoding/json` řadí ze stejného důvodu.
- Rozdíl mezi `nil` a prázdným slicem/mapou je záměrný a odpovídá chování `encoding/json`.
- Escapují se čtyři znaky: `"` → `\"`, `\` → `\\`, nový řádek → `\n`, tabulátor → `\t`.
- Text skládej přes `strings.Builder`.

Až budeš hotový, spusť `go doc -src encoding/json.Marshal` a najdi, kde se ta funkce
větví na reflexi. Kolik řádků má tvoje verze a kolik ta jejich?

```bash
make lesson L=22
```

Až budeš hotový, porovnej se `solutions/` (spoiler).

## Ověření

Po dokončení úkolů spusť v Cursoru **`/go-deep-review`** a zadej třeba jen `22`. AI tě postupně projde body níže, doptá se a ověří pochopení — nestačí jen zelené testy.

- [ ] `make lesson L=22` prochází
- [ ] Umíš vysvětlit, proč `HandlerFunc(f).Handle(x)` zavolá `f(x)`
- [ ] Umíš vysvětlit, proč `*Mux` splňuje `Handler`, aniž jsi to kdekoli napsal
- [ ] Umíš vyjmenovat tři pravidla výběru vzoru v `ServeMux`
- [ ] Umíš vysvětlit, proč `json.Unmarshal` potřebuje ukazatel a ignoruje malá pole
- [ ] Umíš popsat, v jakém pořadí čteš neznámý repozitář
- [ ] Umíš `go doc` použít na typ, metodu i zdroják

## AI režim

`JEN VYSVĚTLENÍ` — viz [docs/ai-playbook.md](../../docs/ai-playbook.md).

Zajímavý prompt pro tuhle lekci: *„Vysvětli, proč má `http.HandlerFunc` metodu se stejným
jménem jako interface, který splňuje. Nenabízej kód."* Zákaz zní: nenech si vygenerovat
`Mux` ani `Marshal` — obojí je krátké, a právě proto se to vyplatí napsat rukou.

## Další čtení

1. [pkg.go.dev — net/http](https://pkg.go.dev/net/http) — proklikej se ze `Handler` do zdrojáku
2. [Go blog — JSON and Go](https://go.dev/blog/json)
3. [Go blog — The Laws of Reflection](https://go.dev/blog/laws-of-reflection)
4. [Effective Go — Interfaces and methods](https://go.dev/doc/effective_go#interface_methods)

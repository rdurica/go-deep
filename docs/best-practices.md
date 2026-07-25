# Best practices — konsolidované idiomy

Rychlá reference přes celý kurz. Sloupec „Lekce" ukazuje, kde se pravidlo probírá
do hloubky. Formát v lekcích: **pravidlo → proč → anti-příklad → idiomatický příklad
→ kdy výjimka**.

## Jazyk a styl

| Pravidlo | Proč | Lekce |
|----------|------|-------|
| Vždy `gofmt` | jedna pravda o formátování, žádné debaty v PR | 01 |
| Zero value má být použitelná | ušetří konstruktor a chybové stavy | 03, 20 |
| Žádné implicitní konverze — konvertuj vědomě | tichá ztráta přesnosti je horší než ukecaný kód | 03 |
| Pointer jen když sdílíš nebo mutuješ | value semantics je default, ne výjimka | 06 |
| `append` nemusí mutovat originál | backing array se může realokovat | 07 |
| Do `nil` mapy se nezapisuje | zero value mapy je jen pro čtení | 08 |
| `len()` na stringu vrací bajty | UTF-8, ne znaky — u češtiny to bolí hned | 09 |
| `defer` pro úklid, ne pro control flow | LIFO, argumenty se vyhodnotí hned | 10 |
| Balíček je jednotka zapouzdření | export je velké písmeno, ne `public` | 11 |
| Malé interfacy (1–3 metody) | snadná implementace i testovatelnost | 12, 13, 33 |
| Interface definuj u konzumenta | ten ví, co potřebuje; implementace ne | 33 |
| Accept interfaces, return structs | volnost na vstupu, konkrétnost na výstupu | 20 |
| Chyba je hodnota, ne výjimka | je součástí signatury a tedy kontraktu | 14 |
| Wrapuj s `%w` a přidej kontext | `errors.Is`/`As` musí projít celým řetězem | 14, 21 |
| Chybové texty malým písmenem, bez tečky | skládají se do delších vět | 21 |
| Generalizuj až při třetím výskytu | generika nejsou cíl, jsou nástroj | 15 |
| MixedCaps, zkratky celé velké (`URL`, `ID`) | konzistence se stdlib | 19 |
| Žádné `utils` / `common` / `helpers` | balíček musí mít účel, ne být skládka | 19, 32 |
| Balíčky podle domény, ne podle vrstvy | layer cake je Symfony reflex | 32 |
| Early return, happy path vlevo | zanoření skrývá logiku | 19 |

## Concurrency

| Pravidlo | Proč | Lekce |
|----------|------|-------|
| Kdo goroutinu spustil, odpovídá za její konec | jinak vzniká leak, který nic nenahlásí | 40 |
| Kanál zavírá odesílatel, nikdy příjemce | zápis do zavřeného kanálu panikuje | 41 |
| Směrové typy `chan<-` / `<-chan` v signatuře | kontrakt je vidět bez čtení těla | 41 |
| `time.After` v cyklu je leak | použij `NewTimer` a `Stop` | 42 |
| Zámek pro stav, kanál pro předání vlastnictví | správný nástroj na správný problém | 43 |
| Strukturu se zámkem nikdy nekopíruj | `go vet` to naštěstí odhalí | 43 |
| `-race` je součást Definition of Done | závod není „občas špatné číslo", je to UB | 44 |
| Omez souběžnost explicitně | neomezený fan-out je výpadek čekající na příležitost | 46 |
| První chyba ruší zbytek | jinak platíš za práci, kterou zahodíš | 47 |
| Viditelnost zajišťuje synchronizace, ne čas | happens-before je uspořádání | 48 |

## HTTP a produkce

| Pravidlo | Proč | Lekce |
|----------|------|-------|
| `net/http` first, framework až když víš, co skrývá | od 1.22 stačí stdlib na většinu služeb | 24, 25 |
| Timeouty na serveru i klientovi | `http.DefaultClient` nemá žádný | 24, 30 |
| Vždy `defer resp.Body.Close()` a tělo dočti | jinak se spojení nerecykluje | 30 |
| Graceful shutdown na signál | v kontejneru to není luxus | 30, 55 |
| `context.Context` první parametr, nikdy ve structu | lifetime requestu ≠ lifetime objektu | 27 |
| Strukturované logy přes `log/slog`, na stdout | grep na volný text je slepá ulička | 29 |
| Tajemství jen z prostředí a nikdy do logu | maskuj je i v `String()` | 28, 37 |
| Validuj na hranici, invarianty vynucuj typem | parse, don't validate | 36 |
| Konstantní porovnání tajemství | `==` na token je časový únik | 37 |
| Nízká kardinalita labelů metrik | `user_id` jako label zabije Prometheus | 37 |
| Peníze v celých minoritních jednotkách | `float64` na fakturách nikdy | 34 |
| Hodnoty jen přes placeholdery | jména sloupců whitelistem, ne konkatenací | 35 |
| Liveness ≠ readiness | jinak Kubernetes restartuje zdravý pod | 55 |

## Testování

| Pravidlo | Proč | Lekce |
|----------|------|-------|
| Table-driven testy a `t.Run` | čitelnost a pokrytí hraničních případů | 17 |
| Fake místo mockovacího frameworku | v Go je fake obvykle kratší než mock setup | 33 |
| Testuj přes veřejné API (`package foo_test`) | testuješ kontrakt, ne implementaci | 17 |
| Jedno měření není benchmark | porovnávej, neoslavuj absolutní čísla | 52 |
| Fuzz na round-trip a „nepanikuje" | najde vstupy, na které bys nepomyslel | 52 |
| Profiluj až když máš reprodukovatelný benchmark | jinak optimalizuješ dojmy | 53 |

## PHP → Go pasty (rychlý přehled)

Detail v [php-to-go.md](php-to-go.md).

1. **Výjimky** → `error` hodnoty a explicitní větvení.
2. **Dědičnost** → kompozice a malé interfacy.
3. **DI kontejner** → konstruktory a ruční wiring v `main`.
4. **Všechno reference** → value semantics jako default.
5. **Bohaté service interfacy** → interface u konzumenta, často jednometodový.
6. **Layer balíčky** (`Controller`/`Service`/`Repository`) → doménové balíčky (`billing`, `auth`).
7. **`null` všude** → zero value, nebo explicitní `(hodnota, bool)`.
8. **Anotace a atributy** → explicitní kód; tagy jen pro serializaci.

## Kdy pravidlo porušit

Idiomy nejsou dogma. Porušení musíš umět **obhájit** — výkon změřený benchmarkem,
generovaný kód, FFI, kompatibilita. „AI to tak napsala" a „takhle to dělám v Symfony"
nestačí.

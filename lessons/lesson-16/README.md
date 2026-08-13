# Lekce 16 — JSON: marshal, unmarshal, tagy

> **Čas:** ~70 min · **Fáze:** 1 — Jazyk a paměťový model · **AI režim:** `ZAKÁZÁNO`

## Co budeš umět

- Namapovat Go struct na konkrétní JSON tvar pomocí tagů a vědět, proč `omitempty` občas zahodí data, která jsi poslat chtěl.
- Rozhodnout mezi `json.Marshal` a `json.Encoder` (a mezi `Unmarshal` a `Decoder`) podle toho, jestli máš v ruce bajty, nebo stream.
- Odložit dekódování části dokumentu přes `json.RawMessage` a rozhodnout se podle diskriminátoru.
- Vysvětlit, proč se v `map[string]any` z čísel stanou `float64`, a kdy místo toho sáhnout po `json.Number`.
- Napsat vlastní `MarshalJSON`/`UnmarshalJSON` pro doménový typ a zapnout striktní dekódování na hranici API.

## Teorie

### Tagy a exportovaná pole

`encoding/json` pracuje přes reflexi, a ta vidí jen **exportovaná** pole. Neexportované
pole (malé počáteční písmeno) balíček tiše přeskočí, ať už mu dáš jakýkoli tag. Tohle je
nejčastější „proč mi tam to pole není" u nováčka.

```go
type Order struct {
	ID     int     `json:"id"`
	Total  float64 `json:"total"`
	note   string  `json:"note"` // NIKDY se neserializuje — je neexportované
	Secret string  `json:"-"`    // vynecháno záměrně
}
```

Tag má tvar `json:"jméno,volby"`. Užitečné volby:

| Tag | Význam |
|-----|--------|
| `json:"name"` | přejmenování klíče |
| `json:"-"` | pole se ignoruje oběma směry |
| `json:"-,"` | klíč se opravdu jmenuje `-` (vzácné, ale existuje) |
| `json:"name,omitempty"` | vynech, pokud je hodnota prázdná |
| `json:"name,string"` | číslo/bool zapiš jako řetězec (`"42"`) |

Bez tagu se použije jméno pole tak, jak je (`Total` → `"Total"`). Při dekódování je
párování klíčů **case-insensitive**, takže `{"id":1}` i `{"ID":1}` naplní `ID`.

### Past jménem `omitempty`

`omitempty` neznamená „vynech, pokud to uživatel nevyplnil". Znamená „vynech, pokud je
hodnota zero value": `0`, `false`, `""`, `nil`, prázdný slice nebo mapa. U struct pole
nedělá **nic** — prázdný struct se serializuje vždy.

```go
type Filter struct {
	Page    int  `json:"page,omitempty"`    // page=0 zmizí
	Enabled bool `json:"enabled,omitempty"` // enabled=false zmizí
}

data, _ := json.Marshal(Filter{Page: 0, Enabled: false})
fmt.Println(string(data)) // {}
```

Pokud musíš rozlišit „nula" od „nevyplněno", modeluj to explicitně pointerem:

```go
type Filter struct {
	Enabled *bool `json:"enabled,omitempty"`
}
```

`nil` = klíč chyběl, `*Enabled == false` = přišlo `false`. Je to stejná úvaha jako
v lekci 03 o zero values, jen na hranici API.

### `Marshal` vs `Encoder`, `Unmarshal` vs `Decoder`

Dvojice funkcí a typů dělá skoro totéž, ale nad jiným vstupem:

```go
// bajty v paměti
data, err := json.Marshal(v)
err = json.Unmarshal(data, &v)

// stream (io.Writer / io.Reader)
err = json.NewEncoder(w).Encode(v)
err = json.NewDecoder(r).Decode(&v)
```

Pravidlo: **máš-li `io.Reader`/`io.Writer`, použij Decoder/Encoder.** V HTTP handleru je
to skoro vždy on — `json.NewDecoder(r.Body).Decode(&req)` nenačte tělo celé do paměti
a `json.NewEncoder(w).Encode(resp)` nemusí stavět mezibuffer.

Dva detaily, které překvapí:

- `Encode` **přidá na konec `\n`**. Když porovnáváš výstup v testu, počítej s tím.
- `Decode` přečte jen **první** JSON hodnotu ze streamu. Zbytek tě nezajímá, dokud si
  ho nevyžádáš přes `dec.More()`. Toho se dá využít pro NDJSON, ale taky to znamená, že
  `{"a":1}garbage` projde bez chyby, kdežto `json.Unmarshal` na tom spadne.

Decoder umí navíc dvě věci, které jinak nemáš:

```go
dec := json.NewDecoder(r)
dec.DisallowUnknownFields() // neznámý klíč = chyba
dec.UseNumber()             // čísla jako json.Number, ne float64
```

`DisallowUnknownFields` je levná validace na hranici: klient, který pošle překlep
`{"emial": "..."}`, dostane 400 místo tichého ignorování.

### `json.RawMessage` a odložené dekódování

Když dopředu nevíš, jaký typ payload má, nedekóduj ho hned. `json.RawMessage` je
`[]byte`, který si syrový JSON jen uschová:

```go
type Event struct {
	Kind    string          `json:"kind"`
	Payload json.RawMessage `json:"payload"`
}

var ev Event
if err := json.Unmarshal(data, &ev); err != nil {
	return err
}
switch ev.Kind {
case "user.created":
	var p UserCreated
	if err := json.Unmarshal(ev.Payload, &p); err != nil {
		return fmt.Errorf("event %q: %w", ev.Kind, err)
	}
	// ...
}
```

Je to Go verze discriminator mapy z API Platform, jen viditelná a laditelná. Výhoda proti
dekódování do `map[string]any`: payload zůstane bajt po bajtu původní, takže se
nezkomolí velká čísla ani pořadí a můžeš ho třeba jen přeposlat dál.

### `map[string]any`, `float64` a `json.Number`

Když dekóduješ do `any`, dostaneš přesně šest typů: `nil`, `bool`, `string`, `float64`,
`[]any` a `map[string]any`. **Všechna čísla jsou `float64`** — JSON žádný `int` nezná
a balíček nemá kam se podívat, co jsi chtěl.

```go
var m map[string]any
json.Unmarshal([]byte(`{"id":9007199254740993}`), &m)
fmt.Println(m["id"]) // 9.007199254740992e+15 — o jedničku vedle
```

Pro velká ID (Twitter/Stripe styl) použij `UseNumber()`; `json.Number` je řetězec
s metodami `Int64()` a `Float64()`, takže o přesnost nepřijdeš.

### Vlastní `MarshalJSON` / `UnmarshalJSON` a `time.Time`

Když chceš, aby doménový typ v JSON vypadal jinak než uvnitř, implementuj rozhraní
`json.Marshaler` a `json.Unmarshaler`. Přesně tak to dělá `time.Time`, který se sám
serializuje do RFC 3339 (`"2024-03-01T12:30:00Z"`) — proto v JSON nevidíš jeho vnitřní
pole.

```go
type Cents int64

func (c Cents) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("%d.%02d", int64(c)/100, int64(c)%100)), nil
}

func (c *Cents) UnmarshalJSON(data []byte) error {
	f, err := strconv.ParseFloat(string(data), 64)
	if err != nil {
		return fmt.Errorf("cents: %w", err)
	}
	*c = Cents(math.Round(f * 100))
	return nil
}
```

Dvě pravidla, na kterých se to nejčastěji láme:

1. `MarshalJSON` piš na **hodnotovém** receiveru, jinak se nepoužije u pole typu `Cents`
   (jen u `*Cents`). `UnmarshalJSON` naopak **musí** být na pointeru — jinak nemá co změnit.
2. Vracený bajtový výstup musí být platný JSON, jinak `json.Marshal` selže.

`time.Time` v JSON má ještě jednu pastičku: zero value se serializuje jako
`"0001-01-01T00:00:00Z"` a `omitempty` na ni **nezabere** (struct není nikdy „empty").
Chceš-li volitelné datum, použij `*time.Time`.

## Rozdíly proti PHP

V Symfony máš Serializer: nakonfiguruješ normalizery, na property dáš atribut a komponenta
za tebe vyřeší viditelnost, přejmenování i skupiny.

```php
final class User
{
    public function __construct(
        #[Groups(['read'])] public int $id,
        #[SerializedName('full_name')] public string $name,
        #[Ignore] public string $password,
    ) {}
}

$json = $serializer->serialize($user, 'json', ['groups' => ['read']]);
```

V Go je to jedna funkce ze standardní knihovny a několik znaků v tagu:

```go
type User struct {
	ID       int    `json:"id"`
	Name     string `json:"full_name"`
	Password string `json:"-"`
}

data, err := json.Marshal(u)
```

Co se mění v uvažování: **Serializer je runtime služba, `encoding/json` je funkce nad
reflexí bez konfigurace.** Nejsou skupiny, nejsou kontexty, není DI. Když potřebuješ dva
různé tvary téhož objektu, nevymýšlíš groups — napíšeš druhý struct (typicky DTO na
hranici HTTP) a mapování mezi nimi napíšeš rukou. Zní to jako krok zpět; ve skutečnosti
tím zmizí celá třída otázek typu „proč se to pole v produkci neserializovalo".

## Časté chyby

| Chyba | Proč vzniká | Jak to udělat správně |
|-------|-------------|------------------------|
| Neexportované pole s tagem | v PHP je `private` jen konvence pro serializer | pole musí začínat velkým písmenem |
| `omitempty` na `bool`/`int` | čeká se „vynech, když chybí" | pointer, nebo tag zahodit |
| `Unmarshal(data, v)` bez `&` | v PHP se objekt předává referencí | vždy pointer: `Unmarshal(data, &v)` |
| Dekódování do `map[string]any` a přetypování na `int` | zvyk na volnou typovost PHP polí | dekóduj do structu, případně `UseNumber()` |
| Tichý průchod překlepu v klíči | Serializer to taky ignoroval | `dec.DisallowUnknownFields()` na hranici |
| `MarshalJSON` na pointer receiveru | reflex „metody na pointeru" | hodnotový receiver pro Marshal, pointer pro Unmarshal |

## AI kvíz

Po přečtení teorie spusť v Cursoru **`/go-deep-quiz 16`**. AI tě ~5 minut prověří mentální model (ne hotové cvičení). Slabiny si uloží do [`GAPS.md`](../../GAPS.md).

## Úkol

Pracuj v `exercise/`. Kontrakt je v komentáři nad funkcí. Stupně jdou od jednodušších
ke složitějším — po každém stupni spusť review, než jdeš dál.

### Jednoduchý

Oprav: `FromJSON` (kód je záměrně vadný — jméno nekontroluje po TrimSpace)

```bash
make lesson L=16 PART=1
```

Pak **`/go-deep-review 16 easy`**.

### Střední

Implementuj: `DecodeEvent` (`json.RawMessage` + diskriminátor)

```bash
make lesson L=16 PART=2
```

Pak **`/go-deep-review 16 medium`**.

### Obtížný

Doplň: `StrictDecode` (`DisallowUnknownFields` + kontrola zbytku vstupu)

```bash
make lesson L=16 PART=3
```

Pak **`/go-deep-review 16 hard`**.

Až budou stupně hotové, porovnej se `solutions/` (spoiler).

## Závěrečné otázky

Spusť **`/go-deep-review 16 final`**. AI projde body níže, doptá se a ověří pochopení. Celé cvičení ověří `make lesson L=16` (+ `make race L=16`, pokud to lekce vyžaduje).

- [ ] Umíš vysvětlit, proč se neexportované pole neserializuje ani s tagem
- [ ] Umíš vysvětlit, kdy `omitempty` zahodí data, o která jsi přijít nechtěl
- [ ] Umíš vysvětlit rozdíl mezi `json.Unmarshal` a `json.Decoder` u zbytku vstupu
- [ ] Umíš vysvětlit, proč jsou čísla v `map[string]any` typu `float64`
- [ ] Umíš vysvětlit, proč je `MarshalJSON` na hodnotě a `UnmarshalJSON` na pointeru

## AI režim

`ZAKÁZÁNO` — viz [docs/ai-playbook.md](../../docs/ai-playbook.md).

Mentor, kvíz i review (dialog) jsou vždy OK; v tomto režimu AI nesmí psát kód cvičení.
## Další čtení

1. [pkg.go.dev — encoding/json](https://pkg.go.dev/encoding/json)
2. [Go blog — JSON and Go](https://go.dev/blog/json)
3. [pkg.go.dev — json.RawMessage](https://pkg.go.dev/encoding/json#RawMessage)
4. [pkg.go.dev — json.Number](https://pkg.go.dev/encoding/json#Number)

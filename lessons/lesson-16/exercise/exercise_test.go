package exercise_test

import (
	"encoding/json"
	"math"
	"strings"
	"testing"
	"time"

	exercise "github.com/rdurica/go-deep/lessons/lesson-16/exercise"
)

func sampleUser() exercise.User {
	return exercise.User{
		ID:        7,
		Name:      "Ada",
		Email:     "ada@example.com",
		Active:    true,
		Tags:      []string{"admin", "beta"},
		CreatedAt: time.Date(2024, time.March, 1, 12, 30, 0, 0, time.UTC),
		Password:  "tajne-heslo",
	}
}

func TestToJSONKeys(t *testing.T) {
	data, err := exercise.ToJSON(sampleUser())
	if err != nil {
		t.Fatalf("ToJSON(...) = _, %v, chci nil", err)
	}

	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("výstup ToJSON není platný JSON: %v (%s)", err, data)
	}

	want := map[string]any{
		"id":         float64(7),
		"name":       "Ada",
		"email":      "ada@example.com",
		"active":     true,
		"created_at": "2024-03-01T12:30:00Z",
	}
	for k, w := range want {
		got, ok := m[k]
		if !ok {
			t.Errorf("ve výstupu chybí klíč %q (%s)", k, data)
			continue
		}
		if got != w {
			t.Errorf("klíč %q = %#v, chci %#v", k, got, w)
		}
	}

	if _, ok := m["password"]; ok {
		t.Errorf("klíč password nemá být ve výstupu vůbec (%s)", data)
	}
	if _, ok := m["Password"]; ok {
		t.Errorf("klíč Password nemá být ve výstupu vůbec (%s)", data)
	}
	if _, ok := m["tags"]; !ok {
		t.Errorf("neprázdné tagy mají být ve výstupu (%s)", data)
	}
	if strings.Contains(string(data), "tajne-heslo") {
		t.Errorf("heslo nesmí uniknout do JSON: %s", data)
	}
}

func TestToJSONOmitempty(t *testing.T) {
	u := exercise.User{
		ID:        1,
		Name:      "Bob",
		CreatedAt: time.Date(2020, time.January, 2, 3, 4, 5, 0, time.UTC),
	}
	data, err := exercise.ToJSON(u)
	if err != nil {
		t.Fatalf("ToJSON(...) = _, %v, chci nil", err)
	}

	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("výstup ToJSON není platný JSON: %v (%s)", err, data)
	}
	if _, ok := m["email"]; ok {
		t.Errorf("prázdný email má být vynechán (%s)", data)
	}
	if _, ok := m["tags"]; ok {
		t.Errorf("nil tags mají být vynechány (%s)", data)
	}
	got, ok := m["active"]
	if !ok {
		t.Fatalf("active nemá omitempty, musí být ve výstupu i s hodnotou false (%s)", data)
	}
	if got != false {
		t.Errorf("active = %#v, chci false", got)
	}
}

func TestFromJSONValid(t *testing.T) {
	in := sampleUser()
	data, err := exercise.ToJSON(in)
	if err != nil {
		t.Fatalf("ToJSON(...) = _, %v, chci nil", err)
	}

	got, err := exercise.FromJSON(data)
	if err != nil {
		t.Fatalf("FromJSON(%s) = _, %v, chci nil", data, err)
	}
	if got.ID != in.ID || got.Name != in.Name || got.Email != in.Email || got.Active != in.Active {
		t.Errorf("FromJSON(...) = %+v, chci skalární pole shodná s %+v", got, in)
	}
	if !got.CreatedAt.Equal(in.CreatedAt) {
		t.Errorf("CreatedAt = %v, chci %v", got.CreatedAt, in.CreatedAt)
	}
	if len(got.Tags) != len(in.Tags) {
		t.Fatalf("Tags = %v, chci %v", got.Tags, in.Tags)
	}
	for i := range in.Tags {
		if got.Tags[i] != in.Tags[i] {
			t.Errorf("Tags[%d] = %q, chci %q", i, got.Tags[i], in.Tags[i])
		}
	}
	if got.Password != "" {
		t.Errorf("Password = %q, chci prázdný řetězec (pole se neserializuje)", got.Password)
	}
}

func TestFromJSONInvalid(t *testing.T) {
	tests := map[string]string{
		"rozbitý JSON":  `{"id": 1,`,
		"chybí id":      `{"name":"Ada"}`,
		"nulové id":     `{"id":0,"name":"Ada"}`,
		"záporné id":    `{"id":-3,"name":"Ada"}`,
		"chybí jméno":   `{"id":1}`,
		"prázdné jméno": `{"id":1,"name":"   "}`,
		"špatný typ id": `{"id":"1","name":"Ada"}`,
	}
	for name, in := range tests {
		t.Run(name, func(t *testing.T) {
			if _, err := exercise.FromJSON([]byte(in)); err == nil {
				t.Errorf("FromJSON(%s) = _, nil, chci chybu", in)
			}
		})
	}
}

func TestDecodeEvent(t *testing.T) {
	t.Run("user.created", func(t *testing.T) {
		in := `{"kind":"user.created","payload":{"id":42,"name":"Grace"}}`
		got, err := exercise.DecodeEvent([]byte(in))
		if err != nil {
			t.Fatalf("DecodeEvent(%s) = _, %v, chci nil", in, err)
		}
		p, ok := got.(exercise.UserCreated)
		if !ok {
			t.Fatalf("DecodeEvent(...) = %T, chci UserCreated", got)
		}
		if p.ID != 42 || p.Name != "Grace" {
			t.Errorf("payload = %+v, chci {ID:42 Name:Grace}", p)
		}
	})

	t.Run("user.deleted", func(t *testing.T) {
		in := `{"kind":"user.deleted","payload":{"id":9,"reason":"gdpr"}}`
		got, err := exercise.DecodeEvent([]byte(in))
		if err != nil {
			t.Fatalf("DecodeEvent(%s) = _, %v, chci nil", in, err)
		}
		p, ok := got.(exercise.UserDeleted)
		if !ok {
			t.Fatalf("DecodeEvent(...) = %T, chci UserDeleted", got)
		}
		if p.ID != 9 || p.Reason != "gdpr" {
			t.Errorf("payload = %+v, chci {ID:9 Reason:gdpr}", p)
		}
	})

	t.Run("payload se dekóduje až podle kind", func(t *testing.T) {
		// Payload, který by se do UserCreated nikdy nevešel, ale kind ho neaktivuje.
		in := `{"kind":"user.deleted","payload":{"id":1,"reason":"spam"},"extra":{"x":[1,2,3]}}`
		got, err := exercise.DecodeEvent([]byte(in))
		if err != nil {
			t.Fatalf("DecodeEvent(%s) = _, %v, chci nil", in, err)
		}
		if _, ok := got.(exercise.UserDeleted); !ok {
			t.Fatalf("DecodeEvent(...) = %T, chci UserDeleted", got)
		}
	})

	chybne := map[string]string{
		"neznámý kind":   `{"kind":"user.renamed","payload":{"id":1}}`,
		"chybějící kind": `{"payload":{"id":1}}`,
		"chybí payload":  `{"kind":"user.created"}`,
		"rozbitý obal":   `{"kind":"user.created", "payload"`,
		"špatný payload": `{"kind":"user.created","payload":{"id":"ne-cislo"}}`,
	}
	for name, in := range chybne {
		t.Run(name, func(t *testing.T) {
			if _, err := exercise.DecodeEvent([]byte(in)); err == nil {
				t.Errorf("DecodeEvent(%s) = _, nil, chci chybu", in)
			}
		})
	}
}

type wallet struct {
	Owner   string          `json:"owner"`
	Balance exercise.Cents  `json:"balance"`
	Fee     *exercise.Cents `json:"fee,omitempty"`
}

func TestCentsMarshalFormat(t *testing.T) {
	tests := []struct {
		in   exercise.Cents
		want string
	}{
		{1999, "19.99"},
		{0, "0.00"},
		{5, "0.05"},
		{100, "1.00"},
		{-250, "-2.50"},
		{123456789, "1234567.89"},
	}
	for _, tt := range tests {
		got, err := json.Marshal(tt.in)
		if err != nil {
			t.Fatalf("json.Marshal(Cents(%d)) = _, %v, chci nil", int64(tt.in), err)
		}
		if string(got) != tt.want {
			t.Errorf("json.Marshal(Cents(%d)) = %s, chci %s", int64(tt.in), got, tt.want)
		}
	}
}

func TestCentsRoundTrip(t *testing.T) {
	for v := int64(-500000); v <= 500000; v += 4111 {
		in := exercise.Cents(v)
		data, err := json.Marshal(wallet{Owner: "ada", Balance: in})
		if err != nil {
			t.Fatalf("json.Marshal(wallet{Balance: %d}) = _, %v, chci nil", v, err)
		}

		var num struct {
			Balance float64 `json:"balance"`
		}
		if err := json.Unmarshal(data, &num); err != nil {
			t.Fatalf("balance není JSON číslo: %v (%s)", err, data)
		}
		if math.Abs(num.Balance-float64(v)/100) > 1e-9 {
			t.Errorf("balance = %v, chci %v (%s)", num.Balance, float64(v)/100, data)
		}

		var back wallet
		if err := json.Unmarshal(data, &back); err != nil {
			t.Fatalf("json.Unmarshal(%s) = %v, chci nil", data, err)
		}
		if back.Balance != in {
			t.Errorf("round-trip Cents(%d) = %d, chci %d (%s)", v, int64(back.Balance), v, data)
		}
	}
}

func TestCentsUnmarshalInput(t *testing.T) {
	tests := map[string]exercise.Cents{
		`19.99`:   1999,
		`"19.99"`: 1999,
		`0`:       0,
		`1`:       100,
		`-2.5`:    -250,
		`0.05`:    5,
	}
	for in, want := range tests {
		var got exercise.Cents
		if err := json.Unmarshal([]byte(in), &got); err != nil {
			t.Errorf("json.Unmarshal(%s) = %v, chci nil", in, err)
			continue
		}
		if got != want {
			t.Errorf("json.Unmarshal(%s) -> Cents(%d), chci Cents(%d)", in, int64(got), int64(want))
		}
	}

	var bad exercise.Cents
	if err := json.Unmarshal([]byte(`"neni-cislo"`), &bad); err == nil {
		t.Error(`json.Unmarshal("neni-cislo") = nil, chci chybu`)
	}
}

func TestStrictDecode(t *testing.T) {
	t.Run("známá pole projdou", func(t *testing.T) {
		var w wallet
		if err := exercise.StrictDecode([]byte(`{"owner":"ada","balance":19.99}`), &w); err != nil {
			t.Fatalf("StrictDecode(...) = %v, chci nil", err)
		}
		if w.Owner != "ada" || w.Balance != exercise.Cents(1999) {
			t.Errorf("StrictDecode(...) -> %+v, chci owner ada a balance 1999", w)
		}
	})

	t.Run("neznámé pole je chyba", func(t *testing.T) {
		var w wallet
		err := exercise.StrictDecode([]byte(`{"owner":"ada","balance":1,"typo":true}`), &w)
		if err == nil {
			t.Fatal("StrictDecode s neznámým polem = nil, chci chybu")
		}
		if !strings.Contains(err.Error(), "typo") {
			t.Errorf("chyba = %q, chci zmínku o neznámém poli typo", err)
		}
	})

	t.Run("data navíc jsou chyba", func(t *testing.T) {
		var w wallet
		if err := exercise.StrictDecode([]byte(`{"owner":"ada"}{"owner":"bob"}`), &w); err == nil {
			t.Fatal("StrictDecode se zbytkem dat = nil, chci chybu")
		}
	})

	t.Run("rozbitý JSON je chyba", func(t *testing.T) {
		var w wallet
		if err := exercise.StrictDecode([]byte(`{"owner":`), &w); err == nil {
			t.Fatal("StrictDecode(rozbitý JSON) = nil, chci chybu")
		}
	})
}

package solutions_test

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	exercise "github.com/rdurica/go-deep/lessons/lesson-16/solutions"
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

func TestUserJSONTags(t *testing.T) {
	data, err := json.Marshal(sampleUser())
	if err != nil {
		t.Fatalf("json.Marshal(User) = _, %v, chci nil", err)
	}

	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("výstup není platný JSON: %v (%s)", err, data)
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
	if strings.Contains(string(data), "tajne-heslo") {
		t.Errorf("heslo nesmí uniknout do JSON: %s", data)
	}
}

func TestUserOmitempty(t *testing.T) {
	u := exercise.User{
		ID:        1,
		Name:      "Bob",
		CreatedAt: time.Date(2020, time.January, 2, 3, 4, 5, 0, time.UTC),
	}
	data, err := json.Marshal(u)
	if err != nil {
		t.Fatalf("json.Marshal(User) = _, %v, chci nil", err)
	}

	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("výstup není platný JSON: %v (%s)", err, data)
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
	data := []byte(`{
		"id": 7,
		"name": "Ada",
		"email": "ada@example.com",
		"active": true,
		"tags": ["admin", "beta"],
		"created_at": "2024-03-01T12:30:00Z",
		"password": "tajne-heslo"
	}`)

	got, err := exercise.FromJSON(data)
	if err != nil {
		t.Fatalf("FromJSON(%s) = _, %v, chci nil", data, err)
	}
	if got.ID != 7 || got.Name != "Ada" || got.Email != "ada@example.com" || !got.Active {
		t.Errorf("FromJSON(...) = %+v, chci id=7 name=Ada email=ada@example.com active=true", got)
	}
	wantAt := time.Date(2024, time.March, 1, 12, 30, 0, 0, time.UTC)
	if !got.CreatedAt.Equal(wantAt) {
		t.Errorf("CreatedAt = %v, chci %v", got.CreatedAt, wantAt)
	}
	wantTags := []string{"admin", "beta"}
	if len(got.Tags) != len(wantTags) {
		t.Fatalf("Tags = %v, chci %v", got.Tags, wantTags)
	}
	for i := range wantTags {
		if got.Tags[i] != wantTags[i] {
			t.Errorf("Tags[%d] = %q, chci %q", i, got.Tags[i], wantTags[i])
		}
	}
	if got.Password != "" {
		t.Errorf("Password = %q, chci prázdný řetězec (json:\"-\" se nenačítá)", got.Password)
	}
}

func TestFromJSONInvalid(t *testing.T) {
	tests := map[string]string{
		"broken JSON":   `{"id": 1,`,
		"chybí id":      `{"name":"Ada"}`,
		"nulové id":     `{"id":0,"name":"Ada"}`,
		"záporné id":    `{"id":-3,"name":"Ada"}`,
		"missing name":  `{"id":1}`,
		"empty name":    `{"id":1,"name":"   "}`,
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
	Owner   string  `json:"owner"`
	Balance float64 `json:"balance"`
}

func TestStrictDecode(t *testing.T) {
	t.Run("known fields pass", func(t *testing.T) {
		var w wallet
		if err := exercise.StrictDecode([]byte(`{"owner":"ada","balance":19.99}`), &w); err != nil {
			t.Fatalf("StrictDecode(...) = %v, chci nil", err)
		}
		if w.Owner != "ada" || w.Balance != 19.99 {
			t.Errorf("StrictDecode(...) -> %+v, chci owner ada a balance 19.99", w)
		}
	})

	t.Run("unknown field is error", func(t *testing.T) {
		var w wallet
		err := exercise.StrictDecode([]byte(`{"owner":"ada","balance":1,"typo":true}`), &w)
		if err == nil {
			t.Fatal("StrictDecode s neznámým polem = nil, chci chybu")
		}
		if !strings.Contains(err.Error(), "typo") {
			t.Errorf("chyba = %q, chci zmínku o neznámém poli typo", err)
		}
	})

	t.Run("extra data is error", func(t *testing.T) {
		var w wallet
		if err := exercise.StrictDecode([]byte(`{"owner":"ada"}{"owner":"bob"}`), &w); err == nil {
			t.Fatal("StrictDecode se zbytkem dat = nil, chci chybu")
		}
	})

	t.Run("broken JSON is error", func(t *testing.T) {
		var w wallet
		if err := exercise.StrictDecode([]byte(`{"owner":`), &w); err == nil {
			t.Fatal("StrictDecode(rozbitý JSON) = nil, chci chybu")
		}
	})
}

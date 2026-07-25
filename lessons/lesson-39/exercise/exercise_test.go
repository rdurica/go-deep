package exercise_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	exercise "github.com/rdurica/go-deep/lessons/lesson-39/exercise"
)

// Adaptér musí splňovat port.
var _ exercise.Repository = (*exercise.MemoryRepo)(nil)

func day(t *testing.T, s string) time.Time {
	t.Helper()
	d, err := time.Parse("2006-01-02", s)
	if err != nil {
		t.Fatalf("nečitelné datum %q: %v", s, err)
	}
	return d
}

func mustRange(t *testing.T, from, to string) exercise.DateRange {
	t.Helper()
	r, err := exercise.NewDateRange(day(t, from), day(t, to))
	if err != nil {
		t.Fatalf("NewDateRange(%s, %s) = %v", from, to, err)
	}
	return r
}

func TestParseRoomID(t *testing.T) {
	ok := map[string]string{
		"A-101":   "A-101",
		"a-101":   "A-101",
		" b-042 ": "B-042",
		"Z-999":   "Z-999",
	}
	for in, want := range ok {
		got, err := exercise.ParseRoomID(in)
		if err != nil {
			t.Errorf("ParseRoomID(%q) = chyba %v", in, err)
			continue
		}
		if string(got) != want {
			t.Errorf("ParseRoomID(%q) = %q, chci %q", in, string(got), want)
		}
	}

	bad := []struct {
		in      string
		wantErr error
	}{
		{"", exercise.ErrEmptyRoomID},
		{"   ", exercise.ErrEmptyRoomID},
		{"A101", exercise.ErrInvalidRoomID},
		{"A-10", exercise.ErrInvalidRoomID},
		{"A-1011", exercise.ErrInvalidRoomID},
		{"1-101", exercise.ErrInvalidRoomID},
		{"A-10X", exercise.ErrInvalidRoomID},
		{"A-000", exercise.ErrInvalidRoomID},
		{"Á-101", exercise.ErrInvalidRoomID},
	}
	for _, tt := range bad {
		got, err := exercise.ParseRoomID(tt.in)
		if !errors.Is(err, tt.wantErr) {
			t.Errorf("ParseRoomID(%q) = chyba %v, chci %v", tt.in, err, tt.wantErr)
		}
		if got != "" {
			t.Errorf("ParseRoomID(%q) vrátil při chybě %q", tt.in, string(got))
		}
	}
}

func TestNewDateRange(t *testing.T) {
	r := mustRange(t, "2024-05-17", "2024-05-20")
	if r.Nights() != 3 {
		t.Errorf("Nights() = %d, chci 3", r.Nights())
	}
	if r.From().Location() != time.UTC || r.To().Location() != time.UTC {
		t.Error("termín musí být zarovnaný do UTC")
	}
	if h := r.From().Hour(); h != 0 {
		t.Errorf("From() = %v, chci zarovnání na celý den", r.From())
	}

	// Čas během dne se zahazuje.
	withTime, err := exercise.NewDateRange(
		time.Date(2024, 5, 17, 18, 30, 0, 0, time.UTC),
		time.Date(2024, 5, 20, 6, 0, 0, 0, time.UTC),
	)
	if err != nil {
		t.Fatalf("NewDateRange s časem = %v", err)
	}
	if withTime.Nights() != 3 {
		t.Errorf("Nights() = %d, chci 3 (čas během dne se ignoruje)", withTime.Nights())
	}

	bad := []struct {
		name    string
		from    string
		to      string
		wantErr error
	}{
		{"konec před začátkem", "2024-05-20", "2024-05-17", exercise.ErrInvalidRange},
		{"stejný den", "2024-05-17", "2024-05-17", exercise.ErrInvalidRange},
		{"příliš dlouhý", "2024-05-01", "2024-06-05", exercise.ErrRangeTooLong},
	}
	for _, tt := range bad {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := exercise.NewDateRange(day(t, tt.from), day(t, tt.to)); !errors.Is(err, tt.wantErr) {
				t.Errorf("NewDateRange = %v, chci %v", err, tt.wantErr)
			}
		})
	}

	// Hranice: přesně MaxNights nocí ještě projde.
	if _, err := exercise.NewDateRange(day(t, "2024-05-01"), day(t, "2024-05-31")); err != nil {
		t.Errorf("30 nocí = %v, chci nil", err)
	}
}

func TestDateRangeOverlaps(t *testing.T) {
	tests := []struct {
		name         string
		aFrom, aTo   string
		bFrom, bTo   string
		wantOverlaps bool
	}{
		{"úplný překryv", "2024-05-17", "2024-05-20", "2024-05-17", "2024-05-20", true},
		{"částečný překryv", "2024-05-17", "2024-05-20", "2024-05-19", "2024-05-22", true},
		{"obsažený", "2024-05-17", "2024-05-25", "2024-05-19", "2024-05-20", true},
		{"navazuje hned", "2024-05-17", "2024-05-20", "2024-05-20", "2024-05-22", false},
		{"navazuje před", "2024-05-20", "2024-05-22", "2024-05-17", "2024-05-20", false},
		{"úplně jinde", "2024-05-17", "2024-05-20", "2024-06-01", "2024-06-03", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := mustRange(t, tt.aFrom, tt.aTo)
			b := mustRange(t, tt.bFrom, tt.bTo)
			if got := a.Overlaps(b); got != tt.wantOverlaps {
				t.Errorf("Overlaps = %v, chci %v", got, tt.wantOverlaps)
			}
			if got := b.Overlaps(a); got != tt.wantOverlaps {
				t.Errorf("Overlaps je asymetrický: %v vs %v", got, tt.wantOverlaps)
			}
		})
	}
}

func validRequest() exercise.CreateBookingRequest {
	return exercise.CreateBookingRequest{
		Ref:         "BK-1",
		Room:        "A-101",
		Guest:       "Radek",
		From:        "2024-05-17",
		To:          "2024-05-20",
		NightlyRate: 1500,
	}
}

func TestValidateOK(t *testing.T) {
	if err := validRequest().Validate(); err != nil {
		t.Errorf("Validate() = %v, chci nil", err)
	}
}

func TestValidateSbiraVsechnyChyby(t *testing.T) {
	req := exercise.CreateBookingRequest{}
	err := req.Validate()
	if err == nil {
		t.Fatal("Validate() = nil, chci ValidationErrors")
	}
	var errs exercise.ValidationErrors
	if !errors.As(err, &errs) {
		t.Fatalf("Validate() = %T, chci ValidationErrors", err)
	}
	for _, field := range []string{"ref", "room", "guest", "from", "to", "nightly_rate"} {
		fe, ok := errs.Get(field)
		if !ok {
			t.Errorf("chybí chyba pole %q, mám %v", field, errs)
			continue
		}
		if fe.Code == "" || fe.Message == "" {
			t.Errorf("neúplná položka: %+v", fe)
		}
	}
}

func TestValidateJednotliveChyby(t *testing.T) {
	tests := []struct {
		name     string
		mutate   func(*exercise.CreateBookingRequest)
		field    string
		wantCode string
	}{
		{"chybí reference", func(r *exercise.CreateBookingRequest) { r.Ref = "  " }, "ref", exercise.CodeRequired},
		{"chybí pokoj", func(r *exercise.CreateBookingRequest) { r.Room = "" }, "room", exercise.CodeRequired},
		{"pokažený pokoj", func(r *exercise.CreateBookingRequest) { r.Room = "AA-1" }, "room", exercise.CodeFormat},
		{"chybí host", func(r *exercise.CreateBookingRequest) { r.Guest = "" }, "guest", exercise.CodeRequired},
		{"krátký host", func(r *exercise.CreateBookingRequest) { r.Guest = "R" }, "guest", exercise.CodeFormat},
		{"chybí datum", func(r *exercise.CreateBookingRequest) { r.From = "" }, "from", exercise.CodeRequired},
		{"pokažené datum", func(r *exercise.CreateBookingRequest) { r.To = "17.5.2024" }, "to", exercise.CodeFormat},
		{"obrácený termín", func(r *exercise.CreateBookingRequest) { r.To = "2024-05-16" }, "to", exercise.CodeRange},
		{"dlouhý pobyt", func(r *exercise.CreateBookingRequest) { r.To = "2024-07-17" }, "to", exercise.CodeRange},
		{"nulová cena", func(r *exercise.CreateBookingRequest) { r.NightlyRate = 0 }, "nightly_rate", exercise.CodeRange},
		{"záporná cena", func(r *exercise.CreateBookingRequest) { r.NightlyRate = -1 }, "nightly_rate", exercise.CodeRange},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := validRequest()
			tt.mutate(&req)

			var errs exercise.ValidationErrors
			if !errors.As(req.Validate(), &errs) {
				t.Fatalf("Validate() nevrátil ValidationErrors pro %+v", req)
			}
			if len(errs) != 1 {
				t.Fatalf("Validate() = %v, chci právě jednu chybu u %q", errs, tt.field)
			}
			fe, ok := errs.Get(tt.field)
			if !ok {
				t.Fatalf("Validate() = %v, chci chybu u %q", errs, tt.field)
			}
			if fe.Code != tt.wantCode {
				t.Errorf("kód = %q, chci %q", fe.Code, tt.wantCode)
			}
		})
	}
}

func TestValidationErrorsError(t *testing.T) {
	errs := exercise.ValidationErrors{
		{Field: "room", Code: exercise.CodeFormat, Message: "špatný tvar"},
		{Field: "guest", Code: exercise.CodeRequired, Message: "chybí"},
	}
	first := errs.Error()
	for i := 0; i < 10; i++ {
		if got := errs.Error(); got != first {
			t.Fatalf("Error() není deterministický: %q vs %q", got, first)
		}
	}
	for _, want := range []string{"room", "guest", "format", "required"} {
		if !strings.Contains(first, want) {
			t.Errorf("Error() = %q, chybí %q", first, want)
		}
	}

	var wrapped error = fmt.Errorf("hranice: %w", errs)
	var target exercise.ValidationErrors
	if !errors.As(wrapped, &target) {
		t.Fatal("ValidationErrors musí jít vytáhnout přes errors.As")
	}
}

func TestMemoryRepo(t *testing.T) {
	ctx := context.Background()
	repo := exercise.NewMemoryRepo()

	b := exercise.Booking{Ref: "BK-1", Room: "A-101", Stay: mustRange(t, "2024-05-17", "2024-05-20"), Guest: "Radek"}
	if err := repo.Save(ctx, b); err != nil {
		t.Fatalf("Save = %v", err)
	}
	if err := repo.Save(ctx, b); !errors.Is(err, exercise.ErrDuplicateRef) {
		t.Errorf("duplicitní Save = %v, chci ErrDuplicateRef", err)
	}

	other := exercise.Booking{Ref: "BK-2", Room: "B-202", Stay: mustRange(t, "2024-05-17", "2024-05-18"), Guest: "Jana"}
	if err := repo.Save(ctx, other); err != nil {
		t.Fatalf("Save = %v", err)
	}

	got, err := repo.ByRoom(ctx, "A-101")
	if err != nil {
		t.Fatalf("ByRoom = %v", err)
	}
	if len(got) != 1 || got[0].Ref != "BK-1" {
		t.Errorf("ByRoom(A-101) = %v, chci jen BK-1", got)
	}
	if empty, err := repo.ByRoom(ctx, "Z-999"); err != nil || len(empty) != 0 {
		t.Errorf("ByRoom(Z-999) = (%v, %v), chci prázdno bez chyby", empty, err)
	}
}

func TestMemoryRepoRadi(t *testing.T) {
	ctx := context.Background()
	repo := exercise.NewMemoryRepo()

	terms := []struct{ ref, from, to string }{
		{"BK-3", "2024-06-01", "2024-06-03"},
		{"BK-1", "2024-05-01", "2024-05-03"},
		{"BK-2", "2024-05-10", "2024-05-12"},
	}
	for _, term := range terms {
		if err := repo.Save(ctx, exercise.Booking{
			Ref: term.ref, Room: "A-101", Stay: mustRange(t, term.from, term.to), Guest: "Host",
		}); err != nil {
			t.Fatalf("Save(%s) = %v", term.ref, err)
		}
	}

	got, err := repo.ByRoom(ctx, "A-101")
	if err != nil {
		t.Fatalf("ByRoom = %v", err)
	}
	want := []string{"BK-1", "BK-2", "BK-3"}
	for i, ref := range want {
		if got[i].Ref != ref {
			t.Fatalf("ByRoom()[%d] = %s, chci %s (seřazeno podle začátku pobytu)", i, got[i].Ref, ref)
		}
	}
}

func TestMemoryRepoZrusenyKontext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	repo := exercise.NewMemoryRepo()
	if err := repo.Save(ctx, exercise.Booking{Ref: "BK-1", Room: "A-101"}); !errors.Is(err, context.Canceled) {
		t.Errorf("Save = %v, chci context.Canceled", err)
	}
	if _, err := repo.ByRoom(ctx, "A-101"); !errors.Is(err, context.Canceled) {
		t.Errorf("ByRoom = %v, chci context.Canceled", err)
	}
}

func TestMetricsNulovaHodnotaAObeh(t *testing.T) {
	var m exercise.Metrics
	if len(m.Snapshot()) != 0 {
		t.Error("nulová hodnota Metrics má být prázdná, ne panika")
	}
	m.Inc("bookings_created_total")
	m.Inc("bookings_created_total")
	if got := m.Snapshot()["bookings_created_total"]; got != 2 {
		t.Errorf("čítač = %d, chci 2", got)
	}

	snap := m.Snapshot()
	snap["bookings_created_total"] = 99
	if again := m.Snapshot()["bookings_created_total"]; again != 2 {
		t.Error("Snapshot musí vracet kopii")
	}
}

func TestMetricsSoubezne(t *testing.T) {
	var m exercise.Metrics
	const n = 200
	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			m.Inc("bookings_created_total")
			m.Snapshot()
		}()
	}
	wg.Wait()

	if got := m.Snapshot()["bookings_created_total"]; got != n {
		t.Errorf("čítač = %d, chci %d", got, n)
	}
}

func TestServiceBook(t *testing.T) {
	ctx := context.Background()
	svc := exercise.NewService(exercise.NewMemoryRepo(), nil)

	b, err := svc.Book(ctx, validRequest())
	if err != nil {
		t.Fatalf("Book = %v", err)
	}
	if b.Room != "A-101" {
		t.Errorf("pokoj = %q, chci normalizovaný A-101", string(b.Room))
	}
	if b.Stay.Nights() != 3 {
		t.Errorf("nocí = %d, chci 3", b.Stay.Nights())
	}
	if b.Total != 4500 {
		t.Errorf("Total = %d, chci 4500 (3 noci × 1500)", b.Total)
	}
	if got := svc.Metrics().Snapshot()[exercise.MetricCreated]; got != 1 {
		t.Errorf("%s = %d, chci 1", exercise.MetricCreated, got)
	}
}

func TestServiceBookObsazenyPokoj(t *testing.T) {
	ctx := context.Background()
	svc := exercise.NewService(exercise.NewMemoryRepo(), nil)
	if _, err := svc.Book(ctx, validRequest()); err != nil {
		t.Fatalf("příprava = %v", err)
	}

	overlapping := validRequest()
	overlapping.Ref = "BK-2"
	overlapping.From = "2024-05-19"
	overlapping.To = "2024-05-22"
	if _, err := svc.Book(ctx, overlapping); !errors.Is(err, exercise.ErrRoomTaken) {
		t.Fatalf("Book přes obsazený termín = %v, chci ErrRoomTaken", err)
	}

	// Navazující termín ve stejném pokoji projít musí.
	next := validRequest()
	next.Ref = "BK-3"
	next.From = "2024-05-20"
	next.To = "2024-05-22"
	if _, err := svc.Book(ctx, next); err != nil {
		t.Errorf("navazující termín = %v, chci nil (interval je polootevřený)", err)
	}

	// Jiný pokoj ve stejném termínu taky.
	otherRoom := validRequest()
	otherRoom.Ref = "BK-4"
	otherRoom.Room = "B-202"
	if _, err := svc.Book(ctx, otherRoom); err != nil {
		t.Errorf("jiný pokoj = %v, chci nil", err)
	}

	snap := svc.Metrics().Snapshot()
	if snap[exercise.MetricCreated] != 3 {
		t.Errorf("%s = %d, chci 3", exercise.MetricCreated, snap[exercise.MetricCreated])
	}
	if snap[exercise.MetricRejected] != 1 {
		t.Errorf("%s = %d, chci 1", exercise.MetricRejected, snap[exercise.MetricRejected])
	}
}

func TestServiceBookNeplatnaData(t *testing.T) {
	ctx := context.Background()
	svc := exercise.NewService(exercise.NewMemoryRepo(), nil)

	req := validRequest()
	req.Room = "nesmysl"
	_, err := svc.Book(ctx, req)

	var errs exercise.ValidationErrors
	if !errors.As(err, &errs) {
		t.Fatalf("Book = %v (%T), chci ValidationErrors", err, err)
	}
	if got := svc.Metrics().Snapshot()[exercise.MetricRejected]; got != 1 {
		t.Errorf("%s = %d, chci 1", exercise.MetricRejected, got)
	}
	if got := svc.Metrics().Snapshot()[exercise.MetricCreated]; got != 0 {
		t.Errorf("%s = %d, chci 0", exercise.MetricCreated, got)
	}
}

func TestServiceBookDuplicitniReference(t *testing.T) {
	ctx := context.Background()
	svc := exercise.NewService(exercise.NewMemoryRepo(), nil)
	if _, err := svc.Book(ctx, validRequest()); err != nil {
		t.Fatalf("příprava = %v", err)
	}

	dup := validRequest()
	dup.Room = "C-303"
	if _, err := svc.Book(ctx, dup); !errors.Is(err, exercise.ErrDuplicateRef) {
		t.Errorf("duplicitní reference = %v, chci ErrDuplicateRef", err)
	}
}

func newServer(t *testing.T) (*httptest.Server, *exercise.Service) {
	t.Helper()
	svc := exercise.NewService(exercise.NewMemoryRepo(), nil)
	srv := httptest.NewServer(exercise.Handler(svc))
	t.Cleanup(srv.Close)
	return srv, svc
}

func post(t *testing.T, srv *httptest.Server, path, body string) (*http.Response, map[string]any) {
	t.Helper()
	resp, err := srv.Client().Post(srv.URL+path, "application/json", strings.NewReader(body))
	if err != nil {
		t.Fatalf("POST %s: %v", path, err)
	}
	t.Cleanup(func() { _ = resp.Body.Close() })

	var decoded map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&decoded); err != nil {
		t.Fatalf("POST %s: tělo není JSON objekt: %v", path, err)
	}
	return resp, decoded
}

const validBody = `{"ref":"BK-1","room":"A-101","guest":"Radek","from":"2024-05-17","to":"2024-05-20","nightly_rate":1500}`

func TestHandlerVytvoreni(t *testing.T) {
	srv, _ := newServer(t)
	resp, body := post(t, srv, "/bookings", validBody)

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("status = %d, chci 201 (tělo: %v)", resp.StatusCode, body)
	}
	if got := resp.Header.Get("Location"); got != "/bookings/BK-1" {
		t.Errorf("Location = %q, chci /bookings/BK-1", got)
	}
	if body["nights"] != float64(3) {
		t.Errorf("nights = %v, chci 3", body["nights"])
	}
	if body["total"] != float64(4500) {
		t.Errorf("total = %v, chci 4500", body["total"])
	}
	if body["room"] != "A-101" {
		t.Errorf("room = %v, chci A-101", body["room"])
	}
}

func TestHandlerChyboveStavy(t *testing.T) {
	t.Run("rozbité tělo", func(t *testing.T) {
		srv, _ := newServer(t)
		resp, _ := post(t, srv, "/bookings", `{"ref":`)
		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("status = %d, chci 400", resp.StatusCode)
		}
		if ct := resp.Header.Get("Content-Type"); ct != exercise.ProblemContentType {
			t.Errorf("Content-Type = %q, chci %q", ct, exercise.ProblemContentType)
		}
	})

	t.Run("neznámé pole", func(t *testing.T) {
		srv, _ := newServer(t)
		resp, _ := post(t, srv, "/bookings", `{"ref":"BK-1","sleva":10}`)
		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("status = %d, chci 400", resp.StatusCode)
		}
	})

	t.Run("neplatná data", func(t *testing.T) {
		srv, _ := newServer(t)
		resp, body := post(t, srv, "/bookings", `{"ref":"","room":"xx","guest":"","from":"","to":"","nightly_rate":0}`)
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("status = %d, chci 422 (tělo: %v)", resp.StatusCode, body)
		}
		errs, ok := body["errors"].([]any)
		if !ok || len(errs) == 0 {
			t.Fatalf("tělo = %v, chci seznam chyb polí", body)
		}
		first, _ := errs[0].(map[string]any)
		for _, key := range []string{"field", "code", "message"} {
			if _, ok := first[key]; !ok {
				t.Errorf("položka chyby nemá %q: %v", key, first)
			}
		}
	})

	t.Run("obsazený pokoj", func(t *testing.T) {
		srv, _ := newServer(t)
		if resp, _ := post(t, srv, "/bookings", validBody); resp.StatusCode != http.StatusCreated {
			t.Fatal("příprava selhala")
		}
		body := strings.Replace(validBody, `"BK-1"`, `"BK-2"`, 1)
		resp, _ := post(t, srv, "/bookings", body)
		if resp.StatusCode != http.StatusConflict {
			t.Errorf("status = %d, chci 409", resp.StatusCode)
		}
	})

	t.Run("duplicitní reference", func(t *testing.T) {
		srv, _ := newServer(t)
		if resp, _ := post(t, srv, "/bookings", validBody); resp.StatusCode != http.StatusCreated {
			t.Fatal("příprava selhala")
		}
		body := strings.Replace(validBody, `"A-101"`, `"C-303"`, 1)
		resp, _ := post(t, srv, "/bookings", body)
		if resp.StatusCode != http.StatusConflict {
			t.Errorf("status = %d, chci 409", resp.StatusCode)
		}
	})
}

func TestHandlerMetriky(t *testing.T) {
	srv, _ := newServer(t)
	if resp, _ := post(t, srv, "/bookings", validBody); resp.StatusCode != http.StatusCreated {
		t.Fatal("příprava selhala")
	}
	if resp, _ := post(t, srv, "/bookings", `{"ref":"","room":"","guest":"","from":"","to":"","nightly_rate":0}`); resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatal("příprava selhala")
	}

	resp, err := srv.Client().Get(srv.URL + "/metrics")
	if err != nil {
		t.Fatalf("GET /metrics: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	var snap map[string]int64
	if err := json.NewDecoder(resp.Body).Decode(&snap); err != nil {
		t.Fatalf("tělo /metrics není JSON: %v", err)
	}
	if snap[exercise.MetricCreated] != 1 {
		t.Errorf("%s = %d, chci 1", exercise.MetricCreated, snap[exercise.MetricCreated])
	}
	if snap[exercise.MetricRejected] != 1 {
		t.Errorf("%s = %d, chci 1", exercise.MetricRejected, snap[exercise.MetricRejected])
	}
}

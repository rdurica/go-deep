package solutions_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"go/parser"
	"go/token"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync"
	"testing"

	exercise "github.com/rdurica/go-deep/lessons/lesson-38/solutions"
)

// ---------- pomocníci ----------

func lines() []exercise.Line {
	return []exercise.Line{
		{SKU: "kniha-go", Quantity: 2, UnitPriceCents: 49900},
		{SKU: "hrnek", Quantity: 1, UnitPriceCents: 19900},
	}
}

func mustOrder(t *testing.T) exercise.Order {
	t.Helper()
	o, err := exercise.NewOrder("ord-1", lines())
	if err != nil {
		t.Fatalf("NewOrder = chyba %v", err)
	}
	return o
}

// stubIDs vydává předvídatelná ID, aby testy nezávisely na náhodě.
type stubIDs struct {
	mu sync.Mutex
	n  int
}

func (s *stubIDs) NewID() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.n++
	return "ord-" + strconv.Itoa(s.n)
}

// fakeRepo je fake portu Repository. Počítá volání, aby šlo ověřit,
// že se při chybě domény nic neukládá.
type fakeRepo struct {
	mu      sync.Mutex
	orders  map[string]exercise.Order
	saves   int
	finds   int
	saveErr error
}

func newFakeRepo() *fakeRepo {
	return &fakeRepo{orders: make(map[string]exercise.Order)}
}

func (f *fakeRepo) Save(ctx context.Context, o exercise.Order) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.saves++
	if f.saveErr != nil {
		return f.saveErr
	}
	f.orders[o.ID] = o
	return nil
}

func (f *fakeRepo) Find(ctx context.Context, id string) (exercise.Order, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.finds++
	o, ok := f.orders[id]
	if !ok {
		return exercise.Order{}, fmt.Errorf("%w: %s", exercise.ErrNotFound, id)
	}
	return o, nil
}

// ---------- A: doména ----------

func TestStatusString(t *testing.T) {
	tests := []struct {
		in   exercise.Status
		want string
	}{
		{exercise.StatusUnknown, "unknown"},
		{exercise.StatusNew, "new"},
		{exercise.StatusPaid, "paid"},
		{exercise.StatusShipped, "shipped"},
		{exercise.StatusCancelled, "cancelled"},
		{exercise.Status(42), "unknown"},
	}
	for _, tt := range tests {
		if got := tt.in.String(); got != tt.want {
			t.Errorf("Status(%d).String() = %q, chci %q", int(tt.in), got, tt.want)
		}
	}
	if got := fmt.Sprintf("%v", exercise.StatusPaid); got != "paid" {
		t.Errorf("fmt.Sprintf(%%v, StatusPaid) = %q, chci %q — chybí Stringer", got, "paid")
	}
}

func TestLineTotalCents(t *testing.T) {
	l := exercise.Line{SKU: "x", Quantity: 3, UnitPriceCents: 1999}
	if got := l.TotalCents(); got != 5997 {
		t.Errorf("Line.TotalCents() = %d, chci 5997", got)
	}
}

func TestNewOrder(t *testing.T) {
	o, err := exercise.NewOrder("  ord-1  ", lines())
	if err != nil {
		t.Fatalf("NewOrder = chyba %v", err)
	}
	if o.ID != "ord-1" {
		t.Errorf("ID = %q, chci %q (ID se má ořezat)", o.ID, "ord-1")
	}
	if o.Status != exercise.StatusNew {
		t.Errorf("Status = %v, chci new", o.Status)
	}
	if got := o.TotalCents(); got != 2*49900+19900 {
		t.Errorf("TotalCents() = %d, chci %d", got, 2*49900+19900)
	}
}

func TestNewOrderInvarianty(t *testing.T) {
	tests := []struct {
		name    string
		id      string
		lines   []exercise.Line
		wantErr error
	}{
		{"prázdné ID", "", lines(), exercise.ErrMissingID},
		{"ID jen mezery", "   ", lines(), exercise.ErrMissingID},
		{"nil položky", "ord-1", nil, exercise.ErrEmptyOrder},
		{"prázdné položky", "ord-1", []exercise.Line{}, exercise.ErrEmptyOrder},
		{"prázdné SKU", "ord-1", []exercise.Line{{SKU: " ", Quantity: 1, UnitPriceCents: 100}}, exercise.ErrInvalidLine},
		{"nulové množství", "ord-1", []exercise.Line{{SKU: "x", Quantity: 0, UnitPriceCents: 100}}, exercise.ErrInvalidLine},
		{"záporné množství", "ord-1", []exercise.Line{{SKU: "x", Quantity: -1, UnitPriceCents: 100}}, exercise.ErrInvalidLine},
		{"nulová cena", "ord-1", []exercise.Line{{SKU: "x", Quantity: 1, UnitPriceCents: 0}}, exercise.ErrInvalidLine},
		{"záporná cena", "ord-1", []exercise.Line{{SKU: "x", Quantity: 1, UnitPriceCents: -5}}, exercise.ErrInvalidLine},
		{"vadná druhá položka", "ord-1", []exercise.Line{
			{SKU: "x", Quantity: 1, UnitPriceCents: 100},
			{SKU: "y", Quantity: 0, UnitPriceCents: 100},
		}, exercise.ErrInvalidLine},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o, err := exercise.NewOrder(tt.id, tt.lines)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("NewOrder = chyba %v, chci %v", err, tt.wantErr)
			}
			if o.ID != "" || o.Status != exercise.StatusUnknown || len(o.Lines) != 0 {
				t.Errorf("při chybě chci nulovou objednávku, mám %+v", o)
			}
		})
	}
}

func TestNewOrderKopirujePolozky(t *testing.T) {
	in := lines()
	o, err := exercise.NewOrder("ord-1", in)
	if err != nil {
		t.Fatalf("NewOrder = chyba %v", err)
	}
	before := o.TotalCents()

	in[0].Quantity = 999
	in[0].SKU = "podvrh"

	if got := o.TotalCents(); got != before {
		t.Errorf("změna vstupního slice prosákla do objednávky: %d → %d", before, got)
	}
	if o.Lines[0].SKU == "podvrh" {
		t.Error("objednávka sdílí slice s volajícím, konstruktor musí dělat kopii")
	}
}

func TestPrechodyStavu(t *testing.T) {
	type step struct {
		name string
		fn   func(exercise.Order) (exercise.Order, error)
	}
	steps := []step{
		{"Pay", exercise.Order.Pay},
		{"Ship", exercise.Order.Ship},
		{"Cancel", exercise.Order.Cancel},
	}

	// stav → povolený přechod → výsledný stav
	allowed := map[exercise.Status]map[string]exercise.Status{
		exercise.StatusNew:       {"Pay": exercise.StatusPaid, "Cancel": exercise.StatusCancelled},
		exercise.StatusPaid:      {"Ship": exercise.StatusShipped, "Cancel": exercise.StatusCancelled},
		exercise.StatusShipped:   {},
		exercise.StatusCancelled: {},
	}

	for status, ok := range allowed {
		for _, s := range steps {
			t.Run(status.String()+"/"+s.name, func(t *testing.T) {
				base := mustOrder(t)
				base.Status = status

				next, err := s.fn(base)
				want, isAllowed := ok[s.name]
				if !isAllowed {
					if !errors.Is(err, exercise.ErrInvalidTransition) {
						t.Fatalf("%s ze stavu %v = %v, chci ErrInvalidTransition", s.name, status, err)
					}
					return
				}
				if err != nil {
					t.Fatalf("%s ze stavu %v = chyba %v, chci úspěch", s.name, status, err)
				}
				if next.Status != want {
					t.Errorf("%s ze stavu %v dal %v, chci %v", s.name, status, next.Status, want)
				}
			})
		}
	}
}

func TestOdeslanouObjednavkuNelzeZrusit(t *testing.T) {
	o := mustOrder(t)
	paid, err := o.Pay()
	if err != nil {
		t.Fatalf("Pay = chyba %v", err)
	}
	shipped, err := paid.Ship()
	if err != nil {
		t.Fatalf("Ship = chyba %v", err)
	}
	if _, err := shipped.Cancel(); !errors.Is(err, exercise.ErrInvalidTransition) {
		t.Errorf("Cancel odeslané objednávky = %v, chci ErrInvalidTransition", err)
	}
}

func TestPrechodNemeniPuvodniObjednavku(t *testing.T) {
	o := mustOrder(t)
	if _, err := o.Pay(); err != nil {
		t.Fatalf("Pay = chyba %v", err)
	}
	if o.Status != exercise.StatusNew {
		t.Errorf("po Pay má původní hodnota stav %v, chci new — přechod musí vracet novou hodnotu", o.Status)
	}
}

// ---------- B: use-casy a in-memory adaptér ----------

func TestServicePlace(t *testing.T) {
	repo := newFakeRepo()
	svc := exercise.NewService(repo, &stubIDs{})

	o, err := svc.Place(context.Background(), lines())
	if err != nil {
		t.Fatalf("Place = chyba %v", err)
	}
	if o.ID != "ord-1" {
		t.Errorf("ID = %q, chci ord-1 — použij port IDGen", o.ID)
	}
	if o.Status != exercise.StatusNew {
		t.Errorf("Status = %v, chci new", o.Status)
	}
	if repo.saves != 1 {
		t.Errorf("Save volán %dkrát, chci 1", repo.saves)
	}

	second, err := svc.Place(context.Background(), lines())
	if err != nil {
		t.Fatalf("Place = chyba %v", err)
	}
	if second.ID == o.ID {
		t.Errorf("druhá objednávka má stejné ID %q", second.ID)
	}
}

func TestServicePlaceNeukladaNeplatnou(t *testing.T) {
	repo := newFakeRepo()
	svc := exercise.NewService(repo, &stubIDs{})

	_, err := svc.Place(context.Background(), nil)
	if !errors.Is(err, exercise.ErrEmptyOrder) {
		t.Fatalf("Place(nil) = %v, chci ErrEmptyOrder", err)
	}
	if repo.saves != 0 {
		t.Errorf("Save volán %dkrát, chci 0 — neplatná objednávka se nesmí uložit", repo.saves)
	}
}

func TestServicePlaceObaliChybuUlozeni(t *testing.T) {
	sentinel := errors.New("disk plný")
	repo := newFakeRepo()
	repo.saveErr = sentinel
	svc := exercise.NewService(repo, &stubIDs{})

	if _, err := svc.Place(context.Background(), lines()); !errors.Is(err, sentinel) {
		t.Errorf("Place = %v, chci chybu obalující chybu portu (%%w)", err)
	}
}

func TestServiceGet(t *testing.T) {
	repo := newFakeRepo()
	svc := exercise.NewService(repo, &stubIDs{})

	if _, err := svc.Get(context.Background(), "nic"); !errors.Is(err, exercise.ErrNotFound) {
		t.Fatalf("Get(neexistující) = %v, chci ErrNotFound", err)
	}

	placed, err := svc.Place(context.Background(), lines())
	if err != nil {
		t.Fatalf("Place = chyba %v", err)
	}
	got, err := svc.Get(context.Background(), placed.ID)
	if err != nil {
		t.Fatalf("Get = chyba %v", err)
	}
	if got.ID != placed.ID || got.TotalCents() != placed.TotalCents() {
		t.Errorf("Get = %+v, chci %+v", got, placed)
	}
}

func TestServiceZivotniCyklus(t *testing.T) {
	repo := newFakeRepo()
	svc := exercise.NewService(repo, &stubIDs{})
	ctx := context.Background()

	o, err := svc.Place(ctx, lines())
	if err != nil {
		t.Fatalf("Place = chyba %v", err)
	}

	paid, err := svc.Pay(ctx, o.ID)
	if err != nil {
		t.Fatalf("Pay = chyba %v", err)
	}
	if paid.Status != exercise.StatusPaid {
		t.Fatalf("po Pay je stav %v, chci paid", paid.Status)
	}

	shipped, err := svc.Ship(ctx, o.ID)
	if err != nil {
		t.Fatalf("Ship = chyba %v", err)
	}
	if shipped.Status != exercise.StatusShipped {
		t.Fatalf("po Ship je stav %v, chci shipped", shipped.Status)
	}

	// Stav musí být opravdu uložený, ne jen vrácený.
	reloaded, err := svc.Get(ctx, o.ID)
	if err != nil {
		t.Fatalf("Get = chyba %v", err)
	}
	if reloaded.Status != exercise.StatusShipped {
		t.Errorf("uložený stav = %v, chci shipped — use-case musí ukládat", reloaded.Status)
	}
}

func TestServiceNepovolenyPrechodNeuklada(t *testing.T) {
	repo := newFakeRepo()
	svc := exercise.NewService(repo, &stubIDs{})
	ctx := context.Background()

	o, err := svc.Place(ctx, lines())
	if err != nil {
		t.Fatalf("Place = chyba %v", err)
	}
	saveBefore := repo.saves

	if _, err := svc.Ship(ctx, o.ID); !errors.Is(err, exercise.ErrInvalidTransition) {
		t.Fatalf("Ship nezaplacené = %v, chci ErrInvalidTransition", err)
	}
	if repo.saves != saveBefore {
		t.Errorf("Save volán %dkrát navíc, chci 0 — zamítnutý přechod se neukládá", repo.saves-saveBefore)
	}
}

func TestServiceNaNeexistujici(t *testing.T) {
	svc := exercise.NewService(newFakeRepo(), &stubIDs{})
	ctx := context.Background()

	for name, fn := range map[string]func(context.Context, string) (exercise.Order, error){
		"Pay":    svc.Pay,
		"Ship":   svc.Ship,
		"Cancel": svc.Cancel,
		"Get":    svc.Get,
	} {
		if _, err := fn(ctx, "neexistuje"); !errors.Is(err, exercise.ErrNotFound) {
			t.Errorf("%s(neexistující) = %v, chci ErrNotFound", name, err)
		}
	}
}

func TestMemoryRepository(t *testing.T) {
	repo := exercise.NewMemoryRepository()
	ctx := context.Background()

	if _, err := repo.Find(ctx, "ord-1"); !errors.Is(err, exercise.ErrNotFound) {
		t.Errorf("Find(prázdné úložiště) = %v, chci ErrNotFound", err)
	}

	o := mustOrder(t)
	if err := repo.Save(ctx, o); err != nil {
		t.Fatalf("Save = chyba %v", err)
	}
	got, err := repo.Find(ctx, "ord-1")
	if err != nil {
		t.Fatalf("Find = chyba %v", err)
	}
	if got.ID != o.ID || got.Status != o.Status || got.TotalCents() != o.TotalCents() {
		t.Errorf("Find = %+v, chci %+v", got, o)
	}

	paid, err := o.Pay()
	if err != nil {
		t.Fatalf("Pay = chyba %v", err)
	}
	if err := repo.Save(ctx, paid); err != nil {
		t.Fatalf("Save = chyba %v", err)
	}
	again, err := repo.Find(ctx, "ord-1")
	if err != nil {
		t.Fatalf("Find = chyba %v", err)
	}
	if again.Status != exercise.StatusPaid {
		t.Errorf("po přepsání je stav %v, chci paid", again.Status)
	}
}

func TestMemoryRepositoryIzolujeData(t *testing.T) {
	repo := exercise.NewMemoryRepository()
	ctx := context.Background()
	o := mustOrder(t)

	if err := repo.Save(ctx, o); err != nil {
		t.Fatalf("Save = chyba %v", err)
	}
	o.Lines[0].Quantity = 999 // volající si slice ponechal a změnil ho

	got, err := repo.Find(ctx, "ord-1")
	if err != nil {
		t.Fatalf("Find = chyba %v", err)
	}
	if got.Lines[0].Quantity == 999 {
		t.Error("úložiště sdílí slice s volajícím — Save musí položky zkopírovat")
	}

	got.Lines[0].Quantity = 111
	again, err := repo.Find(ctx, "ord-1")
	if err != nil {
		t.Fatalf("Find = chyba %v", err)
	}
	if again.Lines[0].Quantity == 111 {
		t.Error("úložiště vrací vnitřní slice — Find musí položky zkopírovat")
	}
}

func TestMemoryRepositorySoubezne(t *testing.T) {
	repo := exercise.NewMemoryRepository()
	ctx := context.Background()

	var wg sync.WaitGroup
	for i := 0; i < 16; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			id := "ord-" + strconv.Itoa(i%4)
			o, err := exercise.NewOrder(id, lines())
			if err != nil {
				t.Errorf("NewOrder = chyba %v", err)
				return
			}
			for j := 0; j < 50; j++ {
				if err := repo.Save(ctx, o); err != nil {
					t.Errorf("Save = chyba %v", err)
					return
				}
				if _, err := repo.Find(ctx, id); err != nil {
					t.Errorf("Find = chyba %v", err)
					return
				}
			}
		}(i)
	}
	wg.Wait()
}

// ---------- hranice balíčků ----------

// TestHraniceBalicku hlídá směr závislostí strojově. Slovní dohoda „doména
// nezná HTTP" vydrží do prvního spěchu; tenhle test vydrží i po něm.
func TestHraniceBalicku(t *testing.T) {
	zakazane := map[string]bool{
		"net/http":      true,
		"encoding/json": true,
		"database/sql":  true,
	}
	// Testy balíčku klidně net/http importovat mohou; hranici porušuje až
	// produkční kód.
	bezTestu := func(fi fs.FileInfo) bool {
		return !strings.HasSuffix(fi.Name(), "_test.go")
	}

	for _, dir := range []string{"order", "app"} {
		fset := token.NewFileSet()
		pkgs, err := parser.ParseDir(fset, dir, bezTestu, parser.ImportsOnly)
		if err != nil {
			t.Fatalf("parsování %q selhalo: %v", dir, err)
		}

		// Bez tohohle počítadla by test prošel i tehdy, kdyby ParseDir
		// nenašel vůbec nic — falešně zelený test je horší než žádný.
		souboru := 0
		for _, pkg := range pkgs {
			for name, file := range pkg.Files {
				souboru++
				for _, spec := range file.Imports {
					path, err := strconv.Unquote(spec.Path.Value)
					if err != nil {
						t.Fatalf("%s: nečitelný import %s", name, spec.Path.Value)
					}
					if zakazane[path] {
						t.Errorf("%s importuje %q — vnitřní vrstva nesmí znát transport ani úložiště", name, path)
					}
				}
			}
		}
		if souboru == 0 {
			t.Errorf("v %q se neprošel žádný soubor", dir)
		}
	}
}

// ---------- C: HTTP adaptér ----------

func newTestHandler() http.Handler {
	svc := exercise.NewService(exercise.NewMemoryRepository(), &stubIDs{})
	return exercise.NewHandler(svc)
}

func do(t *testing.T, h http.Handler, method, path, body string) *httptest.ResponseRecorder {
	t.Helper()
	var req *http.Request
	if body == "" {
		req = httptest.NewRequest(method, path, nil)
	} else {
		req = httptest.NewRequest(method, path, strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

const validBody = `{"lines":[{"sku":"kniha-go","quantity":2,"unit_price_cents":49900}]}`

func decodeBody(t *testing.T, rec *httptest.ResponseRecorder) map[string]any {
	t.Helper()
	var out map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatalf("tělo není JSON objekt (%v): %s", err, rec.Body.String())
	}
	return out
}

func TestHTTPZalozeniObjednavky(t *testing.T) {
	h := newTestHandler()
	rec := do(t, h, http.MethodPost, "/orders", validBody)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, chci 201 (tělo: %s)", rec.Code, rec.Body.String())
	}
	body := decodeBody(t, rec)
	if body["status"] != "new" {
		t.Errorf("status v těle = %v, chci \"new\"", body["status"])
	}
	if body["total_cents"] != float64(99800) {
		t.Errorf("total_cents = %v, chci 99800", body["total_cents"])
	}
	if id, _ := body["id"].(string); id == "" {
		t.Errorf("v těle chybí id: %v", body)
	}
}

func TestHTTPChybneVstupy(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		wantStatus int
	}{
		{"rozbitý JSON", `{"lines":`, http.StatusBadRequest},
		{"prázdné tělo", `{}`, http.StatusUnprocessableEntity},
		{"prázdné položky", `{"lines":[]}`, http.StatusUnprocessableEntity},
		{"nulové množství", `{"lines":[{"sku":"x","quantity":0,"unit_price_cents":100}]}`, http.StatusUnprocessableEntity},
		{"záporná cena", `{"lines":[{"sku":"x","quantity":1,"unit_price_cents":-1}]}`, http.StatusUnprocessableEntity},
		{"prázdné SKU", `{"lines":[{"sku":"","quantity":1,"unit_price_cents":100}]}`, http.StatusUnprocessableEntity},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := do(t, newTestHandler(), http.MethodPost, "/orders", tt.body)
			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, chci %d (tělo: %s)", rec.Code, tt.wantStatus, rec.Body.String())
			}
			if got := rec.Header().Get("Content-Type"); got != exercise.ProblemContentType {
				t.Errorf("Content-Type = %q, chci %q", got, exercise.ProblemContentType)
			}
			body := decodeBody(t, rec)
			if body["title"] == nil || body["title"] == "" {
				t.Errorf("v problem+json chybí title: %v", body)
			}
			if body["status"] != float64(tt.wantStatus) {
				t.Errorf("status v těle = %v, chci %d", body["status"], tt.wantStatus)
			}
		})
	}
}

func TestHTTPNenalezeno(t *testing.T) {
	h := newTestHandler()
	for _, path := range []string{"/orders/neexistuje", "/orders/neexistuje/pay", "/orders/neexistuje/cancel"} {
		method := http.MethodGet
		if strings.Count(path, "/") > 2 {
			method = http.MethodPost
		}
		rec := do(t, h, method, path, "")
		if rec.Code != http.StatusNotFound {
			t.Errorf("%s %s = %d, chci 404 (tělo: %s)", method, path, rec.Code, rec.Body.String())
		}
		if got := rec.Header().Get("Content-Type"); got != exercise.ProblemContentType {
			t.Errorf("%s: Content-Type = %q, chci %q", path, got, exercise.ProblemContentType)
		}
	}
}

func TestHTTPZivotniCyklus(t *testing.T) {
	h := newTestHandler()

	created := do(t, h, http.MethodPost, "/orders", validBody)
	if created.Code != http.StatusCreated {
		t.Fatalf("POST /orders = %d, chci 201", created.Code)
	}
	id, _ := decodeBody(t, created)["id"].(string)
	if id == "" {
		t.Fatal("v odpovědi chybí id")
	}

	steps := []struct {
		path string
		want string
	}{
		{"/orders/" + id + "/pay", "paid"},
		{"/orders/" + id + "/ship", "shipped"},
	}
	for _, s := range steps {
		rec := do(t, h, http.MethodPost, s.path, "")
		if rec.Code != http.StatusOK {
			t.Fatalf("POST %s = %d, chci 200 (tělo: %s)", s.path, rec.Code, rec.Body.String())
		}
		if got := decodeBody(t, rec)["status"]; got != s.want {
			t.Errorf("POST %s: stav = %v, chci %q", s.path, got, s.want)
		}
	}

	got := do(t, h, http.MethodGet, "/orders/"+id, "")
	if got.Code != http.StatusOK {
		t.Fatalf("GET /orders/%s = %d, chci 200", id, got.Code)
	}
	if s := decodeBody(t, got)["status"]; s != "shipped" {
		t.Errorf("GET: stav = %v, chci \"shipped\"", s)
	}
}

func TestHTTPKonfliktStavu(t *testing.T) {
	h := newTestHandler()
	created := do(t, h, http.MethodPost, "/orders", validBody)
	id, _ := decodeBody(t, created)["id"].(string)

	// odeslání bez zaplacení
	if rec := do(t, h, http.MethodPost, "/orders/"+id+"/ship", ""); rec.Code != http.StatusConflict {
		t.Errorf("ship nezaplacené = %d, chci 409 (tělo: %s)", rec.Code, rec.Body.String())
	}

	if rec := do(t, h, http.MethodPost, "/orders/"+id+"/pay", ""); rec.Code != http.StatusOK {
		t.Fatalf("pay = %d, chci 200", rec.Code)
	}
	if rec := do(t, h, http.MethodPost, "/orders/"+id+"/ship", ""); rec.Code != http.StatusOK {
		t.Fatalf("ship = %d, chci 200", rec.Code)
	}

	rec := do(t, h, http.MethodPost, "/orders/"+id+"/cancel", "")
	if rec.Code != http.StatusConflict {
		t.Errorf("cancel odeslané = %d, chci 409 (tělo: %s)", rec.Code, rec.Body.String())
	}
	if got := rec.Header().Get("Content-Type"); got != exercise.ProblemContentType {
		t.Errorf("Content-Type = %q, chci %q", got, exercise.ProblemContentType)
	}
}

func TestHTTPZruseni(t *testing.T) {
	h := newTestHandler()
	created := do(t, h, http.MethodPost, "/orders", validBody)
	id, _ := decodeBody(t, created)["id"].(string)

	rec := do(t, h, http.MethodPost, "/orders/"+id+"/cancel", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("cancel nové objednávky = %d, chci 200 (tělo: %s)", rec.Code, rec.Body.String())
	}
	if got := decodeBody(t, rec)["status"]; got != "cancelled" {
		t.Errorf("stav = %v, chci \"cancelled\"", got)
	}
}

func TestHTTPMetodyARouting(t *testing.T) {
	h := newTestHandler()
	if rec := do(t, h, http.MethodGet, "/healthz", ""); rec.Code != http.StatusOK {
		t.Errorf("GET /healthz = %d, chci 200", rec.Code)
	}
	if rec := do(t, h, http.MethodGet, "/orders", ""); rec.Code == http.StatusOK {
		t.Errorf("GET /orders = 200, chci odmítnutí — trasa je jen pro POST")
	}
	if rec := do(t, h, http.MethodDelete, "/orders/ord-1", ""); rec.Code == http.StatusOK {
		t.Errorf("DELETE /orders/ord-1 = 200, chci odmítnutí")
	}
}

func TestHTTPNeprozradiInterniChybu(t *testing.T) {
	repo := newFakeRepo()
	repo.saveErr = errors.New("pq: connection to 10.0.0.7 refused")
	h := exercise.NewHandler(exercise.NewService(repo, &stubIDs{}))

	rec := do(t, h, http.MethodPost, "/orders", validBody)
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, chci 500 (tělo: %s)", rec.Code, rec.Body.String())
	}
	if strings.Contains(rec.Body.String(), "10.0.0.7") {
		t.Errorf("tělo prozradilo interní chybu: %s", rec.Body.String())
	}
}

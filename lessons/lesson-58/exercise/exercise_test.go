package exercise_test

import (
	"errors"
	"math"
	"reflect"
	"sort"
	"testing"
	"time"

	exercise "github.com/rdurica/go-deep/lessons/lesson-58/exercise"
)

// fakeClock vrací hodiny, které při každém volání posunou čas o step.
func fakeClock(start time.Time, step time.Duration) func() time.Time {
	cur := start.Add(-step)
	return func() time.Time {
		cur = cur.Add(step)
		return cur
	}
}

func TestSeverityString(t *testing.T) {
	tests := map[exercise.Severity]string{
		exercise.SeverityInfo:  "INFO",
		exercise.SeverityWarn:  "WARN",
		exercise.SeverityError: "ERROR",
	}
	for in, want := range tests {
		if got := in.String(); got != want {
			t.Errorf("Severity(%d).String() = %q, chci %q", int(in), got, want)
		}
	}
}

func TestMergeChecklists(t *testing.T) {
	base := []exercise.CheckItem{
		{ID: "err-ignored", Text: "žádné _ = err", Severity: exercise.SeverityError},
		{ID: "ctx-first", Text: "context je první parametr", Severity: exercise.SeverityWarn},
		{ID: "iface-small", Text: "malé interfacy", Severity: exercise.SeverityInfo},
	}
	personal := []exercise.CheckItem{
		{ID: "ctx-first", Text: "context nikdy ve structu (moje slabina)", Severity: exercise.SeverityError},
		{ID: "rows-err", Text: "nezapomeň na rows.Err()", Severity: exercise.SeverityError},
	}

	got := exercise.MergeChecklists(base, personal)
	want := []exercise.CheckItem{
		{ID: "err-ignored", Text: "žádné _ = err", Severity: exercise.SeverityError},
		{ID: "ctx-first", Text: "context nikdy ve structu (moje slabina)", Severity: exercise.SeverityError},
		{ID: "iface-small", Text: "malé interfacy", Severity: exercise.SeverityInfo},
		{ID: "rows-err", Text: "nezapomeň na rows.Err()", Severity: exercise.SeverityError},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("MergeChecklists() =\n%+v\nchci\n%+v", got, want)
	}
}

func TestMergeChecklistsEdgeCases(t *testing.T) {
	t.Run("empty inputs", func(t *testing.T) {
		if got := exercise.MergeChecklists(nil, nil); len(got) != 0 {
			t.Errorf("MergeChecklists(nil, nil) = %+v, chci prázdný výsledek", got)
		}
	})

	t.Run("personal only", func(t *testing.T) {
		personal := []exercise.CheckItem{{ID: "a", Text: "A"}}
		got := exercise.MergeChecklists(nil, personal)
		if !reflect.DeepEqual(got, personal) {
			t.Errorf("MergeChecklists(nil, personal) = %+v, chci %+v", got, personal)
		}
	})

	t.Run("duplicate in base", func(t *testing.T) {
		base := []exercise.CheckItem{
			{ID: "a", Text: "first"},
			{ID: "a", Text: "druhá"},
			{ID: "b", Text: "béčko"},
		}
		got := exercise.MergeChecklists(base, nil)
		want := []exercise.CheckItem{{ID: "a", Text: "first"}, {ID: "b", Text: "béčko"}}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("MergeChecklists() = %+v, chci %+v", got, want)
		}
	})

	t.Run("item without ID is dropped", func(t *testing.T) {
		got := exercise.MergeChecklists(
			[]exercise.CheckItem{{ID: "", Text: "bez ID"}, {ID: "a", Text: "A"}},
			[]exercise.CheckItem{{ID: "", Text: "taky bez ID"}},
		)
		want := []exercise.CheckItem{{ID: "a", Text: "A"}}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("MergeChecklists() = %+v, chci %+v", got, want)
		}
	})
}

func TestRoleString(t *testing.T) {
	tests := map[exercise.Role]string{
		exercise.RoleNone:   "none",
		exercise.RoleSpec:   "spec",
		exercise.RoleTests:  "tests",
		exercise.RoleImpl:   "impl",
		exercise.RoleReview: "review",
		exercise.RoleDone:   "done",
		exercise.Role(42):   "none",
	}
	for in, want := range tests {
		if got := in.String(); got != want {
			t.Errorf("Role(%d).String() = %q, chci %q", int(in), got, want)
		}
	}
}

func TestSessionHappyPath(t *testing.T) {
	start := time.Date(2024, time.May, 1, 9, 0, 0, 0, time.UTC)
	s := exercise.NewSession(fakeClock(start, time.Minute))

	if got := s.Current(); got != exercise.RoleNone {
		t.Errorf("Current() před startem = %v, chci none", got)
	}
	if err := s.Start(exercise.RoleSpec); err != nil {
		t.Fatalf("Start(spec) = %v", err)
	}
	steps := []struct {
		to     exercise.Role
		reason string
	}{
		{exercise.RoleTests, "spec je hotová, píšu acceptance testy"},
		{exercise.RoleImpl, "testy padají, agent smí implementovat"},
		{exercise.RoleReview, "testy prochází, jdu číst diff"},
	}
	for _, st := range steps {
		if err := s.Handoff(st.to, st.reason); err != nil {
			t.Fatalf("Handoff(%v) = %v", st.to, err)
		}
	}
	if err := s.Finish(); err != nil {
		t.Fatalf("Finish() = %v", err)
	}
	if got := s.Current(); got != exercise.RoleDone {
		t.Errorf("Current() po Finish = %v, chci done", got)
	}

	timeline := s.Timeline()
	if len(timeline) != 5 {
		t.Fatalf("Timeline() má %d událostí, chci 5: %+v", len(timeline), timeline)
	}
	wantPairs := [][2]exercise.Role{
		{exercise.RoleNone, exercise.RoleSpec},
		{exercise.RoleSpec, exercise.RoleTests},
		{exercise.RoleTests, exercise.RoleImpl},
		{exercise.RoleImpl, exercise.RoleReview},
		{exercise.RoleReview, exercise.RoleDone},
	}
	for i, want := range wantPairs {
		if timeline[i].From != want[0] || timeline[i].To != want[1] {
			t.Errorf("událost %d = %v→%v, chci %v→%v", i, timeline[i].From, timeline[i].To, want[0], want[1])
		}
		if timeline[i].Reason == "" {
			t.Errorf("událost %d nemá důvod", i)
		}
		wantAt := start.Add(time.Duration(i) * time.Minute)
		if !timeline[i].At.Equal(wantAt) {
			t.Errorf("událost %d má čas %v, chci %v", i, timeline[i].At, wantAt)
		}
	}
}

func TestSessionBackwardHandoff(t *testing.T) {
	s := exercise.NewSession(fakeClock(time.Unix(0, 0).UTC(), time.Second))
	if err := s.Start(exercise.RoleReview); err != nil {
		t.Fatalf("Start(review) = %v", err)
	}
	if err := s.Handoff(exercise.RoleImpl, "review našel chybu v normalizaci URL"); err != nil {
		t.Fatalf("Handoff(review→impl) = %v", err)
	}
	if err := s.Handoff(exercise.RoleTests, "chybí test na utm_ parametry"); err != nil {
		t.Fatalf("Handoff(impl→tests) = %v", err)
	}
	if got := s.Current(); got != exercise.RoleTests {
		t.Errorf("Current() = %v, chci tests", got)
	}
}

func TestSessionInvalidTransitions(t *testing.T) {
	tests := []struct {
		name string
		from exercise.Role
		to   exercise.Role
	}{
		{"skipping tests", exercise.RoleSpec, exercise.RoleImpl},
		{"skok na konec", exercise.RoleSpec, exercise.RoleReview},
		{"back two steps", exercise.RoleReview, exercise.RoleTests},
		{"done via Handoff", exercise.RoleReview, exercise.RoleDone},
		{"na none", exercise.RoleImpl, exercise.RoleNone},
		{"na sebe", exercise.RoleImpl, exercise.RoleImpl},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := exercise.NewSession(fakeClock(time.Unix(0, 0).UTC(), time.Second))
			if err := s.Start(tt.from); err != nil {
				t.Fatalf("Start(%v) = %v", tt.from, err)
			}
			err := s.Handoff(tt.to, "protože")
			if !errors.Is(err, exercise.ErrInvalidTransition) {
				t.Errorf("Handoff(%v→%v) = %v, chci ErrInvalidTransition", tt.from, tt.to, err)
			}
			if got := s.Current(); got != tt.from {
				t.Errorf("neplatný přechod změnil roli na %v, chci %v", got, tt.from)
			}
		})
	}
}

func TestSessionErrors(t *testing.T) {
	clock := fakeClock(time.Unix(0, 0).UTC(), time.Second)

	t.Run("handoff before start", func(t *testing.T) {
		s := exercise.NewSession(clock)
		if err := s.Handoff(exercise.RoleTests, "důvod"); !errors.Is(err, exercise.ErrNotStarted) {
			t.Errorf("Handoff() bez startu = %v, chci ErrNotStarted", err)
		}
		if err := s.Finish(); !errors.Is(err, exercise.ErrNotStarted) {
			t.Errorf("Finish() bez startu = %v, chci ErrNotStarted", err)
		}
	})

	t.Run("double start", func(t *testing.T) {
		s := exercise.NewSession(clock)
		if err := s.Start(exercise.RoleSpec); err != nil {
			t.Fatalf("Start() = %v", err)
		}
		if err := s.Start(exercise.RoleTests); !errors.Is(err, exercise.ErrAlreadyStarted) {
			t.Errorf("druhý Start() = %v, chci ErrAlreadyStarted", err)
		}
	})

	t.Run("start in invalid role", func(t *testing.T) {
		for _, r := range []exercise.Role{exercise.RoleNone, exercise.RoleDone, exercise.Role(99)} {
			s := exercise.NewSession(clock)
			if err := s.Start(r); !errors.Is(err, exercise.ErrInvalidRole) {
				t.Errorf("Start(%d) = %v, chci ErrInvalidRole", int(r), err)
			}
		}
	})

	t.Run("handoff without reason", func(t *testing.T) {
		s := exercise.NewSession(clock)
		if err := s.Start(exercise.RoleSpec); err != nil {
			t.Fatalf("Start() = %v", err)
		}
		if err := s.Handoff(exercise.RoleTests, "   "); !errors.Is(err, exercise.ErrMissingReason) {
			t.Errorf("Handoff() bez důvodu = %v, chci ErrMissingReason", err)
		}
	})

	t.Run("finish mimo review", func(t *testing.T) {
		s := exercise.NewSession(clock)
		if err := s.Start(exercise.RoleImpl); err != nil {
			t.Fatalf("Start() = %v", err)
		}
		if err := s.Finish(); !errors.Is(err, exercise.ErrInvalidTransition) {
			t.Errorf("Finish() z impl = %v, chci ErrInvalidTransition", err)
		}
	})

	t.Run("nothing after finished", func(t *testing.T) {
		s := exercise.NewSession(clock)
		if err := s.Start(exercise.RoleReview); err != nil {
			t.Fatalf("Start() = %v", err)
		}
		if err := s.Finish(); err != nil {
			t.Fatalf("Finish() = %v", err)
		}
		if err := s.Handoff(exercise.RoleImpl, "ještě jedna oprava"); !errors.Is(err, exercise.ErrFinished) {
			t.Errorf("Handoff() po Finish = %v, chci ErrFinished", err)
		}
		if err := s.Finish(); !errors.Is(err, exercise.ErrFinished) {
			t.Errorf("druhý Finish() = %v, chci ErrFinished", err)
		}
	})
}

func TestSessionTimelineIsCopy(t *testing.T) {
	s := exercise.NewSession(fakeClock(time.Unix(0, 0).UTC(), time.Second))
	if err := s.Start(exercise.RoleSpec); err != nil {
		t.Fatalf("Start() = %v", err)
	}
	timeline := s.Timeline()
	if len(timeline) == 0 {
		t.Fatal("Timeline() po Start je prázdná, chci alespoň jednu událost")
	}
	timeline[0].Reason = "podvrženo"

	if got := s.Timeline()[0].Reason; got == "podvrženo" {
		t.Error("Timeline() vrací vnitřní slice, chci kopii")
	}
}

func approx(a, b float64) bool { return math.Abs(a-b) < 1e-9 }

func TestScoreReview(t *testing.T) {
	planted := []exercise.Finding{
		{ID: "f1", Category: "errors"},
		{ID: "f2", Category: "context"},
		{ID: "f3", Category: "concurrency"},
		{ID: "f4", Category: "errors"},
	}
	found := []exercise.Finding{
		{ID: "f1", Category: "errors"},
		{ID: "f2", Category: "context"},
		{ID: "x9", Category: "style"},
	}

	got := exercise.ScoreReview(found, planted)
	if got.TruePositives != 2 || got.FalsePositives != 1 || got.Missed != 2 {
		t.Errorf("ScoreReview() = TP %d, FP %d, missed %d; chci 2, 1, 2",
			got.TruePositives, got.FalsePositives, got.Missed)
	}
	if !approx(got.Precision, 2.0/3.0) {
		t.Errorf("Precision = %v, chci %v", got.Precision, 2.0/3.0)
	}
	if !approx(got.Recall, 0.5) {
		t.Errorf("Recall = %v, chci 0.5", got.Recall)
	}

	want := []string{
		exercise.RecommendLesson("errors"),
		exercise.RecommendLesson("concurrency"),
	}
	sort.Strings(want)
	if len(got.Review) != len(want) {
		t.Fatalf("Review = %+v, chci %+v", got.Review, want)
	}
	for i := range want {
		if got.Review[i] != want[i] {
			t.Errorf("Review[%d] = %q, chci %q (seřazeno abecedně)", i, got.Review[i], want[i])
		}
	}
}

func TestScoreReviewEdgeCases(t *testing.T) {
	t.Run("empty inputs", func(t *testing.T) {
		got := exercise.ScoreReview(nil, nil)
		if got.TruePositives != 0 || got.FalsePositives != 0 || got.Missed != 0 {
			t.Errorf("ScoreReview(nil, nil) = %+v, chci nuly", got)
		}
		if !approx(got.Precision, 0) {
			t.Errorf("Precision = %v, chci 0 (nic nenalezeno)", got.Precision)
		}
		if !approx(got.Recall, 1) {
			t.Errorf("Recall = %v, chci 1 (nebylo co najít)", got.Recall)
		}
		if len(got.Review) != 0 {
			t.Errorf("Review = %+v, chci prázdné", got.Review)
		}
	})

	t.Run("nic nenalezeno", func(t *testing.T) {
		planted := []exercise.Finding{{ID: "f1", Category: "http"}}
		got := exercise.ScoreReview(nil, planted)
		if got.Missed != 1 || !approx(got.Recall, 0) || !approx(got.Precision, 0) {
			t.Errorf("ScoreReview(nil, planted) = %+v, chci missed 1, recall 0, precision 0", got)
		}
		if len(got.Review) != 1 || got.Review[0] != exercise.RecommendLesson("http") {
			t.Errorf("Review = %+v, chci doporučení pro http", got.Review)
		}
	})

	t.Run("only false positives", func(t *testing.T) {
		got := exercise.ScoreReview([]exercise.Finding{{ID: "a"}, {ID: "b"}}, nil)
		if got.FalsePositives != 2 || !approx(got.Precision, 0) || !approx(got.Recall, 1) {
			t.Errorf("ScoreReview() = %+v, chci FP 2, precision 0, recall 1", got)
		}
	})

	t.Run("duplicate findings counted once", func(t *testing.T) {
		planted := []exercise.Finding{{ID: "f1", Category: "errors"}}
		found := []exercise.Finding{{ID: "f1"}, {ID: "f1"}, {ID: "f1"}}
		got := exercise.ScoreReview(found, planted)
		if got.TruePositives != 1 || got.FalsePositives != 0 {
			t.Errorf("ScoreReview() = TP %d, FP %d; chci 1, 0", got.TruePositives, got.FalsePositives)
		}
		if !approx(got.Precision, 1) || !approx(got.Recall, 1) {
			t.Errorf("Precision/Recall = %v/%v, chci 1/1", got.Precision, got.Recall)
		}
	})
}

func TestRecommendLesson(t *testing.T) {
	categories := []string{"errors", "context", "concurrency", "http", "design", "testing"}
	seen := make(map[string]string, len(categories))
	for _, c := range categories {
		rec := exercise.RecommendLesson(c)
		if rec == "" {
			t.Errorf("RecommendLesson(%q) = prázdný řetězec", c)
		}
		if other, dup := seen[rec]; dup {
			t.Errorf("RecommendLesson(%q) vrací totéž co %q: %q", c, other, rec)
		}
		seen[rec] = c
	}
	if exercise.RecommendLesson("ERRORS") != exercise.RecommendLesson("errors") {
		t.Error("RecommendLesson má být case-insensitive")
	}
	if got := exercise.RecommendLesson("něco úplně jiného"); got == "" {
		t.Error("RecommendLesson(neznámá kategorie) = prázdný řetězec, chci výchozí doporučení")
	}
}

package solutions_test

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"testing"

	exercise "github.com/rdurica/go-deep/lessons/lesson-35/solutions"
)

// MemoryRepo musí splňovat port. Tohle je celý „implements" v Go.
var _ exercise.UserRepo = (*exercise.MemoryRepo)(nil)

func user(id, email string) exercise.User {
	return exercise.User{ID: id, Email: email, Name: "Uživatel " + id, Active: true}
}

func TestMemoryRepoUlozACti(t *testing.T) {
	ctx := context.Background()
	repo := exercise.NewMemoryRepo()

	alice := user("u-1", "alice@example.com")
	if err := repo.Save(ctx, alice); err != nil {
		t.Fatalf("Save = %v, chci nil", err)
	}

	got, err := repo.Get(ctx, "u-1")
	if err != nil {
		t.Fatalf("Get = %v, chci nil", err)
	}
	if got != alice {
		t.Errorf("Get = %+v, chci %+v", got, alice)
	}

	updated := alice
	updated.Name = "Alice Nová"
	updated.Active = false
	if err := repo.Save(ctx, updated); err != nil {
		t.Fatalf("Save (upsert) = %v", err)
	}
	got, err = repo.Get(ctx, "u-1")
	if err != nil {
		t.Fatalf("Get po upsertu = %v", err)
	}
	if got != updated {
		t.Errorf("Get = %+v, chci %+v", got, updated)
	}

	if err := repo.Delete(ctx, "u-1"); err != nil {
		t.Fatalf("Delete = %v", err)
	}
	if _, err := repo.Get(ctx, "u-1"); !errors.Is(err, exercise.ErrNotFound) {
		t.Errorf("Get po Delete = %v, chci ErrNotFound", err)
	}
}

func TestMemoryRepoNenalezeno(t *testing.T) {
	ctx := context.Background()
	repo := exercise.NewMemoryRepo()

	if _, err := repo.Get(ctx, "nic"); !errors.Is(err, exercise.ErrNotFound) {
		t.Errorf("Get(nic) = %v, chci ErrNotFound", err)
	}
	if err := repo.Delete(ctx, "nic"); !errors.Is(err, exercise.ErrNotFound) {
		t.Errorf("Delete(nic) = %v, chci ErrNotFound", err)
	}
}

func TestMemoryRepoNeplatnyUzivatel(t *testing.T) {
	ctx := context.Background()
	repo := exercise.NewMemoryRepo()

	tests := []struct {
		name string
		u    exercise.User
	}{
		{"prázdné ID", exercise.User{Email: "a@example.com"}},
		{"ID jen z mezer", exercise.User{ID: "  ", Email: "a@example.com"}},
		{"prázdný e-mail", exercise.User{ID: "u-1"}},
		{"e-mail jen z mezer", exercise.User{ID: "u-1", Email: "\t"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := repo.Save(ctx, tt.u); !errors.Is(err, exercise.ErrInvalidUser) {
				t.Errorf("Save(%+v) = %v, chci ErrInvalidUser", tt.u, err)
			}
		})
	}
}

func TestMemoryRepoListJeSerazeny(t *testing.T) {
	ctx := context.Background()
	repo := exercise.NewMemoryRepo()

	if got, err := repo.List(ctx); err != nil || len(got) != 0 {
		t.Fatalf("List na prázdném repozitáři = (%v, %v), chci (prázdné, nil)", got, err)
	}

	for _, id := range []string{"u-3", "u-1", "u-2"} {
		if err := repo.Save(ctx, user(id, id+"@example.com")); err != nil {
			t.Fatalf("Save = %v", err)
		}
	}

	// Iterace mapy je náhodná, takže test běží víckrát.
	for i := 0; i < 20; i++ {
		list, err := repo.List(ctx)
		if err != nil {
			t.Fatalf("List = %v", err)
		}
		var ids []string
		for _, u := range list {
			ids = append(ids, u.ID)
		}
		if strings.Join(ids, ",") != "u-1,u-2,u-3" {
			t.Fatalf("List = %v, chci [u-1 u-2 u-3]", ids)
		}
	}
}

func TestMemoryRepoRespektujeContext(t *testing.T) {
	repo := exercise.NewMemoryRepo()
	if err := repo.Save(context.Background(), user("u-1", "a@example.com")); err != nil {
		t.Fatalf("Save = %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if _, err := repo.Get(ctx, "u-1"); !errors.Is(err, context.Canceled) {
		t.Errorf("Get se zrušeným contextem = %v, chci context.Canceled", err)
	}
	if err := repo.Save(ctx, user("u-2", "b@example.com")); !errors.Is(err, context.Canceled) {
		t.Errorf("Save se zrušeným contextem = %v, chci context.Canceled", err)
	}
	if err := repo.Delete(ctx, "u-1"); !errors.Is(err, context.Canceled) {
		t.Errorf("Delete se zrušeným contextem = %v, chci context.Canceled", err)
	}
	if _, err := repo.List(ctx); !errors.Is(err, context.Canceled) {
		t.Errorf("List se zrušeným contextem = %v, chci context.Canceled", err)
	}
}

func TestMemoryRepoSoubezne(t *testing.T) {
	ctx := context.Background()
	repo := exercise.NewMemoryRepo()

	const n = 100
	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			id := fmt.Sprintf("u-%03d", i)
			if err := repo.Save(ctx, user(id, id+"@example.com")); err != nil {
				t.Errorf("Save = %v", err)
				return
			}
			if _, err := repo.Get(ctx, id); err != nil {
				t.Errorf("Get = %v", err)
			}
			if _, err := repo.List(ctx); err != nil {
				t.Errorf("List = %v", err)
			}
		}(i)
	}
	wg.Wait()

	list, err := repo.List(ctx)
	if err != nil {
		t.Fatalf("List = %v", err)
	}
	if len(list) != n {
		t.Errorf("len(List()) = %d, chci %d", len(list), n)
	}
}

func TestBuildSelect(t *testing.T) {
	tests := []struct {
		name      string
		table     string
		cols      []string
		filters   map[string]any
		wantQuery string
		wantArgs  []any
	}{
		{
			name:      "bez filtru",
			table:     "users",
			cols:      []string{"id", "email"},
			wantQuery: "SELECT id, email FROM users",
			wantArgs:  nil,
		},
		{
			name:      "jeden filtr",
			table:     "users",
			cols:      []string{"id"},
			filters:   map[string]any{"active": true},
			wantQuery: "SELECT id FROM users WHERE active = $1",
			wantArgs:  []any{true},
		},
		{
			name:      "víc filtrů se řadí abecedně",
			table:     "users",
			cols:      []string{"id", "name"},
			filters:   map[string]any{"name": "Alice", "active": true, "email": "a@example.com"},
			wantQuery: "SELECT id, name FROM users WHERE active = $1 AND email = $2 AND name = $3",
			wantArgs:  []any{true, "a@example.com", "Alice"},
		},
		{
			name:      "jiná tabulka",
			table:     "orders",
			cols:      []string{"id", "total_cents"},
			filters:   map[string]any{"user_id": "u-1"},
			wantQuery: "SELECT id, total_cents FROM orders WHERE user_id = $1",
			wantArgs:  []any{"u-1"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Opakování odhalí nedeterministické pořadí z iterace mapy.
			for i := 0; i < 20; i++ {
				q, args, err := exercise.BuildSelect(tt.table, tt.cols, tt.filters)
				if err != nil {
					t.Fatalf("BuildSelect = %v, chci nil", err)
				}
				if q != tt.wantQuery {
					t.Fatalf("query = %q, chci %q", q, tt.wantQuery)
				}
				if len(args) != len(tt.wantArgs) {
					t.Fatalf("args = %v, chci %v", args, tt.wantArgs)
				}
				for j := range args {
					if args[j] != tt.wantArgs[j] {
						t.Fatalf("args = %v, chci %v", args, tt.wantArgs)
					}
				}
			}
		})
	}
}

func TestBuildSelectChyby(t *testing.T) {
	tests := []struct {
		name    string
		table   string
		cols    []string
		filters map[string]any
		want    error
	}{
		{"neznámá tabulka", "sessions", []string{"id"}, nil, exercise.ErrUnknownTable},
		{"prázdná tabulka", "", []string{"id"}, nil, exercise.ErrUnknownTable},
		{"žádné sloupce", "users", nil, nil, exercise.ErrNoColumns},
		{"prázdný slice sloupců", "users", []string{}, nil, exercise.ErrNoColumns},
		{"neznámý sloupec", "users", []string{"password"}, nil, exercise.ErrUnknownColumn},
		{"sloupec z jiné tabulky", "users", []string{"total_cents"}, nil, exercise.ErrUnknownColumn},
		{"neznámý sloupec ve filtru", "users", []string{"id"}, map[string]any{"password": "x"}, exercise.ErrUnknownColumn},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q, args, err := exercise.BuildSelect(tt.table, tt.cols, tt.filters)
			if !errors.Is(err, tt.want) {
				t.Fatalf("BuildSelect = %v, chci %v", err, tt.want)
			}
			if q != "" || args != nil {
				t.Errorf("při chybě chci prázdný dotaz a nil args, mám (%q, %v)", q, args)
			}
		})
	}
}

func TestBuildSelectOdolavaInjection(t *testing.T) {
	t.Run("injection ve jméně tabulky", func(t *testing.T) {
		_, _, err := exercise.BuildSelect("users; DROP TABLE users --", []string{"id"}, nil)
		if !errors.Is(err, exercise.ErrUnknownTable) {
			t.Errorf("BuildSelect = %v, chci ErrUnknownTable", err)
		}
	})

	t.Run("injection ve jméně sloupce", func(t *testing.T) {
		for _, col := range []string{"id, password", "id; DROP TABLE users", "1=1", "*"} {
			_, _, err := exercise.BuildSelect("users", []string{col}, nil)
			if !errors.Is(err, exercise.ErrUnknownColumn) {
				t.Errorf("BuildSelect(col=%q) = %v, chci ErrUnknownColumn", col, err)
			}
		}
	})

	t.Run("injection ve jméně filtru", func(t *testing.T) {
		filters := map[string]any{"email": "a@example.com", "1=1 OR id": "x"}
		_, _, err := exercise.BuildSelect("users", []string{"id"}, filters)
		if !errors.Is(err, exercise.ErrUnknownColumn) {
			t.Errorf("BuildSelect = %v, chci ErrUnknownColumn", err)
		}
	})

	t.Run("injection v hodnotě skončí v argumentech", func(t *testing.T) {
		payload := "' OR 1=1 --"
		q, args, err := exercise.BuildSelect("users", []string{"id"}, map[string]any{"email": payload})
		if err != nil {
			t.Fatalf("BuildSelect = %v, chci nil — hodnota se nevaliduje, jde do placeholderu", err)
		}
		if q != "SELECT id FROM users WHERE email = $1" {
			t.Errorf("query = %q, chci dotaz s placeholderem", q)
		}
		if strings.Contains(q, payload) || strings.Contains(q, "OR 1=1") {
			t.Errorf("query = %q — hodnota se nikdy nesmí vlepit do textu dotazu", q)
		}
		if len(args) != 1 || args[0] != payload {
			t.Errorf("args = %v, chci [%q]", args, payload)
		}
	})
}

func mig(v int, name string) exercise.Migration {
	return exercise.Migration{Version: v, Name: name, Up: fmt.Sprintf("-- migrace %d", v)}
}

func TestPlan(t *testing.T) {
	all := []exercise.Migration{
		mig(3, "add_orders"),
		mig(1, "create_users"),
		mig(2, "add_email_index"),
	}

	tests := []struct {
		name    string
		applied []int
		want    []int
	}{
		{"nic neaplikováno", nil, []int{1, 2, 3}},
		{"prázdný slice", []int{}, []int{1, 2, 3}},
		{"první hotová", []int{1}, []int{2, 3}},
		{"díra uprostřed", []int{1, 3}, []int{2}},
		{"všechno hotové", []int{1, 2, 3}, nil},
		{"duplicitní záznam v applied", []int{1, 1, 2}, []int{3}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pending, err := exercise.Plan(tt.applied, all)
			if err != nil {
				t.Fatalf("Plan = %v, chci nil", err)
			}
			if len(pending) != len(tt.want) {
				t.Fatalf("Plan = %v, chci verze %v", versions(pending), tt.want)
			}
			for i, m := range pending {
				if m.Version != tt.want[i] {
					t.Fatalf("Plan = %v, chci verze %v", versions(pending), tt.want)
				}
				if m.Up == "" || m.Name == "" {
					t.Errorf("Plan vrátil neúplnou migraci %+v", m)
				}
			}
		})
	}
}

func TestPlanChyby(t *testing.T) {
	tests := []struct {
		name    string
		applied []int
		all     []exercise.Migration
		want    error
	}{
		{
			name: "duplicitní verze",
			all:  []exercise.Migration{mig(1, "create_users"), mig(1, "create_users_again")},
			want: exercise.ErrDuplicateVersion,
		},
		{
			name:    "drift — aplikovaná verze chybí",
			applied: []int{1, 7},
			all:     []exercise.Migration{mig(1, "create_users"), mig(2, "add_index")},
			want:    exercise.ErrDrift,
		},
		{
			name: "nekladná verze",
			all:  []exercise.Migration{mig(0, "nula")},
			want: exercise.ErrInvalidMigration,
		},
		{
			name: "záporná verze",
			all:  []exercise.Migration{mig(-1, "zapor")},
			want: exercise.ErrInvalidMigration,
		},
		{
			name: "prázdné jméno",
			all:  []exercise.Migration{mig(1, "   ")},
			want: exercise.ErrInvalidMigration,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := exercise.Plan(tt.applied, tt.all)
			if !errors.Is(err, tt.want) {
				t.Fatalf("Plan = %v, chci %v", err, tt.want)
			}
			if got != nil {
				t.Errorf("při chybě chci nil plán, mám %v", versions(got))
			}
		})
	}
}

func TestPlanPrazdnySeznam(t *testing.T) {
	got, err := exercise.Plan(nil, nil)
	if err != nil {
		t.Fatalf("Plan(nil, nil) = %v, chci nil", err)
	}
	if len(got) != 0 {
		t.Errorf("Plan(nil, nil) = %v, chci prázdný plán", versions(got))
	}
}

func versions(ms []exercise.Migration) []int {
	out := make([]int, len(ms))
	for i, m := range ms {
		out[i] = m.Version
	}
	return out
}

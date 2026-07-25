package task_test

import (
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/rdurica/go-deep/projects/p02-http-api/internal/task"
)

// fixedClock vrací pevný čas, aby testy nezávisely na skutečných hodinách.
func fixedClock(t time.Time) func() time.Time {
	return func() time.Time { return t }
}

func TestStatusValid(t *testing.T) {
	t.Parallel()

	tests := map[task.Status]bool{
		task.StatusTodo:  true,
		task.StatusDoing: true,
		task.StatusDone:  true,
		"":               false,
		"hotovo":         false,
		"TODO":           false,
	}
	for status, want := range tests {
		if got := status.Valid(); got != want {
			t.Errorf("Status(%q).Valid() = %v, chci %v", status, got, want)
		}
	}
}

func TestCreateValidation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		title   string
		status  task.Status
		wantErr error
	}{
		{"platný úkol", "napsat testy", task.StatusTodo, nil},
		{"prázdný stav se doplní na todo", "napsat testy", "", nil},
		{"název se ořeže", "  napsat testy  ", task.StatusDoing, nil},
		{"prázdný název", "", task.StatusTodo, task.ErrEmptyTitle},
		{"název jen z mezer", "   ", task.StatusTodo, task.ErrEmptyTitle},
		{"příliš dlouhý název", strings.Repeat("a", task.MaxTitleLength+1), task.StatusTodo, task.ErrTitleTooLong},
		{"neznámý stav", "napsat testy", "hotovo", task.ErrInvalidStatus},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			s := task.NewStore()
			got, err := s.Create(tt.title, tt.status)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("Create(%q, %q) = %v, chci %v", tt.title, tt.status, err, tt.wantErr)
			}
			if tt.wantErr != nil {
				if got != (task.Task{}) {
					t.Errorf("při chybě chci nulový Task, mám %+v", got)
				}
				return
			}
			if got.Title != strings.TrimSpace(tt.title) {
				t.Errorf("Title = %q, chci %q", got.Title, strings.TrimSpace(tt.title))
			}
			if got.ID == "" {
				t.Error("Create nepřidělil ID")
			}
			if got.Status == "" {
				t.Error("Create nedoplnil výchozí stav")
			}
		})
	}
}

func TestCreateAssignsUniqueIDs(t *testing.T) {
	t.Parallel()

	s := task.NewStore()
	seen := make(map[string]bool)
	for i := 0; i < 10; i++ {
		got, err := s.Create("úkol", task.StatusTodo)
		if err != nil {
			t.Fatalf("Create selhal: %v", err)
		}
		if seen[got.ID] {
			t.Fatalf("ID %q se opakuje", got.ID)
		}
		seen[got.ID] = true
	}
}

func TestCreateSetsTimestamps(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 7, 25, 10, 0, 0, 0, time.UTC)
	s := task.NewStoreWithClock(fixedClock(now))

	got, err := s.Create("úkol", task.StatusTodo)
	if err != nil {
		t.Fatalf("Create selhal: %v", err)
	}
	if !got.CreatedAt.Equal(now) {
		t.Errorf("CreatedAt = %v, chci %v", got.CreatedAt, now)
	}
	if !got.UpdatedAt.Equal(now) {
		t.Errorf("UpdatedAt = %v, chci %v", got.UpdatedAt, now)
	}
}

func TestGet(t *testing.T) {
	t.Parallel()

	s := task.NewStore()
	created, err := s.Create("úkol", task.StatusTodo)
	if err != nil {
		t.Fatalf("Create selhal: %v", err)
	}

	got, err := s.Get(created.ID)
	if err != nil {
		t.Fatalf("Get(%q) = %v, chci nil", created.ID, err)
	}
	if got != created {
		t.Errorf("Get = %+v, chci %+v", got, created)
	}

	if _, err := s.Get("neexistuje"); !errors.Is(err, task.ErrNotFound) {
		t.Errorf("Get(neexistující) = %v, chci ErrNotFound", err)
	}
}

func TestListKeepsOrder(t *testing.T) {
	t.Parallel()

	s := task.NewStore()
	if got := s.List(); len(got) != 0 {
		t.Errorf("prázdné úložiště vrátilo %d úkolů", len(got))
	}
	if s.List() == nil {
		t.Error("List má vracet prázdný slice, ne nil")
	}

	titles := []string{"první", "druhý", "třetí"}
	for _, title := range titles {
		if _, err := s.Create(title, task.StatusTodo); err != nil {
			t.Fatalf("Create(%q) selhal: %v", title, err)
		}
	}

	got := s.List()
	if len(got) != len(titles) {
		t.Fatalf("List vrátil %d úkolů, chci %d", len(got), len(titles))
	}
	for i, want := range titles {
		if got[i].Title != want {
			t.Errorf("List()[%d].Title = %q, chci %q", i, got[i].Title, want)
		}
	}
}

func TestUpdate(t *testing.T) {
	t.Parallel()

	created := time.Date(2026, 7, 25, 10, 0, 0, 0, time.UTC)
	updated := created.Add(time.Hour)

	clock := created
	s := task.NewStoreWithClock(func() time.Time { return clock })

	orig, err := s.Create("původní", task.StatusTodo)
	if err != nil {
		t.Fatalf("Create selhal: %v", err)
	}

	clock = updated
	got, err := s.Update(orig.ID, "změněný", task.StatusDone)
	if err != nil {
		t.Fatalf("Update selhal: %v", err)
	}
	if got.Title != "změněný" || got.Status != task.StatusDone {
		t.Errorf("Update = %+v, chci title=změněný status=done", got)
	}
	if !got.CreatedAt.Equal(created) {
		t.Errorf("CreatedAt = %v, Update ho nesmí měnit (chci %v)", got.CreatedAt, created)
	}
	if !got.UpdatedAt.Equal(updated) {
		t.Errorf("UpdatedAt = %v, chci %v", got.UpdatedAt, updated)
	}

	// Změna se musí projevit i při dalším čtení.
	reread, err := s.Get(orig.ID)
	if err != nil {
		t.Fatalf("Get po Update selhal: %v", err)
	}
	if reread != got {
		t.Errorf("Get = %+v, chci %+v", reread, got)
	}
}

func TestUpdateErrors(t *testing.T) {
	t.Parallel()

	s := task.NewStore()
	created, err := s.Create("úkol", task.StatusTodo)
	if err != nil {
		t.Fatalf("Create selhal: %v", err)
	}

	if _, err := s.Update("neexistuje", "nový", task.StatusTodo); !errors.Is(err, task.ErrNotFound) {
		t.Errorf("Update(neexistující) = %v, chci ErrNotFound", err)
	}
	if _, err := s.Update(created.ID, "", task.StatusTodo); !errors.Is(err, task.ErrEmptyTitle) {
		t.Errorf("Update s prázdným názvem = %v, chci ErrEmptyTitle", err)
	}
	if _, err := s.Update(created.ID, "úkol", "neznámý"); !errors.Is(err, task.ErrInvalidStatus) {
		t.Errorf("Update s neznámým stavem = %v, chci ErrInvalidStatus", err)
	}
}

func TestDelete(t *testing.T) {
	t.Parallel()

	s := task.NewStore()
	first, err := s.Create("první", task.StatusTodo)
	if err != nil {
		t.Fatalf("Create selhal: %v", err)
	}
	if _, err := s.Create("druhý", task.StatusTodo); err != nil {
		t.Fatalf("Create selhal: %v", err)
	}

	if err := s.Delete(first.ID); err != nil {
		t.Fatalf("Delete = %v, chci nil", err)
	}
	if err := s.Delete(first.ID); !errors.Is(err, task.ErrNotFound) {
		t.Errorf("druhý Delete = %v, chci ErrNotFound", err)
	}
	if _, err := s.Get(first.ID); !errors.Is(err, task.ErrNotFound) {
		t.Errorf("Get po Delete = %v, chci ErrNotFound", err)
	}

	got := s.List()
	if len(got) != 1 || got[0].Title != "druhý" {
		t.Errorf("List po Delete = %+v, chci jen úkol \"druhý\"", got)
	}
}

// TestStoreConcurrent spouštěj s -race: souběžné zápisy i čtení musí projít
// bez závodu a bez ztraceného úkolu.
func TestStoreConcurrent(t *testing.T) {
	t.Parallel()

	s := task.NewStore()

	const writers = 25
	var wg sync.WaitGroup
	wg.Add(writers * 2)

	for i := 0; i < writers; i++ {
		go func() {
			defer wg.Done()
			if _, err := s.Create("souběžný úkol", task.StatusTodo); err != nil {
				t.Errorf("Create selhal: %v", err)
			}
		}()
		go func() {
			defer wg.Done()
			_ = s.List()
		}()
	}
	wg.Wait()

	if got := len(s.List()); got != writers {
		t.Errorf("po %d souběžných zápisech je v úložišti %d úkolů", writers, got)
	}
}

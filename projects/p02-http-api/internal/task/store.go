package task

import (
	"strconv"
	"sync"
	"time"
)

// Store je úložiště úkolů v paměti, bezpečné pro souběžný přístup.
// Zero value použitelná není — vytvoř ho přes NewStore.
type Store struct {
	mu    sync.RWMutex
	tasks map[string]Task
	order []string
	seq   int64
	now   func() time.Time
}

// NewStore vytvoří prázdné úložiště se systémovými hodinami.
func NewStore() *Store {
	return NewStoreWithClock(time.Now)
}

// NewStoreWithClock vytvoří úložiště s vlastním zdrojem času. Testy díky tomu
// nemusí spoléhat na skutečné hodiny.
func NewStoreWithClock(now func() time.Time) *Store {
	return &Store{
		tasks: make(map[string]Task),
		now:   now,
	}
}

// Create založí nový úkol. Vrací ErrEmptyTitle, ErrTitleTooLong nebo ErrInvalidStatus.
func (s *Store) Create(title string, status Status) (Task, error) {
	title, status = Normalize(title, status)
	if err := Validate(title, status); err != nil {
		return Task{}, err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.seq++
	now := s.now().UTC()
	t := Task{
		ID:        strconv.FormatInt(s.seq, 10),
		Title:     title,
		Status:    status,
		CreatedAt: now,
		UpdatedAt: now,
	}
	s.tasks[t.ID] = t
	s.order = append(s.order, t.ID)
	return t, nil
}

// Get vrátí úkol podle ID, nebo ErrNotFound.
func (s *Store) Get(id string) (Task, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	t, ok := s.tasks[id]
	if !ok {
		return Task{}, ErrNotFound
	}
	return t, nil
}

// List vrátí úkoly v pořadí vzniku. Vrácený slice patří volajícímu.
func (s *Store) List() []Task {
	s.mu.RLock()
	defer s.mu.RUnlock()

	out := make([]Task, 0, len(s.order))
	for _, id := range s.order {
		out = append(out, s.tasks[id])
	}
	return out
}

// Update přepíše název a stav existujícího úkolu.
func (s *Store) Update(id, title string, status Status) (Task, error) {
	title, status = Normalize(title, status)
	if err := Validate(title, status); err != nil {
		return Task{}, err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	t, ok := s.tasks[id]
	if !ok {
		return Task{}, ErrNotFound
	}
	t.Title = title
	t.Status = status
	t.UpdatedAt = s.now().UTC()
	s.tasks[id] = t
	return t, nil
}

// Delete smaže úkol podle ID, nebo vrátí ErrNotFound.
func (s *Store) Delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.tasks[id]; !ok {
		return ErrNotFound
	}
	delete(s.tasks, id)
	for i, existing := range s.order {
		if existing == id {
			s.order = append(s.order[:i], s.order[i+1:]...)
			break
		}
	}
	return nil
}

package postgres

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"sort"
	"strings"
)

//go:embed migrations/*.sql
var migrationFS embed.FS

// ApplyMigrations načte a spustí SQL soubory z embedded migrací.
// Pro jednoduchost bonus projektu jsou migrace idempotentní (IF NOT EXISTS).
func (s *Store) ApplyMigrations(ctx context.Context) error {
	entries, err := fs.ReadDir(migrationFS, "migrations")
	if err != nil {
		return fmt.Errorf("postgres: list migrations: %w", err)
	}
	names := make([]string, 0, len(entries))
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".sql") {
			names = append(names, e.Name())
		}
	}
	sort.Strings(names)
	for _, name := range names {
		body, err := migrationFS.ReadFile("migrations/" + name)
		if err != nil {
			return fmt.Errorf("postgres: read %s: %w", name, err)
		}
		if err := s.Migrate(ctx, string(body)); err != nil {
			return fmt.Errorf("postgres: apply %s: %w", name, err)
		}
	}
	return nil
}

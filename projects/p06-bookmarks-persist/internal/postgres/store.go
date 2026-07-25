// Package postgres je adaptér BookmarkStore nad PostgreSQL (pgx + sqlc).
package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/rdurica/go-deep/projects/p06-bookmarks-persist/internal/app"
	"github.com/rdurica/go-deep/projects/p06-bookmarks-persist/internal/bookmark"
	"github.com/rdurica/go-deep/projects/p06-bookmarks-persist/internal/postgres/dbsqlc"
)

// Store je PostgreSQL implementace app.BookmarkStore.
type Store struct {
	pool *pgxpool.Pool
	q    *dbsqlc.Queries
}

// Open připojí pool k DATABASE_URL.
func Open(ctx context.Context, databaseURL string) (*Store, error) {
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return nil, fmt.Errorf("postgres: pool: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("postgres: ping: %w", err)
	}
	return &Store{pool: pool, q: dbsqlc.New(pool)}, nil
}

// Close uzavře pool.
func (s *Store) Close() {
	if s != nil && s.pool != nil {
		s.pool.Close()
	}
}

// Migrate aplikuje SQL migrace z embedded / předaných souborů.
func (s *Store) Migrate(ctx context.Context, sql string) error {
	_, err := s.pool.Exec(ctx, sql)
	if err != nil {
		return fmt.Errorf("postgres: migrate: %w", err)
	}
	return nil
}

// Add vloží záložku.
func (s *Store) Add(ctx context.Context, b bookmark.Bookmark) error {
	if err := b.Validate(); err != nil {
		return err
	}
	err := s.q.InsertBookmark(ctx, dbsqlc.InsertBookmarkParams{
		ID:        b.ID,
		Url:       b.URL,
		Title:     b.Title,
		Tags:      b.Tags,
		CreatedAt: pgtype.Timestamptz{Time: b.CreatedAt.UTC(), Valid: true},
	})
	if err != nil {
		return mapWriteErr(err)
	}
	return nil
}

// Get načte záložku podle ID.
func (s *Store) Get(ctx context.Context, id string) (bookmark.Bookmark, error) {
	row, err := s.q.GetBookmark(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return bookmark.Bookmark{}, bookmark.ErrNotFound
		}
		return bookmark.Bookmark{}, fmt.Errorf("postgres: get: %w", err)
	}
	return fromRow(row), nil
}

// Delete smaže záložku podle ID.
func (s *Store) Delete(ctx context.Context, id string) error {
	n, err := s.q.DeleteBookmark(ctx, id)
	if err != nil {
		return fmt.Errorf("postgres: delete: %w", err)
	}
	if n == 0 {
		return bookmark.ErrNotFound
	}
	return nil
}

// Search vyhledá záložky a stránkuje v Go (cursor podle ID).
func (s *Store) Search(ctx context.Context, q app.Query) (app.Page, error) {
	if q.Limit < 0 || q.Limit > bookmark.MaxLimit {
		return app.Page{}, bookmark.ErrInvalidQuery
	}
	limit := q.Limit
	if limit == 0 {
		limit = bookmark.DefaultLimit
	}

	text := strings.TrimSpace(q.Text)
	tag := strings.ToLower(strings.TrimSpace(q.Tag))
	if tag != "" {
		norm, err := bookmark.NormalizeTags([]string{tag})
		if err != nil || len(norm) != 1 {
			return app.Page{}, bookmark.ErrInvalidQuery
		}
		tag = norm[0]
	}

	rows, err := s.q.SearchBookmarks(ctx, dbsqlc.SearchBookmarksParams{
		QueryText: text,
		Tag:       tag,
	})
	if err != nil {
		return app.Page{}, fmt.Errorf("postgres: search: %w", err)
	}

	matched := make([]bookmark.Bookmark, 0, len(rows))
	for _, row := range rows {
		matched = append(matched, fromRow(row))
	}

	page := app.Page{Total: len(matched)}
	start := 0
	if q.Cursor != "" {
		idx := -1
		for i, b := range matched {
			if b.ID == q.Cursor {
				idx = i
				break
			}
		}
		if idx < 0 {
			return app.Page{}, bookmark.ErrInvalidCursor
		}
		start = idx + 1
	}

	end := start + limit
	if end > len(matched) {
		end = len(matched)
	}
	page.Items = matched[start:end]
	if end < len(matched) && end > start {
		page.NextCursor = matched[end-1].ID
	}
	return page, nil
}

// Ready pingne databázi.
func (s *Store) Ready(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	if err := s.pool.Ping(ctx); err != nil {
		return fmt.Errorf("postgres: not ready: %w", err)
	}
	return nil
}

func fromRow(row dbsqlc.Bookmark) bookmark.Bookmark {
	tags := row.Tags
	if tags == nil {
		tags = []string{}
	}
	created := time.Time{}
	if row.CreatedAt.Valid {
		created = row.CreatedAt.Time.UTC()
	}
	return bookmark.Bookmark{
		ID:        row.ID,
		URL:       row.Url,
		Title:     row.Title,
		Tags:      tags,
		CreatedAt: created,
	}
}

func mapWriteErr(err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		switch pgErr.ConstraintName {
		case "bookmarks_pkey":
			return bookmark.ErrDuplicateID
		case "bookmarks_url_key":
			return bookmark.ErrDuplicateURL
		default:
			if strings.Contains(pgErr.ConstraintName, "url") {
				return bookmark.ErrDuplicateURL
			}
			return bookmark.ErrDuplicateID
		}
	}
	return fmt.Errorf("postgres: insert: %w", err)
}

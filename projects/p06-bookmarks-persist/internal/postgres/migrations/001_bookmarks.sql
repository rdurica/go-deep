CREATE TABLE IF NOT EXISTS bookmarks (
    id         TEXT PRIMARY KEY,
    url        TEXT NOT NULL UNIQUE,
    title      TEXT NOT NULL,
    tags       TEXT[] NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX IF NOT EXISTS bookmarks_tags_gin ON bookmarks USING GIN (tags);
CREATE INDEX IF NOT EXISTS bookmarks_created_at_idx ON bookmarks (created_at DESC, id ASC);

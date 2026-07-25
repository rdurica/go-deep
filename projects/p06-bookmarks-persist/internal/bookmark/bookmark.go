// Package bookmark je doména záložek: model, validace a normalizace.
//
// Záměrně neimportuje net/http ani encoding/json — formát na drátě patří
// do HTTP adaptéru, ne do domény.
package bookmark

import (
	"errors"
	"net/url"
	"sort"
	"strings"
	"time"
	"unicode/utf8"
)

// Limity domény.
const (
	MaxTitleLen  = 200
	MaxTags      = 10
	DefaultLimit = 20
	MaxLimit     = 100
)

// Doménové chyby.
var (
	ErrEmptyID       = errors.New("bookmark: prázdné ID")
	ErrInvalidURL    = errors.New("bookmark: neplatná URL")
	ErrEmptyTitle    = errors.New("bookmark: prázdný titulek")
	ErrTitleTooLong  = errors.New("bookmark: příliš dlouhý titulek")
	ErrInvalidTag    = errors.New("bookmark: neplatný tag")
	ErrTooManyTags   = errors.New("bookmark: příliš mnoho tagů")
	ErrDuplicateTag  = errors.New("bookmark: duplicitní tag")
	ErrDuplicateID   = errors.New("bookmark: ID už existuje")
	ErrDuplicateURL  = errors.New("bookmark: URL už existuje")
	ErrNotFound      = errors.New("bookmark: záložka nenalezena")
	ErrInvalidQuery  = errors.New("bookmark: neplatný dotaz")
	ErrInvalidCursor = errors.New("bookmark: neplatný cursor")
)

// Bookmark je uložená záložka.
type Bookmark struct {
	ID        string
	URL       string
	Title     string
	Tags      []string
	CreatedAt time.Time
}

// NormalizeURL převede URL na kanonický tvar: malé schéma a host, bez výchozího portu,
// bez utm_ parametrů, bez fragmentu, s seřazeným query a bez koncového lomítka.
func NormalizeURL(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", ErrInvalidURL
	}

	u, err := url.Parse(raw)
	if err != nil {
		return "", ErrInvalidURL
	}

	scheme := strings.ToLower(u.Scheme)
	if scheme != "http" && scheme != "https" {
		return "", ErrInvalidURL
	}
	u.Scheme = scheme

	host := strings.ToLower(u.Hostname())
	if host == "" {
		return "", ErrInvalidURL
	}
	port := u.Port()
	if (scheme == "http" && port == "80") || (scheme == "https" && port == "443") {
		port = ""
	}
	u.Host = host
	if port != "" {
		u.Host = host + ":" + port
	}

	u.Fragment = ""
	u.RawFragment = ""

	if u.RawQuery != "" {
		values := u.Query()
		for key := range values {
			if strings.HasPrefix(strings.ToLower(key), "utm_") {
				values.Del(key)
			}
		}
		u.RawQuery = values.Encode()
	}

	if u.Path != "" {
		u.Path = strings.TrimRight(u.Path, "/")
		u.RawPath = ""
	}
	return u.String(), nil
}

func validTag(tag string) bool {
	if tag == "" {
		return false
	}
	for _, r := range tag {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9', r == '-':
		default:
			return false
		}
	}
	return true
}

// NormalizeTags ořeže, převede na malá písmena, odstraní duplicity a seřadí tagy.
func NormalizeTags(tags []string) ([]string, error) {
	seen := make(map[string]struct{}, len(tags))
	out := make([]string, 0, len(tags))
	for _, tag := range tags {
		tag = strings.ToLower(strings.TrimSpace(tag))
		if tag == "" {
			continue
		}
		if !validTag(tag) {
			return nil, ErrInvalidTag
		}
		if _, dup := seen[tag]; dup {
			continue
		}
		seen[tag] = struct{}{}
		out = append(out, tag)
	}
	if len(out) > MaxTags {
		return nil, ErrTooManyTags
	}
	sort.Strings(out)
	return out, nil
}

// Validate zkontroluje, že je záložka kompletní a normalizovaná.
func (b Bookmark) Validate() error {
	if strings.TrimSpace(b.ID) == "" {
		return ErrEmptyID
	}
	norm, err := NormalizeURL(b.URL)
	if err != nil {
		return err
	}
	if norm != b.URL {
		return ErrInvalidURL
	}
	if strings.TrimSpace(b.Title) == "" {
		return ErrEmptyTitle
	}
	if utf8.RuneCountInString(b.Title) > MaxTitleLen {
		return ErrTitleTooLong
	}
	if len(b.Tags) > MaxTags {
		return ErrTooManyTags
	}
	seen := make(map[string]struct{}, len(b.Tags))
	for _, tag := range b.Tags {
		if !validTag(tag) {
			return ErrInvalidTag
		}
		if _, dup := seen[tag]; dup {
			return ErrDuplicateTag
		}
		seen[tag] = struct{}{}
	}
	return nil
}

// New sestaví normalizovanou a ověřenou záložku.
// Prázdný titulek nahradí hostem z URL.
func New(id, rawURL, title string, tags []string, createdAt time.Time) (Bookmark, error) {
	normURL, err := NormalizeURL(rawURL)
	if err != nil {
		return Bookmark{}, err
	}
	normTags, err := NormalizeTags(tags)
	if err != nil {
		return Bookmark{}, err
	}

	title = strings.TrimSpace(title)
	if title == "" {
		u, err := url.Parse(normURL)
		if err != nil || u.Hostname() == "" {
			return Bookmark{}, ErrEmptyTitle
		}
		title = u.Hostname()
	}

	b := Bookmark{
		ID:        strings.TrimSpace(id),
		URL:       normURL,
		Title:     title,
		Tags:      normTags,
		CreatedAt: createdAt,
	}
	if err := b.Validate(); err != nil {
		return Bookmark{}, err
	}
	return b, nil
}

// Clone vrátí hlubokou kopii záložky.
func Clone(b Bookmark) Bookmark {
	out := b
	if b.Tags != nil {
		out.Tags = make([]string, len(b.Tags))
		copy(out.Tags, b.Tags)
	}
	return out
}

package pagination

import (
	"encoding/base64"
	"net/http"
	"strconv"

	"github.com/google/uuid"
)

const (
	DefaultLimit = 20
	MaxLimit     = 100
)

// Generic response envelope for paginated lists
type Page[T any] struct {
	Items      []T    `json:"items"`
	NextCursor string `json:"next_cursor,omitempty"`
	HasMore    bool   `json:"has_more"`
}

// Extracts cursor and limit from query parameters
func ParseCursor(r *http.Request) (cursor string, limit int) {
	cursor = r.URL.Query().Get("cursor")
	limit = DefaultLimit
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}
	if limit > MaxLimit {
		limit = MaxLimit
	}
	return cursor, limit
}

// Turns a UUID into a base64 string for use as a cursor
func EncodeCursor(id uuid.UUID) string {
	return base64.URLEncoding.EncodeToString(id[:])
}

// Turns a cursor string back into a UUID
func DecodeCursor(cursor string) (uuid.UUID, bool) {
	if cursor == "" {
		return uuid.Nil, false
	}
	b, err := base64.URLEncoding.DecodeString(cursor)
	if err != nil || len(b) != 16 {
		return uuid.Nil, false
	}
	id, err := uuid.FromBytes(b)
	if err != nil {
		return uuid.Nil, false
	}
	return id, true
}

// Constructs a page from a slice of items
func NewPage[T any](items []T, limit int, cursorFn func(T) uuid.UUID) Page[T] {
	hasMore := len(items) > limit
	if hasMore {
		items = items[:limit] // trim the extra probe item
	}
	var nextCursor string
	if hasMore && len(items) > 0 {
		lastItem := items[len(items)-1]
		nextCursor = EncodeCursor(cursorFn(lastItem))
	}
	return Page[T]{
		Items:      items,
		NextCursor: nextCursor,
		HasMore:    hasMore,
	}
}

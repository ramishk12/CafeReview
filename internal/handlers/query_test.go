package handlers

import (
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestPageParams(t *testing.T) {
	tests := []struct {
		name       string
		target     string
		wantLimit  int
		wantOffset int
		wantPaged  bool
		wantOK     bool
	}{
		{"no paging", "/x", 0, 0, false, true},
		{"limit only", "/x?limit=10", 10, 0, true, true},
		{"limit and offset", "/x?limit=5&offset=20", 5, 20, true, true},
		{"offset without limit", "/x?offset=3", 0, 3, false, true},
		{"max limit", "/x?limit=100", 100, 0, true, true},
		{"limit zero", "/x?limit=0", 0, 0, false, false},
		{"limit above max", "/x?limit=101", 0, 0, false, false},
		{"limit not a number", "/x?limit=ten", 0, 0, false, false},
		{"negative offset", "/x?offset=-1", 0, 0, false, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, w := newTestContext(t, http.MethodGet, tt.target, nil, nil, 0, "")
			limit, offset, paged, ok := pageParams(c)

			assert.Equal(t, tt.wantOK, ok)
			if tt.wantOK {
				assert.Equal(t, tt.wantLimit, limit)
				assert.Equal(t, tt.wantOffset, offset)
				assert.Equal(t, tt.wantPaged, paged)
			} else {
				assert.Equal(t, http.StatusBadRequest, w.Code)
			}
		})
	}
}

func TestAppendPaging(t *testing.T) {
	tests := []struct {
		name      string
		limit     int
		offset    int
		paged     bool
		wantQuery string
		wantArgs  []any
	}{
		{"no paging", 0, 0, false, "SELECT 1", []any{"a"}},
		{"limit only", 10, 0, true, "SELECT 1 LIMIT $2", []any{"a", 10}},
		{"offset only", 0, 5, false, "SELECT 1 OFFSET $2", []any{"a", 5}},
		{"limit and offset", 10, 5, true, "SELECT 1 LIMIT $2 OFFSET $3", []any{"a", 10, 5}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query, args := appendPaging("SELECT 1", []any{"a"}, tt.limit, tt.offset, tt.paged)
			assert.Equal(t, tt.wantQuery, query)
			assert.Equal(t, tt.wantArgs, args)
		})
	}
}

func TestOrderByParam(t *testing.T) {
	clauses := map[string]string{"name": "c.name ASC", "newest": "c.created_at DESC"}

	t.Run("Returns the clause for a known sort", func(t *testing.T) {
		c, w := newTestContext(t, http.MethodGet, "/x?sort=newest", nil, nil, 0, "")
		clause, ok := orderByParam(c, clauses, "name")
		assert.True(t, ok)
		assert.Equal(t, "c.created_at DESC", clause)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Uses the default when sort is absent", func(t *testing.T) {
		c, _ := newTestContext(t, http.MethodGet, "/x", nil, nil, 0, "")
		clause, ok := orderByParam(c, clauses, "name")
		assert.True(t, ok)
		assert.Equal(t, "c.name ASC", clause)
	})

	t.Run("Rejects an unknown sort with the allowed values listed", func(t *testing.T) {
		c, w := newTestContext(t, http.MethodGet, "/x?sort=name%20DESC%20--x", nil, nil, 0, "")
		_, ok := orderByParam(c, clauses, "name")
		assert.False(t, ok)
		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "sort must be one of: name, newest")
	})
}

func TestParseID(t *testing.T) {
	tests := []struct {
		value  string
		wantID int
		wantOK bool
	}{
		{"1", 1, true},
		{"42", 42, true},
		{"0", 0, false},
		{"-3", 0, false},
		{"abc", 0, false},
		{"1.5", 0, false},
	}
	for _, tt := range tests {
		t.Run(tt.value, func(t *testing.T) {
			c, _ := newTestContext(t, http.MethodGet, "/x", nil, gin.Params{{Key: "id", Value: tt.value}}, 0, "")
			id, ok := parseID(c, "id")
			assert.Equal(t, tt.wantOK, ok)
			assert.Equal(t, tt.wantID, id)
		})
	}
}

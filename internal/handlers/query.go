package handlers

import (
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"cafe-review/internal/database"

	"github.com/gin-gonic/gin"
)

const (
	maxPageSize     = 100
	maxSearchLength = 100

	// totalCountHeader carries the number of matching rows, so paged clients know how many pages exist.
	totalCountHeader = "X-Total-Count"
)

// likeEscaper escapes LIKE wildcards so user input is matched literally.
var likeEscaper = strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`)

// pageParams reads the optional limit and offset query parameters. When limit
// is absent, paged is false and the full list is returned. It writes a 400 and
// returns ok=false for invalid values.
func pageParams(c *gin.Context) (limit, offset int, paged, ok bool) {
	if raw := c.Query("limit"); raw != "" {
		n, err := strconv.Atoi(raw)
		if err != nil || n < 1 || n > maxPageSize {
			c.JSON(http.StatusBadRequest, gin.H{"error": "limit must be between 1 and 100"})
			return 0, 0, false, false
		}
		limit, paged = n, true
	}
	if raw := c.Query("offset"); raw != "" {
		n, err := strconv.Atoi(raw)
		if err != nil || n < 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "offset must be 0 or greater"})
			return 0, 0, false, false
		}
		offset = n
	}
	return limit, offset, paged, true
}

// appendPaging adds LIMIT and OFFSET to query when requested and appends their values to args.
func appendPaging(query string, args []any, limit, offset int, paged bool) (string, []any) {
	if paged {
		args = append(args, limit)
		query += fmt.Sprintf(" LIMIT $%d", len(args))
	}
	if offset > 0 {
		args = append(args, offset)
		query += fmt.Sprintf(" OFFSET $%d", len(args))
	}
	return query, args
}

// orderByParam reads the sort query parameter and returns its SQL ORDER BY
// clause. Only keys in clauses are accepted, so the SQL is never built from raw input.
func orderByParam(c *gin.Context, clauses map[string]string, defaultSort string) (string, bool) {
	clause, ok := clauses[c.DefaultQuery("sort", defaultSort)]
	if ok {
		return clause, true
	}
	keys := make([]string, 0, len(clauses))
	for k := range clauses {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	c.JSON(http.StatusBadRequest, gin.H{"error": "sort must be one of: " + strings.Join(keys, ", ")})
	return "", false
}

// setTotalCount writes the total number of matching rows to the response header.
func setTotalCount(c *gin.Context, total int) {
	c.Header(totalCountHeader, strconv.Itoa(total))
}

// countRows runs a COUNT(*) query and returns the result.
func countRows(query string, args ...any) (int, error) {
	var n int
	err := database.DB.QueryRow(query, args...).Scan(&n)
	return n, err
}

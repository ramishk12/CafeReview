package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"cafe-review/internal/apierror"
	"cafe-review/internal/database"
	"cafe-review/internal/models"

	"github.com/gin-gonic/gin"
)

var validTagScopes = map[string]bool{"cafe": true, "review": true, "image": true}

// ListTags returns approved tags, optionally filtered by scope.
func ListTags(c *gin.Context) {
	scope := strings.TrimSpace(c.Query("scope"))
	if scope != "" && !validTagScopes[scope] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "scope must be cafe, review, or image"})
		return
	}

	query := `SELECT id, name, category, scope FROM tags WHERE status = 'approved'`
	var args []any
	if scope != "" {
		args = append(args, scope)
		query += fmt.Sprintf(" AND scope = $%d", len(args))
	}
	query += " ORDER BY category, name"

	rows, err := database.DB.Query(query, args...)
	if err != nil {
		apierror.Internal(c, err)
		return
	}
	defer rows.Close()

	tags := []models.Tag{}
	for rows.Next() {
		var t models.Tag
		if err := rows.Scan(&t.ID, &t.Name, &t.Category, &t.Scope); err != nil {
			apierror.Internal(c, err)
			return
		}
		tags = append(tags, t)
	}
	if err := rows.Err(); err != nil {
		apierror.Internal(c, err)
		return
	}
	c.JSON(http.StatusOK, tags)
}

// SetCafeTags replaces a cafe's tags. Admin-only.
func SetCafeTags(c *gin.Context) {
	cafeID, ok := parseID(c, "id")
	if !ok {
		return
	}

	var req struct {
		TagIDs []int `json:"tag_ids"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	var exists bool
	if err := database.DB.QueryRow(`SELECT EXISTS(SELECT 1 FROM cafes WHERE id = $1)`, cafeID).Scan(&exists); err != nil {
		apierror.Internal(c, err)
		return
	}
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Cafe not found"})
		return
	}

	ids := uniqueInts(req.TagIDs)
	if len(ids) > 0 {
		placeholders := make([]string, len(ids))
		args := make([]any, len(ids))
		for i, id := range ids {
			args[i] = id
			placeholders[i] = fmt.Sprintf("$%d", i+1)
		}
		var count int
		query := fmt.Sprintf(
			`SELECT COUNT(*) FROM tags WHERE scope = 'cafe' AND status = 'approved' AND id IN (%s)`,
			strings.Join(placeholders, ", "))
		if err := database.DB.QueryRow(query, args...).Scan(&count); err != nil {
			apierror.Internal(c, err)
			return
		}
		if count != len(ids) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "One or more tags are invalid"})
			return
		}
	}

	tx, err := database.DB.Begin()
	if err != nil {
		apierror.Internal(c, err)
		return
	}
	defer tx.Rollback()

	if _, err := tx.Exec(`DELETE FROM cafe_tags WHERE cafe_id = $1`, cafeID); err != nil {
		apierror.Internal(c, err)
		return
	}
	for _, tagID := range ids {
		if _, err := tx.Exec(`INSERT INTO cafe_tags (cafe_id, tag_id) VALUES ($1, $2)`, cafeID, tagID); err != nil {
			apierror.Internal(c, err)
			return
		}
	}
	if err := tx.Commit(); err != nil {
		apierror.Internal(c, err)
		return
	}

	cafe, err := fetchCafe(cafeID)
	if err != nil {
		apierror.Internal(c, err)
		return
	}
	tagsByCafe, err := loadTagsForCafes([]int{cafeID})
	if err != nil {
		apierror.Internal(c, err)
		return
	}
	cafe.Tags = tagsByCafe[cafeID]
	if cafe.Tags == nil {
		cafe.Tags = []models.Tag{}
	}
	c.JSON(http.StatusOK, cafe)
}

// loadTagsForCafes returns the tags for each cafe, grouped by cafe ID.
func loadTagsForCafes(cafeIDs []int) (map[int][]models.Tag, error) {
	result := make(map[int][]models.Tag)
	if len(cafeIDs) == 0 {
		return result, nil
	}

	placeholders := make([]string, len(cafeIDs))
	args := make([]any, len(cafeIDs))
	for i, id := range cafeIDs {
		args[i] = id
		placeholders[i] = fmt.Sprintf("$%d", i+1)
	}
	query := fmt.Sprintf(`
		SELECT ct.cafe_id, t.id, t.name, t.category, t.scope
		FROM cafe_tags ct
		JOIN tags t ON t.id = ct.tag_id
		WHERE ct.cafe_id IN (%s)
		ORDER BY t.category, t.name`, strings.Join(placeholders, ", "))

	rows, err := database.DB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var cafeID int
		var t models.Tag
		if err := rows.Scan(&cafeID, &t.ID, &t.Name, &t.Category, &t.Scope); err != nil {
			return nil, err
		}
		result[cafeID] = append(result[cafeID], t)
	}
	return result, rows.Err()
}

// loadTagsForImages returns the tags for each review image, grouped by image ID.
func loadTagsForImages(imageIDs []int) (map[int][]models.Tag, error) {
	result := make(map[int][]models.Tag)
	if len(imageIDs) == 0 {
		return result, nil
	}

	placeholders := make([]string, len(imageIDs))
	args := make([]any, len(imageIDs))
	for i, id := range imageIDs {
		args[i] = id
		placeholders[i] = fmt.Sprintf("$%d", i+1)
	}
	query := fmt.Sprintf(`
		SELECT rit.image_id, t.id, t.name, t.category, t.scope
		FROM review_image_tags rit
		JOIN tags t ON t.id = rit.tag_id
		WHERE rit.image_id IN (%s)
		ORDER BY t.category, t.name`, strings.Join(placeholders, ", "))

	rows, err := database.DB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var imageID int
		var t models.Tag
		if err := rows.Scan(&imageID, &t.ID, &t.Name, &t.Category, &t.Scope); err != nil {
			return nil, err
		}
		result[imageID] = append(result[imageID], t)
	}
	return result, rows.Err()
}

// parseTagIDs reads the comma-separated tags query parameter. It writes a 400
// and returns ok=false when a value is not a positive integer.
func parseTagIDs(c *gin.Context) ([]int, bool) {
	raw := strings.TrimSpace(c.Query("tags"))
	if raw == "" {
		return nil, true
	}
	var ids []int
	seen := map[int]bool{}
	for _, part := range strings.Split(raw, ",") {
		id, err := strconv.Atoi(strings.TrimSpace(part))
		if err != nil || id <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "tags must be a comma-separated list of tag IDs"})
			return nil, false
		}
		if !seen[id] {
			seen[id] = true
			ids = append(ids, id)
		}
	}
	return ids, true
}

func uniqueInts(values []int) []int {
	seen := map[int]bool{}
	out := make([]int, 0, len(values))
	for _, v := range values {
		if v > 0 && !seen[v] {
			seen[v] = true
			out = append(out, v)
		}
	}
	return out
}

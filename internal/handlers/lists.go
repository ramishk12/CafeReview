package handlers

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"cafe-review/internal/apierror"
	"cafe-review/internal/auth"
	"cafe-review/internal/database"
	"cafe-review/internal/models"

	"github.com/gin-gonic/gin"
)

const (
	maxListsPerUser = 50
	maxCafesPerList = 500
	maxListNameLen  = 100
)

// requireListOwner loads a list's owner and writes 404 or 403 if the current
// user may not access it. It returns false when a response was written.
func requireListOwner(c *gin.Context, listID int) bool {
	var ownerID int
	err := database.DB.QueryRow(`SELECT user_id FROM lists WHERE id = $1`, listID).Scan(&ownerID)
	if errors.Is(err, sql.ErrNoRows) {
		c.JSON(http.StatusNotFound, gin.H{"error": "List not found"})
		return false
	}
	if err != nil {
		apierror.Internal(c, err)
		return false
	}
	if ownerID != auth.CurrentUserID(c) {
		c.JSON(http.StatusForbidden, gin.H{"error": "This list belongs to another user"})
		return false
	}
	return true
}

// fetchList returns a list with its cafe count.
func fetchList(listID int) (models.List, error) {
	var l models.List
	err := database.DB.QueryRow(`
		SELECT l.id, l.name, l.is_public, l.created_at, COUNT(lc.cafe_id)
		FROM lists l
		LEFT JOIN list_cafes lc ON lc.list_id = l.id
		WHERE l.id = $1
		GROUP BY l.id`, listID,
	).Scan(&l.ID, &l.Name, &l.IsPublic, &l.CreatedAt, &l.CafeCount)
	return l, err
}

func validateListName(name string) (string, string) {
	name = strings.TrimSpace(name)
	if name == "" || len([]rune(name)) > maxListNameLen {
		return "", "Name must be 1-100 characters"
	}
	return name, ""
}

// ListMyLists returns the signed-in user's lists. With ?cafe=<id> each list
// reports whether it contains that cafe.
func ListMyLists(c *gin.Context) {
	userID := auth.CurrentUserID(c)

	cafeID := 0
	if raw := strings.TrimSpace(c.Query("cafe")); raw != "" {
		id, err := strconv.Atoi(raw)
		if err != nil || id <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "cafe must be a valid ID"})
			return
		}
		cafeID = id
	}

	rows, err := database.DB.Query(`
		SELECT l.id, l.name, l.is_public, l.created_at, COUNT(lc.cafe_id)
		FROM lists l
		LEFT JOIN list_cafes lc ON lc.list_id = l.id
		WHERE l.user_id = $1
		GROUP BY l.id
		ORDER BY l.created_at DESC`, userID)
	if err != nil {
		apierror.Internal(c, err)
		return
	}
	defer rows.Close()

	lists := []models.List{}
	for rows.Next() {
		var l models.List
		if err := rows.Scan(&l.ID, &l.Name, &l.IsPublic, &l.CreatedAt, &l.CafeCount); err != nil {
			apierror.Internal(c, err)
			return
		}
		lists = append(lists, l)
	}
	if err := rows.Err(); err != nil {
		apierror.Internal(c, err)
		return
	}

	if cafeID > 0 {
		contained := map[int]bool{}
		membership, err := database.DB.Query(`SELECT list_id FROM list_cafes WHERE cafe_id = $1`, cafeID)
		if err != nil {
			apierror.Internal(c, err)
			return
		}
		defer membership.Close()
		for membership.Next() {
			var listID int
			if err := membership.Scan(&listID); err != nil {
				apierror.Internal(c, err)
				return
			}
			contained[listID] = true
		}
		if err := membership.Err(); err != nil {
			apierror.Internal(c, err)
			return
		}
		for i := range lists {
			lists[i].ContainsCafe = contained[lists[i].ID]
		}
	}

	c.JSON(http.StatusOK, lists)
}

// CreateList creates a list for the signed-in user.
func CreateList(c *gin.Context) {
	var req struct {
		Name string `json:"name"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}
	name, msg := validateListName(req.Name)
	if msg != "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": msg})
		return
	}

	userID := auth.CurrentUserID(c)
	var count int
	if err := database.DB.QueryRow(`SELECT COUNT(*) FROM lists WHERE user_id = $1`, userID).Scan(&count); err != nil {
		apierror.Internal(c, err)
		return
	}
	if count >= maxListsPerUser {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("You can have at most %d lists", maxListsPerUser)})
		return
	}

	var l models.List
	err := database.DB.QueryRow(
		`INSERT INTO lists (user_id, name) VALUES ($1, $2) RETURNING id, name, is_public, created_at`,
		userID, name,
	).Scan(&l.ID, &l.Name, &l.IsPublic, &l.CreatedAt)
	if err != nil {
		apierror.Internal(c, err)
		return
	}
	c.JSON(http.StatusCreated, l)
}

// GetList returns a list and its cafes. Owner only for now (public lists later).
func GetList(c *gin.Context) {
	listID, ok := parseID(c, "id")
	if !ok {
		return
	}
	if !requireListOwner(c, listID) {
		return
	}

	list, err := fetchList(listID)
	if err != nil {
		apierror.Internal(c, err)
		return
	}

	cafes, err := cafesInList(listID)
	if err != nil {
		apierror.Internal(c, err)
		return
	}
	c.JSON(http.StatusOK, models.ListDetail{List: list, Cafes: cafes})
}

// UpdateList renames a list.
func UpdateList(c *gin.Context) {
	listID, ok := parseID(c, "id")
	if !ok {
		return
	}
	if !requireListOwner(c, listID) {
		return
	}

	var req struct {
		Name string `json:"name"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}
	name, msg := validateListName(req.Name)
	if msg != "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": msg})
		return
	}

	if _, err := database.DB.Exec(`UPDATE lists SET name = $1 WHERE id = $2`, name, listID); err != nil {
		apierror.Internal(c, err)
		return
	}
	list, err := fetchList(listID)
	if err != nil {
		apierror.Internal(c, err)
		return
	}
	c.JSON(http.StatusOK, list)
}

// DeleteList removes a list and its items (cascade).
func DeleteList(c *gin.Context) {
	listID, ok := parseID(c, "id")
	if !ok {
		return
	}
	if !requireListOwner(c, listID) {
		return
	}
	if _, err := database.DB.Exec(`DELETE FROM lists WHERE id = $1`, listID); err != nil {
		apierror.Internal(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// AddCafeToList adds a cafe to a list.
func AddCafeToList(c *gin.Context) {
	listID, ok := parseID(c, "id")
	if !ok {
		return
	}
	if !requireListOwner(c, listID) {
		return
	}

	var req struct {
		CafeID int `json:"cafe_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.CafeID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "A valid cafe_id is required"})
		return
	}

	var count int
	if err := database.DB.QueryRow(`SELECT COUNT(*) FROM list_cafes WHERE list_id = $1`, listID).Scan(&count); err != nil {
		apierror.Internal(c, err)
		return
	}
	if count >= maxCafesPerList {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("A list can hold at most %d cafes", maxCafesPerList)})
		return
	}

	_, err := database.DB.Exec(`INSERT INTO list_cafes (list_id, cafe_id) VALUES ($1, $2)`, listID, req.CafeID)
	if database.IsForeignKeyViolation(err) {
		c.JSON(http.StatusNotFound, gin.H{"error": "Cafe not found"})
		return
	}
	if database.IsUniqueViolation(err) {
		c.JSON(http.StatusConflict, gin.H{"error": "This cafe is already in the list"})
		return
	}
	if err != nil {
		apierror.Internal(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// RemoveCafeFromList removes a cafe from a list.
func RemoveCafeFromList(c *gin.Context) {
	listID, ok := parseID(c, "id")
	if !ok {
		return
	}
	cafeID, ok := parseID(c, "cafeId")
	if !ok {
		return
	}
	if !requireListOwner(c, listID) {
		return
	}

	res, err := database.DB.Exec(`DELETE FROM list_cafes WHERE list_id = $1 AND cafe_id = $2`, listID, cafeID)
	if err != nil {
		apierror.Internal(c, err)
		return
	}
	affected, err := res.RowsAffected()
	if err != nil {
		apierror.Internal(c, err)
		return
	}
	if affected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "This cafe is not in the list"})
		return
	}
	c.Status(http.StatusNoContent)
}

// cafesInList returns the cafes in a list, with the same fields as other cafe
// responses.
func cafesInList(listID int) ([]models.Cafe, error) {
	rows, err := database.DB.Query(`
		SELECT c.id, c.name, c.address, c.description, c.created_at, c.closed_at,
		       c.background_image_id, c.background_image_path, bri.filename,
		       COALESCE(AVG(r.rating), 0), COUNT(r.id), COUNT(DISTINCT r.user_id)
		FROM list_cafes lc
		JOIN cafes c ON c.id = lc.cafe_id
		LEFT JOIN review_images bri ON bri.id = c.background_image_id
		LEFT JOIN reviews r ON r.cafe_id = c.id
		WHERE lc.list_id = $1
		GROUP BY c.id
		ORDER BY lc.added_at DESC`, listID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	cafes := []models.Cafe{}
	ids := []int{}
	for rows.Next() {
		cafe, err := scanCafe(rows)
		if err != nil {
			return nil, err
		}
		cafes = append(cafes, cafe)
		ids = append(ids, cafe.ID)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	tagsByCafe, err := loadTagsForCafes(ids)
	if err != nil {
		return nil, err
	}
	for i := range cafes {
		cafes[i].Tags = tagsByCafe[cafes[i].ID]
		if cafes[i].Tags == nil {
			cafes[i].Tags = []models.Tag{}
		}
	}
	return cafes, nil
}

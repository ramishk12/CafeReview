package handlers

import (
	"database/sql"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"unicode/utf8"

	"cafe-review/internal/database"
	"cafe-review/internal/models"

	"github.com/gin-gonic/gin"
)

const cafeSelect = `
	SELECT c.id, c.name, c.address, c.description, c.created_at,
	       COALESCE(AVG(r.rating), 0), COUNT(r.id)
	FROM cafes c
	LEFT JOIN reviews r ON r.cafe_id = c.id`

// cafeSortClauses maps the accepted ?sort= values to ORDER BY clauses.
var cafeSortClauses = map[string]string{
	"name":    "c.name ASC, c.id ASC",
	"rating":  "COALESCE(AVG(r.rating), 0) DESC, COUNT(r.id) DESC, c.name ASC",
	"reviews": "COUNT(r.id) DESC, COALESCE(AVG(r.rating), 0) DESC, c.name ASC",
	"newest":  "c.created_at DESC, c.id DESC",
}

// parseID reads a positive integer route parameter. It writes a 400 response
// and returns false when the value is invalid.
func parseID(c *gin.Context, name string) (int, bool) {
	id, err := strconv.Atoi(c.Param(name))
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return 0, false
	}
	return id, true
}

func scanCafe(row interface{ Scan(...any) error }) (models.Cafe, error) {
	var cafe models.Cafe
	err := row.Scan(
		&cafe.ID, &cafe.Name, &cafe.Address, &cafe.Description, &cafe.CreatedAt,
		&cafe.AvgRating, &cafe.ReviewCount,
	)
	return cafe, err
}

// fetchCafe returns one cafe with its rating summary.
func fetchCafe(id int) (models.Cafe, error) {
	return scanCafe(database.DB.QueryRow(cafeSelect+`
		WHERE c.id = $1
		GROUP BY c.id`, id))
}

// ListCafes supports ?q= (name or address search), ?sort=, and optional ?limit= and ?offset=.
// Without limit, every matching cafe is returned.
func ListCafes(c *gin.Context) {
	orderBy, ok := orderByParam(c, cafeSortClauses, "name")
	if !ok {
		return
	}
	limit, offset, paged, ok := pageParams(c)
	if !ok {
		return
	}
	search := strings.TrimSpace(c.Query("q"))
	if utf8.RuneCountInString(search) > maxSearchLength {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Search must be 100 characters or fewer"})
		return
	}

	var where string
	var args []any
	if search != "" {
		where = ` WHERE c.name ILIKE $1 OR c.address ILIKE $1`
		args = append(args, "%"+likeEscaper.Replace(search)+"%")
	}

	total, err := countRows(`SELECT COUNT(*) FROM cafes c`+where, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	query, args := appendPaging(cafeSelect+where+` GROUP BY c.id ORDER BY `+orderBy, args, limit, offset, paged)
	rows, err := database.DB.Query(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	cafes := []models.Cafe{}
	for rows.Next() {
		cafe, err := scanCafe(rows)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		cafes = append(cafes, cafe)
	}
	if err := rows.Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	setTotalCount(c, total)
	c.JSON(http.StatusOK, cafes)
}

func GetCafe(c *gin.Context) {
	id, ok := parseID(c, "id")
	if !ok {
		return
	}

	cafe, err := fetchCafe(id)
	if errors.Is(err, sql.ErrNoRows) {
		c.JSON(http.StatusNotFound, gin.H{"error": "Cafe not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, cafe)
}

type cafeRequest struct {
	Name        string `json:"name"`
	Address     string `json:"address"`
	Description string `json:"description"`
}

// validateCafeRequest trims the fields in place and returns an error message
// if the input is invalid, or "" if it is valid.
func validateCafeRequest(req *cafeRequest) string {
	req.Name = strings.TrimSpace(req.Name)
	req.Address = strings.TrimSpace(req.Address)
	req.Description = strings.TrimSpace(req.Description)
	if req.Name == "" || len(req.Name) > 255 {
		return "Name must be 1-255 characters"
	}
	return ""
}

// CreateCafe is admin-only; the route must be guarded by auth.RequireAdmin.
func CreateCafe(c *gin.Context) {
	var req cafeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}
	if msg := validateCafeRequest(&req); msg != "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": msg})
		return
	}

	var id int
	err := database.DB.QueryRow(
		`INSERT INTO cafes (name, address, description) VALUES ($1, $2, $3) RETURNING id`,
		req.Name, req.Address, req.Description,
	).Scan(&id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Return the full record, including DB-generated timestamp.
	cafe, err := fetchCafe(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, cafe)
}

// UpdateCafe replaces a cafe's name, address, and description. Admin-only.
func UpdateCafe(c *gin.Context) {
	id, ok := parseID(c, "id")
	if !ok {
		return
	}

	var req cafeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}
	if msg := validateCafeRequest(&req); msg != "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": msg})
		return
	}

	res, err := database.DB.Exec(
		`UPDATE cafes SET name = $1, address = $2, description = $3 WHERE id = $4`,
		req.Name, req.Address, req.Description, id,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	affected, err := res.RowsAffected()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if affected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Cafe not found"})
		return
	}

	cafe, err := fetchCafe(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, cafe)
}

// DeleteCafe removes a cafe along with its reviews and their photos. Admin-only.
func DeleteCafe(c *gin.Context) {
	id, ok := parseID(c, "id")
	if !ok {
		return
	}

	// Collect photo filenames before the cascade removes their rows.
	filenames, err := imageFilenames(`
		SELECT ri.filename
		FROM review_images ri
		JOIN reviews r ON r.id = ri.review_id
		WHERE r.cafe_id = $1`, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	res, err := database.DB.Exec(`DELETE FROM cafes WHERE id = $1`, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	affected, err := res.RowsAffected()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if affected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Cafe not found"})
		return
	}

	removeImageFiles(filenames) // best effort; the rows are already gone
	c.Status(http.StatusNoContent)
}

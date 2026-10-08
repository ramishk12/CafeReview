package handlers

import (
	"database/sql"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"cafe-review/internal/database"
	"cafe-review/internal/models"

	"github.com/gin-gonic/gin"
)

const cafeSelect = `
	SELECT c.id, c.name, c.address, c.description, c.created_at,
	       COALESCE(AVG(r.rating), 0), COUNT(r.id)
	FROM cafes c
	LEFT JOIN reviews r ON r.cafe_id = c.id`

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

func ListCafes(c *gin.Context) {
	rows, err := database.DB.Query(cafeSelect + `
		GROUP BY c.id
		ORDER BY c.name`)
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
	c.JSON(http.StatusOK, cafes)
}

func GetCafe(c *gin.Context) {
	id, ok := parseID(c, "id")
	if !ok {
		return
	}

	cafe, err := scanCafe(database.DB.QueryRow(cafeSelect+`
		WHERE c.id = $1
		GROUP BY c.id`, id))
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

type createCafeRequest struct {
	Name        string `json:"name"`
	Address     string `json:"address"`
	Description string `json:"description"`
}

// CreateCafe is admin-only; the route must be guarded by auth.RequireAdmin.
func CreateCafe(c *gin.Context) {
	var req createCafeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}
	req.Name = strings.TrimSpace(req.Name)
	req.Address = strings.TrimSpace(req.Address)
	req.Description = strings.TrimSpace(req.Description)

	if req.Name == "" || len(req.Name) > 255 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Name must be 1-255 characters"})
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
	cafe, err := scanCafe(database.DB.QueryRow(cafeSelect+`
		WHERE c.id = $1
		GROUP BY c.id`, id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, cafe)
}

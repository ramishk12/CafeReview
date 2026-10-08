package handlers

import (
	"database/sql"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"cafe-review/internal/auth"
	"cafe-review/internal/database"
	"cafe-review/internal/models"
	"cafe-review/internal/storage"

	"github.com/gin-gonic/gin"
)

// Uploads stores review images. main sets it at startup.
var Uploads storage.Storage

const maxReviewBodyLength = 2000

const reviewSelect = `
	SELECT r.id, r.cafe_id, r.user_id, u.display_name, r.rating, r.body, r.created_at, r.updated_at
	FROM reviews r
	JOIN users u ON u.id = r.user_id`

type reviewRequest struct {
	Rating int    `json:"rating"`
	Body   string `json:"body"`
}

// validateReviewRequest trims the body in place and returns an error message
// if the input is invalid, or "" if it is valid.
func validateReviewRequest(req *reviewRequest) string {
	req.Body = strings.TrimSpace(req.Body)
	if req.Rating < 1 || req.Rating > 5 {
		return "Rating must be between 1 and 5"
	}
	if req.Body == "" {
		return "Review text is required"
	}
	if len(req.Body) > maxReviewBodyLength {
		return "Review text must be at most 2000 characters"
	}
	return ""
}

func scanReview(row interface{ Scan(...any) error }) (models.Review, error) {
	var r models.Review
	err := row.Scan(
		&r.ID, &r.CafeID, &r.UserID, &r.AuthorName, &r.Rating, &r.Body,
		&r.CreatedAt, &r.UpdatedAt,
	)
	return r, err
}

// fetchReview returns one review with its author name and images.
func fetchReview(id int) (models.Review, error) {
	r, err := scanReview(database.DB.QueryRow(reviewSelect+` WHERE r.id = $1`, id))
	if err != nil {
		return r, err
	}
	images, err := loadImagesForReviews([]int{id})
	if err != nil {
		return r, err
	}
	r.Images = images[id]
	if r.Images == nil {
		r.Images = []models.ReviewImage{}
	}
	return r, nil
}

// loadImagesForReviews returns images grouped by review ID.
func loadImagesForReviews(reviewIDs []int) (map[int][]models.ReviewImage, error) {
	result := make(map[int][]models.ReviewImage)
	if len(reviewIDs) == 0 {
		return result, nil
	}

	query := `SELECT id, review_id, filename, created_at FROM review_images WHERE review_id IN (`
	args := make([]any, 0, len(reviewIDs))
	for i, id := range reviewIDs {
		if i > 0 {
			query += ", "
		}
		query += "$" + strconv.Itoa(i+1)
		args = append(args, id)
	}
	query += ") ORDER BY id"

	rows, err := database.DB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var img models.ReviewImage
		var filename string
		if err := rows.Scan(&img.ID, &img.ReviewID, &filename, &img.CreatedAt); err != nil {
			return nil, err
		}
		img.URL = Uploads.URL(filename)
		result[img.ReviewID] = append(result[img.ReviewID], img)
	}
	return result, rows.Err()
}

// requireReviewOwner loads the review's owner and writes 404 or 403 if the
// current user may not modify it. It returns false when a response was written.
func requireReviewOwner(c *gin.Context, reviewID int) bool {
	var ownerID int
	err := database.DB.QueryRow(`SELECT user_id FROM reviews WHERE id = $1`, reviewID).Scan(&ownerID)
	if errors.Is(err, sql.ErrNoRows) {
		c.JSON(http.StatusNotFound, gin.H{"error": "Review not found"})
		return false
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return false
	}
	if ownerID != auth.CurrentUserID(c) {
		c.JSON(http.StatusForbidden, gin.H{"error": "You can only modify your own reviews"})
		return false
	}
	return true
}

func ListReviews(c *gin.Context) {
	cafeID, ok := parseID(c, "id")
	if !ok {
		return
	}

	rows, err := database.DB.Query(reviewSelect+` WHERE r.cafe_id = $1 ORDER BY r.created_at DESC, r.id DESC`, cafeID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	reviews := []models.Review{}
	ids := []int{}
	for rows.Next() {
		r, err := scanReview(rows)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		reviews = append(reviews, r)
		ids = append(ids, r.ID)
	}
	if err := rows.Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	images, err := loadImagesForReviews(ids)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	for i := range reviews {
		reviews[i].Images = images[reviews[i].ID]
		if reviews[i].Images == nil {
			reviews[i].Images = []models.ReviewImage{}
		}
	}
	c.JSON(http.StatusOK, reviews)
}

func CreateReview(c *gin.Context) {
	cafeID, ok := parseID(c, "id")
	if !ok {
		return
	}

	var req reviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}
	if msg := validateReviewRequest(&req); msg != "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": msg})
		return
	}

	var id int
	err := database.DB.QueryRow(
		`INSERT INTO reviews (cafe_id, user_id, rating, body) VALUES ($1, $2, $3, $4) RETURNING id`,
		cafeID, auth.CurrentUserID(c), req.Rating, req.Body,
	).Scan(&id)
	if database.IsForeignKeyViolation(err) {
		c.JSON(http.StatusNotFound, gin.H{"error": "Cafe not found"})
		return
	}
	if database.IsUniqueViolation(err) {
		c.JSON(http.StatusConflict, gin.H{"error": "You have already reviewed this cafe"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	review, err := fetchReview(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, review)
}

func UpdateReview(c *gin.Context) {
	reviewID, ok := parseID(c, "id")
	if !ok {
		return
	}
	if !requireReviewOwner(c, reviewID) {
		return
	}

	var req reviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}
	if msg := validateReviewRequest(&req); msg != "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": msg})
		return
	}

	_, err := database.DB.Exec(
		`UPDATE reviews SET rating = $1, body = $2, updated_at = (CURRENT_TIMESTAMP AT TIME ZONE 'UTC')
		 WHERE id = $3`,
		req.Rating, req.Body, reviewID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	review, err := fetchReview(reviewID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, review)
}

func DeleteReview(c *gin.Context) {
	reviewID, ok := parseID(c, "id")
	if !ok {
		return
	}
	if !requireReviewOwner(c, reviewID) {
		return
	}

	// Collect image filenames first so the files can be removed after the
	// rows are deleted by the cascade.
	rows, err := database.DB.Query(`SELECT filename FROM review_images WHERE review_id = $1`, reviewID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	var filenames []string
	for rows.Next() {
		var f string
		if err := rows.Scan(&f); err != nil {
			rows.Close()
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		filenames = append(filenames, f)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	rows.Close()

	if _, err := database.DB.Exec(`DELETE FROM reviews WHERE id = $1`, reviewID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	for _, f := range filenames {
		_ = Uploads.Delete(f) // best effort; the review is already gone
	}
	c.Status(http.StatusNoContent)
}

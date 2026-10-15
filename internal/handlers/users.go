package handlers

import (
	"database/sql"
	"errors"
	"net/http"
	"time"

	"cafe-review/internal/apierror"
	"cafe-review/internal/database"
	"cafe-review/internal/models"

	"github.com/gin-gonic/gin"
)

// GetUserProfile returns a user's public profile: display name, join date, and
// review count. Email and other private data are never included.
func GetUserProfile(c *gin.Context) {
	id, ok := parseID(c, "id")
	if !ok {
		return
	}

	var (
		user        models.User
		reviewCount int
	)
	err := database.DB.QueryRow(`
		SELECT u.id, u.display_name, u.created_at, COUNT(r.id)
		FROM users u
		LEFT JOIN reviews r ON r.user_id = u.id
		WHERE u.id = $1
		GROUP BY u.id`, id,
	).Scan(&user.ID, &user.DisplayName, &user.CreatedAt, &reviewCount)
	if errors.Is(err, sql.ErrNoRows) {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}
	if err != nil {
		apierror.Internal(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":           user.ID,
		"display_name": user.DisplayName,
		"created_at":   user.CreatedAt,
		"review_count": reviewCount,
	})
}

// ListUserReviews returns a user's reviews, newest first, with the cafe name so
// they can be listed outside a cafe page. Paged with limit and offset.
func ListUserReviews(c *gin.Context) {
	id, ok := parseID(c, "id")
	if !ok {
		return
	}
	limit, offset, paged, ok := pageParams(c)
	if !ok {
		return
	}

	var exists bool
	if err := database.DB.QueryRow(`SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)`, id).Scan(&exists); err != nil {
		apierror.Internal(c, err)
		return
	}
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	total, err := countRows(`SELECT COUNT(*) FROM reviews WHERE user_id = $1`, id)
	if err != nil {
		apierror.Internal(c, err)
		return
	}

	query, args := appendPaging(`
		SELECT r.id, r.cafe_id, c.name, r.user_id, u.display_name, r.rating, r.body, r.visited_on, r.created_at, r.updated_at
		FROM reviews r
		JOIN users u ON u.id = r.user_id
		JOIN cafes c ON c.id = r.cafe_id
		WHERE r.user_id = $1
		ORDER BY r.created_at DESC, r.id DESC`, []any{id}, limit, offset, paged)

	rows, err := database.DB.Query(query, args...)
	if err != nil {
		apierror.Internal(c, err)
		return
	}
	defer rows.Close()

	reviews := []models.Review{}
	ids := []int{}
	for rows.Next() {
		var r models.Review
		var visited time.Time
		if err := rows.Scan(
			&r.ID, &r.CafeID, &r.CafeName, &r.UserID, &r.AuthorName, &r.Rating, &r.Body, &visited, &r.CreatedAt, &r.UpdatedAt,
		); err != nil {
			apierror.Internal(c, err)
			return
		}
		r.VisitedOn = visited.Format("2006-01-02")
		reviews = append(reviews, r)
		ids = append(ids, r.ID)
	}
	if err := rows.Err(); err != nil {
		apierror.Internal(c, err)
		return
	}

	images, err := loadImagesForReviews(ids)
	if err != nil {
		apierror.Internal(c, err)
		return
	}
	tags, err := loadTagsForReviews(ids)
	if err != nil {
		apierror.Internal(c, err)
		return
	}
	for i := range reviews {
		reviews[i].Images = images[reviews[i].ID]
		if reviews[i].Images == nil {
			reviews[i].Images = []models.ReviewImage{}
		}
		reviews[i].Tags = tags[reviews[i].ID]
		if reviews[i].Tags == nil {
			reviews[i].Tags = []models.Tag{}
		}
	}

	setTotalCount(c, total)
	c.JSON(http.StatusOK, reviews)
}

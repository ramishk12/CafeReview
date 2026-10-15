package handlers

import (
	"database/sql"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"cafe-review/internal/apierror"
	"cafe-review/internal/auth"
	"cafe-review/internal/database"

	"github.com/gin-gonic/gin"
)

// reviewVoteCount returns the number of helpful votes on a review.
func reviewVoteCount(reviewID int) (int, error) {
	var n int
	err := database.DB.QueryRow(`SELECT COUNT(*) FROM review_votes WHERE review_id = $1`, reviewID).Scan(&n)
	return n, err
}

// VoteReview marks a review as helpful for the signed-in user. Authors cannot
// vote on their own reviews. Voting twice is a no-op.
func VoteReview(c *gin.Context) {
	reviewID, ok := parseID(c, "id")
	if !ok {
		return
	}

	var authorID int
	err := database.DB.QueryRow(`SELECT user_id FROM reviews WHERE id = $1`, reviewID).Scan(&authorID)
	if errors.Is(err, sql.ErrNoRows) {
		c.JSON(http.StatusNotFound, gin.H{"error": "Review not found"})
		return
	}
	if err != nil {
		apierror.Internal(c, err)
		return
	}
	if authorID == auth.CurrentUserID(c) {
		c.JSON(http.StatusForbidden, gin.H{"error": "You cannot mark your own review as helpful"})
		return
	}

	if _, err := database.DB.Exec(
		`INSERT INTO review_votes (review_id, user_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`,
		reviewID, auth.CurrentUserID(c),
	); err != nil {
		apierror.Internal(c, err)
		return
	}

	count, err := reviewVoteCount(reviewID)
	if err != nil {
		apierror.Internal(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"helpful_count": count, "voted": true})
}

// UnvoteReview removes the signed-in user's helpful vote.
func UnvoteReview(c *gin.Context) {
	reviewID, ok := parseID(c, "id")
	if !ok {
		return
	}

	var exists bool
	if err := database.DB.QueryRow(`SELECT EXISTS(SELECT 1 FROM reviews WHERE id = $1)`, reviewID).Scan(&exists); err != nil {
		apierror.Internal(c, err)
		return
	}
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Review not found"})
		return
	}

	if _, err := database.DB.Exec(
		`DELETE FROM review_votes WHERE review_id = $1 AND user_id = $2`, reviewID, auth.CurrentUserID(c),
	); err != nil {
		apierror.Internal(c, err)
		return
	}

	count, err := reviewVoteCount(reviewID)
	if err != nil {
		apierror.Internal(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"helpful_count": count, "voted": false})
}

// ListMyReviewVotes returns the review IDs the signed-in user has voted on,
// optionally limited to one cafe with ?cafe=<id>.
func ListMyReviewVotes(c *gin.Context) {
	userID := auth.CurrentUserID(c)

	query := `
		SELECT rv.review_id
		FROM review_votes rv
		JOIN reviews r ON r.id = rv.review_id
		WHERE rv.user_id = $1`
	args := []any{userID}

	if raw := strings.TrimSpace(c.Query("cafe")); raw != "" {
		cafeID, err := strconv.Atoi(raw)
		if err != nil || cafeID <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "cafe must be a valid ID"})
			return
		}
		args = append(args, cafeID)
		query += " AND r.cafe_id = $2"
	}

	rows, err := database.DB.Query(query, args...)
	if err != nil {
		apierror.Internal(c, err)
		return
	}
	defer rows.Close()

	ids := []int{}
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			apierror.Internal(c, err)
			return
		}
		ids = append(ids, id)
	}
	if err := rows.Err(); err != nil {
		apierror.Internal(c, err)
		return
	}
	c.JSON(http.StatusOK, ids)
}

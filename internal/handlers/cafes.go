package handlers

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"unicode/utf8"

	"cafe-review/internal/apierror"
	"cafe-review/internal/database"
	"cafe-review/internal/models"

	"github.com/gin-gonic/gin"
)

// cafeSelect is grouped by c.id, so every selected column must come from cafes
// or be an aggregate. The background filename is fetched with a scalar subquery
// rather than a join, because a joined column cannot be selected under GROUP BY c.id.
const cafeSelect = `
	SELECT c.id, c.name, c.address, c.description, c.created_at, c.closed_at,
	       c.background_image_id, c.background_image_path,
	       (SELECT ri.filename FROM review_images ri WHERE ri.id = c.background_image_id),
	       COALESCE(AVG(r.rating), 0), COUNT(r.id), COUNT(DISTINCT r.user_id)
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
	var (
		closedAt   sql.NullTime
		background sql.NullString
		bgID       sql.NullInt64
		bgImage    sql.NullString
	)
	err := row.Scan(
		&cafe.ID, &cafe.Name, &cafe.Address, &cafe.Description, &cafe.CreatedAt, &closedAt,
		&bgID, &background, &bgImage,
		&cafe.AvgRating, &cafe.ReviewCount, &cafe.ReviewerCount,
	)
	if closedAt.Valid {
		cafe.Closed = true
		cafe.ClosedAt = &closedAt.Time
	}
	if bgID.Valid {
		id := int(bgID.Int64)
		cafe.BackgroundImageID = &id
	}
	// A chosen review image wins over an uploaded background.
	switch {
	case bgImage.Valid && bgImage.String != "":
		cafe.BackgroundImage = Uploads.URL(bgImage.String)
		cafe.BackgroundSource = "review_image"
	case background.Valid && background.String != "":
		cafe.BackgroundImage = Uploads.URL(background.String)
		cafe.BackgroundSource = "uploaded"
	}
	return cafe, err
}

// fetchCafe returns one cafe with its rating summary.
func fetchCafe(id int) (models.Cafe, error) {
	return scanCafe(database.DB.QueryRow(cafeSelect+`
		WHERE c.id = $1
		GROUP BY c.id`, id))
}

// ListCafes supports ?q= (name or address search), ?tags= (comma-separated tag
// IDs, matched with AND), ?sort=, and optional ?limit= and ?offset=. Without
// limit, every matching cafe is returned.
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
	tagIDs, ok := parseTagIDs(c)
	if !ok {
		return
	}

	var conditions []string
	var args []any
	if search != "" {
		args = append(args, "%"+likeEscaper.Replace(search)+"%")
		conditions = append(conditions, fmt.Sprintf("(c.name ILIKE $%d OR c.address ILIKE $%d)", len(args), len(args)))
	}
	if len(tagIDs) > 0 {
		placeholders := make([]string, len(tagIDs))
		for i, id := range tagIDs {
			args = append(args, id)
			placeholders[i] = fmt.Sprintf("$%d", len(args))
		}
		// A cafe must have every requested tag.
		conditions = append(conditions, fmt.Sprintf(
			`c.id IN (SELECT cafe_id FROM cafe_tags WHERE tag_id IN (%s) GROUP BY cafe_id HAVING COUNT(DISTINCT tag_id) = %d)`,
			strings.Join(placeholders, ", "), len(tagIDs)))
	}
	where := ""
	if len(conditions) > 0 {
		where = " WHERE " + strings.Join(conditions, " AND ")
	}

	total, err := countRows(`SELECT COUNT(*) FROM cafes c`+where, args...)
	if err != nil {
		apierror.Internal(c, err)
		return
	}

	query, queryArgs := appendPaging(cafeSelect+where+` GROUP BY c.id ORDER BY `+orderBy, args, limit, offset, paged)
	rows, err := database.DB.Query(query, queryArgs...)
	if err != nil {
		apierror.Internal(c, err)
		return
	}
	defer rows.Close()

	cafes := []models.Cafe{}
	ids := []int{}
	for rows.Next() {
		cafe, err := scanCafe(rows)
		if err != nil {
			apierror.Internal(c, err)
			return
		}
		cafes = append(cafes, cafe)
		ids = append(ids, cafe.ID)
	}
	if err := rows.Err(); err != nil {
		apierror.Internal(c, err)
		return
	}

	tagsByCafe, err := loadTagsForCafes(ids)
	if err != nil {
		apierror.Internal(c, err)
		return
	}
	for i := range cafes {
		cafes[i].Tags = tagsByCafe[cafes[i].ID]
		if cafes[i].Tags == nil {
			cafes[i].Tags = []models.Tag{}
		}
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
		apierror.Internal(c, err)
		return
	}

	tagsByCafe, err := loadTagsForCafes([]int{id})
	if err != nil {
		apierror.Internal(c, err)
		return
	}
	cafe.Tags = tagsByCafe[id]
	if cafe.Tags == nil {
		cafe.Tags = []models.Tag{}
	}
	c.JSON(http.StatusOK, cafe)
}

// maxSimilarCafes caps how many possible duplicates the check returns.
const maxSimilarCafes = 5

// ListSimilarCafes returns existing cafes whose name or address looks like the
// one being added, so an admin can check before creating a duplicate. The check
// is advisory: nothing here blocks a create. Query: ?name= and optional
// ?address=. Matching is case-insensitive and uses trigrams (pg_trgm) plus
// substring matching; the address is matched the same way when one is given,
// and matches that are close in both fields sort first.
func ListSimilarCafes(c *gin.Context) {
	name := strings.TrimSpace(c.Query("name"))
	if name == "" || utf8.RuneCountInString(name) > 255 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Name must be 1-255 characters"})
		return
	}
	address := strings.TrimSpace(c.Query("address"))

	likeName := "%" + likeEscaper.Replace(name) + "%"
	likeAddress := "%" + likeEscaper.Replace(address) + "%"

	rows, err := database.DB.Query(`
		SELECT c.id, c.name, c.address
		FROM cafes c
		WHERE lower(c.name) % lower($1)
		   OR c.name ILIKE $2
		   OR ($3 <> '' AND (lower(c.address) % lower($3) OR c.address ILIKE $4))
		ORDER BY (similarity(lower(c.name), lower($1)) +
		          CASE WHEN $3 = '' THEN 0 ELSE similarity(lower(c.address), lower($3)) END) DESC,
		         c.name ASC, c.id ASC
		LIMIT $5`,
		name, likeName, address, likeAddress, maxSimilarCafes)
	if err != nil {
		apierror.Internal(c, err)
		return
	}
	defer rows.Close()

	similar := []models.SimilarCafe{}
	for rows.Next() {
		var match models.SimilarCafe
		if err := rows.Scan(&match.ID, &match.Name, &match.Address); err != nil {
			apierror.Internal(c, err)
			return
		}
		similar = append(similar, match)
	}
	if err := rows.Err(); err != nil {
		apierror.Internal(c, err)
		return
	}
	c.JSON(http.StatusOK, similar)
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
	// Count characters, not bytes, so the limit matches the form's maxLength.
	if req.Name == "" || utf8.RuneCountInString(req.Name) > 255 {
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
		apierror.Internal(c, err)
		return
	}

	// Return the full record, including DB-generated timestamp.
	cafe, err := fetchCafe(id)
	if err != nil {
		apierror.Internal(c, err)
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
		apierror.Internal(c, err)
		return
	}
	affected, err := res.RowsAffected()
	if err != nil {
		apierror.Internal(c, err)
		return
	}
	if affected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Cafe not found"})
		return
	}

	cafe, err := fetchCafe(id)
	if err != nil {
		apierror.Internal(c, err)
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
		apierror.Internal(c, err)
		return
	}

	// The background image is stored on the cafe, so collect it too.
	var background sql.NullString
	if err := database.DB.QueryRow(`SELECT background_image_path FROM cafes WHERE id = $1`, id).Scan(&background); err != nil && !errors.Is(err, sql.ErrNoRows) {
		apierror.Internal(c, err)
		return
	}
	if background.Valid && background.String != "" {
		filenames = append(filenames, background.String)
	}

	res, err := database.DB.Exec(`DELETE FROM cafes WHERE id = $1`, id)
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
		c.JSON(http.StatusNotFound, gin.H{"error": "Cafe not found"})
		return
	}

	removeImageFiles(filenames) // best effort; the rows are already gone
	c.Status(http.StatusNoContent)
}

// CloseCafe marks a cafe closed. Closed cafes keep their reviews but accept no
// new reviews and no edits. Admin-only.
func CloseCafe(c *gin.Context) {
	setCafeClosed(c, true)
}

// ReopenCafe clears a cafe's closed state. Admin-only.
func ReopenCafe(c *gin.Context) {
	setCafeClosed(c, false)
}

func setCafeClosed(c *gin.Context, closed bool) {
	id, ok := parseID(c, "id")
	if !ok {
		return
	}

	query := `UPDATE cafes SET closed_at = NULL WHERE id = $1`
	if closed {
		query = `UPDATE cafes SET closed_at = (CURRENT_TIMESTAMP AT TIME ZONE 'UTC') WHERE id = $1`
	}
	res, err := database.DB.Exec(query, id)
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
		c.JSON(http.StatusNotFound, gin.H{"error": "Cafe not found"})
		return
	}

	cafe, err := fetchCafe(id)
	if err != nil {
		apierror.Internal(c, err)
		return
	}
	c.JSON(http.StatusOK, cafe)
}

// SetCafeBackground sets or replaces a cafe's background image. Admin-only.
func SetCafeBackground(c *gin.Context) {
	id, ok := parseID(c, "id")
	if !ok {
		return
	}

	var old sql.NullString
	err := database.DB.QueryRow(`SELECT background_image_path FROM cafes WHERE id = $1`, id).Scan(&old)
	if errors.Is(err, sql.ErrNoRows) {
		c.JSON(http.StatusNotFound, gin.H{"error": "Cafe not found"})
		return
	}
	if err != nil {
		apierror.Internal(c, err)
		return
	}

	filename, ok := storeUploadedImage(c)
	if !ok {
		return
	}

	if _, err := database.DB.Exec(`UPDATE cafes SET background_image_path = $1 WHERE id = $2`, filename, id); err != nil {
		_ = Uploads.Delete(filename)
		apierror.Internal(c, err)
		return
	}
	if old.Valid && old.String != "" {
		_ = Uploads.Delete(old.String) // best effort; the row already points at the new file
	}

	cafe, err := fetchCafe(id)
	if err != nil {
		apierror.Internal(c, err)
		return
	}
	c.JSON(http.StatusOK, cafe)
}

// DeleteCafeBackground clears a cafe's background image. Admin-only.
func DeleteCafeBackground(c *gin.Context) {
	id, ok := parseID(c, "id")
	if !ok {
		return
	}

	var old sql.NullString
	err := database.DB.QueryRow(`SELECT background_image_path FROM cafes WHERE id = $1`, id).Scan(&old)
	if errors.Is(err, sql.ErrNoRows) {
		c.JSON(http.StatusNotFound, gin.H{"error": "Cafe not found"})
		return
	}
	if err != nil {
		apierror.Internal(c, err)
		return
	}

	if _, err := database.DB.Exec(`UPDATE cafes SET background_image_path = NULL WHERE id = $1`, id); err != nil {
		apierror.Internal(c, err)
		return
	}
	if old.Valid && old.String != "" {
		_ = Uploads.Delete(old.String)
	}

	cafe, err := fetchCafe(id)
	if err != nil {
		apierror.Internal(c, err)
		return
	}
	c.JSON(http.StatusOK, cafe)
}

// SetCafeReviewImageBackground sets a cafe's background to one of its review
// images. The image must belong to a review of this cafe. Admin-only.
func SetCafeReviewImageBackground(c *gin.Context) {
	cafeID, ok := parseID(c, "id")
	if !ok {
		return
	}

	var req struct {
		ImageID int `json:"image_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.ImageID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "A valid image_id is required"})
		return
	}

	// The image must belong to a review of this cafe.
	var exists bool
	err := database.DB.QueryRow(`
		SELECT EXISTS(
			SELECT 1 FROM review_images ri
			JOIN reviews r ON r.id = ri.review_id
			WHERE ri.id = $1 AND r.cafe_id = $2
		)`, req.ImageID, cafeID).Scan(&exists)
	if err != nil {
		apierror.Internal(c, err)
		return
	}
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "That image does not belong to this cafe"})
		return
	}

	if _, err := database.DB.Exec(`UPDATE cafes SET background_image_id = $1 WHERE id = $2`, req.ImageID, cafeID); err != nil {
		apierror.Internal(c, err)
		return
	}

	cafe, err := fetchCafe(cafeID)
	if err != nil {
		apierror.Internal(c, err)
		return
	}
	c.JSON(http.StatusOK, cafe)
}

// ClearCafeReviewImageBackground clears a cafe's chosen review-image background,
// falling back to any uploaded background. Admin-only.
func ClearCafeReviewImageBackground(c *gin.Context) {
	cafeID, ok := parseID(c, "id")
	if !ok {
		return
	}

	res, err := database.DB.Exec(`UPDATE cafes SET background_image_id = NULL WHERE id = $1`, cafeID)
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
		c.JSON(http.StatusNotFound, gin.H{"error": "Cafe not found"})
		return
	}

	cafe, err := fetchCafe(cafeID)
	if err != nil {
		apierror.Internal(c, err)
		return
	}
	c.JSON(http.StatusOK, cafe)
}

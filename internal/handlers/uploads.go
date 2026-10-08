package handlers

import (
	"errors"
	"io"
	"net/http"

	"cafe-review/internal/database"
	"cafe-review/internal/models"

	"github.com/gin-gonic/gin"
)

const maxImageBytes = 5 << 20 // 5 MB per file

// allowedImageTypes maps sniffed content types to file extensions.
var allowedImageTypes = map[string]string{
	"image/jpeg": ".jpg",
	"image/png":  ".png",
	"image/webp": ".webp",
}

// UploadReviewImage accepts a multipart "image" field and attaches it to a
// review owned by the current user.
func UploadReviewImage(c *gin.Context) {
	reviewID, ok := parseID(c, "id")
	if !ok {
		return
	}
	if !requireReviewOwner(c, reviewID) {
		return
	}

	// Allow some slack for multipart boundaries and headers; the per-file
	// size check below is the real limit.
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxImageBytes+64<<10)

	fh, err := c.FormFile("image")
	if err != nil {
		var tooLarge *http.MaxBytesError
		if errors.As(err, &tooLarge) {
			c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": "Image must be 5 MB or smaller"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": "An image file is required in the 'image' field"})
		return
	}
	if fh.Size > maxImageBytes {
		c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": "Image must be 5 MB or smaller"})
		return
	}

	f, err := fh.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer f.Close()

	// Sniff the real content type instead of trusting the client header.
	head := make([]byte, 512)
	n, err := io.ReadFull(f, head)
	if err != nil && !errors.Is(err, io.ErrUnexpectedEOF) && !errors.Is(err, io.EOF) {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ext, allowed := allowedImageTypes[http.DetectContentType(head[:n])]
	if !allowed {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Only JPEG, PNG, and WebP images are allowed"})
		return
	}
	if _, err := f.Seek(0, io.SeekStart); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	filename, err := Uploads.Save(ext, f)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not save image"})
		return
	}

	var img models.ReviewImage
	err = database.DB.QueryRow(
		`INSERT INTO review_images (review_id, filename) VALUES ($1, $2) RETURNING id, created_at`,
		reviewID, filename,
	).Scan(&img.ID, &img.CreatedAt)
	if err != nil {
		_ = Uploads.Delete(filename)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	img.ReviewID = reviewID
	img.URL = Uploads.URL(filename)
	c.JSON(http.StatusCreated, img)
}

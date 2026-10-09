package server

import (
	"net/http"

	"cafe-review/internal/auth"
	"cafe-review/internal/handlers"

	"github.com/gin-gonic/gin"
)

// NewRouter builds the HTTP router. uploadDir is served at /uploads.
func NewRouter(uploadDir string) *gin.Engine {
	r := gin.Default()

	r.GET("/api/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	r.Static("/uploads", uploadDir)

	api := r.Group("/api")

	api.POST("/auth/register", auth.Register)
	api.POST("/auth/login", auth.Login)

	authed := api.Group("")
	authed.Use(auth.RequireAuth())
	authed.GET("/auth/me", auth.Me)

	api.GET("/cafes", handlers.ListCafes)
	api.GET("/cafes/:id", handlers.GetCafe)
	api.GET("/cafes/:id/reviews", handlers.ListReviews)

	authed.POST("/cafes/:id/reviews", handlers.CreateReview)
	authed.PUT("/reviews/:id", handlers.UpdateReview)
	authed.DELETE("/reviews/:id", handlers.DeleteReview)
	authed.POST("/reviews/:id/images", handlers.UploadReviewImage)
	authed.DELETE("/reviews/:id/images/:imageId", handlers.DeleteReviewImage)

	admin := authed.Group("")
	admin.Use(auth.RequireAdmin())
	admin.POST("/cafes", handlers.CreateCafe)
	admin.PUT("/cafes/:id", handlers.UpdateCafe)
	admin.DELETE("/cafes/:id", handlers.DeleteCafe)

	return r
}

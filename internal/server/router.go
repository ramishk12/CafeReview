package server

import (
	"net/http"

	"cafe-review/internal/apierror"
	"cafe-review/internal/auth"
	"cafe-review/internal/handlers"

	"github.com/gin-gonic/gin"
)

// NewRouter builds the HTTP router. uploadDir is served at /uploads.
func NewRouter(uploadDir string) *gin.Engine {
	r := gin.Default()
	r.Use(apierror.RequestID())
	r.Use(securityHeaders())

	r.GET("/api/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	r.Static("/uploads", uploadDir)

	api := r.Group("/api")

	api.POST("/auth/register", auth.RegisterLimiter.Middleware(), auth.Register)
	api.POST("/auth/login", auth.LoginLimiter.Middleware(), auth.Login)
	api.POST("/auth/verify/resend", auth.ResendLimiter.Middleware(), auth.ResendVerification)
	api.GET("/auth/verify", auth.Verify)

	authed := api.Group("")
	authed.Use(auth.RequireAuth())
	// RequireCSRF skips safe methods, so GET /auth/me passes and writes need the header.
	authed.Use(auth.RequireCSRF())
	authed.GET("/auth/me", auth.Me)
	authed.POST("/auth/logout", auth.Logout)
	authed.PUT("/review-images/:id/tags", handlers.SetReviewImageTags)
	authed.GET("/me/lists", handlers.ListMyLists)
	authed.POST("/me/lists", handlers.CreateList)
	authed.GET("/lists/:id", handlers.GetList)
	authed.PUT("/lists/:id", handlers.UpdateList)
	authed.DELETE("/lists/:id", handlers.DeleteList)
	authed.POST("/lists/:id/cafes", handlers.AddCafeToList)
	authed.DELETE("/lists/:id/cafes/:cafeId", handlers.RemoveCafeFromList)

	api.GET("/tags", handlers.ListTags)
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
	admin.PUT("/cafes/:id/tags", handlers.SetCafeTags)
	admin.PUT("/cafes/:id/background", handlers.SetCafeBackground)
	admin.PUT("/cafes/:id/background/review-image", handlers.SetCafeReviewImageBackground)
	admin.DELETE("/cafes/:id/background/review-image", handlers.ClearCafeReviewImageBackground)
	admin.DELETE("/cafes/:id/background", handlers.DeleteCafeBackground)
	admin.POST("/cafes/:id/close", handlers.CloseCafe)
	admin.POST("/cafes/:id/reopen", handlers.ReopenCafe)
	admin.PUT("/cafes/:id", handlers.UpdateCafe)
	admin.DELETE("/cafes/:id", handlers.DeleteCafe)

	return r
}

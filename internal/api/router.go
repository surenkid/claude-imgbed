package api

import (
	"embed"
	"claude-imgbed/internal/auth"
	"claude-imgbed/internal/config"
	"claude-imgbed/internal/models"
	"claude-imgbed/internal/ratelimit"
	"io/fs"
	"net/http"

	"github.com/gin-gonic/gin"
)

func SetupRouter(cfg *config.Config, recentUploads *models.RecentUploads, staticFS embed.FS) *gin.Engine {
	// Set Gin mode
	gin.SetMode(gin.ReleaseMode)

	router := gin.Default()

	// CORS middleware
	router.Use(CORSMiddleware())

	// Initialize components
	authenticator := auth.NewAuthenticator(cfg.Auth.Token)
	rateLimiter := ratelimit.NewIPRateLimiter(cfg.RateLimit.RequestsPerMinute, cfg.RateLimit.Burst)
	rateLimiter.Cleanup() // Start cleanup goroutine

	handler := NewHandler(cfg, recentUploads)

	// Health check (no auth required)
	router.GET("/health", handler.HealthCheck)

	// Image serving (no auth required, public access)
	router.GET("/images/:year/:month/:filename", handler.ServeImage)

	// API routes (require authentication and rate limiting)
	api := router.Group("/api")
	api.Use(AuthMiddleware(authenticator))
	api.Use(RateLimitMiddleware(rateLimiter))
	{
		api.POST("/upload", handler.UploadImage)
		api.GET("/recent", handler.GetRecentUploads)
	}

	// Serve embedded frontend static files
	// Strip the "web/dist" prefix from the embedded filesystem
	distFS, err := fs.Sub(staticFS, "web/dist")
	if err == nil {
		router.NoRoute(gin.WrapH(http.FileServer(http.FS(distFS))))
	}

	return router
}

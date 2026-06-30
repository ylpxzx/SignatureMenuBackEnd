package app

import (
	"net/http"

	"signature-menu-backend/internal/auth"
	"signature-menu-backend/internal/config"
	"signature-menu-backend/internal/home"
	"signature-menu-backend/internal/middleware"
	"signature-menu-backend/internal/recipe"
	"signature-menu-backend/internal/store"
	"signature-menu-backend/pkg/token"

	"github.com/gin-gonic/gin"
)

func NewRouter(cfg config.Config, dataStore *store.Store, tokens *token.Manager) *gin.Engine {
	if cfg.AppEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.Default()
	router.Use(middleware.CORS(cfg.AllowedOrigins))

	router.GET("/api/v1/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"data": gin.H{"status": "ok", "app": "Signature Menu"}})
	})

	api := router.Group("/api/v1")
	protected := api.Group("")
	protected.Use(middleware.Auth(tokens))

	auth.NewHandler(dataStore, tokens).RegisterRoutes(api, protected)
	recipe.NewHandler(dataStore).RegisterRoutes(protected)
	home.NewHandler(dataStore).RegisterRoutes(protected)

	return router
}

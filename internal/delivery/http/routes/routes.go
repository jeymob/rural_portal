package routes

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/jeymob/rural-portal/internal/config"
)

func SetupRoutes(r *gin.Engine, db *gorm.DB, cfg *config.Config) {
	// Health-check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"version": "0.1.0-mvp",
		})
	})

	// Пока только health
	// Здесь позже добавим /api/v1/auth, /api/v1/ads и т.д.
}

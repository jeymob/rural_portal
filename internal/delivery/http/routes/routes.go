package routes

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/jeymob/rural-portal/internal/auth"
	"github.com/jeymob/rural-portal/internal/config"
	"github.com/jeymob/rural-portal/internal/delivery/http/handlers"

	"github.com/jeymob/rural-portal/internal/middleware"
)

func SetupRoutes(r *gin.Engine, db *gorm.DB, cfg *config.Config) {
	// Health-check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"version": "0.1.0-mvp",
		})
	})

	r.GET("/api/health", handlers.HealthCheck) // например

	// Открытые роуты для Яндекс OAuth
	r.GET("/api/auth/yandex", auth.YandexLogin(cfg))
	r.GET("/api/auth/yandex/callback", auth.YandexCallback(db, cfg))

	// Открытые роуты для ВК OAuth
	r.GET("/api/auth/vk", auth.VkLogin(cfg))
	r.GET("/api/auth/vk/callback", auth.VkCallback(db, cfg))

	// Вот начинается самое важное — группа защищённых роутов
	protected := r.Group("/api")                  // все роуты будут начинаться с /api/...
	protected.Use(middleware.AuthMiddleware(cfg)) // ← ШАГ 3: подключаем проверку токена ко всей группе

	// Теперь любой роут внутри этой группы будет проверять токен
	protected.GET("/me", handlers.GetMe(db)) // информация о себе

	// Можно создать ещё одну группу только для админов
	admin := protected.Group("/admin")      // /api/admin/...
	admin.Use(middleware.AdminMiddleware()) // дополнительная проверка роли
}

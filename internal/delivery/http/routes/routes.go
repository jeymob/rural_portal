package routes

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

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

	// Открытые роуты — сюда может зайти любой, без токена
	r.POST("/api/register", handlers.Register(db, cfg))
	r.POST("/api/login", handlers.Login(db, cfg))
	r.GET("/api/health", handlers.HealthCheck) // например

	// Вот начинается самое важное — группа защищённых роутов
	protected := r.Group("/api")                  // все роуты будут начинаться с /api/...
	protected.Use(middleware.AuthMiddleware(cfg)) // ← ШАГ 3: подключаем проверку токена ко всей группе

	// Теперь любой роут внутри этой группы будет проверять токен
	protected.GET("/me", handlers.GetMe(db)) // информация о себе

	// Можно создать ещё одну группу только для админов
	admin := protected.Group("/admin")            // /api/admin/...
	admin.Use(middleware.AdminMiddleware())       // дополнительная проверка роли
	admin.GET("/users", handlers.GetAllUsers(db)) // список всех пользователей — только админ
}

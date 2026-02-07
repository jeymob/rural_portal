package middleware

import (
	"net/http"

	"github.com/jeymob/rural-portal/internal/config" // предполагаю, что у тебя есть config с JWTSecret
	"github.com/jeymob/rural-portal/internal/utils"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware — проверяет JWT-токен в заголовке Authorization: Bearer ...
func AuthMiddleware(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString, err := c.Cookie("access_token")
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Не авторизован"})
			return
		}

		claims, err := utils.ValidateAccessToken(tokenString, cfg.JWTSecret)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Неверный или просроченный токен"})
			return
		}

		c.Set("user_id", claims.UserID)
		c.Set("role", claims.Role)
		c.Next()
	}
}

func AdminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, _ := c.Get("role")
		if role != "admin" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Доступ только для администраторов"})
			return
		}
		c.Next()
	}
}

package middleware

import (
	"net/http"
	"strings"

	"github.com/jeymob/rural-portal/internal/config" // предполагаю, что у тебя есть config с JWTSecret

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// AuthMiddleware — проверяет JWT-токен в заголовке Authorization: Bearer ...
func AuthMiddleware(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Требуется авторизация"})
			return
		}

		// Ожидаем формат: Bearer <token>
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Неверный формат токена"})
			return
		}

		tokenString := parts[1]

		// Парсим токен
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// Проверяем, что метод подписи правильный
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			// Возвращаем секретный ключ
			return []byte(cfg.JWTSecret), nil
		})

		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Недействительный токен"})
			return
		}

		if !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Токен просрочен или недействителен"})
			return
		}

		// Токен валидный — сохраняем полезные данные в контексте Gin
		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			c.Set("user_id", claims["user_id"])

			if role, ok := claims["role"].(string); ok {
				c.Set("role", role)
			}
			// Можно сохранить и другие поля, если добавлял
		}

		// Всё хорошо — продолжаем обработку запроса
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

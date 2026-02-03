package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jeymob/rural-portal/internal/infrastructure/persistence/models"
	"gorm.io/gorm"
)

func GetMe(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Достаём user_id из контекста (middleware уже проверил токен)
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Не удалось определить пользователя"})
			return
		}

		var user models.User
		if err := db.First(&user, userID).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Пользователь не найден"})
			return
		}

		// Возвращаем данные без пароля
		c.JSON(http.StatusOK, gin.H{
			"id":       user.ID,
			"username": user.Username,
			"email":    user.Email,
			"phone":    user.Phone,
			"region":   user.Region,
			"role":     user.Role,
		})
	}
}

// GetAllUsers — возвращает список всех пользователей (только для админа)
func GetAllUsers(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var users []models.User

		// Находим всех пользователей, но без паролей
		if err := db.Select("id", "username", "email", "phone", "region", "role", "created_at").
			Find(&users).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось получить список пользователей"})
			return
		}

		// Формируем ответ без чувствительных данных
		type UserResponse struct {
			ID        uint   `json:"id"`
			Username  string `json:"username"`
			Email     string `json:"email"`
			Phone     string `json:"phone,omitempty"`
			Region    string `json:"region,omitempty"`
			Role      string `json:"role"`
			CreatedAt string `json:"created_at"`
		}

		var response []UserResponse
		for _, u := range users {
			response = append(response, UserResponse{
				ID:        u.ID,
				Username:  u.Username,
				Email:     u.Email,
				Phone:     u.Phone,
				Region:    u.Region,
				Role:      u.Role,
				CreatedAt: u.CreatedAt.Format("2006-01-02 15:04:05"),
			})
		}

		c.JSON(http.StatusOK, gin.H{
			"users": response,
			"count": len(response),
		})
	}
}

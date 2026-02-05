package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jeymob/rural-portal/internal/infrastructure/persistence/models"
	"gorm.io/gorm"
)

func GetMe(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Не авторизован"})
			return
		}

		var user models.User
		if err := db.First(&user, userID).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Пользователь не найден"})
			return
		}

		c.JSON(http.StatusOK, user)
	}
}

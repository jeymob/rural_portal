package handlers

import (
	"net/http"
	"time"

	"github.com/google/uuid"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/jeymob/rural-portal/internal/config" // если у тебя есть JWT_SECRET в конфиге
	"github.com/jeymob/rural-portal/internal/infrastructure/persistence/models"
)

type RegisterInput struct {
	Username string `json:"username" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
	Phone    string `json:"phone,omitempty"`
	Region   string `json:"region,omitempty"`
}

type LoginInput struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type AuthResponse struct {
	Token string `json:"token"`
}

// Register — регистрация нового пользователя
func Register(db *gorm.DB, cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		var input RegisterInput
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Проверяем, существует ли уже пользователь
		var existingUser models.User
		if err := db.Where("email = ?", input.Email).First(&existingUser).Error; err == nil {
			c.JSON(http.StatusConflict, gin.H{"error": "Пользователь с таким email уже существует"})
			return
		}

		// Хэшируем пароль
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка сервера"})
			return
		}

		user := models.User{
			Username: input.Username,
			Email:    input.Email,
			Password: string(hashedPassword),
			Phone:    input.Phone,
			Region:   input.Region,
			Role:     "user", // по умолчанию обычный пользователь
		}

		if err := db.Create(&user).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось создать пользователя"})
			return
		}

		// Создаём токен
		tokenString, err := generateJWT(user.ID, user.Role, cfg.JWTSecret)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка создания токена"})
			return
		}

		c.JSON(http.StatusCreated, AuthResponse{Token: tokenString})
	}
}

func Login(db *gorm.DB, cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		var input LoginInput
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var user models.User
		if err := db.Where("email = ?", input.Email).First(&user).Error; err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Неверный email или пароль"})
			return
		}

		if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password)); err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Неверный email или пароль"})
			return
		}

		accessToken, _ := generateJWT(user.ID, user.Role, cfg.JWTSecret)

		// Генерируем refresh-токен (лучше использовать uuid)
		refreshToken := uuid.New().String()

		// Сохраняем в базу
		db.Create(&models.RefreshToken{
			UserID:    user.ID,
			Token:     refreshToken,
			ExpiresAt: time.Now().Add(30 * 24 * time.Hour), // 30 дней
		})

		// Устанавливаем HttpOnly cookie
		c.SetCookie(
			"refresh_token",
			refreshToken,
			30*24*60*60, // 30 дней
			"/",
			"",   // домен (пустой = текущий)
			true, // Secure — в продакшене true (только HTTPS)
			true, // HttpOnly — JS не видит
		)

		c.JSON(http.StatusOK, gin.H{
			"access_token": accessToken,
		})
	}
}

func Refresh(db *gorm.DB, cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		refreshToken, err := c.Cookie("refresh_token")
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Refresh-токен не найден"})
			return
		}

		var rt models.RefreshToken
		if err := db.Where("token = ? AND revoked = ? AND expires_at > ?", refreshToken, false, time.Now()).
			First(&rt).Error; err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Недействительный или просроченный refresh-токен"})
			return
		}

		var user models.User
		if err := db.First(&user, rt.UserID).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Пользователь не найден"})
			return
		}

		// Выдаём новый access-токен
		newAccessToken, _ := generateJWT(user.ID, user.Role, cfg.JWTSecret)

		// Можно выдать новый refresh-токен (рекомендуется для ротации)
		newRefreshToken := uuid.New().String()
		db.Model(&rt).Updates(models.RefreshToken{
			Token:     newRefreshToken,
			ExpiresAt: time.Now().Add(30 * 24 * time.Hour),
		})

		c.SetCookie("refresh_token", newRefreshToken, 30*24*60*60, "/", "", true, true)

		c.JSON(http.StatusOK, gin.H{
			"access_token": newAccessToken,
		})
	}
}

// Вспомогательная функция для генерации JWT
func generateJWT(userID uint, role string, secret string) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"role":    role,
		"exp":     time.Now().Add(time.Hour * 24 * 7).Unix(), // 7 дней
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

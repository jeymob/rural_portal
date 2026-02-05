package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
	"gorm.io/gorm"

	"github.com/jeymob/rural-portal/internal/config"
	"github.com/jeymob/rural-portal/internal/infrastructure/persistence/models"
	"github.com/jeymob/rural-portal/internal/utils"
)

// YandexOAuthConfig возвращает конфигурацию для Яндекс OAuth
func YandexOAuthConfig(cfg *config.Config) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     cfg.YandexClientID,
		ClientSecret: cfg.YandexClientSecret,
		RedirectURL:  cfg.YandexRedirectURI,
		Scopes:       []string{"login:email", "login:info"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://oauth.yandex.ru/authorize",
			TokenURL: "https://oauth.yandex.ru/token",
		},
	}
}

// YandexLogin — начало авторизации через Яндекс
func YandexLogin(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		oauthConfig := YandexOAuthConfig(cfg)
		// state — защита от CSRF
		state := "yandex-state" // можно сгенерировать уникальный
		url := oauthConfig.AuthCodeURL(state, oauth2.AccessTypeOffline)
		c.Redirect(http.StatusTemporaryRedirect, url)
	}
}

// YandexCallback — обработка ответа от Яндекса
func YandexCallback(db *gorm.DB, cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		code := c.Query("code")
		state := c.Query("state")

		if code == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Код авторизации не получен"})
			return
		}

		// Проверяем state (защита от CSRF)
		if state != "yandex-state" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный state-параметр"})
			return
		}

		oauthConfig := YandexOAuthConfig(cfg)

		// Обмен кода на токен
		token, err := oauthConfig.Exchange(c.Request.Context(), code)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка обмена кода на токен"})
			return
		}

		// Получаем данные пользователя
		client := oauthConfig.Client(context.Background(), token)
		resp, err := client.Get("https://login.yandex.ru/info?format=json")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка получения данных пользователя"})
			return
		}
		defer resp.Body.Close()

		body, _ := io.ReadAll(resp.Body)
		var yandexUser struct {
			ID            string `json:"id"`
			Login         string `json:"login"`
			DefaultEmail  string `json:"default_email"`
			DisplayName   string `json:"display_name"`
			DefaultAvatar string `json:"default_avatar_id"`
		}

		if err := json.Unmarshal(body, &yandexUser); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка парсинга данных"})
			return
		}

		// Ищем пользователя по email
		var user models.User
		err = db.Where("email = ?", yandexUser.DefaultEmail).First(&user).Error

		if err == gorm.ErrRecordNotFound {
			// Создаём нового пользователя
			user = models.User{
				Username: yandexUser.Login,
				Email:    yandexUser.DefaultEmail,
				// Можно сохранить имя, аватар и т.д.
				Role: "user",
			}
			if err := db.Create(&user).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка создания пользователя"})
				return
			}
		} else if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка базы данных"})
			return
		}

		// Генерируем свой access-токен
		accessToken, err := utils.GenerateAccessToken(user.ID, user.Role, cfg.JWTSecret)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка создания токена"})
			return
		}

		// Редирект на фронт с токеном в query-параметре
		redirectURL := fmt.Sprintf("http://localhost:5173/?access_token=%s", accessToken)
		c.Redirect(http.StatusTemporaryRedirect, redirectURL)
	}
}

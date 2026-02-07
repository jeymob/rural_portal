package auth

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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

func YandexLogin(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		oauthConfig := YandexOAuthConfig(cfg)

		// Генерируем уникальный state
		state := uuid.New().String()

		// Сохраняем state в cookie (короткий срок жизни)
		c.SetCookie(
			"oauth_state", // имя куки
			state,         // значение
			300,           // 5 минут
			"/",           // путь
			"",            // домен (текущий)
			false,         // Secure — true в продакшене (HTTPS)
			true,          // HttpOnly — защита от XSS
		)

		// Формируем URL авторизации с state
		url := oauthConfig.AuthCodeURL(state, oauth2.AccessTypeOffline)
		c.Redirect(http.StatusTemporaryRedirect, url)
	}
}

// YandexCallback — обработка ответа от Яндекса
func YandexCallback(db *gorm.DB, cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		code := c.Query("code")
		receivedState := c.Query("state")

		if code == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Код авторизации не получен"})
			return
		}

		expectedState, err := c.Cookie("oauth_state")
		if err != nil || receivedState != expectedState {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный state"})
			return
		}

		c.SetCookie("oauth_state", "", -1, "/", "", false, true)

		oauthConfig := YandexOAuthConfig(cfg)
		token, err := oauthConfig.Exchange(c.Request.Context(), code)
		if err != nil {
			log.Printf("Ошибка обмена кода: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка авторизации"})
			return
		}

		client := oauthConfig.Client(context.Background(), token)
		resp, err := client.Get("https://login.yandex.ru/info?format=json")
		if err != nil {
			log.Printf("Ошибка получения данных: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка профиля"})
			return
		}
		defer resp.Body.Close()

		var yandexUser struct {
			ID          string `json:"id"`
			DisplayName string `json:"display_name"`
		}

		if err := json.NewDecoder(resp.Body).Decode(&yandexUser); err != nil {
			log.Printf("Ошибка парсинга: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка данных"})
			return
		}

		var user models.User
		err = db.Where("yandex_id = ?", yandexUser.ID).First(&user).Error

		if errors.Is(err, gorm.ErrRecordNotFound) {
			user = models.User{
				YandexID: yandexUser.ID,
				Name:     yandexUser.DisplayName,
				Role:     "user",
			}

			if user.Name == "" {
				user.Name = "Пользователь"
			}

			if err := db.Create(&user).Error; err != nil {
				log.Printf("Ошибка создания пользователя: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка создания аккаунта"})
				return
			}
		} else if err == nil {
			// Обновляем имя, если изменилось
			newName := yandexUser.DisplayName
			if user.Name != newName {
				user.Name = newName
				db.Save(&user)
			}
		} else {
			log.Printf("Ошибка поиска: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка базы"})
			return
		}

		accessToken, err := utils.GenerateAccessToken(user.ID, user.Name, cfg.JWTSecret)
		if err != nil {
			log.Printf("Ошибка токена: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка сессии"})
			return
		}

		// Кладём токен в HttpOnly cookie
		c.SetCookie(
			"access_token",
			accessToken,
			900,
			"/",
			"",
			strings.HasPrefix(cfg.FrontendURL, "https://"), // Secure только на https
			true,
		)

		c.Redirect(http.StatusTemporaryRedirect, cfg.FrontendURL)
	}
}

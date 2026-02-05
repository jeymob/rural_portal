package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

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

		// Получаем сохранённый state из cookie
		expectedState, err := c.Cookie("oauth_state")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Отсутствует сохранённый state (возможно, cookie не передалась)"})
			return
		}

		// Проверяем, что пришедший state совпадает с сохранённым
		if receivedState == "" || receivedState != expectedState {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный или отсутствующий state-параметр"})
			return
		}

		// Удаляем state-cookie после проверки (не нужен больше)
		c.SetCookie("oauth_state", "", -1, "/", "", false, true)

		oauthConfig := YandexOAuthConfig(cfg)

		// Обмен кода на токен
		token, err := oauthConfig.Exchange(c.Request.Context(), code)
		if err != nil {
			log.Printf("Ошибка обмена кода на токен: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка авторизации"})
			return
		}

		// Получаем данные пользователя
		client := oauthConfig.Client(context.Background(), token)
		resp, err := client.Get("https://login.yandex.ru/info?format=json")
		if err != nil {
			log.Printf("Ошибка получения данных от Яндекса: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка получения профиля"})
			return
		}
		defer resp.Body.Close()

		var yandexUser struct {
			ID            string `json:"id"`
			Login         string `json:"login"`
			DefaultEmail  string `json:"default_email"`
			DisplayName   string `json:"display_name"`
			DefaultAvatar string `json:"default_avatar_id"`
		}

		if err := json.NewDecoder(resp.Body).Decode(&yandexUser); err != nil {
			log.Printf("Ошибка парсинга ответа Яндекса: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка обработки данных"})
			return
		}

		// Ищем пользователя по yandex_id
		var user models.User
		err = db.Where("yandex_id = ?", yandexUser.ID).First(&user).Error

		if err == gorm.ErrRecordNotFound {
			// Новый пользователь
			user = models.User{
				YandexID: yandexUser.ID,
				Name:     yandexUser.DisplayName,
				Role:     "user",
			}

			if err := db.Create(&user).Error; err != nil {
				log.Printf("Ошибка создания пользователя: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка создания аккаунта"})
				return
			}
		} else if err != nil {
			log.Printf("Ошибка поиска пользователя: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка базы данных"})
			return
		}

		// Генерируем свой access-токен
		accessToken, err := utils.GenerateAccessToken(user.ID, user.Name, cfg.JWTSecret)
		if err != nil {
			log.Printf("Ошибка генерации токена: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка создания сессии"})
			return
		}

		// Замени SetCookie на это:
		redirectURL := fmt.Sprintf("http://localhost:5173/?access_token=%s", accessToken)
		c.Redirect(http.StatusTemporaryRedirect, redirectURL)

		/*// Сохраняем токен в HttpOnly cookie (рекомендуется)
		c.SetCookie(
			"access_token",
			accessToken,
			3600*24*7, // 7 дней
			"/",
			"",
			false, // Secure = true в продакшене
			true,  // HttpOnly
		)

		// Редирект на главную без параметров в URL
		c.Redirect(http.StatusTemporaryRedirect, "http://localhost:5173/")*/
	}
}

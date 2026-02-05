package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/oauth2"
	"gorm.io/gorm"

	"github.com/jeymob/rural-portal/internal/config"
	"github.com/jeymob/rural-portal/internal/infrastructure/persistence/models"
	"github.com/jeymob/rural-portal/internal/utils"
)

func VkOAuthConfig(cfg *config.Config) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     cfg.VkClientID,
		ClientSecret: cfg.VkClientSecret,
		RedirectURL:  cfg.VkRedirectURI,
		Scopes:       []string{"email"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://oauth.vk.com/authorize",
			TokenURL: "https://oauth.vk.com/access_token",
		},
	}
}

func VkLogin(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		oauthConfig := VkOAuthConfig(cfg)
		state := "vk-state-" + uuid.New().String()
		url := oauthConfig.AuthCodeURL(state, oauth2.AccessTypeOffline)
		c.Redirect(http.StatusTemporaryRedirect, url)
	}
}

func VkCallback(db *gorm.DB, cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		code := c.Query("code")
		if code == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Код не получен"})
			return
		}

		oauthConfig := VkOAuthConfig(cfg)
		token, err := oauthConfig.Exchange(c.Request.Context(), code)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка обмена кода"})
			return
		}

		// Получаем данные пользователя
		client := oauthConfig.Client(context.Background(), token)
		resp, err := client.Get("https://api.vk.com/method/users.get?fields=photo_200,email&access_token=" + token.AccessToken + "&v=5.199")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка получения данных"})
			return
		}
		defer resp.Body.Close()

		body, _ := io.ReadAll(resp.Body)
		var vkResp struct {
			Response []struct {
				ID        int    `json:"id"`
				FirstName string `json:"first_name"`
				LastName  string `json:"last_name"`
				Photo200  string `json:"photo_200"`
			} `json:"response"`
		}
		json.Unmarshal(body, &vkResp)

		if len(vkResp.Response) == 0 {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Нет данных пользователя"})
			return
		}

		vkUser := vkResp.Response[0]

		email := ""
		if emailFromToken, ok := token.Extra("email").(string); ok {
			email = emailFromToken
		}

		// Ищем или создаём пользователя
		var user models.User
		err = db.Where("email = ?", email).First(&user).Error

		if err == gorm.ErrRecordNotFound {
			user = models.User{
				Username: fmt.Sprintf("%s %s", vkUser.FirstName, vkUser.LastName),
				Email:    email,
				Role:     "user",
			}
			db.Create(&user)
		}

		accessToken, _ := utils.GenerateAccessToken(user.ID, user.Role, cfg.JWTSecret)

		redirectURL := fmt.Sprintf("http://localhost:5173/?access_token=%s", accessToken)
		c.Redirect(http.StatusTemporaryRedirect, redirectURL)
	}
}

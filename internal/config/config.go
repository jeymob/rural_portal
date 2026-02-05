package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port        string
	DatabaseURL string
	JWTSecret   string

	YandexClientID     string `env:"YANDEX_CLIENT_ID"`
	YandexClientSecret string `env:"YANDEX_CLIENT_SECRET"`
	YandexRedirectURI  string `env:"YANDEX_REDIRECT_URI"`

	VkClientID     string `env:"VK_CLIENT_ID"`
	VkClientSecret string `env:"VK_CLIENT_SECRET"`
	VkRedirectURI  string `env:"VK_REDIRECT_URI"`
}

func Load() *Config {
	_ = godotenv.Load() // игнорируем ошибку, если .env нет

	cfg := &Config{
		Port:        getEnv("PORT", "8080"),
		DatabaseURL: getEnv("DATABASE_URL", ""),
		JWTSecret:   getEnv("JWT_SECRET", "default-secret-change-me"),

		YandexClientID:     getEnv("YANDEX_CLIENT_ID", ""),
		YandexClientSecret: getEnv("YANDEX_CLIENT_SECRET", ""),
		YandexRedirectURI:  getEnv("YANDEX_REDIRECT_URI", "http://localhost:8080/api/auth/yandex/callback"),

		VkClientID:     getEnv("VK_CLIENT_ID", ""),
		VkClientSecret: getEnv("VK_CLIENT_SECRET", ""),
		VkRedirectURI:  getEnv("VK_REDIRECT_URI", "http://localhost:8080/api/auth/vk/callback"),
	}

	if cfg.DatabaseURL == "" {
		log.Fatal("DATABASE_URL is required")
	}

	return cfg
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

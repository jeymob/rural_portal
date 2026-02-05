package main

import (
	"context"
	"database/sql"
	"html/template"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/jeymob/rural-portal/internal/config"
	"github.com/jeymob/rural-portal/internal/delivery/http/routes"
)

// RunServer настроит роутер, загрузит шаблоны и запустит HTTP сервер с graceful shutdown.
func RunServer(cfg *config.Config, db *gorm.DB, sqlDB *sql.DB) error {
	r := gin.Default()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	// Проверяем и загружаем шаблоны
	wd, _ := os.Getwd()
	log.Printf("📂 Рабочая директория: %s", wd)

	templatePath := "internal/delivery/http/templates"
	if _, err := os.Stat(templatePath); os.IsNotExist(err) {
		log.Printf("⚠️  Папка с шаблонами НЕ НАЙДЕНА: %s (продолжаем без HTML-шаблонов)", templatePath)
	} else {
		indexPath := filepath.Join(templatePath, "index.html")
		if _, err := os.Stat(indexPath); os.IsNotExist(err) {
			log.Printf("⚠️  index.html НЕ НАЙДЕН: %s", indexPath)
		} else {
			log.Println("✅ index.html найден")
			templates := template.Must(template.ParseGlob("internal/delivery/http/templates/*.html"))
			templates = template.Must(templates.ParseGlob("internal/delivery/http/templates/partials/*.html"))
			r.SetHTMLTemplate(templates)
			log.Println("✅ Шаблоны успешно загружены")
		}
	}

	// Настройка CORS — разрешаем запросы с фронта
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173", "http://127.0.0.1:5173"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Подключаем дополнительные роуты
	routes.SetupRoutes(r, db, cfg)

	r.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "Страница не найдена",
			"path":    c.Request.URL.Path,
			"method":  c.Request.Method,
			"suggest": "Попробуй /health или /",
		})
	})

	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      r,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Запускаем сервер
	go func() {
		log.Printf("🚀 Сервер запущен на http://localhost:%s", cfg.Port)
		log.Printf("📊 Health-check: http://localhost:%s/health", cfg.Port)
		log.Printf("🏠 Главная: http://localhost:%s/", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("💥 Ошибка запуска сервера: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("🛑 Получен сигнал завершения (Ctrl+C). Останавливаем...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		return err
	}

	if sqlDB != nil {
		sqlDB.Close()
	}
	log.Println("✅ Сервер остановлен корректно")
	return nil
}

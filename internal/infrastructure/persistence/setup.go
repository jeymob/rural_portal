package persistence

import (
	"database/sql"
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/jeymob/rural-portal/internal/config"
	"github.com/jeymob/rural-portal/internal/infrastructure/persistence/models"
)

// InitDB открывает соединение к базе, выполняет automigrate и создаёт первого администратора при необходимости.
func InitDB(cfg *config.Config) (*gorm.DB, *sql.DB, error) {
	db, err := gorm.Open(postgres.Open(cfg.DatabaseURL), &gorm.Config{})
	if err != nil {
		return nil, nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, nil, err
	}
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)

	if err := db.AutoMigrate(&models.User{}); err != nil {
		return nil, nil, err
	}
	log.Println("✅ Миграции выполнены")

	var adminCount int64
	db.Model(&models.User{}).Where("role = ?", "admin").Count(&adminCount)
	if adminCount == 0 {
		admin := &models.User{
			Username: "admin",
			Email:    "admin@rural.local",
			Password: "admin123",
			Role:     "admin",
			Region:   "Вся Россия",
			Phone:    "+7 (999) 123-45-67",
		}
		if err := db.Create(admin).Error; err != nil {
			return nil, nil, err
		}
		log.Println("✅ Создан первый администратор: login=admin, pass=admin123")
		log.Println("⚠️  ВРЕМЕННО пароль в открытом виде! Добавь bcrypt!")
	} else {
		log.Println("ℹ️  Администратор уже существует")
	}

	return db, sqlDB, nil
}

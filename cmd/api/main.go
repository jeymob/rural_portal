package main

import (
	"log"

	"github.com/jeymob/rural-portal/internal/config"
	"github.com/jeymob/rural-portal/internal/infrastructure/persistence"
)

func main() {
	cfg := config.Load()
	log.Printf("Конфигурация загружена: PORT=%s", cfg.Port)

	db, sqlDB, err := persistence.InitDB(cfg)
	if err != nil {
		log.Fatalf("Не удалось инициализировать БД: %v", err)
	}

	if err := RunServer(cfg, db, sqlDB); err != nil {
		log.Fatalf("Сервер завершился с ошибкой: %v", err)
	}
}

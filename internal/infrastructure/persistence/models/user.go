package models

import "gorm.io/gorm"

type User struct {
	gorm.Model
	YandexID string `gorm:"size:64;index"`     // индекс для быстрого поиска, НЕ unique
	VkID     string `gorm:"size:64;index"`     // индекс для быстрого поиска, НЕ unique
	Name     string `gorm:"size:100;not null"` // имя, которое отображается на сайте
	Region   string `gorm:"size:200"`
	Role     string `gorm:"size:20;default:'user'"`
}

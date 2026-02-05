package models

import "gorm.io/gorm"

type User struct {
	gorm.Model
	YandexID string `gorm:"uniqueIndex;size:64"` // uniqueIndex — создаёт индекс, а не constraint unique на уровне БД
	VkID     string `gorm:"uniqueIndex;size:64"`
	Name     string `gorm:"unique;not null;size:100"`
	Role     string `gorm:"default:'user';size:20"`
	Region   string `gorm:"size:200"`
}

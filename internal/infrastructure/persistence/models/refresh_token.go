package models

import (
	"time"

	"gorm.io/gorm"
)

type RefreshToken struct {
	gorm.Model
	UserID    uint      `gorm:"index;not null"`
	Token     string    `gorm:"unique;size:255;not null"` // сам токен или его хэш
	ExpiresAt time.Time `gorm:"index;not null"`
	Revoked   bool      `gorm:"default:false"`
}

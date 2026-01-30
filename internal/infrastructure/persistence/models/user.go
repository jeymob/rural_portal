package models

import "gorm.io/gorm"

type User struct {
	gorm.Model
	Username string `gorm:"unique;not null;size:100"`
	Email    string `gorm:"unique;not null;size:150"`
	Password string `gorm:"not null;size:255"`    // bcrypt hash
	Phone    string `gorm:"size:30;default:null"` // ← добавлено, nullable
	Role     string `gorm:"default:'user';size:20"`
	Region   string `gorm:"size:200"`
}

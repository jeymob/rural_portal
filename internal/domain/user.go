package domain

import "time"

type UserRole string

const (
	RoleUser  UserRole = "user"
	RoleAdmin UserRole = "admin"
)

type User struct {
	ID        uint
	Username  string
	Email     string
	Password  string // hashed
	Role      UserRole
	Region    string // область/район/село
	CreatedAt time.Time
	UpdatedAt time.Time
}

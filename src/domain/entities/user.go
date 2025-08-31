package entities

import (
	"time"
)

type UserRole string

const (
	RoleViewer UserRole = "viewer"
	RoleAdmin  UserRole = "admin"
)

type User struct {
	ID        string    `json:"id" gorm:"primaryKey;type:uuid;default:uuid_generate_v4()"`
	Username  string    `json:"username" gorm:"uniqueIndex;not null"`
	Email     string    `json:"email" gorm:"uniqueIndex;not null"`
	Password  string    `json:"-" gorm:"not null"`
	Role      UserRole  `json:"role" gorm:"default:viewer"`
	Active    bool      `json:"active" gorm:"default:true"`
	LastLogin *time.Time `json:"lastLogin"`
	CreatedAt time.Time `json:"createdAt" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updatedAt" gorm:"autoUpdateTime"`
}

func (User) TableName() string {
	return "users"
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type LoginResponse struct {
	Token        string    `json:"token"`
	RefreshToken string    `json:"refreshToken"`
	ExpiresAt    time.Time `json:"expiresAt"`
	User         *User     `json:"user"`
}

type JWTClaims struct {
	UserID   string   `json:"user_id"`
	Username string   `json:"username"`
	Role     UserRole `json:"role"`
	Exp      int64    `json:"exp"`
	Iat      int64    `json:"iat"`
}

func (c JWTClaims) Valid() error {
	return nil
}
package interfaces

import (
	"context"

	"api/domain/entities"
)

type AuthService interface {
	Login(ctx context.Context, request *entities.LoginRequest) (*entities.LoginResponse, error)
	ValidateToken(ctx context.Context, token string) (*entities.JWTClaims, error)
	RefreshToken(ctx context.Context, refreshToken string) (*entities.LoginResponse, error)
	HashPassword(password string) (string, error)
	VerifyPassword(hashedPassword, password string) bool
	GenerateTokens(ctx context.Context, user *entities.User) (accessToken, refreshToken string, expiresAt int64, error error)
}
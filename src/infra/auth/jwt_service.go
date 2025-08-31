package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"api/domain/entities"
	"api/domain/interfaces"
	"api/infra/config"

	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"
)

type jwtService struct {
	userRepo   interfaces.UserRepository
	jwtSecret  string
	expiration time.Duration
}

func NewJWTService(userRepo interfaces.UserRepository, cfg *config.Config) interfaces.AuthService {
	return &jwtService{
		userRepo:   userRepo,
		jwtSecret:  cfg.JWT.Secret,
		expiration: cfg.JWT.AccessTokenExpiry,
	}
}

func (s *jwtService) Login(ctx context.Context, request *entities.LoginRequest) (*entities.LoginResponse, error) {
	user, err := s.userRepo.GetByEmail(ctx, request.Email)
	if err != nil {
		return nil, errors.New("invalid credentials")
	}

	if !user.Active {
		return nil, errors.New("account is inactive")
	}

	if !s.VerifyPassword(user.Password, request.Password) {
		return nil, errors.New("invalid credentials")
	}

	accessToken, refreshToken, expiresAt, err := s.GenerateTokens(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	err = s.userRepo.UpdateLastLogin(ctx, user.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to update last login: %w", err)
	}

	user.Password = ""

	return &entities.LoginResponse{
		Token:        accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    time.Unix(expiresAt, 0),
		User:         user,
	}, nil
}

func (s *jwtService) ValidateToken(ctx context.Context, tokenString string) (*entities.JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &entities.JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.jwtSecret), nil
	})

	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	claims, ok := token.Claims.(*entities.JWTClaims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token claims")
	}

	if claims.Exp < time.Now().Unix() {
		return nil, errors.New("token has expired")
	}

	return claims, nil
}

func (s *jwtService) RefreshToken(ctx context.Context, refreshToken string) (*entities.LoginResponse, error) {
	claims, err := s.ValidateToken(ctx, refreshToken)
	if err != nil {
		return nil, fmt.Errorf("invalid refresh token: %w", err)
	}

	user, err := s.userRepo.GetByID(ctx, claims.UserID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	if !user.Active {
		return nil, errors.New("account is inactive")
	}

	accessToken, newRefreshToken, expiresAt, err := s.GenerateTokens(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	user.Password = ""

	return &entities.LoginResponse{
		Token:        accessToken,
		RefreshToken: newRefreshToken,
		ExpiresAt:    time.Unix(expiresAt, 0),
		User:         user,
	}, nil
}

func (s *jwtService) HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

func (s *jwtService) VerifyPassword(hashedPassword, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}

func (s *jwtService) GenerateTokens(ctx context.Context, user *entities.User) (accessToken, refreshToken string, expiresAt int64, err error) {
	now := time.Now()
	exp := now.Add(s.expiration)
	refreshExp := now.Add(7 * 24 * time.Hour) // 7 days for refresh token

	accessClaims := &entities.JWTClaims{
		UserID:   user.ID,
		Username: user.Username,
		Role:     user.Role,
		Exp:      exp.Unix(),
		Iat:      now.Unix(),
	}

	refreshClaims := &entities.JWTClaims{
		UserID:   user.ID,
		Username: user.Username,
		Role:     user.Role,
		Exp:      refreshExp.Unix(),
		Iat:      now.Unix(),
	}

	accessToken, err = s.generateToken(accessClaims)
	if err != nil {
		return "", "", 0, err
	}

	refreshToken, err = s.generateToken(refreshClaims)
	if err != nil {
		return "", "", 0, err
	}

	return accessToken, refreshToken, exp.Unix(), nil
}

func (s *jwtService) generateToken(claims *entities.JWTClaims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtSecret))
}
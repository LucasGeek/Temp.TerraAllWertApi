package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"terra-allwert/domain/entities"
)

// JWTService handles JWT token operations
type JWTService struct {
	secretKey             string
	accessTokenDuration   time.Duration
	refreshTokenDuration  time.Duration
}

// TokenPair represents access and refresh tokens
type TokenPair struct {
	AccessToken           string    `json:"access_token"`
	RefreshToken          string    `json:"refresh_token"`
	AccessTokenExpiresAt  time.Time `json:"access_token_expires_at"`
	RefreshTokenExpiresAt time.Time `json:"refresh_token_expires_at"`
	TokenType             string    `json:"token_type"`
}

// Claims represents JWT token claims
type Claims struct {
	UserID       uuid.UUID          `json:"user_id"`
	Email        string             `json:"email"`
	Role         entities.UserRole  `json:"role"`
	EnterpriseID uuid.UUID          `json:"enterprise_id"`
	TokenType    string             `json:"token_type"` // "access" or "refresh"
	jwt.RegisteredClaims
}

// RefreshTokenClaims represents refresh token specific claims
type RefreshTokenClaims struct {
	UserID    uuid.UUID `json:"user_id"`
	TokenID   string    `json:"token_id"` // Unique token identifier
	TokenType string    `json:"token_type"`
	jwt.RegisteredClaims
}

// NewJWTService creates a new JWT service
func NewJWTService(secretKey string, accessTokenHours, refreshTokenHours int) *JWTService {
	return &JWTService{
		secretKey:             secretKey,
		accessTokenDuration:   time.Duration(accessTokenHours) * time.Hour,
		refreshTokenDuration:  time.Duration(refreshTokenHours) * time.Hour,
	}
}

// GenerateTokenPair creates both access and refresh tokens
func (j *JWTService) GenerateTokenPair(user *entities.User) (*TokenPair, error) {
	now := time.Now()
	tokenID := uuid.New().String()

	// Generate access token
	accessToken, accessExpiresAt, err := j.generateAccessToken(user, now)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	// Generate refresh token
	refreshToken, refreshExpiresAt, err := j.generateRefreshToken(user.ID, tokenID, now)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	return &TokenPair{
		AccessToken:           accessToken,
		RefreshToken:          refreshToken,
		AccessTokenExpiresAt:  accessExpiresAt,
		RefreshTokenExpiresAt: refreshExpiresAt,
		TokenType:             "Bearer",
	}, nil
}

// generateAccessToken creates a new access token
func (j *JWTService) generateAccessToken(user *entities.User, now time.Time) (string, time.Time, error) {
	expiresAt := now.Add(j.accessTokenDuration)

	claims := &Claims{
		UserID:    user.ID,
		Email:     user.Email,
		Role:      user.Role,
		TokenType: "access",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "terra-allwert-api",
			Subject:   user.ID.String(),
		},
	}

	// Add enterprise ID
	claims.EnterpriseID = user.EnterpriseID

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(j.secretKey))
	if err != nil {
		return "", time.Time{}, err
	}

	return tokenString, expiresAt, nil
}

// generateRefreshToken creates a new refresh token
func (j *JWTService) generateRefreshToken(userID uuid.UUID, tokenID string, now time.Time) (string, time.Time, error) {
	expiresAt := now.Add(j.refreshTokenDuration)

	claims := &RefreshTokenClaims{
		UserID:    userID,
		TokenID:   tokenID,
		TokenType: "refresh",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "terra-allwert-api",
			Subject:   userID.String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(j.secretKey))
	if err != nil {
		return "", time.Time{}, err
	}

	return tokenString, expiresAt, nil
}

// ValidateAccessToken validates and parses an access token
func (j *JWTService) ValidateAccessToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(j.secretKey), nil
	})

	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	if !token.Valid {
		return nil, fmt.Errorf("token is not valid")
	}

	claims, ok := token.Claims.(*Claims)
	if !ok {
		return nil, fmt.Errorf("invalid token claims")
	}

	if claims.TokenType != "access" {
		return nil, fmt.Errorf("token is not an access token")
	}

	return claims, nil
}

// ValidateRefreshToken validates and parses a refresh token
func (j *JWTService) ValidateRefreshToken(tokenString string) (*RefreshTokenClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &RefreshTokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(j.secretKey), nil
	})

	if err != nil {
		return nil, fmt.Errorf("invalid refresh token: %w", err)
	}

	if !token.Valid {
		return nil, fmt.Errorf("refresh token is not valid")
	}

	claims, ok := token.Claims.(*RefreshTokenClaims)
	if !ok {
		return nil, fmt.Errorf("invalid refresh token claims")
	}

	if claims.TokenType != "refresh" {
		return nil, fmt.Errorf("token is not a refresh token")
	}

	return claims, nil
}

// RefreshAccessToken generates a new access token using a valid refresh token
func (j *JWTService) RefreshAccessToken(refreshTokenString string, getUserByID func(uuid.UUID) (*entities.User, error)) (*TokenPair, error) {
	// Validate refresh token
	refreshClaims, err := j.ValidateRefreshToken(refreshTokenString)
	if err != nil {
		return nil, fmt.Errorf("invalid refresh token: %w", err)
	}

	// Get user data
	user, err := getUserByID(refreshClaims.UserID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	// Generate new token pair
	tokenPair, err := j.GenerateTokenPair(user)
	if err != nil {
		return nil, fmt.Errorf("failed to generate new tokens: %w", err)
	}

	return tokenPair, nil
}

// ExtractTokenFromAuthHeader extracts token from Authorization header
func ExtractTokenFromAuthHeader(authHeader string) (string, error) {
	if authHeader == "" {
		return "", fmt.Errorf("authorization header is empty")
	}

	const bearerPrefix = "Bearer "
	if len(authHeader) < len(bearerPrefix) || authHeader[:len(bearerPrefix)] != bearerPrefix {
		return "", fmt.Errorf("invalid authorization header format")
	}

	return authHeader[len(bearerPrefix):], nil
}

// TokenBlacklist interface for managing revoked tokens
type TokenBlacklist interface {
	AddToken(tokenID string, expiresAt time.Time) error
	IsTokenBlacklisted(tokenID string) (bool, error)
	CleanupExpiredTokens() error
}

// RevokeToken adds a token to the blacklist
func (j *JWTService) RevokeToken(tokenString string, blacklist TokenBlacklist) error {
	// Parse token to get claims without validation (token might be expired)
	token, _, err := new(jwt.Parser).ParseUnverified(tokenString, &Claims{})
	if err != nil {
		return fmt.Errorf("failed to parse token: %w", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok {
		return fmt.Errorf("invalid token claims")
	}

	// Add to blacklist with expiration time
	return blacklist.AddToken(claims.ID, claims.ExpiresAt.Time)
}

// IsTokenRevoked checks if a token is in the blacklist
func (j *JWTService) IsTokenRevoked(tokenID string, blacklist TokenBlacklist) (bool, error) {
	return blacklist.IsTokenBlacklisted(tokenID)
}
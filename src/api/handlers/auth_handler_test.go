package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http/httptest"
	"testing"

	"api/domain/entities"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockAuthService struct {
	mock.Mock
}

func (m *MockAuthService) Login(ctx context.Context, request *entities.LoginRequest) (*entities.LoginResponse, error) {
	args := m.Called(ctx, request)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.LoginResponse), args.Error(1)
}

func (m *MockAuthService) ValidateToken(ctx context.Context, token string) (*entities.JWTClaims, error) {
	args := m.Called(ctx, token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.JWTClaims), args.Error(1)
}

func (m *MockAuthService) RefreshToken(ctx context.Context, refreshToken string) (*entities.LoginResponse, error) {
	args := m.Called(ctx, refreshToken)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.LoginResponse), args.Error(1)
}

func (m *MockAuthService) HashPassword(password string) (string, error) {
	args := m.Called(password)
	return args.String(0), args.Error(1)
}

func (m *MockAuthService) VerifyPassword(hashedPassword, password string) bool {
	args := m.Called(hashedPassword, password)
	return args.Bool(0)
}

func (m *MockAuthService) GenerateTokens(ctx context.Context, user *entities.User) (accessToken, refreshToken string, expiresAt int64, err error) {
	args := m.Called(ctx, user)
	return args.String(0), args.String(1), args.Get(2).(int64), args.Error(3)
}

func TestAuthHandler_Login_Success(t *testing.T) {
	mockAuthService := new(MockAuthService)
	handler := NewAuthHandler(mockAuthService)

	app := fiber.New()
	app.Post("/login", handler.Login)

	loginRequest := entities.LoginRequest{
		Username: "testuser",
		Password: "testpass",
	}

	loginResponse := &entities.LoginResponse{
		Token:        "test-token",
		RefreshToken: "test-refresh-token",
		User: &entities.User{
			ID:       "123",
			Username: "testuser",
			Role:     entities.RoleViewer,
		},
	}

	mockAuthService.On("Login", mock.Anything, &loginRequest).Return(loginResponse, nil)

	body, _ := json.Marshal(loginRequest)
	req := httptest.NewRequest("POST", "/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var response entities.LoginResponse
	err = json.NewDecoder(resp.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Equal(t, "test-token", response.Token)
	assert.Equal(t, "test-refresh-token", response.RefreshToken)
	assert.Equal(t, "testuser", response.User.Username)

	mockAuthService.AssertExpectations(t)
}

func TestAuthHandler_Login_InvalidBody(t *testing.T) {
	mockAuthService := new(MockAuthService)
	handler := NewAuthHandler(mockAuthService)

	app := fiber.New()
	app.Post("/login", handler.Login)

	req := httptest.NewRequest("POST", "/login", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode)

	var errorResponse map[string]string
	err = json.NewDecoder(resp.Body).Decode(&errorResponse)
	assert.NoError(t, err)
	assert.Equal(t, "Invalid request body", errorResponse["error"])
}

func TestAuthHandler_Login_ValidationError(t *testing.T) {
	mockAuthService := new(MockAuthService)
	handler := NewAuthHandler(mockAuthService)

	app := fiber.New()
	app.Post("/login", handler.Login)

	loginRequest := entities.LoginRequest{
		Username: "", // Missing required field
		Password: "testpass",
	}

	body, _ := json.Marshal(loginRequest)
	req := httptest.NewRequest("POST", "/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode)

	var errorResponse map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&errorResponse)
	assert.NoError(t, err)
	assert.Equal(t, "Validation failed", errorResponse["error"])
}

func TestAuthHandler_RefreshToken_Success(t *testing.T) {
	mockAuthService := new(MockAuthService)
	handler := NewAuthHandler(mockAuthService)

	app := fiber.New()
	app.Post("/refresh", handler.RefreshToken)

	refreshRequest := map[string]string{
		"refreshToken": "test-refresh-token",
	}

	refreshResponse := &entities.LoginResponse{
		Token:        "new-test-token",
		RefreshToken: "new-test-refresh-token",
		User: &entities.User{
			ID:       "123",
			Username: "testuser",
			Role:     entities.RoleViewer,
		},
	}

	mockAuthService.On("RefreshToken", mock.Anything, "test-refresh-token").Return(refreshResponse, nil)

	body, _ := json.Marshal(refreshRequest)
	req := httptest.NewRequest("POST", "/refresh", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var response entities.LoginResponse
	err = json.NewDecoder(resp.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Equal(t, "new-test-token", response.Token)
	assert.Equal(t, "new-test-refresh-token", response.RefreshToken)

	mockAuthService.AssertExpectations(t)
}
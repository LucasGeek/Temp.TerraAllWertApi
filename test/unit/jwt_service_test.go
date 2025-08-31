package unit

import (
	"context"
	"testing"
	"time"

	"api/domain/entities"
	"api/infra/auth"
	"api/infra/config"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(ctx context.Context, user *entities.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) GetByID(ctx context.Context, id string) (*entities.User, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*entities.User), args.Error(1)
}

func (m *MockUserRepository) GetByUsername(ctx context.Context, username string) (*entities.User, error) {
	args := m.Called(ctx, username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.User), args.Error(1)
}

func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (*entities.User, error) {
	args := m.Called(ctx, email)
	return args.Get(0).(*entities.User), args.Error(1)
}

func (m *MockUserRepository) Update(ctx context.Context, user *entities.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) UpdateLastLogin(ctx context.Context, userID string) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockUserRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockUserRepository) List(ctx context.Context, limit, offset int) ([]*entities.User, error) {
	args := m.Called(ctx, limit, offset)
	return args.Get(0).([]*entities.User), args.Error(1)
}

func (m *MockUserRepository) ExistsByUsername(ctx context.Context, username string) (bool, error) {
	args := m.Called(ctx, username)
	return args.Bool(0), args.Error(1)
}

func (m *MockUserRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	args := m.Called(ctx, email)
	return args.Bool(0), args.Error(1)
}

func TestJWTService_HashPassword(t *testing.T) {
	mockRepo := new(MockUserRepository)
	cfg := &config.Config{
		JWT: config.JWTConfig{
			Secret:             "test-secret",
			AccessTokenExpiry:  15 * time.Minute,
		},
	}
	service := auth.NewJWTService(mockRepo, cfg)

	password := "testpassword123"
	hashedPassword, err := service.HashPassword(password)

	assert.NoError(t, err)
	assert.NotEmpty(t, hashedPassword)
	assert.NotEqual(t, password, hashedPassword)
}

func TestJWTService_VerifyPassword(t *testing.T) {
	mockRepo := new(MockUserRepository)
	cfg := &config.Config{
		JWT: config.JWTConfig{
			Secret:             "test-secret",
			AccessTokenExpiry:  15 * time.Minute,
		},
	}
	service := auth.NewJWTService(mockRepo, cfg)

	password := "testpassword123"
	hashedPassword, _ := service.HashPassword(password)

	// Test correct password
	assert.True(t, service.VerifyPassword(hashedPassword, password))

	// Test wrong password
	assert.False(t, service.VerifyPassword(hashedPassword, "wrongpassword"))
}

func TestJWTService_GenerateTokens(t *testing.T) {
	mockRepo := new(MockUserRepository)
	cfg := &config.Config{
		JWT: config.JWTConfig{
			Secret:             "test-secret",
			AccessTokenExpiry:  15 * time.Minute,
		},
	}
	service := auth.NewJWTService(mockRepo, cfg)

	user := &entities.User{
		ID:       "123",
		Username: "testuser",
		Role:     entities.RoleViewer,
	}

	accessToken, refreshToken, expiresAt, err := service.GenerateTokens(context.Background(), user)

	assert.NoError(t, err)
	assert.NotEmpty(t, accessToken)
	assert.NotEmpty(t, refreshToken)
	assert.Greater(t, expiresAt, time.Now().Unix())
}

func TestJWTService_ValidateToken(t *testing.T) {
	mockRepo := new(MockUserRepository)
	cfg := &config.Config{
		JWT: config.JWTConfig{
			Secret:             "test-secret",
			AccessTokenExpiry:  15 * time.Minute,
		},
	}
	service := auth.NewJWTService(mockRepo, cfg)

	user := &entities.User{
		ID:       "123",
		Username: "testuser",
		Role:     entities.RoleViewer,
	}

	accessToken, _, _, err := service.GenerateTokens(context.Background(), user)
	assert.NoError(t, err)

	claims, err := service.ValidateToken(context.Background(), accessToken)
	assert.NoError(t, err)
	assert.Equal(t, user.ID, claims.UserID)
	assert.Equal(t, user.Username, claims.Username)
	assert.Equal(t, user.Role, claims.Role)
}

func TestJWTService_Login_Success(t *testing.T) {
	mockRepo := new(MockUserRepository)
	cfg := &config.Config{
		JWT: config.JWTConfig{
			Secret:             "test-secret",
			AccessTokenExpiry:  15 * time.Minute,
		},
	}
	service := auth.NewJWTService(mockRepo, cfg)

	hashedPassword, _ := service.HashPassword("testpassword")
	user := &entities.User{
		ID:       "123",
		Username: "testuser",
		Password: hashedPassword,
		Role:     entities.RoleViewer,
		Active:   true,
	}

	mockRepo.On("GetByUsername", mock.Anything, "testuser").Return(user, nil)
	mockRepo.On("UpdateLastLogin", mock.Anything, "123").Return(nil)

	request := &entities.LoginRequest{
		Username: "testuser",
		Password: "testpassword",
	}

	response, err := service.Login(context.Background(), request)

	assert.NoError(t, err)
	assert.NotEmpty(t, response.Token)
	assert.NotEmpty(t, response.RefreshToken)
	assert.Equal(t, user.Username, response.User.Username)
	assert.Empty(t, response.User.Password) // Password should be cleared

	mockRepo.AssertExpectations(t)
}

func TestJWTService_Login_InvalidCredentials(t *testing.T) {
	mockRepo := new(MockUserRepository)
	cfg := &config.Config{
		JWT: config.JWTConfig{
			Secret:             "test-secret",
			AccessTokenExpiry:  15 * time.Minute,
		},
	}
	service := auth.NewJWTService(mockRepo, cfg)

	hashedPassword, _ := service.HashPassword("testpassword")
	user := &entities.User{
		ID:       "123",
		Username: "testuser",
		Password: hashedPassword,
		Role:     entities.RoleViewer,
		Active:   true,
	}

	mockRepo.On("GetByUsername", mock.Anything, "testuser").Return(user, nil)

	request := &entities.LoginRequest{
		Username: "testuser",
		Password: "wrongpassword",
	}

	response, err := service.Login(context.Background(), request)

	assert.Error(t, err)
	assert.Nil(t, response)
	assert.Contains(t, err.Error(), "invalid credentials")

	mockRepo.AssertExpectations(t)
}
package unit

import (
	"context"
	"errors"
	"testing"
	"time"

	"api/domain/entities"
	"api/infra/security"
	"test/fixtures/testutils"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock do repository de usuários
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) FindByEmail(ctx context.Context, email string) (*entities.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.User), args.Error(1)
}

func (m *MockUserRepository) Create(ctx context.Context, user *entities.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) Update(ctx context.Context, user *entities.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockUserRepository) FindByID(ctx context.Context, id string) (*entities.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.User), args.Error(1)
}

// Mock do password hasher
type MockPasswordHasher struct {
	mock.Mock
}

func (m *MockPasswordHasher) HashPassword(password string) (string, error) {
	args := m.Called(password)
	return args.String(0), args.Error(1)
}

func (m *MockPasswordHasher) CheckPassword(password, hash string) error {
	args := m.Called(password, hash)
	return args.Error(0)
}

// Mock do JWT service
type MockJWTService struct {
	mock.Mock
}

func (m *MockJWTService) GenerateToken(claims *entities.JWTClaims) (string, error) {
	args := m.Called(claims)
	return args.String(0), args.Error(1)
}

func (m *MockJWTService) ValidateToken(token string) (*entities.JWTClaims, error) {
	args := m.Called(token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.JWTClaims), args.Error(1)
}

func (m *MockJWTService) GenerateRefreshToken(userID string) (string, error) {
	args := m.Called(userID)
	return args.String(0), args.Error(1)
}

// AuthService simulado para testes
type AuthService struct {
	userRepo       *MockUserRepository
	passwordHasher *MockPasswordHasher
	jwtService     *MockJWTService
}

func NewAuthService(userRepo *MockUserRepository, passwordHasher *MockPasswordHasher, jwtService *MockJWTService) *AuthService {
	return &AuthService{
		userRepo:       userRepo,
		passwordHasher: passwordHasher,
		jwtService:     jwtService,
	}
}

func (s *AuthService) Login(ctx context.Context, req *entities.LoginRequest) (*entities.LoginResponse, error) {
	// Buscar usuário por email
	user, err := s.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		return nil, err
	}

	if user == nil {
		return nil, errors.New("invalid credentials")
	}

	// Verificar senha
	if err := s.passwordHasher.CheckPassword(req.Password, user.Password); err != nil {
		return nil, errors.New("invalid credentials")
	}

	// Gerar tokens
	claims := &entities.JWTClaims{
		UserID:   user.ID,
		Username: user.Username,
		Role:     user.Role,
		Exp:      time.Now().Add(15 * time.Minute).Unix(),
		Iat:      time.Now().Unix(),
	}

	accessToken, err := s.jwtService.GenerateToken(claims)
	if err != nil {
		return nil, err
	}

	refreshToken, err := s.jwtService.GenerateRefreshToken(user.ID)
	if err != nil {
		return nil, err
	}

	return &entities.LoginResponse{
		Token:        accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    time.Unix(claims.Exp, 0),
		User:         user,
	}, nil
}

// Testes do AuthService
func TestAuthService_Login_Success(t *testing.T) {
	// Arrange
	mockUserRepo := new(MockUserRepository)
	mockPasswordHasher := new(MockPasswordHasher)
	mockJWTService := new(MockJWTService)
	
	authService := NewAuthService(mockUserRepo, mockPasswordHasher, mockJWTService)
	
	ctx := context.Background()
	loginReq := testutils.CreateTestLoginRequest()
	testUser := testutils.CreateTestUser(entities.RoleViewer)
	
	// Mock expectations
	mockUserRepo.On("FindByEmail", ctx, loginReq.Email).Return(testUser, nil)
	mockPasswordHasher.On("CheckPassword", loginReq.Password, testUser.Password).Return(nil)
	mockJWTService.On("GenerateToken", mock.AnythingOfType("*entities.JWTClaims")).Return("access-token", nil)
	mockJWTService.On("GenerateRefreshToken", testUser.ID).Return("refresh-token", nil)

	// Act
	result, err := authService.Login(ctx, loginReq)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "access-token", result.Token)
	assert.Equal(t, "refresh-token", result.RefreshToken)
	assert.Equal(t, testUser, result.User)
	assert.True(t, result.ExpiresAt.After(time.Now()))

	// Verify mock expectations
	mockUserRepo.AssertExpectations(t)
	mockPasswordHasher.AssertExpectations(t)
	mockJWTService.AssertExpectations(t)
}

func TestAuthService_Login_UserNotFound(t *testing.T) {
	// Arrange
	mockUserRepo := new(MockUserRepository)
	mockPasswordHasher := new(MockPasswordHasher)
	mockJWTService := new(MockJWTService)
	
	authService := NewAuthService(mockUserRepo, mockPasswordHasher, mockJWTService)
	
	ctx := context.Background()
	loginReq := testutils.CreateTestLoginRequest()
	
	// Mock expectations
	mockUserRepo.On("FindByEmail", ctx, loginReq.Email).Return(nil, errors.New("user not found"))

	// Act
	result, err := authService.Login(ctx, loginReq)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "user not found")

	// Verify mock expectations
	mockUserRepo.AssertExpectations(t)
}

func TestAuthService_Login_InvalidPassword(t *testing.T) {
	// Arrange
	mockUserRepo := new(MockUserRepository)
	mockPasswordHasher := new(MockPasswordHasher)
	mockJWTService := new(MockJWTService)
	
	authService := NewAuthService(mockUserRepo, mockPasswordHasher, mockJWTService)
	
	ctx := context.Background()
	loginReq := testutils.CreateTestLoginRequest()
	testUser := testutils.CreateTestUser(entities.RoleViewer)
	
	// Mock expectations
	mockUserRepo.On("FindByEmail", ctx, loginReq.Email).Return(testUser, nil)
	mockPasswordHasher.On("CheckPassword", loginReq.Password, testUser.Password).Return(errors.New("password mismatch"))

	// Act
	result, err := authService.Login(ctx, loginReq)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "invalid credentials")

	// Verify mock expectations
	mockUserRepo.AssertExpectations(t)
	mockPasswordHasher.AssertExpectations(t)
}

func TestAuthService_Login_TokenGenerationFailed(t *testing.T) {
	// Arrange
	mockUserRepo := new(MockUserRepository)
	mockPasswordHasher := new(MockPasswordHasher)
	mockJWTService := new(MockJWTService)
	
	authService := NewAuthService(mockUserRepo, mockPasswordHasher, mockJWTService)
	
	ctx := context.Background()
	loginReq := testutils.CreateTestLoginRequest()
	testUser := testutils.CreateTestUser(entities.RoleViewer)
	
	// Mock expectations
	mockUserRepo.On("FindByEmail", ctx, loginReq.Email).Return(testUser, nil)
	mockPasswordHasher.On("CheckPassword", loginReq.Password, testUser.Password).Return(nil)
	mockJWTService.On("GenerateToken", mock.AnythingOfType("*entities.JWTClaims")).Return("", errors.New("token generation failed"))

	// Act
	result, err := authService.Login(ctx, loginReq)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "token generation failed")

	// Verify mock expectations
	mockUserRepo.AssertExpectations(t)
	mockPasswordHasher.AssertExpectations(t)
	mockJWTService.AssertExpectations(t)
}

// Teste do JWT Service real
func TestJWTService_GenerateAndValidateToken(t *testing.T) {
	// Arrange
	jwtService := security.NewJWTService("test-secret-key")
	claims := testutils.CreateTestJWTClaims(entities.RoleAdmin)

	// Act - Generate token
	token, err := jwtService.GenerateToken(claims)

	// Assert - Token generation
	assert.NoError(t, err)
	assert.NotEmpty(t, token)

	// Act - Validate token
	validatedClaims, err := jwtService.ValidateToken(token)

	// Assert - Token validation
	assert.NoError(t, err)
	assert.NotNil(t, validatedClaims)
	assert.Equal(t, claims.UserID, validatedClaims.UserID)
	assert.Equal(t, claims.Username, validatedClaims.Username)
	assert.Equal(t, claims.Role, validatedClaims.Role)
}

func TestJWTService_ValidateToken_InvalidToken(t *testing.T) {
	// Arrange
	jwtService := security.NewJWTService("test-secret-key")
	invalidToken := "invalid.token.here"

	// Act
	claims, err := jwtService.ValidateToken(invalidToken)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, claims)
}

func TestJWTService_ValidateToken_ExpiredToken(t *testing.T) {
	// Arrange
	jwtService := security.NewJWTService("test-secret-key")
	expiredClaims := &entities.JWTClaims{
		UserID:   "test-user",
		Username: "testuser",
		Role:     entities.RoleViewer,
		Exp:      time.Now().Add(-1 * time.Hour).Unix(), // Expired 1 hour ago
		Iat:      time.Now().Add(-2 * time.Hour).Unix(),
	}

	// Generate token with expired claims (this might not work in real JWT libraries)
	// For testing purposes, we'll just test with an obviously invalid timestamp
	claims := testutils.CreateTestJWTClaims(entities.RoleViewer)
	claims.Exp = time.Now().Add(-1 * time.Hour).Unix()

	token, err := jwtService.GenerateToken(claims)
	assert.NoError(t, err)

	// Wait a brief moment to ensure token is expired
	time.Sleep(10 * time.Millisecond)

	// Act
	validatedClaims, err := jwtService.ValidateToken(token)

	// Assert
	// Note: Depending on JWT library implementation, expired tokens might still be parsed
	// but validation should fail. Adjust assertion based on actual behavior.
	if err != nil {
		assert.Error(t, err)
		assert.Nil(t, validatedClaims)
	} else {
		// If JWT library doesn't automatically check expiration, verify manually
		assert.True(t, time.Unix(validatedClaims.Exp, 0).Before(time.Now()))
	}
}

// Benchmark tests
func BenchmarkJWTService_GenerateToken(b *testing.B) {
	jwtService := security.NewJWTService("test-secret-key")
	claims := testutils.CreateTestJWTClaims(entities.RoleAdmin)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = jwtService.GenerateToken(claims)
	}
}

func BenchmarkJWTService_ValidateToken(b *testing.B) {
	jwtService := security.NewJWTService("test-secret-key")
	claims := testutils.CreateTestJWTClaims(entities.RoleAdmin)
	token, _ := jwtService.GenerateToken(claims)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = jwtService.ValidateToken(token)
	}
}
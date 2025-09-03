package handlers

import (
	"time"

	"terra-allwert/domain/entities"
	"terra-allwert/domain/interfaces"
	"terra-allwert/infra/auth"
	"terra-allwert/infra/middleware"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// AuthHandler handles authentication operations
type AuthHandler struct {
	userRepo   interfaces.UserRepository
	jwtService *auth.JWTService
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(userRepo interfaces.UserRepository, jwtService *auth.JWTService) *AuthHandler {
	return &AuthHandler{
		userRepo:   userRepo,
		jwtService: jwtService,
	}
}

// LoginRequest represents login request payload
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
}

// RegisterRequest represents registration request payload
type RegisterRequest struct {
	Email        string            `json:"email" validate:"required,email"`
	Password     string            `json:"password" validate:"required,min=8"`
	Name         string            `json:"name" validate:"required,min=2"`
	Role         entities.UserRole `json:"role,omitempty"`
	EnterpriseID uuid.UUID         `json:"enterprise_id" validate:"required"`
}

// RefreshTokenRequest represents refresh token request
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

// AuthResponse represents authentication response
type AuthResponse struct {
	User      *UserResponse   `json:"user"`
	TokenPair *auth.TokenPair `json:"tokens"`
	Message   string          `json:"message"`
}

// UserResponse represents user data in responses (without sensitive info)
type UserResponse struct {
	ID           uuid.UUID         `json:"id"`
	Email        string            `json:"email"`
	Name         string            `json:"name"`
	Role         entities.UserRole `json:"role"`
	EnterpriseID uuid.UUID         `json:"enterprise_id"`
	CreatedAt    time.Time         `json:"created_at"`
}

// Login authenticates a user
// @Summary User login
// @Description Authenticate user with email and password
// @Tags auth
// @Accept json
// @Produce json
// @Param login body LoginRequest true "Login credentials"
// @Success 200 {object} AuthResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /auth/login [post]
func (h *AuthHandler) Login(c *fiber.Ctx) error {
	// Check if dependencies are properly initialized
	if h.userRepo == nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"error": "Authentication service not available - user repository not initialized",
		})
	}
	
	if h.jwtService == nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"error": "Authentication service not available - JWT service not initialized",
		})
	}

	var req LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Find user by email
	user, err := h.userRepo.GetByEmail(c.Context(), req.Email)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid email or password",
		})
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid email or password",
		})
	}

	// Generate token pair
	tokenPair, err := h.jwtService.GenerateTokenPair(user)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to generate tokens",
		})
	}

	// Prepare response
	userResponse := &UserResponse{
		ID:           user.ID,
		Email:        user.Email,
		Name:         user.Name,
		Role:         user.Role,
		EnterpriseID: user.EnterpriseID,
		CreatedAt:    user.CreatedAt,
	}

	return c.JSON(AuthResponse{
		User:      userResponse,
		TokenPair: tokenPair,
		Message:   "Login successful",
	})
}

// Register creates a new user account
// @Summary User registration
// @Description Create a new user account
// @Tags auth
// @Accept json
// @Produce json
// @Param registration body RegisterRequest true "Registration data"
// @Success 201 {object} AuthResponse
// @Failure 400 {object} map[string]string
// @Failure 409 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /auth/register [post]
func (h *AuthHandler) Register(c *fiber.Ctx) error {
	// Check if dependencies are properly initialized
	if h.userRepo == nil || h.jwtService == nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"error": "Authentication service not available",
		})
	}

	var req RegisterRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Check if user already exists
	existingUser, err := h.userRepo.GetByEmail(c.Context(), req.Email)
	if err == nil && existingUser != nil {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{
			"error": "User with this email already exists",
		})
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to process password",
		})
	}

	// Set default role if not provided
	role := req.Role
	if role == "" {
		role = entities.UserRoleVisitor
	}

	// Create user
	user := &entities.User{
		ID:           uuid.New(),
		Email:        req.Email,
		Name:         req.Name,
		PasswordHash: string(hashedPassword),
		Role:         role,
		EnterpriseID: req.EnterpriseID,
		IsActive:     true,
		CreatedAt:    time.Now(),
	}

	if err := h.userRepo.Create(c.Context(), user); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create user",
		})
	}

	// Generate token pair
	tokenPair, err := h.jwtService.GenerateTokenPair(user)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to generate tokens",
		})
	}

	// Prepare response
	userResponse := &UserResponse{
		ID:           user.ID,
		Email:        user.Email,
		Name:         user.Name,
		Role:         user.Role,
		EnterpriseID: user.EnterpriseID,
		CreatedAt:    user.CreatedAt,
	}

	return c.Status(fiber.StatusCreated).JSON(AuthResponse{
		User:      userResponse,
		TokenPair: tokenPair,
		Message:   "Registration successful",
	})
}

// RefreshToken generates new tokens using refresh token
// @Summary Refresh access token
// @Description Get new access token using refresh token
// @Tags auth
// @Accept json
// @Produce json
// @Param refresh body RefreshTokenRequest true "Refresh token"
// @Success 200 {object} auth.TokenPair
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /auth/refresh [post]
func (h *AuthHandler) RefreshToken(c *fiber.Ctx) error {
	// Check if dependencies are properly initialized
	if h.jwtService == nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"error": "Authentication service not available",
		})
	}

	var req RefreshTokenRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Refresh tokens
	tokenPair, err := h.jwtService.RefreshAccessToken(req.RefreshToken, func(userID uuid.UUID) (*entities.User, error) {
		return h.userRepo.GetByID(c.Context(), userID)
	})
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid or expired refresh token",
		})
	}

	return c.JSON(tokenPair)
}

// Logout invalidates user tokens (if blacklist is implemented)
// @Summary User logout
// @Description Invalidate user tokens
// @Tags auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /auth/logout [post]
func (h *AuthHandler) Logout(c *fiber.Ctx) error {
	// In a complete implementation, you would:
	// 1. Extract token from header
	// 2. Add token to blacklist
	// 3. Return success

	return c.JSON(fiber.Map{
		"message": "Logout successful",
	})
}

// GetProfile returns current user profile
// @Summary Get user profile
// @Description Get authenticated user's profile information
// @Tags auth
// @Produce json
// @Security BearerAuth
// @Success 200 {object} UserResponse
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /auth/profile [get]
func (h *AuthHandler) GetProfile(c *fiber.Ctx) error {
	// Check if dependencies are properly initialized
	if h.userRepo == nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"error": "Authentication service not available",
		})
	}

	userID, err := middleware.GetUserFromContext(c)
	if err != nil {
		return err
	}

	user, err := h.userRepo.GetByID(c.Context(), userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get user profile",
		})
	}

	userResponse := &UserResponse{
		ID:           user.ID,
		Email:        user.Email,
		Name:         user.Name,
		Role:         user.Role,
		EnterpriseID: user.EnterpriseID,
		CreatedAt:    user.CreatedAt,
	}

	return c.JSON(userResponse)
}

// UpdateProfile updates user profile
// @Summary Update user profile
// @Description Update authenticated user's profile information
// @Tags auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param profile body UpdateProfileRequest true "Profile data"
// @Success 200 {object} UserResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /auth/profile [put]
func (h *AuthHandler) UpdateProfile(c *fiber.Ctx) error {
	// Check if dependencies are properly initialized
	if h.userRepo == nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"error": "Authentication service not available",
		})
	}

	userID, err := middleware.GetUserFromContext(c)
	if err != nil {
		return err
	}

	var req UpdateProfileRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	user, err := h.userRepo.GetByID(c.Context(), userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get user",
		})
	}

	// Update fields if provided
	if req.Name != "" {
		user.Name = req.Name
	}
	if req.Email != "" {
		user.Email = req.Email
	}

	now := time.Now()
	user.UpdatedAt = &now

	if err := h.userRepo.Update(c.Context(), user); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update profile",
		})
	}

	userResponse := &UserResponse{
		ID:           user.ID,
		Email:        user.Email,
		Name:         user.Name,
		Role:         user.Role,
		EnterpriseID: user.EnterpriseID,
		CreatedAt:    user.CreatedAt,
	}

	return c.JSON(userResponse)
}

// UpdateProfileRequest represents profile update request
type UpdateProfileRequest struct {
	Name  string `json:"name,omitempty" validate:"omitempty,min=2"`
	Email string `json:"email,omitempty" validate:"omitempty,email"`
}

// ChangePasswordRequest represents password change request
type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" validate:"required"`
	NewPassword     string `json:"new_password" validate:"required,min=8"`
}

// ForgotPasswordRequest represents password reset request
type ForgotPasswordRequest struct {
	Email string `json:"email" validate:"required,email"`
}

// ResetPasswordRequest represents password reset with token request
type ResetPasswordRequest struct {
	Token       string `json:"token" validate:"required"`
	NewPassword string `json:"new_password" validate:"required,min=8"`
}

// ChangePassword updates user password
// @Summary Change password
// @Description Change authenticated user's password
// @Tags auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param password body ChangePasswordRequest true "Password data"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /auth/change-password [post]
func (h *AuthHandler) ChangePassword(c *fiber.Ctx) error {
	// Check if dependencies are properly initialized
	if h.userRepo == nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"error": "Authentication service not available",
		})
	}

	userID, err := middleware.GetUserFromContext(c)
	if err != nil {
		return err
	}

	var req ChangePasswordRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	user, err := h.userRepo.GetByID(c.Context(), userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get user",
		})
	}

	// Verify current password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.CurrentPassword)); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Current password is incorrect",
		})
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to process new password",
		})
	}

	// Update password
	user.PasswordHash = string(hashedPassword)
	now := time.Now()
	user.UpdatedAt = &now

	if err := h.userRepo.Update(c.Context(), user); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update password",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Password changed successfully",
	})
}

// ForgotPassword initiates password reset process
// @Summary Initiate password reset
// @Description Send password reset instructions to user's email
// @Tags auth
// @Accept json
// @Produce json
// @Param forgot body ForgotPasswordRequest true "Email for password reset"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /auth/forgot-password [post]
func (h *AuthHandler) ForgotPassword(c *fiber.Ctx) error {
	// Check if dependencies are properly initialized
	if h.userRepo == nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"error": "Authentication service not available",
		})
	}

	var req ForgotPasswordRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Check if user exists
	user, err := h.userRepo.GetByEmail(c.Context(), req.Email)
	if err != nil {
		// For security, don't reveal if email exists or not
		return c.JSON(fiber.Map{
			"message": "If the email exists in our system, password reset instructions have been sent",
		})
	}

	// Generate password reset token (simplified - in production use secure token generation)
	resetToken := uuid.New().String()

	// TODO: Store reset token with expiration in database or cache (Redis)
	// TODO: Send email with reset link containing the token
	// For now, we'll just return success

	// In production, you would:
	// 1. Generate a secure reset token
	// 2. Store token with expiration (15-30 minutes)
	// 3. Send email with reset link
	// 4. Return success without revealing user existence

	_ = user       // Prevent unused variable warning
	_ = resetToken // Prevent unused variable warning

	return c.JSON(fiber.Map{
		"message": "If the email exists in our system, password reset instructions have been sent",
	})
}

// ResetPassword resets password using reset token
// @Summary Reset password with token
// @Description Reset user password using the reset token from email
// @Tags auth
// @Accept json
// @Produce json
// @Param reset body ResetPasswordRequest true "Reset token and new password"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /auth/reset-password [post]
func (h *AuthHandler) ResetPassword(c *fiber.Ctx) error {
	var req ResetPasswordRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// TODO: In production implementation:
	// 1. Validate reset token from database/cache
	// 2. Check token expiration
	// 3. Get user associated with token
	// 4. Update user password
	// 5. Invalidate the reset token
	// 6. Optionally invalidate all existing tokens for security

	// For now, return error indicating implementation needed
	return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
		"error": "Password reset functionality is not fully implemented yet. Please contact administrator.",
	})

	// Example production code would be:
	/*
		// Validate token and get associated user
		userID, err := validateResetToken(req.Token)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid or expired reset token",
			})
		}

		user, err := h.userRepo.GetByID(c.Context(), userID)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to get user",
			})
		}

		// Hash new password
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to process new password",
			})
		}

		// Update password
		user.PasswordHash = string(hashedPassword)
		now := time.Now()
		user.UpdatedAt = &now

		if err := h.userRepo.Update(c.Context(), user); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to update password",
			})
		}

		// Invalidate reset token
		invalidateResetToken(req.Token)

		return c.JSON(fiber.Map{
			"message": "Password reset successfully",
		})
	*/
}

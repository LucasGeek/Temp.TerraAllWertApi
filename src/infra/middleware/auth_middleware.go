package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"terra-allwert/domain/entities"
	"terra-allwert/domain/interfaces"
	"terra-allwert/infra/auth"
)

// AuthMiddleware handles JWT authentication
type AuthMiddleware struct {
	jwtService *auth.JWTService
	userRepo   interfaces.UserRepository
}

// NewAuthMiddleware creates a new auth middleware
func NewAuthMiddleware(jwtService *auth.JWTService, userRepo interfaces.UserRepository) *AuthMiddleware {
	return &AuthMiddleware{
		jwtService: jwtService,
		userRepo:   userRepo,
	}
}

// RequireAuth middleware that requires valid authentication
func (am *AuthMiddleware) RequireAuth() fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Authorization header is required",
			})
		}

		// Extract token from header
		tokenString, err := auth.ExtractTokenFromAuthHeader(authHeader)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid authorization header format",
			})
		}

		// Validate token
		claims, err := am.jwtService.ValidateAccessToken(tokenString)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid or expired token",
			})
		}

		// Store user information in context for use in handlers
		c.Locals("user_id", claims.UserID.String())
		c.Locals("user_uuid", claims.UserID)
		c.Locals("user_email", claims.Email)
		c.Locals("user_role", claims.Role)
		c.Locals("enterprise_id", claims.EnterpriseID.String())
		c.Locals("enterprise_uuid", claims.EnterpriseID)

		return c.Next()
	}
}

// RequireRole middleware that requires specific user role
func (am *AuthMiddleware) RequireRole(requiredRoles ...entities.UserRole) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// First run auth middleware
		if err := am.RequireAuth()(c); err != nil {
			return err
		}

		userRole, ok := c.Locals("user_role").(entities.UserRole)
		if !ok {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Unable to determine user role",
			})
		}

		// Check if user has required role
		for _, role := range requiredRoles {
			if userRole == role {
				return c.Next()
			}
		}

		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "Insufficient permissions",
			"required_roles": requiredRoles,
			"user_role": userRole,
		})
	}
}

// RequireEnterprise middleware that requires user to belong to an enterprise
func (am *AuthMiddleware) RequireEnterprise() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// First run auth middleware
		if err := am.RequireAuth()(c); err != nil {
			return err
		}

		enterpriseID := c.Locals("enterprise_id")
		if enterpriseID == nil {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "Enterprise access required",
			})
		}

		return c.Next()
	}
}

// OptionalAuth middleware that extracts user info if token is present but doesn't require it
func (am *AuthMiddleware) OptionalAuth() fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Next()
		}

		// Extract token from header
		tokenString, err := auth.ExtractTokenFromAuthHeader(authHeader)
		if err != nil {
			return c.Next() // Continue without auth info
		}

		// Validate token
		claims, err := am.jwtService.ValidateAccessToken(tokenString)
		if err != nil {
			return c.Next() // Continue without auth info
		}

		// Store user information in context
		c.Locals("user_id", claims.UserID.String())
		c.Locals("user_uuid", claims.UserID)
		c.Locals("user_email", claims.Email)
		c.Locals("user_role", claims.Role)
		c.Locals("enterprise_id", claims.EnterpriseID.String())
		c.Locals("enterprise_uuid", claims.EnterpriseID)

		return c.Next()
	}
}

// GetUserFromContext extracts user ID from context
func GetUserFromContext(c *fiber.Ctx) (uuid.UUID, error) {
	userUUID, ok := c.Locals("user_uuid").(uuid.UUID)
	if !ok {
		userIDStr, ok := c.Locals("user_id").(string)
		if !ok {
			return uuid.Nil, fiber.NewError(fiber.StatusUnauthorized, "User not authenticated")
		}
		
		userID, err := uuid.Parse(userIDStr)
		if err != nil {
			return uuid.Nil, fiber.NewError(fiber.StatusInternalServerError, "Invalid user ID format")
		}
		return userID, nil
	}
	return userUUID, nil
}

// GetUserRoleFromContext extracts user role from context
func GetUserRoleFromContext(c *fiber.Ctx) (entities.UserRole, error) {
	userRole, ok := c.Locals("user_role").(entities.UserRole)
	if !ok {
		return "", fiber.NewError(fiber.StatusUnauthorized, "User role not found")
	}
	return userRole, nil
}

// GetEnterpriseFromContext extracts enterprise ID from context
func GetEnterpriseFromContext(c *fiber.Ctx) (uuid.UUID, error) {
	enterpriseUUID, ok := c.Locals("enterprise_uuid").(uuid.UUID)
	if !ok {
		enterpriseIDStr, ok := c.Locals("enterprise_id").(string)
		if !ok {
			return uuid.Nil, fiber.NewError(fiber.StatusForbidden, "Enterprise access required")
		}
		
		enterpriseID, err := uuid.Parse(enterpriseIDStr)
		if err != nil {
			return uuid.Nil, fiber.NewError(fiber.StatusInternalServerError, "Invalid enterprise ID format")
		}
		return enterpriseID, nil
	}
	return enterpriseUUID, nil
}

// CORS middleware with authentication considerations
func CORSWithAuth() fiber.Handler {
	return func(c *fiber.Ctx) error {
		origin := c.Get("Origin")
		
		// Set CORS headers
		c.Set("Access-Control-Allow-Origin", origin)
		c.Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")
		c.Set("Access-Control-Allow-Credentials", "true")
		c.Set("Access-Control-Max-Age", "86400")

		// Handle preflight requests
		if c.Method() == "OPTIONS" {
			return c.SendStatus(fiber.StatusNoContent)
		}

		return c.Next()
	}
}

// SecurityHeaders middleware adds security headers
func SecurityHeaders() fiber.Handler {
	return func(c *fiber.Ctx) error {
		c.Set("X-Content-Type-Options", "nosniff")
		c.Set("X-Frame-Options", "DENY")
		c.Set("X-XSS-Protection", "1; mode=block")
		c.Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		c.Set("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Set("Content-Security-Policy", "default-src 'self'")
		
		return c.Next()
	}
}

// RateLimitByUser applies different rate limits based on user role
func (am *AuthMiddleware) RateLimitByUser() fiber.Handler {
	return func(c *fiber.Ctx) error {
		userRole, _ := c.Locals("user_role").(entities.UserRole)
		
		// Apply different limits based on role
		var limit int
		switch userRole {
		case entities.UserRoleAdmin:
			limit = 1000 // High limit for admins
		case entities.UserRoleManager:
			limit = 500  // Medium limit for managers
		case entities.UserRoleVisitor:
			limit = 100  // Lower limit for visitors
		default:
			limit = 50   // Very low limit for unauthenticated users
		}

		// Store limit in context for other middleware to use
		c.Locals("rate_limit", limit)
		
		return c.Next()
	}
}

// LogUserActivity logs user actions for audit purposes
func (am *AuthMiddleware) LogUserActivity() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Continue processing
		err := c.Next()

		// Log after request processing
		userID := c.Locals("user_id")
		if userID != nil && c.Method() != "GET" {
			// Log non-GET requests for audit
			go func() {
				// Here you could log to database or external service
				// For now, we'll just structure the log data
				logData := map[string]interface{}{
					"user_id":    userID,
					"method":     c.Method(),
					"path":       c.Path(),
					"ip":         c.IP(),
					"user_agent": c.Get("User-Agent"),
					"timestamp":  c.Context().Time(),
					"status":     c.Response().StatusCode(),
				}
				_ = logData // Placeholder for actual logging implementation
			}()
		}

		return err
	}
}

// ValidateResourceOwnership ensures user can only access their own resources
func (am *AuthMiddleware) ValidateResourceOwnership(getResourceUserID func(*fiber.Ctx) (uuid.UUID, error)) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get authenticated user
		userID, err := GetUserFromContext(c)
		if err != nil {
			return err
		}

		// Get resource owner
		resourceUserID, err := getResourceUserID(c)
		if err != nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Resource not found",
			})
		}

		// Check if user owns the resource (or is admin)
		userRole, _ := GetUserRoleFromContext(c)
		if userRole != entities.UserRoleAdmin && userID != resourceUserID {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "Access denied: You can only access your own resources",
			})
		}

		return c.Next()
	}
}
package middleware

import (
	"context"
	"strings"

	"api/domain/entities"
	"api/domain/interfaces"

	"github.com/gofiber/fiber/v2"
)

type AuthMiddleware struct {
	authService interfaces.AuthService
}

func NewAuthMiddleware(authService interfaces.AuthService) *AuthMiddleware {
	return &AuthMiddleware{
		authService: authService,
	}
}

func (m *AuthMiddleware) RequireAuth() fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(401).JSON(fiber.Map{
				"error": "Authorization header required",
			})
		}

		token := strings.TrimPrefix(authHeader, "Bearer ")
		if token == authHeader {
			return c.Status(401).JSON(fiber.Map{
				"error": "Invalid authorization format",
			})
		}

		claims, err := m.authService.ValidateToken(c.Context(), token)
		if err != nil {
			return c.Status(401).JSON(fiber.Map{
				"error": "Invalid or expired token",
			})
		}

		c.Locals("user", claims)
		return c.Next()
	}
}

func (m *AuthMiddleware) RequireRole(role entities.UserRole) fiber.Handler {
	return func(c *fiber.Ctx) error {
		claims, ok := c.Locals("user").(*entities.JWTClaims)
		if !ok {
			return c.Status(401).JSON(fiber.Map{
				"error": "Authentication required",
			})
		}

		if claims.Role != role && claims.Role != entities.RoleAdmin {
			return c.Status(403).JSON(fiber.Map{
				"error": "Insufficient permissions",
			})
		}

		return c.Next()
	}
}

func (m *AuthMiddleware) RequireAdmin() fiber.Handler {
	return m.RequireRole(entities.RoleAdmin)
}

func GetUserFromContext(ctx context.Context) *entities.JWTClaims {
	if user, ok := ctx.Value("user").(*entities.JWTClaims); ok {
		return user
	}
	return nil
}

func GetUserFromFiberContext(c *fiber.Ctx) *entities.JWTClaims {
	if user, ok := c.Locals("user").(*entities.JWTClaims); ok {
		return user
	}
	return nil
}
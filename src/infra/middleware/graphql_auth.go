package middleware

import (
	"context"
	"strings"

	"api/domain/entities"
	"api/domain/interfaces"

	"github.com/99designs/gqlgen/graphql"
	"github.com/gofiber/fiber/v2"
)

// GraphQLAuthMiddleware adiciona autenticação ao contexto GraphQL
type GraphQLAuthMiddleware struct {
	authService interfaces.AuthService
}

// NewGraphQLAuthMiddleware cria novo middleware de auth GraphQL
func NewGraphQLAuthMiddleware(authService interfaces.AuthService) *GraphQLAuthMiddleware {
	return &GraphQLAuthMiddleware{
		authService: authService,
	}
}

// UserFromContext extrai o usuário do contexto
func UserFromContext(ctx context.Context) (*entities.User, bool) {
	user, ok := ctx.Value("user").(*entities.User)
	return user, ok
}

// RequireAuth verifica se o usuário está autenticado
func RequireAuth() graphql.FieldMiddleware {
	return graphql.FieldMiddleware(func(ctx context.Context, next graphql.Resolver) (interface{}, error) {
		user, ok := UserFromContext(ctx)
		if !ok || user == nil {
			return nil, fiber.NewError(fiber.StatusUnauthorized, "authentication required")
		}
		return next(ctx)
	})
}

// RequireAdmin verifica se o usuário é admin
func RequireAdmin() graphql.FieldMiddleware {
	return graphql.FieldMiddleware(func(ctx context.Context, next graphql.Resolver) (interface{}, error) {
		user, ok := UserFromContext(ctx)
		if !ok || user == nil {
			return nil, fiber.NewError(fiber.StatusUnauthorized, "authentication required")
		}

		if user.Role != entities.RoleAdmin {
			return nil, fiber.NewError(fiber.StatusForbidden, "admin access required")
		}

		return next(ctx)
	})
}

// ExtractUser extrai o usuário do token no cabeçalho Authorization
func (m *GraphQLAuthMiddleware) ExtractUser(ctx context.Context) (*entities.User, error) {
	// Esta função será chamada pelo handler HTTP para extrair o usuário
	// e adicionar ao contexto antes de passar para o GraphQL
	return nil, nil
}

// HTTPAuthMiddleware middleware HTTP que extrai o token e adiciona usuário ao contexto
func (m *GraphQLAuthMiddleware) HTTPAuthMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Pegar o token do cabeçalho Authorization
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			// Se não tem token, continua sem autenticação (operações públicas)
			return c.Next()
		}

		// Formato: "Bearer <token>"
		if !strings.HasPrefix(authHeader, "Bearer ") {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "invalid authorization header format",
			})
		}

		token := strings.TrimPrefix(authHeader, "Bearer ")

		// Validar o token
		claims, err := m.authService.ValidateToken(c.Context(), token)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "invalid or expired token",
			})
		}

		// Criar objeto user a partir das claims
		user := &entities.User{
			ID:       claims.UserID,
			Username: claims.Username,
			Role:     claims.Role,
		}

		// Adicionar usuário ao contexto local do Fiber
		c.Locals("user", user)

		return c.Next()
	}
}

// WithUser adiciona o usuário do Fiber locals ao contexto GraphQL
func WithUser(ctx context.Context, c *fiber.Ctx) context.Context {
	if user := c.Locals("user"); user != nil {
		return context.WithValue(ctx, "user", user)
	}
	return ctx
}

// IsPublicQuery verifica se a operação é pública (não requer autenticação)
func IsPublicQuery(operationName string, query string) bool {
	publicOperations := map[string]bool{
		"login":        true,
		"refreshToken": true,

		// Queries públicas (visualização)
		"towers":           true,
		"tower":            true,
		"floors":           true,
		"floor":            true,
		"apartments":       true,
		"apartment":        true,
		"searchApartments": true,
		"galleryImages":    true,
		"galleryImage":     true,
		"galleryRoutes":    true,
		"imagePins":        true,
		"imagePin":         true,
		"appConfig":        true,

		// Introspection queries (GraphQL Playground)
		"__schema": true,
		"__type":   true,
	}

	// Se tem nome da operação, verifica
	if operationName != "" {
		return publicOperations[operationName]
	}

	// Verifica se a query contém operações públicas
	query = strings.ToLower(strings.TrimSpace(query))

	// Login/refresh sempre públicos
	if strings.Contains(query, "login") || strings.Contains(query, "refreshtoken") {
		return true
	}

	// Queries de visualização são públicas
	if strings.HasPrefix(query, "query") || strings.HasPrefix(query, "{") {
		return true
	}

	// Mutations (exceto login) requerem auth
	return false
}

// IsAdminOnly verifica se a operação requer permissões de admin
func IsAdminOnly(operationName string, query string) bool {
	adminOperations := map[string]bool{
		// User management
		"createUser":     true,
		"updateUser":     true,
		"deleteUser":     true,
		"changePassword": true,
		"users":          true,
		"user":           true,

		// Tower management
		"createTower": true,
		"updateTower": true,
		"deleteTower": true,

		// Floor management
		"createFloor": true,
		"updateFloor": true,
		"deleteFloor": true,

		// Apartment management
		"createApartment":        true,
		"updateApartment":        true,
		"deleteApartment":        true,
		"addApartmentImage":      true,
		"removeApartmentImage":   true,
		"reorderApartmentImages": true,

		// Gallery management
		"createGalleryImage":   true,
		"updateGalleryImage":   true,
		"deleteGalleryImage":   true,
		"reorderGalleryImages": true,

		// Pin management
		"createImagePin": true,
		"updateImagePin": true,
		"deleteImagePin": true,

		// Config management
		"updateAppConfig": true,

		// File management
		"generateSignedUploadUrl": true,
		"generateBulkDownload":    true,
	}

	if operationName != "" {
		return adminOperations[operationName]
	}

	// Se é uma mutation e não é login/refresh, provavelmente é admin
	query = strings.ToLower(strings.TrimSpace(query))
	if strings.HasPrefix(query, "mutation") &&
		!strings.Contains(query, "login") &&
		!strings.Contains(query, "refreshtoken") &&
		!strings.Contains(query, "logout") {
		return true
	}

	return false
}

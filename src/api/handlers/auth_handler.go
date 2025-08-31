package handlers

import (
	"api/domain/entities"
	"api/domain/interfaces"
	"api/infra/middleware"

	"github.com/gofiber/fiber/v2"
	"github.com/go-playground/validator/v10"
)

type AuthHandler struct {
	authService interfaces.AuthService
	validator   *validator.Validate
}

func NewAuthHandler(authService interfaces.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		validator:   validator.New(),
	}
}

func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var request entities.LoginRequest
	if err := c.BodyParser(&request); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if err := h.validator.Struct(request); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Validation failed",
			"details": err.Error(),
		})
	}

	response, err := h.authService.Login(c.Context(), &request)
	if err != nil {
		return c.Status(401).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(response)
}

func (h *AuthHandler) RefreshToken(c *fiber.Ctx) error {
	var request struct {
		RefreshToken string `json:"refreshToken" validate:"required"`
	}

	if err := c.BodyParser(&request); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if err := h.validator.Struct(request); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Validation failed",
			"details": err.Error(),
		})
	}

	response, err := h.authService.RefreshToken(c.Context(), request.RefreshToken)
	if err != nil {
		return c.Status(401).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(response)
}

func (h *AuthHandler) GetProfile(c *fiber.Ctx) error {
	user := middleware.GetUserFromFiberContext(c)
	if user == nil {
		return c.Status(401).JSON(fiber.Map{
			"error": "Authentication required",
		})
	}

	return c.JSON(fiber.Map{
		"user": user,
	})
}

func (h *AuthHandler) Logout(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"message": "Logout successful",
	})
}
package handlers

import (
	"strconv"

	"terra-allwert/domain/entities"
	"terra-allwert/domain/interfaces"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type EnterpriseHandler struct {
	enterpriseRepo interfaces.EnterpriseRepository
}

func NewEnterpriseHandler(enterpriseRepo interfaces.EnterpriseRepository) *EnterpriseHandler {
	return &EnterpriseHandler{
		enterpriseRepo: enterpriseRepo,
	}
}

// CreateEnterprise creates a new enterprise
// @Summary Create a new enterprise
// @Description Create a new enterprise with the provided data
// @Tags enterprises
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param enterprise body entities.Enterprise true "Enterprise data"
// @Success 201 {object} entities.Enterprise
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /enterprises [post]
func (h *EnterpriseHandler) CreateEnterprise(c *fiber.Ctx) error {
	var enterprise entities.Enterprise

	if err := c.BodyParser(&enterprise); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if err := h.enterpriseRepo.Create(c.Context(), &enterprise); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create enterprise",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(enterprise)
}

// GetEnterpriseByID gets an enterprise by ID
// @Summary Get enterprise by ID
// @Description Get a single enterprise by its ID
// @Tags enterprises
// @Produce json
// @Security BearerAuth
// @Param id path string true "Enterprise ID"
// @Success 200 {object} entities.Enterprise
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /enterprises/{id} [get]
func (h *EnterpriseHandler) GetEnterpriseByID(c *fiber.Ctx) error {
	idParam := c.Params("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid enterprise ID",
		})
	}

	enterprise, err := h.enterpriseRepo.GetByID(c.Context(), id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Enterprise not found",
		})
	}

	return c.JSON(enterprise)
}

// GetEnterpriseBySlug gets an enterprise by slug
// @Summary Get enterprise by slug
// @Description Get a single enterprise by its slug
// @Tags enterprises
// @Produce json
// @Security BearerAuth
// @Param slug path string true "Enterprise slug"
// @Success 200 {object} entities.Enterprise
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /enterprises/slug/{slug} [get]
func (h *EnterpriseHandler) GetEnterpriseBySlug(c *fiber.Ctx) error {
	slug := c.Params("slug")

	enterprise, err := h.enterpriseRepo.GetBySlug(c.Context(), slug)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Enterprise not found",
		})
	}

	return c.JSON(enterprise)
}

// GetEnterprises gets all enterprises with pagination
// @Summary Get all enterprises
// @Description Get all enterprises with optional pagination
// @Tags enterprises
// @Produce json
// @Security BearerAuth
// @Param limit query int false "Limit" default(10)
// @Param offset query int false "Offset" default(0)
// @Success 200 {array} entities.Enterprise
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /enterprises [get]
func (h *EnterpriseHandler) GetEnterprises(c *fiber.Ctx) error {
	limit, _ := strconv.Atoi(c.Query("limit", "10"))
	offset, _ := strconv.Atoi(c.Query("offset", "0"))

	enterprises, err := h.enterpriseRepo.GetAll(c.Context(), limit, offset)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch enterprises",
		})
	}

	return c.JSON(enterprises)
}

// UpdateEnterprise updates an existing enterprise
// @Summary Update enterprise
// @Description Update an existing enterprise
// @Tags enterprises
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Enterprise ID"
// @Param enterprise body entities.Enterprise true "Enterprise data"
// @Success 200 {object} entities.Enterprise
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /enterprises/{id} [put]
func (h *EnterpriseHandler) UpdateEnterprise(c *fiber.Ctx) error {
	idParam := c.Params("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid enterprise ID",
		})
	}

	var enterprise entities.Enterprise
	if err := c.BodyParser(&enterprise); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	enterprise.ID = id
	if err := h.enterpriseRepo.Update(c.Context(), &enterprise); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update enterprise",
		})
	}

	return c.JSON(enterprise)
}

// DeleteEnterprise deletes an enterprise
// @Summary Delete enterprise
// @Description Delete an enterprise by ID
// @Tags enterprises
// @Security BearerAuth
// @Param id path string true "Enterprise ID"
// @Success 204 "No Content"
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /enterprises/{id} [delete]
func (h *EnterpriseHandler) DeleteEnterprise(c *fiber.Ctx) error {
	idParam := c.Params("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid enterprise ID",
		})
	}

	if err := h.enterpriseRepo.Delete(c.Context(), id); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to delete enterprise",
		})
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// SearchEnterprises searches enterprises
// @Summary Search enterprises
// @Description Search enterprises by query string
// @Tags enterprises
// @Produce json
// @Security BearerAuth
// @Param q query string true "Search query"
// @Param limit query int false "Limit" default(10)
// @Param offset query int false "Offset" default(0)
// @Success 200 {array} entities.Enterprise
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /enterprises/search [get]
func (h *EnterpriseHandler) SearchEnterprises(c *fiber.Ctx) error {
	query := c.Query("q")
	if query == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Search query is required",
		})
	}

	limit, _ := strconv.Atoi(c.Query("limit", "10"))
	offset, _ := strconv.Atoi(c.Query("offset", "0"))

	enterprises, err := h.enterpriseRepo.Search(c.Context(), query, limit, offset)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to search enterprises",
		})
	}

	return c.JSON(enterprises)
}

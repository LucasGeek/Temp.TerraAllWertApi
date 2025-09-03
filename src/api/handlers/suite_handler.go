package handlers

import (
	"strconv"

	"terra-allwert/domain/entities"
	"terra-allwert/domain/interfaces"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type SuiteHandler struct {
	suiteRepo interfaces.SuiteRepository
}

func NewSuiteHandler(suiteRepo interfaces.SuiteRepository) *SuiteHandler {
	return &SuiteHandler{
		suiteRepo: suiteRepo,
	}
}

// CreateSuite creates a new suite
// @Summary Create a new suite
// @Description Create a new suite with the provided data
// @Tags suites
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param suite body entities.Suite true "Suite data"
// @Success 201 {object} entities.Suite
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /suites [post]
func (h *SuiteHandler) CreateSuite(c *fiber.Ctx) error {
	var suite entities.Suite

	if err := c.BodyParser(&suite); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if err := h.suiteRepo.Create(c.Context(), &suite); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create suite",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(suite)
}

// GetSuiteByID gets a suite by ID
// @Summary Get suite by ID
// @Description Get a single suite by its ID
// @Tags suites
// @Produce json
// @Security BearerAuth
// @Param id path string true "Suite ID"
// @Success 200 {object} entities.Suite
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /suites/{id} [get]
func (h *SuiteHandler) GetSuiteByID(c *fiber.Ctx) error {
	idParam := c.Params("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid suite ID",
		})
	}

	suite, err := h.suiteRepo.GetByID(c.Context(), id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Suite not found",
		})
	}

	return c.JSON(suite)
}

// GetSuites gets all suites with pagination
// @Summary Get all suites
// @Description Get all suites with optional pagination
// @Tags suites
// @Produce json
// @Security BearerAuth
// @Param limit query int false "Limit" default(10)
// @Param offset query int false "Offset" default(0)
// @Success 200 {array} entities.Suite
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /suites [get]
func (h *SuiteHandler) GetSuites(c *fiber.Ctx) error {
	limit, _ := strconv.Atoi(c.Query("limit", "10"))
	offset, _ := strconv.Atoi(c.Query("offset", "0"))

	suites, err := h.suiteRepo.GetAll(c.Context(), limit, offset)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch suites",
		})
	}

	return c.JSON(suites)
}

// GetSuitesByFloor gets suites by floor ID
// @Summary Get suites by floor
// @Description Get all suites for a specific floor
// @Tags suites
// @Produce json
// @Security BearerAuth
// @Param floorId path string true "Floor ID"
// @Param limit query int false "Limit" default(10)
// @Param offset query int false "Offset" default(0)
// @Success 200 {array} entities.Suite
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /floors/{floorId}/suites [get]
func (h *SuiteHandler) GetSuitesByFloor(c *fiber.Ctx) error {
	floorIDParam := c.Params("floorId")
	floorID, err := uuid.Parse(floorIDParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid floor ID",
		})
	}

	limit, _ := strconv.Atoi(c.Query("limit", "10"))
	offset, _ := strconv.Atoi(c.Query("offset", "0"))

	suites, err := h.suiteRepo.GetByFloorID(c.Context(), floorID, limit, offset)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch suites",
		})
	}

	return c.JSON(suites)
}

// SearchSuites searches suites with filters
// @Summary Search suites
// @Description Search suites with various filters
// @Tags suites
// @Produce json
// @Security BearerAuth
// @Param min_bedrooms query int false "Minimum bedrooms"
// @Param max_bedrooms query int false "Maximum bedrooms"
// @Param min_area query number false "Minimum area (sqm)"
// @Param max_area query number false "Maximum area (sqm)"
// @Param min_price query number false "Minimum price"
// @Param max_price query number false "Maximum price"
// @Param status query string false "Suite status" Enums(available, reserved, sold, unavailable)
// @Param limit query int false "Limit" default(10)
// @Param offset query int false "Offset" default(0)
// @Success 200 {array} entities.Suite
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /suites/search [get]
func (h *SuiteHandler) SearchSuites(c *fiber.Ctx) error {
	filters := interfaces.SuiteSearchFilters{}

	if minBedrooms := c.Query("min_bedrooms"); minBedrooms != "" {
		if val, err := strconv.Atoi(minBedrooms); err == nil {
			filters.MinBedrooms = &val
		}
	}

	if maxBedrooms := c.Query("max_bedrooms"); maxBedrooms != "" {
		if val, err := strconv.Atoi(maxBedrooms); err == nil {
			filters.MaxBedrooms = &val
		}
	}

	if minArea := c.Query("min_area"); minArea != "" {
		if val, err := strconv.ParseFloat(minArea, 64); err == nil {
			filters.MinArea = &val
		}
	}

	if maxArea := c.Query("max_area"); maxArea != "" {
		if val, err := strconv.ParseFloat(maxArea, 64); err == nil {
			filters.MaxArea = &val
		}
	}

	if minPrice := c.Query("min_price"); minPrice != "" {
		if val, err := strconv.ParseFloat(minPrice, 64); err == nil {
			filters.MinPrice = &val
		}
	}

	if maxPrice := c.Query("max_price"); maxPrice != "" {
		if val, err := strconv.ParseFloat(maxPrice, 64); err == nil {
			filters.MaxPrice = &val
		}
	}

	if status := c.Query("status"); status != "" {
		suiteStatus := entities.SuiteStatus(status)
		filters.Status = &suiteStatus
	}

	limit, _ := strconv.Atoi(c.Query("limit", "10"))
	offset, _ := strconv.Atoi(c.Query("offset", "0"))

	suites, err := h.suiteRepo.Search(c.Context(), filters, limit, offset)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to search suites",
		})
	}

	return c.JSON(suites)
}

// UpdateSuite updates an existing suite
// @Summary Update suite
// @Description Update an existing suite
// @Tags suites
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Suite ID"
// @Param suite body entities.Suite true "Suite data"
// @Success 200 {object} entities.Suite
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /suites/{id} [put]
func (h *SuiteHandler) UpdateSuite(c *fiber.Ctx) error {
	idParam := c.Params("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid suite ID",
		})
	}

	var suite entities.Suite
	if err := c.BodyParser(&suite); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	suite.ID = id
	if err := h.suiteRepo.Update(c.Context(), &suite); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update suite",
		})
	}

	return c.JSON(suite)
}

// UpdateSuiteStatus updates suite status
// @Summary Update suite status
// @Description Update the status of a suite
// @Tags suites
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Suite ID"
// @Param status body object{status=string} true "Status data"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /suites/{id}/status [patch]
func (h *SuiteHandler) UpdateSuiteStatus(c *fiber.Ctx) error {
	idParam := c.Params("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid suite ID",
		})
	}

	var body struct {
		Status entities.SuiteStatus `json:"status"`
	}

	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if err := h.suiteRepo.UpdateStatus(c.Context(), id, body.Status); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update suite status",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Suite status updated successfully",
	})
}

// DeleteSuite deletes a suite
// @Summary Delete suite
// @Description Delete a suite by ID
// @Tags suites
// @Security BearerAuth
// @Param id path string true "Suite ID"
// @Success 204 "No Content"
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /suites/{id} [delete]
func (h *SuiteHandler) DeleteSuite(c *fiber.Ctx) error {
	idParam := c.Params("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid suite ID",
		})
	}

	if err := h.suiteRepo.Delete(c.Context(), id); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to delete suite",
		})
	}

	return c.SendStatus(fiber.StatusNoContent)
}

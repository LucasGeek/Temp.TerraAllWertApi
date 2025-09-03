package handlers

import (
	"strconv"

	"terra-allwert/domain/entities"
	"terra-allwert/domain/interfaces"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type FloorHandler struct {
	floorRepo interfaces.FloorRepository
}

func NewFloorHandler(floorRepo interfaces.FloorRepository) *FloorHandler {
	return &FloorHandler{
		floorRepo: floorRepo,
	}
}

// CreateFloor creates a new floor
// @Summary Create a new floor
// @Description Create a new floor with the provided data
// @Tags floors
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param floor body entities.Floor true "Floor data"
// @Success 201 {object} entities.Floor
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /floors [post]
func (h *FloorHandler) CreateFloor(c *fiber.Ctx) error {
	var floor entities.Floor

	if err := c.BodyParser(&floor); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if err := h.floorRepo.Create(c.Context(), &floor); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create floor",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(floor)
}

// GetFloorByID gets a floor by ID
// @Summary Get floor by ID
// @Description Get a single floor by its ID
// @Tags floors
// @Produce json
// @Security BearerAuth
// @Param id path string true "Floor ID"
// @Success 200 {object} entities.Floor
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /floors/{id} [get]
func (h *FloorHandler) GetFloorByID(c *fiber.Ctx) error {
	idParam := c.Params("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid floor ID",
		})
	}

	floor, err := h.floorRepo.GetByID(c.Context(), id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Floor not found",
		})
	}

	return c.JSON(floor)
}

// GetFloors gets all floors with pagination
// @Summary Get all floors
// @Description Get all floors with optional pagination
// @Tags floors
// @Produce json
// @Security BearerAuth
// @Param limit query int false "Limit" default(10)
// @Param offset query int false "Offset" default(0)
// @Success 200 {array} entities.Floor
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /floors [get]
func (h *FloorHandler) GetFloors(c *fiber.Ctx) error {
	limit, _ := strconv.Atoi(c.Query("limit", "10"))
	offset, _ := strconv.Atoi(c.Query("offset", "0"))

	floors, err := h.floorRepo.GetAll(c.Context(), limit, offset)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch floors",
		})
	}

	return c.JSON(floors)
}

// GetFloorsByTower gets floors by tower ID
// @Summary Get floors by tower
// @Description Get all floors for a specific tower
// @Tags floors
// @Produce json
// @Security BearerAuth
// @Param towerId path string true "Tower ID"
// @Param limit query int false "Limit" default(10)
// @Param offset query int false "Offset" default(0)
// @Success 200 {array} entities.Floor
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /towers/{towerId}/floors [get]
func (h *FloorHandler) GetFloorsByTower(c *fiber.Ctx) error {
	towerIDParam := c.Params("towerId")
	towerID, err := uuid.Parse(towerIDParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid tower ID",
		})
	}

	limit, _ := strconv.Atoi(c.Query("limit", "10"))
	offset, _ := strconv.Atoi(c.Query("offset", "0"))

	floors, err := h.floorRepo.GetByTowerID(c.Context(), towerID, limit, offset)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch floors",
		})
	}

	return c.JSON(floors)
}

// GetFloorByNumber gets a floor by tower ID and floor number
// @Summary Get floor by number
// @Description Get a specific floor by tower ID and floor number
// @Tags floors
// @Produce json
// @Security BearerAuth
// @Param towerId path string true "Tower ID"
// @Param floorNumber path int true "Floor Number"
// @Success 200 {object} entities.Floor
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /towers/{towerId}/floors/{floorNumber} [get]
func (h *FloorHandler) GetFloorByNumber(c *fiber.Ctx) error {
	towerIDParam := c.Params("towerId")
	towerID, err := uuid.Parse(towerIDParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid tower ID",
		})
	}

	floorNumberParam := c.Params("floorNumber")
	floorNumber, err := strconv.Atoi(floorNumberParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid floor number",
		})
	}

	floor, err := h.floorRepo.GetByFloorNumber(c.Context(), towerID, floorNumber)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Floor not found",
		})
	}

	return c.JSON(floor)
}

// UpdateFloor updates an existing floor
// @Summary Update floor
// @Description Update an existing floor
// @Tags floors
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Floor ID"
// @Param floor body entities.Floor true "Floor data"
// @Success 200 {object} entities.Floor
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /floors/{id} [put]
func (h *FloorHandler) UpdateFloor(c *fiber.Ctx) error {
	idParam := c.Params("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid floor ID",
		})
	}

	var floor entities.Floor
	if err := c.BodyParser(&floor); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	floor.ID = id
	if err := h.floorRepo.Update(c.Context(), &floor); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update floor",
		})
	}

	return c.JSON(floor)
}

// DeleteFloor deletes a floor
// @Summary Delete floor
// @Description Delete a floor by ID
// @Tags floors
// @Security BearerAuth
// @Param id path string true "Floor ID"
// @Success 204 "No Content"
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /floors/{id} [delete]
func (h *FloorHandler) DeleteFloor(c *fiber.Ctx) error {
	idParam := c.Params("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid floor ID",
		})
	}

	if err := h.floorRepo.Delete(c.Context(), id); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to delete floor",
		})
	}

	return c.SendStatus(fiber.StatusNoContent)
}

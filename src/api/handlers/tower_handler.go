package handlers

import (
	"strconv"

	"terra-allwert/domain/entities"
	"terra-allwert/domain/interfaces"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type TowerHandler struct {
	towerRepo interfaces.TowerRepository
}

func NewTowerHandler(towerRepo interfaces.TowerRepository) *TowerHandler {
	return &TowerHandler{
		towerRepo: towerRepo,
	}
}

// CreateTower creates a new tower
// @Summary Create a new tower
// @Description Create a new tower with the provided data
// @Tags towers
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param tower body entities.Tower true "Tower data"
// @Success 201 {object} entities.Tower
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /towers [post]
func (h *TowerHandler) CreateTower(c *fiber.Ctx) error {
	var tower entities.Tower

	if err := c.BodyParser(&tower); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if err := h.towerRepo.Create(c.Context(), &tower); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create tower",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(tower)
}

// GetTowerByID gets a tower by ID
// @Summary Get tower by ID
// @Description Get a single tower by its ID
// @Tags towers
// @Produce json
// @Security BearerAuth
// @Param id path string true "Tower ID"
// @Success 200 {object} entities.Tower
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /towers/{id} [get]
func (h *TowerHandler) GetTowerByID(c *fiber.Ctx) error {
	idParam := c.Params("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid tower ID",
		})
	}

	tower, err := h.towerRepo.GetByID(c.Context(), id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Tower not found",
		})
	}

	return c.JSON(tower)
}

// GetTowers gets all towers with pagination
// @Summary Get all towers
// @Description Get all towers with optional pagination
// @Tags towers
// @Produce json
// @Security BearerAuth
// @Param limit query int false "Limit" default(10)
// @Param offset query int false "Offset" default(0)
// @Success 200 {array} entities.Tower
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /towers [get]
func (h *TowerHandler) GetTowers(c *fiber.Ctx) error {
	limit, _ := strconv.Atoi(c.Query("limit", "10"))
	offset, _ := strconv.Atoi(c.Query("offset", "0"))

	towers, err := h.towerRepo.GetAll(c.Context(), limit, offset)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch towers",
		})
	}

	return c.JSON(towers)
}

// GetTowersByMenuFloorPlan gets towers by menu floor plan ID
// @Summary Get towers by menu floor plan
// @Description Get all towers for a specific menu floor plan
// @Tags towers
// @Produce json
// @Security BearerAuth
// @Param menuFloorPlanId path string true "Menu Floor Plan ID"
// @Param limit query int false "Limit" default(10)
// @Param offset query int false "Offset" default(0)
// @Success 200 {array} entities.Tower
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /menu-floor-plans/{menuFloorPlanId}/towers [get]
func (h *TowerHandler) GetTowersByMenuFloorPlan(c *fiber.Ctx) error {
	menuFloorPlanIDParam := c.Params("menuFloorPlanId")
	menuFloorPlanID, err := uuid.Parse(menuFloorPlanIDParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid menu floor plan ID",
		})
	}

	limit, _ := strconv.Atoi(c.Query("limit", "10"))
	offset, _ := strconv.Atoi(c.Query("offset", "0"))

	towers, err := h.towerRepo.GetByMenuFloorPlanID(c.Context(), menuFloorPlanID, limit, offset)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch towers",
		})
	}

	return c.JSON(towers)
}

// UpdateTower updates an existing tower
// @Summary Update tower
// @Description Update an existing tower
// @Tags towers
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Tower ID"
// @Param tower body entities.Tower true "Tower data"
// @Success 200 {object} entities.Tower
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /towers/{id} [put]
func (h *TowerHandler) UpdateTower(c *fiber.Ctx) error {
	idParam := c.Params("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid tower ID",
		})
	}

	var tower entities.Tower
	if err := c.BodyParser(&tower); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	tower.ID = id
	if err := h.towerRepo.Update(c.Context(), &tower); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update tower",
		})
	}

	return c.JSON(tower)
}

// UpdateTowerPosition updates tower position
// @Summary Update tower position
// @Description Update the position of a tower
// @Tags towers
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Tower ID"
// @Param position body object{position=int} true "Position data"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /towers/{id}/position [patch]
func (h *TowerHandler) UpdateTowerPosition(c *fiber.Ctx) error {
	idParam := c.Params("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid tower ID",
		})
	}

	var body struct {
		Position int `json:"position"`
	}

	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if err := h.towerRepo.UpdatePosition(c.Context(), id, body.Position); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update tower position",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Tower position updated successfully",
	})
}

// DeleteTower deletes a tower
// @Summary Delete tower
// @Description Delete a tower by ID
// @Tags towers
// @Security BearerAuth
// @Param id path string true "Tower ID"
// @Success 204 "No Content"
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /towers/{id} [delete]
func (h *TowerHandler) DeleteTower(c *fiber.Ctx) error {
	idParam := c.Params("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid tower ID",
		})
	}

	if err := h.towerRepo.Delete(c.Context(), id); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to delete tower",
		})
	}

	return c.SendStatus(fiber.StatusNoContent)
}

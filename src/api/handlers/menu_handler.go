package handlers

import (
	"strconv"

	"terra-allwert/domain/entities"
	"terra-allwert/domain/interfaces"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type MenuHandler struct {
	menuRepo interfaces.MenuRepository
}

func NewMenuHandler(menuRepo interfaces.MenuRepository) *MenuHandler {
	return &MenuHandler{
		menuRepo: menuRepo,
	}
}

// CreateMenu creates a new menu
// @Summary Create a new menu
// @Description Create a new menu with the provided data
// @Tags menus
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param menu body entities.Menu true "Menu data"
// @Success 201 {object} entities.Menu
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /menus [post]
func (h *MenuHandler) CreateMenu(c *fiber.Ctx) error {
	var menu entities.Menu

	if err := c.BodyParser(&menu); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if err := h.menuRepo.Create(c.Context(), &menu); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create menu",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(menu)
}

// GetMenuByID gets a menu by ID
// @Summary Get menu by ID
// @Description Get a single menu by its ID
// @Tags menus
// @Produce json
// @Security BearerAuth
// @Param id path string true "Menu ID"
// @Success 200 {object} entities.Menu
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /menus/{id} [get]
func (h *MenuHandler) GetMenuByID(c *fiber.Ctx) error {
	idParam := c.Params("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid menu ID",
		})
	}

	menu, err := h.menuRepo.GetByID(c.Context(), id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Menu not found",
		})
	}

	return c.JSON(menu)
}

// GetMenus gets all menus with pagination
// @Summary Get all menus
// @Description Get all menus with optional pagination
// @Tags menus
// @Produce json
// @Security BearerAuth
// @Param limit query int false "Limit" default(10)
// @Param offset query int false "Offset" default(0)
// @Success 200 {array} entities.Menu
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /menus [get]
func (h *MenuHandler) GetMenus(c *fiber.Ctx) error {
	limit, _ := strconv.Atoi(c.Query("limit", "10"))
	offset, _ := strconv.Atoi(c.Query("offset", "0"))

	menus, err := h.menuRepo.GetAll(c.Context(), limit, offset)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch menus",
		})
	}

	return c.JSON(menus)
}

// GetMenusByEnterprise gets menus by enterprise ID
// @Summary Get menus by enterprise
// @Description Get all menus for a specific enterprise
// @Tags menus
// @Produce json
// @Security BearerAuth
// @Param enterpriseId path string true "Enterprise ID"
// @Param limit query int false "Limit" default(10)
// @Param offset query int false "Offset" default(0)
// @Success 200 {array} entities.Menu
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /enterprises/{enterpriseId}/menus [get]
func (h *MenuHandler) GetMenusByEnterprise(c *fiber.Ctx) error {
	enterpriseIDParam := c.Params("enterpriseId")
	enterpriseID, err := uuid.Parse(enterpriseIDParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid enterprise ID",
		})
	}

	limit, _ := strconv.Atoi(c.Query("limit", "10"))
	offset, _ := strconv.Atoi(c.Query("offset", "0"))

	menus, err := h.menuRepo.GetByEnterpriseID(c.Context(), enterpriseID, limit, offset)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch menus",
		})
	}

	return c.JSON(menus)
}

// GetMenuHierarchy gets full menu hierarchy for an enterprise
// @Summary Get menu hierarchy
// @Description Get complete menu hierarchy for a specific enterprise
// @Tags menus
// @Produce json
// @Security BearerAuth
// @Param enterpriseId path string true "Enterprise ID"
// @Success 200 {array} entities.Menu
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /enterprises/{enterpriseId}/menus/hierarchy [get]
func (h *MenuHandler) GetMenuHierarchy(c *fiber.Ctx) error {
	enterpriseIDParam := c.Params("enterpriseId")
	enterpriseID, err := uuid.Parse(enterpriseIDParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid enterprise ID",
		})
	}

	menus, err := h.menuRepo.GetMenuHierarchy(c.Context(), enterpriseID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch menu hierarchy",
		})
	}

	return c.JSON(menus)
}

// GetChildMenus gets child menus of a parent menu
// @Summary Get child menus
// @Description Get all child menus for a specific parent menu
// @Tags menus
// @Produce json
// @Security BearerAuth
// @Param parentId path string true "Parent Menu ID"
// @Param limit query int false "Limit" default(10)
// @Param offset query int false "Offset" default(0)
// @Success 200 {array} entities.Menu
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /menus/{parentId}/children [get]
func (h *MenuHandler) GetChildMenus(c *fiber.Ctx) error {
	parentIDParam := c.Params("parentId")
	parentID, err := uuid.Parse(parentIDParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid parent menu ID",
		})
	}

	limit, _ := strconv.Atoi(c.Query("limit", "10"))
	offset, _ := strconv.Atoi(c.Query("offset", "0"))

	menus, err := h.menuRepo.GetChildren(c.Context(), parentID, limit, offset)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch child menus",
		})
	}

	return c.JSON(menus)
}

// UpdateMenu updates an existing menu
// @Summary Update menu
// @Description Update an existing menu
// @Tags menus
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Menu ID"
// @Param menu body entities.Menu true "Menu data"
// @Success 200 {object} entities.Menu
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /menus/{id} [put]
func (h *MenuHandler) UpdateMenu(c *fiber.Ctx) error {
	idParam := c.Params("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid menu ID",
		})
	}

	var menu entities.Menu
	if err := c.BodyParser(&menu); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	menu.ID = id
	if err := h.menuRepo.Update(c.Context(), &menu); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update menu",
		})
	}

	return c.JSON(menu)
}

// UpdateMenuPosition updates menu position
// @Summary Update menu position
// @Description Update the position of a menu in the hierarchy
// @Tags menus
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Menu ID"
// @Param position body object{position=int} true "Position data"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /menus/{id}/position [patch]
func (h *MenuHandler) UpdateMenuPosition(c *fiber.Ctx) error {
	idParam := c.Params("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid menu ID",
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

	if err := h.menuRepo.UpdatePosition(c.Context(), id, body.Position); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update menu position",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Menu position updated successfully",
	})
}

// DeleteMenu deletes a menu
// @Summary Delete menu
// @Description Delete a menu by ID
// @Tags menus
// @Security BearerAuth
// @Param id path string true "Menu ID"
// @Success 204 "No Content"
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /menus/{id} [delete]
func (h *MenuHandler) DeleteMenu(c *fiber.Ctx) error {
	idParam := c.Params("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid menu ID",
		})
	}

	if err := h.menuRepo.Delete(c.Context(), id); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to delete menu",
		})
	}

	return c.SendStatus(fiber.StatusNoContent)
}

package handlers

import (
	"strconv"

	"terra-allwert/domain/entities"
	"terra-allwert/domain/interfaces"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type PinsHandler struct {
	pinsRepo        interfaces.MenuPinsRepository
	markerRepo      interfaces.PinMarkerRepository
	markerImageRepo interfaces.PinMarkerImageRepository
}

func NewPinsHandler(
	pinsRepo interfaces.MenuPinsRepository,
	markerRepo interfaces.PinMarkerRepository,
	markerImageRepo interfaces.PinMarkerImageRepository,
) *PinsHandler {
	return &PinsHandler{
		pinsRepo:        pinsRepo,
		markerRepo:      markerRepo,
		markerImageRepo: markerImageRepo,
	}
}

// ============== MENU PINS ENDPOINTS ==============

// CreateMenuPins creates a new menu pins
// @Summary Create a new menu pins
// @Description Create a new menu pins with the provided data
// @Tags pins
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param pins body entities.MenuPins true "Menu Pins data"
// @Success 201 {object} entities.MenuPins
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /menu-pins [post]
func (h *PinsHandler) CreateMenuPins(c *fiber.Ctx) error {
	var pins entities.MenuPins

	if err := c.BodyParser(&pins); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if err := h.pinsRepo.Create(c.Context(), &pins); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create menu pins",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(pins)
}

// GetMenuPinsByID gets a menu pins by ID
// @Summary Get menu pins by ID
// @Description Get a single menu pins by its ID
// @Tags pins
// @Produce json
// @Security BearerAuth
// @Param id path string true "Menu Pins ID"
// @Success 200 {object} entities.MenuPins
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /menu-pins/{id} [get]
func (h *PinsHandler) GetMenuPinsByID(c *fiber.Ctx) error {
	idParam := c.Params("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid menu pins ID",
		})
	}

	pins, err := h.pinsRepo.GetByID(c.Context(), id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Menu pins not found",
		})
	}

	return c.JSON(pins)
}

// GetMenuPinsByMenuID gets a menu pins by menu ID
// @Summary Get menu pins by menu ID
// @Description Get menu pins for a specific menu
// @Tags pins
// @Produce json
// @Security BearerAuth
// @Param menuId path string true "Menu ID"
// @Success 200 {object} entities.MenuPins
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /menus/{menuId}/pins [get]
func (h *PinsHandler) GetMenuPinsByMenuID(c *fiber.Ctx) error {
	menuIDParam := c.Params("menuId")
	menuID, err := uuid.Parse(menuIDParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid menu ID",
		})
	}

	pins, err := h.pinsRepo.GetByMenuID(c.Context(), menuID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Menu pins not found",
		})
	}

	return c.JSON(pins)
}

// UpdateMenuPins updates an existing menu pins
// @Summary Update menu pins
// @Description Update an existing menu pins
// @Tags pins
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Menu Pins ID"
// @Param pins body entities.MenuPins true "Menu Pins data"
// @Success 200 {object} entities.MenuPins
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /menu-pins/{id} [put]
func (h *PinsHandler) UpdateMenuPins(c *fiber.Ctx) error {
	idParam := c.Params("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid menu pins ID",
		})
	}

	var pins entities.MenuPins
	if err := c.BodyParser(&pins); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	pins.ID = id
	if err := h.pinsRepo.Update(c.Context(), &pins); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update menu pins",
		})
	}

	return c.JSON(pins)
}

// DeleteMenuPins deletes a menu pins
// @Summary Delete menu pins
// @Description Delete a menu pins by ID
// @Tags pins
// @Security BearerAuth
// @Param id path string true "Menu Pins ID"
// @Success 204 "No Content"
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /menu-pins/{id} [delete]
func (h *PinsHandler) DeleteMenuPins(c *fiber.Ctx) error {
	idParam := c.Params("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid menu pins ID",
		})
	}

	if err := h.pinsRepo.Delete(c.Context(), id); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to delete menu pins",
		})
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// ============== PIN MARKER ENDPOINTS ==============

// CreatePinMarker creates a new pin marker
// @Summary Create a new pin marker
// @Description Create a new pin marker with the provided data
// @Tags pin-markers
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param marker body entities.PinMarker true "Pin Marker data"
// @Success 201 {object} entities.PinMarker
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /pin-markers [post]
func (h *PinsHandler) CreatePinMarker(c *fiber.Ctx) error {
	var marker entities.PinMarker

	if err := c.BodyParser(&marker); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if err := h.markerRepo.Create(c.Context(), &marker); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create pin marker",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(marker)
}

// GetPinMarkersByMenuPin gets pin markers by menu pins ID
// @Summary Get pin markers by menu pins
// @Description Get all pin markers for a specific menu pins
// @Tags pin-markers
// @Produce json
// @Security BearerAuth
// @Param menuPinId path string true "Menu Pins ID"
// @Param limit query int false "Limit" default(10)
// @Param offset query int false "Offset" default(0)
// @Param visible_only query boolean false "Only visible markers"
// @Success 200 {array} entities.PinMarker
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /menu-pins/{menuPinId}/markers [get]
func (h *PinsHandler) GetPinMarkersByMenuPin(c *fiber.Ctx) error {
	menuPinIDParam := c.Params("menuPinId")
	menuPinID, err := uuid.Parse(menuPinIDParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid menu pins ID",
		})
	}

	limit, _ := strconv.Atoi(c.Query("limit", "10"))
	offset, _ := strconv.Atoi(c.Query("offset", "0"))
	visibleOnly, _ := strconv.ParseBool(c.Query("visible_only", "false"))

	var markers []*entities.PinMarker
	if visibleOnly {
		markers, err = h.markerRepo.GetVisibleMarkers(c.Context(), menuPinID, limit, offset)
	} else {
		markers, err = h.markerRepo.GetByMenuPinID(c.Context(), menuPinID, limit, offset)
	}

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch pin markers",
		})
	}

	return c.JSON(markers)
}

// GetPinMarkersByPosition gets pin markers by position range
// @Summary Get pin markers by position
// @Description Get pin markers within a specific position range
// @Tags pin-markers
// @Produce json
// @Security BearerAuth
// @Param menuPinId path string true "Menu Pins ID"
// @Param min_x query number true "Minimum X position"
// @Param max_x query number true "Maximum X position"
// @Param min_y query number true "Minimum Y position"
// @Param max_y query number true "Maximum Y position"
// @Param limit query int false "Limit" default(10)
// @Param offset query int false "Offset" default(0)
// @Success 200 {array} entities.PinMarker
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /menu-pins/{menuPinId}/markers/search [get]
func (h *PinsHandler) GetPinMarkersByPosition(c *fiber.Ctx) error {
	menuPinIDParam := c.Params("menuPinId")
	menuPinID, err := uuid.Parse(menuPinIDParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid menu pins ID",
		})
	}

	minX, err := strconv.ParseFloat(c.Query("min_x"), 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid min_x parameter",
		})
	}

	maxX, err := strconv.ParseFloat(c.Query("max_x"), 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid max_x parameter",
		})
	}

	minY, err := strconv.ParseFloat(c.Query("min_y"), 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid min_y parameter",
		})
	}

	maxY, err := strconv.ParseFloat(c.Query("max_y"), 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid max_y parameter",
		})
	}

	limit, _ := strconv.Atoi(c.Query("limit", "10"))
	offset, _ := strconv.Atoi(c.Query("offset", "0"))

	markers, err := h.markerRepo.GetByPosition(c.Context(), menuPinID, minX, maxX, minY, maxY, limit, offset)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch pin markers",
		})
	}

	return c.JSON(markers)
}

// UpdatePinMarker updates an existing pin marker
// @Summary Update pin marker
// @Description Update an existing pin marker
// @Tags pin-markers
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Pin Marker ID"
// @Param marker body entities.PinMarker true "Pin Marker data"
// @Success 200 {object} entities.PinMarker
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /pin-markers/{id} [put]
func (h *PinsHandler) UpdatePinMarker(c *fiber.Ctx) error {
	idParam := c.Params("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid pin marker ID",
		})
	}

	var marker entities.PinMarker
	if err := c.BodyParser(&marker); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	marker.ID = id
	if err := h.markerRepo.Update(c.Context(), &marker); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update pin marker",
		})
	}

	return c.JSON(marker)
}

// DeletePinMarker deletes a pin marker
// @Summary Delete pin marker
// @Description Delete a pin marker by ID
// @Tags pin-markers
// @Security BearerAuth
// @Param id path string true "Pin Marker ID"
// @Success 204 "No Content"
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /pin-markers/{id} [delete]
func (h *PinsHandler) DeletePinMarker(c *fiber.Ctx) error {
	idParam := c.Params("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid pin marker ID",
		})
	}

	if err := h.markerRepo.Delete(c.Context(), id); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to delete pin marker",
		})
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// ============== PIN MARKER IMAGE ENDPOINTS ==============

// CreatePinMarkerImage creates a new pin marker image
// @Summary Create a new pin marker image
// @Description Create a new pin marker image with the provided data
// @Tags pin-marker-images
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param image body entities.PinMarkerImage true "Pin Marker Image data"
// @Success 201 {object} entities.PinMarkerImage
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /pin-marker-images [post]
func (h *PinsHandler) CreatePinMarkerImage(c *fiber.Ctx) error {
	var image entities.PinMarkerImage

	if err := c.BodyParser(&image); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if err := h.markerImageRepo.Create(c.Context(), &image); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create pin marker image",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(image)
}

// GetPinMarkerImagesByMarker gets pin marker images by pin marker ID
// @Summary Get pin marker images by marker
// @Description Get all images for a specific pin marker
// @Tags pin-marker-images
// @Produce json
// @Security BearerAuth
// @Param markerId path string true "Pin Marker ID"
// @Param limit query int false "Limit" default(10)
// @Param offset query int false "Offset" default(0)
// @Success 200 {array} entities.PinMarkerImage
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /pin-markers/{markerId}/images [get]
func (h *PinsHandler) GetPinMarkerImagesByMarker(c *fiber.Ctx) error {
	markerIDParam := c.Params("markerId")
	markerID, err := uuid.Parse(markerIDParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid pin marker ID",
		})
	}

	limit, _ := strconv.Atoi(c.Query("limit", "10"))
	offset, _ := strconv.Atoi(c.Query("offset", "0"))

	images, err := h.markerImageRepo.GetByPinMarkerID(c.Context(), markerID, limit, offset)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch pin marker images",
		})
	}

	return c.JSON(images)
}

// UpdatePinMarkerImage updates an existing pin marker image
// @Summary Update pin marker image
// @Description Update an existing pin marker image
// @Tags pin-marker-images
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Pin Marker Image ID"
// @Param image body entities.PinMarkerImage true "Pin Marker Image data"
// @Success 200 {object} entities.PinMarkerImage
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /pin-marker-images/{id} [put]
func (h *PinsHandler) UpdatePinMarkerImage(c *fiber.Ctx) error {
	idParam := c.Params("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid pin marker image ID",
		})
	}

	var image entities.PinMarkerImage
	if err := c.BodyParser(&image); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	image.ID = id
	if err := h.markerImageRepo.Update(c.Context(), &image); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update pin marker image",
		})
	}

	return c.JSON(image)
}

// UpdatePinMarkerImagePosition updates pin marker image position
// @Summary Update pin marker image position
// @Description Update the position of a pin marker image
// @Tags pin-marker-images
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Pin Marker Image ID"
// @Param position body object{position=int} true "Position data"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /pin-marker-images/{id}/position [patch]
func (h *PinsHandler) UpdatePinMarkerImagePosition(c *fiber.Ctx) error {
	idParam := c.Params("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid pin marker image ID",
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

	if err := h.markerImageRepo.UpdatePosition(c.Context(), id, body.Position); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update pin marker image position",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Pin marker image position updated successfully",
	})
}

// DeletePinMarkerImage deletes a pin marker image
// @Summary Delete pin marker image
// @Description Delete a pin marker image by ID
// @Tags pin-marker-images
// @Security BearerAuth
// @Param id path string true "Pin Marker Image ID"
// @Success 204 "No Content"
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /pin-marker-images/{id} [delete]
func (h *PinsHandler) DeletePinMarkerImage(c *fiber.Ctx) error {
	idParam := c.Params("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid pin marker image ID",
		})
	}

	if err := h.markerImageRepo.Delete(c.Context(), id); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to delete pin marker image",
		})
	}

	return c.SendStatus(fiber.StatusNoContent)
}

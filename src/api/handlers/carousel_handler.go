package handlers

import (
	"strconv"

	"terra-allwert/domain/entities"
	"terra-allwert/domain/interfaces"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type CarouselHandler struct {
	carouselRepo     interfaces.MenuCarouselRepository
	carouselItemRepo interfaces.CarouselItemRepository
	textOverlayRepo  interfaces.CarouselTextOverlayRepository
}

func NewCarouselHandler(
	carouselRepo interfaces.MenuCarouselRepository,
	carouselItemRepo interfaces.CarouselItemRepository,
	textOverlayRepo interfaces.CarouselTextOverlayRepository,
) *CarouselHandler {
	return &CarouselHandler{
		carouselRepo:     carouselRepo,
		carouselItemRepo: carouselItemRepo,
		textOverlayRepo:  textOverlayRepo,
	}
}

// ============== MENU CAROUSEL ENDPOINTS ==============

// CreateMenuCarousel creates a new menu carousel
// @Summary Create a new menu carousel
// @Description Create a new menu carousel with the provided data
// @Tags carousels
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param carousel body entities.MenuCarousel true "Menu Carousel data"
// @Success 201 {object} entities.MenuCarousel
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /menu-carousels [post]
func (h *CarouselHandler) CreateMenuCarousel(c *fiber.Ctx) error {
	var carousel entities.MenuCarousel

	if err := c.BodyParser(&carousel); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if err := h.carouselRepo.Create(c.Context(), &carousel); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create menu carousel",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(carousel)
}

// GetMenuCarouselByID gets a menu carousel by ID
// @Summary Get menu carousel by ID
// @Description Get a single menu carousel by its ID
// @Tags carousels
// @Produce json
// @Security BearerAuth
// @Param id path string true "Menu Carousel ID"
// @Success 200 {object} entities.MenuCarousel
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /menu-carousels/{id} [get]
func (h *CarouselHandler) GetMenuCarouselByID(c *fiber.Ctx) error {
	idParam := c.Params("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid menu carousel ID",
		})
	}

	carousel, err := h.carouselRepo.GetByID(c.Context(), id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Menu carousel not found",
		})
	}

	return c.JSON(carousel)
}

// GetMenuCarouselByMenuID gets a menu carousel by menu ID
// @Summary Get menu carousel by menu ID
// @Description Get menu carousel for a specific menu
// @Tags carousels
// @Produce json
// @Security BearerAuth
// @Param menuId path string true "Menu ID"
// @Success 200 {object} entities.MenuCarousel
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /menus/{menuId}/carousel [get]
func (h *CarouselHandler) GetMenuCarouselByMenuID(c *fiber.Ctx) error {
	menuIDParam := c.Params("menuId")
	menuID, err := uuid.Parse(menuIDParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid menu ID",
		})
	}

	carousel, err := h.carouselRepo.GetByMenuID(c.Context(), menuID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Menu carousel not found",
		})
	}

	return c.JSON(carousel)
}

// UpdateMenuCarousel updates an existing menu carousel
// @Summary Update menu carousel
// @Description Update an existing menu carousel
// @Tags carousels
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Menu Carousel ID"
// @Param carousel body entities.MenuCarousel true "Menu Carousel data"
// @Success 200 {object} entities.MenuCarousel
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /menu-carousels/{id} [put]
func (h *CarouselHandler) UpdateMenuCarousel(c *fiber.Ctx) error {
	idParam := c.Params("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid menu carousel ID",
		})
	}

	var carousel entities.MenuCarousel
	if err := c.BodyParser(&carousel); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	carousel.ID = id
	if err := h.carouselRepo.Update(c.Context(), &carousel); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update menu carousel",
		})
	}

	return c.JSON(carousel)
}

// DeleteMenuCarousel deletes a menu carousel
// @Summary Delete menu carousel
// @Description Delete a menu carousel by ID
// @Tags carousels
// @Security BearerAuth
// @Param id path string true "Menu Carousel ID"
// @Success 204 "No Content"
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /menu-carousels/{id} [delete]
func (h *CarouselHandler) DeleteMenuCarousel(c *fiber.Ctx) error {
	idParam := c.Params("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid menu carousel ID",
		})
	}

	if err := h.carouselRepo.Delete(c.Context(), id); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to delete menu carousel",
		})
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// ============== CAROUSEL ITEM ENDPOINTS ==============

// CreateCarouselItem creates a new carousel item
// @Summary Create a new carousel item
// @Description Create a new carousel item with the provided data
// @Tags carousel-items
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param item body entities.CarouselItem true "Carousel Item data"
// @Success 201 {object} entities.CarouselItem
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /carousel-items [post]
func (h *CarouselHandler) CreateCarouselItem(c *fiber.Ctx) error {
	var item entities.CarouselItem

	if err := c.BodyParser(&item); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if err := h.carouselItemRepo.Create(c.Context(), &item); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create carousel item",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(item)
}

// GetCarouselItemsByCarousel gets carousel items by menu carousel ID
// @Summary Get carousel items by carousel
// @Description Get all carousel items for a specific menu carousel
// @Tags carousel-items
// @Produce json
// @Security BearerAuth
// @Param carouselId path string true "Menu Carousel ID"
// @Param limit query int false "Limit" default(10)
// @Param offset query int false "Offset" default(0)
// @Param active_only query boolean false "Only active items"
// @Success 200 {array} entities.CarouselItem
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /menu-carousels/{carouselId}/items [get]
func (h *CarouselHandler) GetCarouselItemsByCarousel(c *fiber.Ctx) error {
	carouselIDParam := c.Params("carouselId")
	carouselID, err := uuid.Parse(carouselIDParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid carousel ID",
		})
	}

	limit, _ := strconv.Atoi(c.Query("limit", "10"))
	offset, _ := strconv.Atoi(c.Query("offset", "0"))
	activeOnly, _ := strconv.ParseBool(c.Query("active_only", "false"))

	var items []*entities.CarouselItem
	if activeOnly {
		items, err = h.carouselItemRepo.GetActiveItems(c.Context(), carouselID, limit, offset)
	} else {
		items, err = h.carouselItemRepo.GetByMenuCarouselID(c.Context(), carouselID, limit, offset)
	}

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch carousel items",
		})
	}

	return c.JSON(items)
}

// UpdateCarouselItem updates an existing carousel item
// @Summary Update carousel item
// @Description Update an existing carousel item
// @Tags carousel-items
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Carousel Item ID"
// @Param item body entities.CarouselItem true "Carousel Item data"
// @Success 200 {object} entities.CarouselItem
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /carousel-items/{id} [put]
func (h *CarouselHandler) UpdateCarouselItem(c *fiber.Ctx) error {
	idParam := c.Params("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid carousel item ID",
		})
	}

	var item entities.CarouselItem
	if err := c.BodyParser(&item); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	item.ID = id
	if err := h.carouselItemRepo.Update(c.Context(), &item); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update carousel item",
		})
	}

	return c.JSON(item)
}

// UpdateCarouselItemPosition updates carousel item position
// @Summary Update carousel item position
// @Description Update the position of a carousel item
// @Tags carousel-items
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Carousel Item ID"
// @Param position body object{position=int} true "Position data"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /carousel-items/{id}/position [patch]
func (h *CarouselHandler) UpdateCarouselItemPosition(c *fiber.Ctx) error {
	idParam := c.Params("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid carousel item ID",
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

	if err := h.carouselItemRepo.UpdatePosition(c.Context(), id, body.Position); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update carousel item position",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Carousel item position updated successfully",
	})
}

// DeleteCarouselItem deletes a carousel item
// @Summary Delete carousel item
// @Description Delete a carousel item by ID
// @Tags carousel-items
// @Security BearerAuth
// @Param id path string true "Carousel Item ID"
// @Success 204 "No Content"
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /carousel-items/{id} [delete]
func (h *CarouselHandler) DeleteCarouselItem(c *fiber.Ctx) error {
	idParam := c.Params("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid carousel item ID",
		})
	}

	if err := h.carouselItemRepo.Delete(c.Context(), id); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to delete carousel item",
		})
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// ============== TEXT OVERLAY ENDPOINTS ==============

// CreateTextOverlay creates a new text overlay
// @Summary Create a new text overlay
// @Description Create a new text overlay with the provided data
// @Tags text-overlays
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param overlay body entities.CarouselTextOverlay true "Text Overlay data"
// @Success 201 {object} entities.CarouselTextOverlay
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /text-overlays [post]
func (h *CarouselHandler) CreateTextOverlay(c *fiber.Ctx) error {
	var overlay entities.CarouselTextOverlay

	if err := c.BodyParser(&overlay); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if err := h.textOverlayRepo.Create(c.Context(), &overlay); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create text overlay",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(overlay)
}

// GetTextOverlaysByItem gets text overlays by carousel item ID
// @Summary Get text overlays by carousel item
// @Description Get all text overlays for a specific carousel item
// @Tags text-overlays
// @Produce json
// @Security BearerAuth
// @Param itemId path string true "Carousel Item ID"
// @Param limit query int false "Limit" default(10)
// @Param offset query int false "Offset" default(0)
// @Success 200 {array} entities.CarouselTextOverlay
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /carousel-items/{itemId}/text-overlays [get]
func (h *CarouselHandler) GetTextOverlaysByItem(c *fiber.Ctx) error {
	itemIDParam := c.Params("itemId")
	itemID, err := uuid.Parse(itemIDParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid carousel item ID",
		})
	}

	limit, _ := strconv.Atoi(c.Query("limit", "10"))
	offset, _ := strconv.Atoi(c.Query("offset", "0"))

	overlays, err := h.textOverlayRepo.GetByCarouselItemID(c.Context(), itemID, limit, offset)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch text overlays",
		})
	}

	return c.JSON(overlays)
}

// UpdateTextOverlay updates an existing text overlay
// @Summary Update text overlay
// @Description Update an existing text overlay
// @Tags text-overlays
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Text Overlay ID"
// @Param overlay body entities.CarouselTextOverlay true "Text Overlay data"
// @Success 200 {object} entities.CarouselTextOverlay
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /text-overlays/{id} [put]
func (h *CarouselHandler) UpdateTextOverlay(c *fiber.Ctx) error {
	idParam := c.Params("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid text overlay ID",
		})
	}

	var overlay entities.CarouselTextOverlay
	if err := c.BodyParser(&overlay); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	overlay.ID = id
	if err := h.textOverlayRepo.Update(c.Context(), &overlay); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update text overlay",
		})
	}

	return c.JSON(overlay)
}

// DeleteTextOverlay deletes a text overlay
// @Summary Delete text overlay
// @Description Delete a text overlay by ID
// @Tags text-overlays
// @Security BearerAuth
// @Param id path string true "Text Overlay ID"
// @Success 204 "No Content"
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /text-overlays/{id} [delete]
func (h *CarouselHandler) DeleteTextOverlay(c *fiber.Ctx) error {
	idParam := c.Params("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid text overlay ID",
		})
	}

	if err := h.textOverlayRepo.Delete(c.Context(), id); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to delete text overlay",
		})
	}

	return c.SendStatus(fiber.StatusNoContent)
}

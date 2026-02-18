package recommended

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
)

type Handler struct {
	service *Service
}

func NewHandler(s *Service) *Handler {
	return &Handler{service: s}
}

func (h *Handler) RegisterPublicRoutes(app *fiber.App) {
	app.Get("/api/v1/product/recommended", h.getRecommended)
}

func (h *Handler) getRecommended(c *fiber.Ctx) error {
	// support pagination: ?limit=12&offset=0
	limit := 12
	offset := 0
	if l := c.Query("limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil && v > 0 {
			limit = v
		}
	}
	if o := c.Query("offset"); o != "" {
		if v, err := strconv.Atoi(o); err == nil && v >= 0 {
			offset = v
		}
	}
	items := h.service.List(limit, offset)
	return c.JSON(items)
}

package category

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
	app.Get("/api/v1/product/category", h.getCategories)
}

func (h *Handler) getCategories(c *fiber.Ctx) error {
	// debug log: endpoint hit
	println("DEBUG: getCategories called")
	limit := 100
	if l := c.Query("limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil && v > 0 {
			limit = v
		}
	}
	items := h.service.List(limit)
	return c.JSON(items)
}

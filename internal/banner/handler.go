package banner

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
	app.Get("/api/v1/product/banner", h.getBanner)
}

func (h *Handler) getBanner(c *fiber.Ctx) error {
	// debug log: endpoint hit
	println("DEBUG: getBanner called")
	limit := 10
	if l := c.Query("limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil && v > 0 {
			limit = v
		}
	}
	items := h.service.List(limit)
	// If DB/table is empty it's fine to return an empty array — frontend can render fallback images if desired.
	return c.JSON(items)
}

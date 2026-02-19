package shoppingmall

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
)

type Handler struct {
	service *Service
}

func NewHandler(s *Service) *Handler { return &Handler{service: s} }

func (h *Handler) RegisterPublicRoutes(app *fiber.App) {
	app.Get("/api/v1/shopping-mall", h.getShoppingMall)
	app.Get("/api/v1/product/shopping-mall", h.getProductShoppingMall)
}

func (h *Handler) getShoppingMall(c *fiber.Ctx) error {
	limit := 100
	if l := c.Query("limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil && v > 0 {
			limit = v
		}
	}
	items := h.service.List(limit)
	return c.JSON(items)
}

func (h *Handler) getProductShoppingMall(c *fiber.Ctx) error {
	limit := 100
	if l := c.Query("limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil && v > 0 {
			limit = v
		}
	}
	items := h.service.ListLite(limit)
	return c.JSON(items)
}

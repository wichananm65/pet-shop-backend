package order

import (
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/wichananm65/pet-shop-backend/internal/user"
)

// Handler delegates order operations to the order service.
// It also needs the user service to update user order lists.

type Handler struct {
	service     *Service
	userService user.ServiceInterface
}

func NewHandler(s *Service, us user.ServiceInterface) *Handler {
	return &Handler{service: s, userService: us}
}

func (h *Handler) RegisterProtectedRoutes(app *fiber.App) {
	app.Post("/api/v1/orders", h.createOrder)
}

type createOrderRequest struct {
	Cart          map[string]int `json:"cart"`
	Quantity      int            `json:"quantity"`
	TotalPrice    float64        `json:"totalPrice"`
	ShippingPrice float64        `json:"shippingPrice"`
	GrandPrice    float64        `json:"grandPrice"`
}

func (h *Handler) createOrder(c *fiber.Ctx) error {
	payload := new(createOrderRequest)
	if err := c.BodyParser(payload); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": err.Error()})
	}
	if len(payload.Cart) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "cart cannot be empty"})
	}
	if payload.Quantity <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "quantity must be positive"})
	}
	if payload.TotalPrice < 0 || payload.ShippingPrice < 0 || payload.GrandPrice < 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "prices must be non-negative"})
	}

	userID, err := user.GetUserIDFromCtx(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "unauthorized"})
	}

	order := Order{
		Cart:          payload.Cart,
		Quantity:      payload.Quantity,
		TotalPrice:    payload.TotalPrice,
		ShippingPrice: payload.ShippingPrice,
		GrandPrice:    payload.GrandPrice,
		CreatedAt:     time.Now().UTC().Format(time.RFC3339),
		UpdatedAt:     time.Now().UTC().Format(time.RFC3339),
	}

	created, err := h.service.Create(order, userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": err.Error()})
	}

	// append orderID to user's order list via userService
	if _, err2 := h.userService.AppendOrderID(userID, created.OrderID); err2 != nil {
		fmt.Printf("warning: could not append orderID to user %d: %v\n", userID, err2)
	}
	return c.Status(fiber.StatusOK).JSON(created)
}

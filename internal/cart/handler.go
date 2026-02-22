package cart

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/wichananm65/pet-shop-backend/internal/user"
)

// Handler delegates cart operations to the cart service.
// This keeps cart-specific HTTP routing isolated.
type Handler struct {
	service *Service
}

func NewHandler(s *Service) *Handler {
	return &Handler{service: s}
}

func (h *Handler) RegisterProtectedRoutes(app *fiber.App) {
	app.Get("/api/v1/cart", h.getCart)
	app.Post("/api/v1/product/cart", h.addToCart)
}

type cartRequest struct {
	ProductID int `json:"productID"`
	Quantity  int `json:"quantity,omitempty"`
}

func (h *Handler) addToCart(c *fiber.Ctx) error {
	payload := new(cartRequest)
	if err := c.BodyParser(payload); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": err.Error()})
	}
	if payload.ProductID <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "invalid productID"})
	}
	// allow negative quantities; zero will simply return current cart
	// (service handles qty==0 case)
	userID, err := user.GetUserIDFromCtx(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "unauthorized"})
	}

	items, err := h.service.AddToCart(userID, payload.ProductID, payload.Quantity)
	if err != nil {
		switch err {
		case user.ErrNotFound:
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "user not found"})
		default:
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": err.Error()})
		}
	}

	return c.Status(fiber.StatusOK).JSON(items)
}

func (h *Handler) getCart(c *fiber.Ctx) error {
	fmt.Printf("[DEBUG] cart.getCart invoked, remote=%s\n", c.IP())
	userID, err := user.GetUserIDFromCtx(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "unauthorized"})
	}

	cart, err := h.service.GetCart(userID)
	if err != nil {
		switch err {
		case user.ErrNotFound:
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "user not found"})
		default:
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": err.Error()})
		}
	}

	return c.JSON(cart)
}

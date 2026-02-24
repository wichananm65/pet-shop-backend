package order

import (
	"fmt"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/wichananm65/pet-shop-backend/internal/product"
	"github.com/wichananm65/pet-shop-backend/internal/user"
)

// Handler delegates order operations to the order service.
// It also needs the user service to update user order lists.

type Handler struct {
	service        *Service
	userService    user.ServiceInterface
	productService product.ServiceInterface
}

func NewHandler(s *Service, us user.ServiceInterface, ps product.ServiceInterface) *Handler {
	return &Handler{service: s, userService: us, productService: ps}
}

func (h *Handler) RegisterProtectedRoutes(app *fiber.App) {
	app.Post("/api/v1/orders", h.createOrder)
	app.Get("/api/v1/orders", h.getOrders)
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

// getOrders returns all orders belonging to the currently authenticated user.
// It uses the user service to obtain the list of order IDs stored on the
// user row, then asks the order service for the matching orders.
func (h *Handler) getOrders(c *fiber.Ctx) error {
	userID, err := user.GetUserIDFromCtx(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "unauthorized"})
	}

	usr, err := h.userService.GetByID(userID)
	if err != nil {
		switch err {
		case user.ErrNotFound:
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "user not found"})
		default:
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": err.Error()})
		}
	}

	if len(usr.OrderIDs) == 0 {
		return c.JSON([]Order{})
	}

	orders, err := h.service.ListByIDs(usr.OrderIDs)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": err.Error()})
	}

	// enrich cart entries with product details if service available
	if h.productService != nil && len(orders) > 0 {
		// collect unique product IDs from all carts
		idSet := map[int]struct{}{}
		for _, ord := range orders {
			for pidStr := range ord.Cart {
				if id, err := strconv.Atoi(pidStr); err == nil {
					idSet[id] = struct{}{}
				}
			}
		}
		ids := make([]int, 0, len(idSet))
		for id := range idSet {
			ids = append(ids, id)
		}
		if len(ids) > 0 {
			prods, err2 := h.productService.ListV1ByIDs(ids)
			if err2 == nil {
				prodMap := map[string]product.ProductV1{}
				for _, p := range prods {
					prodMap[strconv.Itoa(p.ProductID)] = p
				}
				for i := range orders {
					orders[i].CartProducts = map[string]product.ProductV1{}
					for pidStr := range orders[i].Cart {
						if p, ok := prodMap[pidStr]; ok {
							orders[i].CartProducts[pidStr] = p
						}
					}
				}
			}
		}
	}

	return c.JSON(orders)
}

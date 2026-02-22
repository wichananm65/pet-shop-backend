package address

import (
    "fmt"

    "github.com/gofiber/fiber/v2"
    "github.com/wichananm65/pet-shop-backend/internal/user"
)

// Handler delegates address operations to the address service.

type Handler struct {
    service *Service
}

func NewHandler(s *Service) *Handler {
    return &Handler{service: s}
}

func (h *Handler) RegisterProtectedRoutes(app *fiber.App) {
    app.Get("/api/v1/address", h.getAddresses)
}

func (h *Handler) getAddresses(c *fiber.Ctx) error {
    fmt.Printf("[DEBUG] address.getAddresses invoked, remote=%s\n", c.IP())
    userID, err := user.GetUserIDFromCtx(c)
    if err != nil {
        return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "unauthorized"})
    }

    addrs, err := h.service.GetAddresses(userID)
    if err != nil {
        switch err {
        case ErrNotFound, user.ErrNotFound:
            return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "user not found"})
        default:
            return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": err.Error()})
        }
    }

    return c.JSON(addrs)
}

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
	app.Post("/api/v1/address", h.addAddress)
	app.Patch("/api/v1/address", h.updateAddress)
	app.Delete("/api/v1/address", h.deleteAddress)
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

// request payloads

type addressCreateRequest struct {
	AddressDesc string `json:"addressDesc"`
	Phone       string `json:"phone"`
	AddressName string `json:"addressName"`
}

type addressUpdateRequest struct {
	AddressID   int    `json:"addressId"`
	AddressDesc string `json:"addressDesc"`
	Phone       string `json:"phone"`
	AddressName string `json:"addressName"`
}

type addressDeleteRequest struct {
	AddressID int `json:"addressId"`
}

func (h *Handler) addAddress(c *fiber.Ctx) error {
	payload := new(addressCreateRequest)
	if err := c.BodyParser(payload); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": err.Error()})
	}
	if payload.AddressDesc == "" && payload.AddressName == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "addressDesc or addressName required"})
	}
	userID, err := user.GetUserIDFromCtx(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "unauthorized"})
	}

	addr, err := h.service.AddAddress(userID, payload.AddressDesc, payload.Phone, payload.AddressName)
	if err != nil {
		switch err {
		case ErrNotFound, user.ErrNotFound:
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "user not found"})
		default:
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": err.Error()})
		}
	}
	return c.Status(fiber.StatusOK).JSON(addr)
}

func (h *Handler) updateAddress(c *fiber.Ctx) error {
	payload := new(addressUpdateRequest)
	if err := c.BodyParser(payload); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": err.Error()})
	}
	if payload.AddressID <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "invalid addressId"})
	}
	if payload.AddressDesc == "" && payload.AddressName == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "addressDesc or addressName required"})
	}
	userID, err := user.GetUserIDFromCtx(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "unauthorized"})
	}
	addr, err := h.service.UpdateAddress(userID, payload.AddressID, payload.AddressDesc, payload.Phone, payload.AddressName)
	if err != nil {
		switch err {
		case ErrNotFound, user.ErrNotFound:
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "not found"})
		default:
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": err.Error()})
		}
	}
	return c.Status(fiber.StatusOK).JSON(addr)
}

func (h *Handler) deleteAddress(c *fiber.Ctx) error {
	payload := new(addressDeleteRequest)
	if err := c.BodyParser(payload); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": err.Error()})
	}
	if payload.AddressID <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "invalid addressId"})
	}
	userID, err := user.GetUserIDFromCtx(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "unauthorized"})
	}
	if err := h.service.DeleteAddress(userID, payload.AddressID); err != nil {
		switch err {
		case ErrNotFound, user.ErrNotFound:
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "not found"})
		default:
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": err.Error()})
		}
	}
	return c.SendStatus(fiber.StatusOK)
}

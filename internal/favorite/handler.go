package favorite

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/wichananm65/pet-shop-backend/internal/user"
)

// Handler delegates favorite operations to the favorite service.
// This keeps favorite-specific HTTP routing isolated from the user handler.
type Handler struct {
	service *Service
}

func NewHandler(s *Service) *Handler {
	return &Handler{service: s}
}

func (h *Handler) RegisterProtectedRoutes(app *fiber.App) {
	app.Get("/api/v1/favorites", h.getFavorites)
	app.Post("/api/v1/favorites", h.addFavorite)
	app.Delete("/api/v1/favorites", h.removeFavorite)
}

type favoriteRequest struct {
	ProductID int `json:"productId"`
}

func (h *Handler) addFavorite(c *fiber.Ctx) error {
	payload := new(favoriteRequest)
	if err := c.BodyParser(payload); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": err.Error()})
	}
	if payload.ProductID <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "invalid productId"})
	}
	userID, err := user.GetUserIDFromCtx(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "unauthorized"})
	}

	fav, err := h.service.AddFavorite(userID, payload.ProductID)
	if err != nil {
		switch err {
		case user.ErrNotFound:
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "user not found"})
		case ErrAlreadyFavorite:
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{"message": "product already in favorites"})
		default:
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": err.Error()})
		}
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"productId": payload.ProductID, "favoriteProductId": fav})
}

func (h *Handler) removeFavorite(c *fiber.Ctx) error {
	payload := new(favoriteRequest)
	if err := c.BodyParser(payload); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": err.Error()})
	}
	if payload.ProductID <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "invalid productId"})
	}
	userID, err := user.GetUserIDFromCtx(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "unauthorized"})
	}

	fav, err := h.service.RemoveFavorite(userID, payload.ProductID)
	if err != nil {
		switch err {
		case user.ErrNotFound:
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "user not found"})
		case ErrNotFavorite:
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "product not in favorites"})
		default:
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": err.Error()})
		}
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"productId": payload.ProductID, "favoriteProductId": fav})
}

func (h *Handler) getFavorites(c *fiber.Ctx) error {
	fmt.Printf("[DEBUG] favorite.getFavorites invoked, remote=%s\n", c.IP())
	userID, err := user.GetUserIDFromCtx(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "unauthorized"})
	}

	favs, err := h.service.GetFavorites(userID)
	if err != nil {
		switch err {
		case user.ErrNotFound:
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "user not found"})
		default:
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": err.Error()})
		}
	}

	return c.JSON(favs)
}

package product

import (
	"os"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) RegisterPublicRoutes(app *fiber.App) {
	app.Get("/products", h.getProducts)
	app.Get("/product/:id", h.getProduct)

	// dev-only endpoint to reset products — enabled when ALLOW_RESET_PRODUCTS=1
	app.Post("/dev/reset-products", h.resetProducts)
}

func (h *Handler) RegisterProtectedRoutes(app *fiber.App) {
	app.Post("/products", h.createProduct)
	app.Put("/product/:id", h.updateProduct)
	app.Delete("/product/:id", h.deleteProduct)
}

func (h *Handler) getProducts(c *fiber.Ctx) error {
	products := h.service.List()
	return c.JSON(products)
}

func (h *Handler) getProduct(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).SendString(err.Error())
	}

	p, err := h.service.GetByID(id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).SendString("Product not found")
	}
	return c.JSON(p)
}

// resetProducts clears the product table and inserts the provided list (or a default sample list).
// This endpoint is protected by ALLOW_RESET_PRODUCTS environment variable; set it to "1" to allow.
func (h *Handler) resetProducts(c *fiber.Ctx) error {
	if os.Getenv("ALLOW_RESET_PRODUCTS") != "1" {
		return c.Status(fiber.StatusForbidden).SendString("reset not allowed")
	}

	var products []Product
	err := c.BodyParser(&products)
	now := time.Now().UTC().Format(time.RFC3339)
	// If body parsing fails, fallback to default sample products.
	// If parsing succeeds and client sends an empty array, treat it as "delete all" (no re-seeding).
	if err != nil {
		// fallback: use default 4 sample products
		sample := []Product{
			{
				Name:        "Cat Scratcher Bed",
				Description: "Comfortable cardboard cat bed",
				Price:       840,
				Score:       5,
				Category:    ptrString("Pet Supplies"),
				Pic:         ptrString("/shopping/cat-bed.svg"),
				CreatedAt:   &now,
				UpdatedAt:   &now,
			},
			{
				Name:        "Double Food Bowl",
				Description: "Wooden elevated double food bowl",
				Price:       420,
				Score:       5,
				Category:    ptrString("Pet Supplies"),
				Pic:         ptrString("/shopping/double-bowl.svg"),
				CreatedAt:   &now,
				UpdatedAt:   &now,
			},
			{
				Name:        "Cat Sweater",
				Description: "Warm knitted cat sweater",
				Price:       260,
				Score:       4,
				Category:    ptrString("Clothes and accessories"),
				Pic:         ptrString("/shopping/cat-sweater.svg"),
				CreatedAt:   &now,
				UpdatedAt:   &now,
			},
			{
				Name:        "Cheese Cat House",
				Description: "Cute cardboard cat house",
				Price:       399,
				Score:       5,
				Category:    ptrString("Cat exercise"),
				Pic:         ptrString("/shopping/cheese-house.svg"),
				CreatedAt:   &now,
				UpdatedAt:   &now,
			},
		}
		products = sample
	}

	// call ResetProducts — an empty `products` slice will now clear the table without inserting rows.
	if err := h.service.ResetProducts(products); err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
	}
	return c.JSON(products)
}

func validateProductPayload(p *Product) map[string]string {
	errs := map[string]string{}
	if p.Name == "" {
		errs["productName"] = "productName is required"
	}
	if p.Price < 0 {
		errs["productPrice"] = "productPrice must be >= 0"
	}
	if p.Score < 0 || p.Score > 5 {
		errs["score"] = "score must be between 0 and 5"
	}
	if p.Category != nil {
		valid := false
		for _, c := range AllowedCategories {
			if *p.Category == c {
				valid = true
				break
			}
		}
		if !valid {
			errs["category"] = "invalid category"
		}
	}
	return errs
}

func ptrString(s string) *string { return &s }

func (h *Handler) createProduct(c *fiber.Ctx) error {
	p := new(Product)
	if err := c.BodyParser(p); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": err.Error()})
	}

	// validate payload and return all validation errors together
	if ves := validateProductPayload(p); len(ves) > 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"errors": ves})
	}

	if p.CreatedAt == nil {
		now := time.Now().UTC().Format(time.RFC3339)
		p.CreatedAt = &now
	}
	if p.UpdatedAt == nil {
		now := time.Now().UTC().Format(time.RFC3339)
		p.UpdatedAt = &now
	}

	created, err := h.service.Create(*p)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": err.Error()})
	}
	return c.Status(fiber.StatusCreated).JSON(created)
}

func (h *Handler) updateProduct(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": err.Error()})
	}

	p := new(Product)
	if err := c.BodyParser(p); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": err.Error()})
	}

	// validate payload before attempting update
	if ves := validateProductPayload(p); len(ves) > 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"errors": ves})
	}

	now := time.Now().UTC().Format(time.RFC3339)
	p.UpdatedAt = &now

	updated, err := h.service.Update(id, *p)
	if err != nil {
		return c.Status(fiber.StatusNotFound).SendString("Product not found")
	}
	return c.JSON(updated)
}

func (h *Handler) deleteProduct(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).SendString(err.Error())
	}
	if err := h.service.Delete(id); err != nil {
		return c.Status(fiber.StatusNotFound).SendString("Product not found")
	}
	return c.SendString("Product deleted")
}

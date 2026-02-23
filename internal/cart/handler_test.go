package cart

import (
	"io"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
	"github.com/wichananm65/pet-shop-backend/internal/user"
)

func makeAppWithCartHandler(cHandler *Handler) *fiber.App {
	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		if v := c.Get("X-User-ID"); v != "" {
			id, err := strconv.Atoi(v)
			if err == nil {
				claims := jwt.MapClaims{"user_id": id}
				tok := &jwt.Token{Claims: claims}
				c.Locals("user", tok)
			}
		}
		return c.Next()
	})
	cHandler.RegisterProtectedRoutes(app)
	return app
}

func TestCartRoutes_Basic(t *testing.T) {
	seed := []user.User{{ID: 42, Cart: map[int]int{1: 1}}}
	repo := NewInMemoryRepository(seed)
	service := NewService(repo)
	handler := NewHandler(service)
	app := makeAppWithCartHandler(handler)

	// ensure routes registered
	routes := map[string]bool{}
	for _, grp := range app.Stack() {
		for _, r := range grp {
			routes[r.Path] = true
		}
	}
	if !routes["/api/v1/cart"] {
		t.Fatalf("expected route '/api/v1/cart' to be registered")
	}
	if !routes["/api/v1/product/cart"] {
		t.Fatalf("expected route '/api/v1/product/cart' to be registered")
	}

	// unauthorized access should be blocked
	req := httptest.NewRequest("GET", "/api/v1/cart", nil)
	res, _ := app.Test(req)
	if res.StatusCode != fiber.StatusUnauthorized {
		t.Fatalf("expected 401 for unauthenticated GET, got %d", res.StatusCode)
	}
	req2 := httptest.NewRequest("POST", "/api/v1/product/cart", strings.NewReader(`{"productID":2}`))
	req2.Header.Set("Content-Type", "application/json")
	res2, _ := app.Test(req2)
	if res2.StatusCode != fiber.StatusUnauthorized {
		t.Fatalf("expected 401 for unauthenticated POST, got %d", res2.StatusCode)
	}

	// authorized GET should succeed and return JSON
	req3 := httptest.NewRequest("GET", "/api/v1/cart", nil)
	req3.Header.Set("X-User-ID", "42")
	res3, _ := app.Test(req3)
	if res3.StatusCode != fiber.StatusOK {
		t.Fatalf("expected 200 for authenticated GET, got %d", res3.StatusCode)
	}

	// authorized POST add product with explicit quantity=2
	req4 := httptest.NewRequest("POST", "/api/v1/product/cart", strings.NewReader(`{"productID":3,"quantity":2}`))
	req4.Header.Set("Content-Type", "application/json")
	req4.Header.Set("X-User-ID", "42")
	res4, _ := app.Test(req4)
	if res4.StatusCode != fiber.StatusOK {
		t.Fatalf("expected 200 for adding to cart, got %d", res4.StatusCode)
	}
	// verify response body has items and correct quantity
	b, _ := io.ReadAll(res4.Body)
	if !strings.Contains(string(b), "quantity") {
		t.Fatalf("response missing quantity field: %s", string(b))
	}

	// add same product again, should increment quantity
	req5 := httptest.NewRequest("POST", "/api/v1/product/cart", strings.NewReader(`{"productID":3,"quantity":1}`))
	req5.Header.Set("Content-Type", "application/json")
	req5.Header.Set("X-User-ID", "42")
	res5, _ := app.Test(req5)
	if res5.StatusCode != fiber.StatusOK {
		t.Fatalf("expected 200 for second add, got %d", res5.StatusCode)
	}
	b5, _ := io.ReadAll(res5.Body)
	if !strings.Contains(string(b5), `"quantity":3`) {
		t.Fatalf("expected quantity 3 after second add, got %s", string(b5))
	}

	// decrease quantity by one using negative quantity
	req6 := httptest.NewRequest("POST", "/api/v1/product/cart", strings.NewReader(`{"productID":3,"quantity":-1}`))
	req6.Header.Set("Content-Type", "application/json")
	req6.Header.Set("X-User-ID", "42")
	res6, _ := app.Test(req6)
	if res6.StatusCode != fiber.StatusOK {
		t.Fatalf("expected 200 for decrement, got %d", res6.StatusCode)
	}
	b6, _ := io.ReadAll(res6.Body)
	if !strings.Contains(string(b6), `"quantity":2`) {
		t.Fatalf("expected quantity 2 after decrement, got %s", string(b6))
	}

	// reduce to zero and ensure item removed
	req7 := httptest.NewRequest("POST", "/api/v1/product/cart", strings.NewReader(`{"productID":3,"quantity":-2}`))
	req7.Header.Set("Content-Type", "application/json")
	req7.Header.Set("X-User-ID", "42")
	res7, _ := app.Test(req7)
	if res7.StatusCode != fiber.StatusOK {
		t.Fatalf("expected 200 for remove, got %d", res7.StatusCode)
	}
	b7, _ := io.ReadAll(res7.Body)
	if strings.Contains(string(b7), `"productID":3`) {
		t.Fatalf("expected product 3 to be removed after quantity zero, got %s", string(b7))
	}

	// clear the cart via DELETE endpoint
	req8 := httptest.NewRequest("DELETE", "/api/v1/cart", nil)
	req8.Header.Set("X-User-ID", "42")
	res8, _ := app.Test(req8)
	if res8.StatusCode != fiber.StatusNoContent {
		t.Fatalf("expected 204 for clear cart, got %d", res8.StatusCode)
	}
	// after clearing, GET should return empty
	req9 := httptest.NewRequest("GET", "/api/v1/cart", nil)
	req9.Header.Set("X-User-ID", "42")
	res9, _ := app.Test(req9)
	if res9.StatusCode != fiber.StatusOK {
		t.Fatalf("expected 200 after clearing, got %d", res9.StatusCode)
	}
	b9, _ := io.ReadAll(res9.Body)
	if strings.Contains(string(b9), "productID") {
		t.Fatalf("expected empty cart after clear, got %s", string(b9))
	}
}

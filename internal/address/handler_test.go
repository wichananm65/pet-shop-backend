package address

import (
	"io"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
)

func makeAppWithAddressHandler(a *Handler) *fiber.App {
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
	a.RegisterProtectedRoutes(app)
	return app
}

func TestAddressRoute(t *testing.T) {
	seed := map[int][]Address{
		42: {{AddressID: 1, UserID: 42, AddressDesc: "123 Main", Phone: "555-1234", AddressName: "Home"}},
	}
	repo := NewInMemoryRepository(seed)
	svc := NewService(repo)
	handler := NewHandler(svc)
	app := makeAppWithAddressHandler(handler)

	// route exists
	routes := map[string]bool{}
	for _, grp := range app.Stack() {
		for _, r := range grp {
			routes[r.Path] = true
		}
	}
	if !routes["/api/v1/address"] {
		t.Fatalf("expected /api/v1/address registered")
	}

	// unauthorized
	req := httptest.NewRequest("GET", "/api/v1/address", nil)
	res, _ := app.Test(req)
	if res.StatusCode != fiber.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", res.StatusCode)
	}

	// authorized GET returns existing
	req2 := httptest.NewRequest("GET", "/api/v1/address", nil)
	req2.Header.Set("X-User-ID", "42")
	res2, _ := app.Test(req2)
	if res2.StatusCode != fiber.StatusOK {
		t.Fatalf("expected 200, got %d", res2.StatusCode)
	}
	b, _ := io.ReadAll(res2.Body)
	if !strings.Contains(string(b), "addressDesc") {
		t.Fatalf("unexpected body: %s", string(b))
	}

	// POST new address
	req3 := httptest.NewRequest("POST", "/api/v1/address", strings.NewReader(`{"addressDesc":"foo","phone":"123"}`))
	req3.Header.Set("Content-Type", "application/json")
	req3.Header.Set("X-User-ID", "42")
	res3, _ := app.Test(req3)
	if res3.StatusCode != fiber.StatusOK {
		t.Fatalf("expected 200 for add, got %d", res3.StatusCode)
	}
	b3, _ := io.ReadAll(res3.Body)
	if !strings.Contains(string(b3), "foo") {
		t.Fatalf("add response unexpected: %s", string(b3))
	}
	// parse returned id
	// update with patch
	req4 := httptest.NewRequest("PATCH", "/api/v1/address", strings.NewReader(`{"addressId":2,"addressDesc":"bar"}`))
	req4.Header.Set("Content-Type", "application/json")
	req4.Header.Set("X-User-ID", "42")
	res4, _ := app.Test(req4)
	if res4.StatusCode != fiber.StatusOK {
		t.Fatalf("expected 200 for patch, got %d", res4.StatusCode)
	}
	b4, _ := io.ReadAll(res4.Body)
	if !strings.Contains(string(b4), "bar") {
		t.Fatalf("patch response unexpected: %s", string(b4))
	}

	// delete the newly added address
	req5 := httptest.NewRequest("DELETE", "/api/v1/address", strings.NewReader(`{"addressId":2}`))
	req5.Header.Set("Content-Type", "application/json")
	req5.Header.Set("X-User-ID", "42")
	res5, _ := app.Test(req5)
	if res5.StatusCode != fiber.StatusOK {
		t.Fatalf("expected 200 for delete, got %d", res5.StatusCode)
	}
	// confirm gone by GET
	req6 := httptest.NewRequest("GET", "/api/v1/address", nil)
	req6.Header.Set("X-User-ID", "42")
	res6, _ := app.Test(req6)
	b6, _ := io.ReadAll(res6.Body)
	if strings.Contains(string(b6), "bar") {
		t.Fatalf("delete did not remove entry: %s", string(b6))
	}
}

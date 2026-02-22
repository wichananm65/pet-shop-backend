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

    // authorized
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
}

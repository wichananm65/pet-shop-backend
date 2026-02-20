package product

import (
	"io"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/wichananm65/pet-shop-backend/internal/favorite"
	"github.com/wichananm65/pet-shop-backend/internal/user"
)

func TestProductV1AndFavoriteRoutes_DoNotCollide(t *testing.T) {
	// prepare in-memory repositories and handlers
	prodSeed := []Product{{ID: 12, Name: "Cat Sweater", Price: 260, Score: 4, Description: "Warm knitted cat sweater", Pic: ptrString("/api/v1/product/12/image")}}
	pRepo := NewInMemoryRepository(prodSeed)
	pHandler := NewHandler(NewService(pRepo))

	userSeed := []user.User{{ID: 1, Email: "u@example.com", Password: "pass", FavoriteProductIDs: []int{12}}}
	favRepo := favorite.NewInMemoryRepository(userSeed)
	favService := favorite.NewService(favRepo)
	favHandler := favorite.NewHandler(favService)

	app := fiber.New()
	// register both handlers on the same app (as in main.go)
	pHandler.RegisterPublicRoutes(app)
	favHandler.RegisterProtectedRoutes(app)

	// 1) route registration check: both routes must exist and be distinct
	routes := map[string]bool{}
	for _, grp := range app.Stack() {
		for _, r := range grp {
			routes[r.Path] = true
		}
	}

	if !routes["/api/v1/product/:id<[0-9]+>"] {
		t.Fatalf("expected route '/api/v1/product/:id<[0-9]+>' to be registered")
	}
	if !routes["/api/v1/favorites"] {
		t.Fatalf("expected route '/api/v1/favorites' to be registered")
	}

	// 2) endpoint behavior check: numeric product returns JSON; favorite does NOT return product JSON
	// product detail (public)
	req := httptest.NewRequest("GET", "/api/v1/product/12", nil)
	res, err := app.Test(req)
	if err != nil {
		t.Fatalf("product request failed: %v", err)
	}
	if res.StatusCode != 200 {
		t.Fatalf("expected product handler to return 200, got %d", res.StatusCode)
	}

	// favorites (protected) — without JWT this should not return product JSON; expect unauthorized (401)
	req2 := httptest.NewRequest("GET", "/api/v1/favorites", nil)
	res2, err := app.Test(req2)
	if err != nil {
		t.Fatalf("favorite request failed: %v", err)
	}
	// Accept either 401 (unauthorized) or 302/307 (redirect) but definitely NOT 200 product JSON
	if res2.StatusCode == 200 {
		// read body to make sure it's not product JSON
		b, _ := io.ReadAll(res2.Body)
		body := string(b)
		if strings.Contains(body, "productID") {
			t.Fatalf("favorite route appears to be handled by product handler (body contains product data)")
		}
	}
}

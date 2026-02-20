package product

import (
	"testing"

	"github.com/gofiber/fiber/v2"
)

// Ensure product handler does NOT register the literal /api/v1/favorites path.
func TestProductHandler_DoesNotRegisterFavoriteRoute(t *testing.T) {
	pHandler := NewHandler(NewService(NewInMemoryRepository(nil)))
	app := fiber.New()
	pHandler.RegisterPublicRoutes(app)

	routes := map[string]bool{}
	for _, grp := range app.Stack() {
		for _, r := range grp {
			routes[r.Path] = true
		}
	}

	if routes["/api/v1/favorites"] {
		t.Fatalf("product handler must not register '/api/v1/favorites' route")
	}
}

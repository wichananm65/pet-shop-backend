package order

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
	"github.com/wichananm65/pet-shop-backend/internal/product"
	"github.com/wichananm65/pet-shop-backend/internal/user"
)

// dummy repo implementing Repository
type dummyRepo struct{}

// implement ListByIDs for retrieving orders in tests
func (r *dummyRepo) ListByIDs(ids []int) ([]Order, error) {
	orders := make([]Order, 0, len(ids))
	for _, id := range ids {
		orders = append(orders, Order{OrderID: id, Cart: map[string]int{"1": 1}, Quantity: 1, TotalPrice: 10, ShippingPrice: 2, GrandPrice: 12})
	}
	return orders, nil
}

// dummy user service with AppendOrderID stub
// import user to satisfy type

type dummyUserService struct{}

// dummy product service for testing
// returns a simple product for any requested id

type dummyProductService struct{}

func (d *dummyProductService) List() []product.Product { return nil }
func (d *dummyProductService) GetByID(id int) (product.Product, error) {
	return product.Product{ID: id, Name: "p"}, nil
}
func (d *dummyProductService) GetV1ByID(id int) (product.ProductV1, error) {
	return product.ProductV1{ProductID: id, ProductName: ptrString("p"), ProductPrice: ptrInt(10)}, nil
}
func (d *dummyProductService) ListV1ByIDs(ids []int) ([]product.ProductV1, error) {
	out := make([]product.ProductV1, 0, len(ids))
	for _, id := range ids {
		out = append(out, product.ProductV1{ProductID: id, ProductName: ptrString("p"), ProductPrice: ptrInt(10)})
	}
	return out, nil
}
func (d *dummyProductService) Create(p product.Product) (product.Product, error) { return p, nil }
func (d *dummyProductService) Update(id int, p product.Product) (product.Product, error) {
	return p, nil
}
func (d *dummyProductService) Delete(id int) error                            { return nil }
func (d *dummyProductService) ListByCategoryID(catID int) []product.Product   { return nil }
func (d *dummyProductService) ResetProducts(products []product.Product) error { return nil }

func ptrString(s string) *string { return &s }
func ptrInt(i int) *int          { return &i }

func (d *dummyUserService) List() []user.User {
	return []user.User{}
}

func (d *dummyUserService) GetByID(id int) (user.User, error) {
	// return a user that already has one order in its list
	return user.User{ID: id, OrderIDs: []int{123}}, nil
}

func (d *dummyUserService) Create(u user.User) (user.User, error) {
	return u, nil
}

func (d *dummyUserService) Update(id int, u user.User) (user.User, error) {
	return u, nil
}

func (d *dummyUserService) Delete(id int) error {
	return nil
}

func (d *dummyUserService) Register(u user.User) (user.User, error) {
	return u, nil
}

func (d *dummyUserService) Authenticate(email, password string) (user.User, error) {
	return user.User{Email: email}, nil
}

func (d *dummyUserService) AppendOrderID(userID int, orderID int) (user.User, error) {
	return user.User{ID: userID}, nil
}

// Ensure dummyUserService implements user.ServiceInterface
var _ user.ServiceInterface = (*dummyUserService)(nil)

func (r *dummyRepo) Create(ord Order) (Order, error) {
	ord.OrderID = 123
	return ord, nil
}

// makeAppWithAuth returns an app wired with the order handler plus a
// tiny piece of middleware that emulates JWT parsing by reading an
// "X-User-ID" header and populating c.Locals("user").  This mirrors the
// pattern used in other package tests such as cart/handler_test.go.
func makeAppWithAuth() *fiber.App {
	a := fiber.New()
	a.Use(func(c *fiber.Ctx) error {
		if v := c.Get("X-User-ID"); v != "" {
			// simple conversion, ignore error for test simplicity
			var id int
			fmt.Sscanf(v, "%d", &id)
			claims := jwt.MapClaims{"user_id": id}
			tok := &jwt.Token{Claims: claims}
			c.Locals("user", tok)
		}
		return c.Next()
	})
	prdService := &dummyProductService{}
	h := NewHandler(NewService(&dummyRepo{}), &dummyUserService{}, prdService)
	h.RegisterProtectedRoutes(a)
	return a
}

func TestCreateOrder_Success(t *testing.T) {
	a := makeAppWithAuth()

	reqBody := map[string]interface{}{
		"cart":          map[string]int{"1": 2},
		"quantity":      2,
		"totalPrice":    100.0,
		"shippingPrice": 10.0,
		"grandPrice":    110.0,
	}
	b, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/api/v1/orders", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-User-ID", "42")

	res, err := a.Test(req, -1)
	if err != nil {
		t.Fatal(err)
	}
	if res.StatusCode != 200 {
		t.Fatalf("expected 200 got %d", res.StatusCode)
	}

	var ord Order
	json.NewDecoder(res.Body).Decode(&ord)
	if ord.OrderID != 123 {
		t.Errorf("expected orderID 123, got %d", ord.OrderID)
	}
	if len(ord.CartProducts) != 0 {
		t.Errorf("expected no cartProducts on created order, got %+v", ord.CartProducts)
	}
}

func TestGetOrders_Success(t *testing.T) {
	a := makeAppWithAuth()

	req := httptest.NewRequest("GET", "/api/v1/orders", nil)
	req.Header.Set("X-User-ID", "42")

	res, err := a.Test(req, -1)
	if err != nil {
		t.Fatal(err)
	}
	if res.StatusCode != 200 {
		t.Fatalf("expected 200 got %d", res.StatusCode)
	}

	var orders []Order
	json.NewDecoder(res.Body).Decode(&orders)
	if len(orders) != 1 || orders[0].OrderID != 123 {
		t.Errorf("expected one order with ID 123, got %+v", orders)
	}
	if qp, ok := orders[0].CartProducts["1"]; !ok || qp.ProductID != 1 {
		t.Errorf("cart products not populated: %+v", orders[0].CartProducts)
	}
}

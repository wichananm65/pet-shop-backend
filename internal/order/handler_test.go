package order

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/wichananm65/pet-shop-backend/internal/user"
)

// dummy repo implementing Repository
type dummyRepo struct{}

// dummy user service with AppendOrderID stub
// import user to satisfy type

type dummyUserService struct{}

func (d *dummyUserService) List() []user.User {
	return []user.User{}
}

func (d *dummyUserService) GetByID(id int) (user.User, error) {
	return user.User{ID: id}, nil
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

func setupApp() *fiber.App {
	a := fiber.New()
	h := NewHandler(NewService(&dummyRepo{}), &dummyUserService{})
	h.RegisterProtectedRoutes(a)
	return a
}

func TestCreateOrder_Success(t *testing.T) {
	a := setupApp()

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

	// attach dummy userID into context via middleware-like hack
	req.Header.Set("Authorization", "Bearer dummy")
	// our GetUserIDFromCtx will fail; skip auth in test by overriding

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
}

package order

import "github.com/wichananm65/pet-shop-backend/internal/product"

// Order represents a purchase made by a user.
type Order struct {
	OrderID       int                          `json:"orderID"`
	UserID        int                          `json:"userID,omitempty"`
	Cart          map[string]int               `json:"cart"`
	CartProducts  map[string]product.ProductV1 `json:"cartProducts,omitempty"`
	Quantity      int                          `json:"quantity"`
	TotalPrice    float64                      `json:"totalPrice"`
	ShippingPrice float64                      `json:"shippingPrice"`
	GrandPrice    float64                      `json:"grandPrice"`
	Status        string                       `json:"status"`
	CreatedAt     string                       `json:"createdAt"`
	UpdatedAt     string                       `json:"updatedAt"`
}

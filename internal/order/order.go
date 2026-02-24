package order

import "github.com/wichananm65/pet-shop-backend/internal/product"

// Order represents a purchase made by a user.
// The Cart field holds a mapping from productID to quantity; when the
// API returns orders we also include an optional CartProducts field that
// contains lookup information about each product.  The field is left out
// when creating an order.
type Order struct {
	OrderID       int                          `json:"orderID"`
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

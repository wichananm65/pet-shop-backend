package order

// Order represents a purchase made by a user.
type Order struct {
    OrderID       int            `json:"orderID"`
    Cart          map[string]int `json:"cart"`
    Quantity      int            `json:"quantity"`
    TotalPrice    float64        `json:"totalPrice"`
    ShippingPrice float64        `json:"shippingPrice"`
    GrandPrice    float64        `json:"grandPrice"`
    Status        string         `json:"status"`
    CreatedAt     string         `json:"createdAt"`
    UpdatedAt     string         `json:"updatedAt"`
}

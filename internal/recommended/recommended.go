package recommended

// RecommendedItem is the public DTO returned by the recommended API.
// Response fields use the `products`-style names requested by the client.
type RecommendedItem struct {
	ProductID     int     `json:"productID"`
	ProductImg    *string `json:"productImg,omitempty"`
	ProductName   *string `json:"productName,omitempty"`   // English or primary name
	ProductNameTH *string `json:"productNameTH,omitempty"` // Thai / localized name
	ProductPrice  *int    `json:"productPrice,omitempty"`
	Score         *int    `json:"score,omitempty"`
}

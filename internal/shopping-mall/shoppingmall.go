package shoppingmall

// ShoppingMallItem is the detailed DTO returned by GET /api/v1/shopping-mall
// JSON tags use camelCase to match the frontend.
type ShoppingMallItem struct {
	ProductID     int     `json:"productID"`
	ProductImg    *string `json:"productImg,omitempty"`
	Price         *int    `json:"price,omitempty"`
	Score         *int    `json:"score,omitempty"`
	ProductName   *string `json:"productName,omitempty"`
	ProductNameTH *string `json:"productNameTH,omitempty"`
}

// LiteItem is the lightweight DTO returned by GET /api/v1/product/shopping-mall
type LiteItem struct {
	ProductId  int     `json:"productId"`
	ProductPic *string `json:"productPic,omitempty"`
}

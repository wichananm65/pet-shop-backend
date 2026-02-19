package product

// Product represents a product in the system and maps to the `public.product` table.
// JSON tags follow the camelCase convention used elsewhere in the project.
type Product struct {
	ID            int     `json:"productId"`
	Name          string  `json:"productName"`
	NameEn        *string `json:"productNameEn,omitempty"`
	Price         int     `json:"productPrice"`
	Score         int     `json:"score"`
	Description   string  `json:"productDesc"`
	DescriptionEn *string `json:"productDescEn,omitempty"`
	Category      *string `json:"category,omitempty"`
	Pic           *string `json:"productPic,omitempty"`
	PicSecond     *string `json:"productPicSecond,omitempty"`
	CreatedAt     *string `json:"createdAt,omitempty"`
	UpdatedAt     *string `json:"updatedAt,omitempty"`
}

// ProductV1 is the API v1 product detail shape (used by `/api/v1/product/:id`).
// Field names follow the `products`-style contract used by other v1 endpoints.
type ProductV1 struct {
	ProductID     int     `json:"productID"`
	ProductName   *string `json:"productName,omitempty"`
	ProductNameTH *string `json:"productNameTH,omitempty"`
	ProductPrice  *int    `json:"productPrice,omitempty"`
	ProductImg    *string `json:"productImg,omitempty"`
	ProductDesc   *string `json:"productDesc,omitempty"`
	ProductDescTH *string `json:"productDescTH,omitempty"`
	Score         *int    `json:"score,omitempty"`
	Category      *string `json:"category,omitempty"`
}

// AllowedCategories contains the supported product categories used across the app.
var AllowedCategories = []string{
	"Animal Food",
	"Pet Supplies",
	"Clothes and accessories",
	"Cleaning equipment",
	"Sand and bathroom",
	"Hygiene care",
	"Cat snacks",
	"Cat exercise",
}

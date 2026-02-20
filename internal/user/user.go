package user

type FavoriteProduct struct {
	ProductID     int     `json:"productID"`
	ProductName   *string `json:"productName,omitempty"`
	ProductNameTH *string `json:"productNameTH,omitempty"`
	ProductDesc   *string `json:"productDesc,omitempty"`
	ProductDescTH *string `json:"productDescTH,omitempty"`
	ProductPrice  *int    `json:"productPrice,omitempty"`
	ProductImg    *string `json:"productImg,omitempty"`
	Score         *int    `json:"score,omitempty"`
}

type User struct {
	ID            int    `json:"userId"`
	Email         string `json:"email"`
	Password      string `json:"password,omitempty"`
	FirstName     string `json:"firstName"`
	LastName      string `json:"lastName"`
	Phone         string `json:"phone"`
	Gender        string `json:"gender"`
	MainAddressID *int   `json:"mainAddressId,omitempty"`
	AddressIDs    []int  `json:"addressId,omitempty"`

	OrderIDs           []int   `json:"orderId,omitempty"`
	FavoriteProductIDs []int   `json:"favoriteProductId,omitempty"`
	AvatarPic          *string `json:"avatarPic,omitempty"`
	CreatedAt          string  `json:"createAt,omitempty"`
	UpdatedAt          string  `json:"updateAt,omitempty"`
}

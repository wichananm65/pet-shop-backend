package user

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

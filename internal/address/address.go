package address

type Address struct {
    AddressID   int    `json:"addressId"`
    UserID      int    `json:"userId"`
    AddressDesc string `json:"addressDesc"`
    Phone       string `json:"phone"`
    AddressName string `json:"addressName"` // new column as requested
    CreatedAt   string `json:"createdAt,omitempty"`
    UpdatedAt   string `json:"updatedAt,omitempty"`
}

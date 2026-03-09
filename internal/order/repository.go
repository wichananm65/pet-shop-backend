package order

// Repository defines persistence operations for orders.
type Repository interface {
	Create(ord Order, userID int) (Order, error)
	ListByIDs(ids []int) ([]Order, error)
	ListByUserID(userID int) ([]Order, error)
}

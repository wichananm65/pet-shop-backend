package order

// Repository defines persistence operations for orders.
type Repository interface {
	Create(ord Order) (Order, error)
	// additional methods like ListByUser could be added later
}

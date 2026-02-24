package order

// Repository defines persistence operations for orders.
type Repository interface {
	Create(ord Order) (Order, error)
	// additional methods like ListByUser could be added later

	// ListByIDs returns the orders whose orderID is present in the
	// provided slice.  The returned slice is ordered the same way as
	// the ids argument.  If ids is empty, the implementation should
	// return an empty slice without performing a database query.
	ListByIDs(ids []int) ([]Order, error)
}

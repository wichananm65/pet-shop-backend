package order

import (
	"database/sql"
	"encoding/json"

	"github.com/lib/pq"
)

type PostgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) Create(ord Order) (Order, error) {
	cartJSON, err := json.Marshal(ord.Cart)
	if err != nil {
		return Order{}, err
	}

	err = r.db.QueryRow(`INSERT INTO orders (cart, quantity, "totalPrice", "shippingPrice", "grandPrice", status, "createdAt", "updatedAt")
        VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
        RETURNING "orderID", cart, quantity, "totalPrice", "shippingPrice", "grandPrice", status, "createdAt", "updatedAt"`,
		cartJSON, ord.Quantity, ord.TotalPrice, ord.ShippingPrice, ord.GrandPrice, ord.Status, ord.CreatedAt, ord.UpdatedAt).Scan(
		&ord.OrderID, &cartJSON, &ord.Quantity, &ord.TotalPrice, &ord.ShippingPrice, &ord.GrandPrice, &ord.Status, &ord.CreatedAt, &ord.UpdatedAt)
	if err != nil {
		return Order{}, err
	}

	// unmarshal stored cart back
	json.Unmarshal(cartJSON, &ord.Cart)
	return ord, nil
}

// ListByIDs returns orders matching the given orderIDs.  The results are
// ordered according to the sequence of ids in the slice.  An empty slice
// leads to an immediate empty result.
func (r *PostgresRepository) ListByIDs(ids []int) ([]Order, error) {
	if len(ids) == 0 {
		return []Order{}, nil
	}

	query := `SELECT "orderID", cart, quantity, "totalPrice", "shippingPrice", "grandPrice", status, "createdAt", "updatedAt"
		FROM orders
		WHERE "orderID" = ANY($1::int[])
		ORDER BY array_position($1::int[], "orderID")`

	rows, err := r.db.Query(query, pq.Array(ids))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	orders := make([]Order, 0)
	for rows.Next() {
		var ord Order
		var cartJSON []byte
		if err := rows.Scan(&ord.OrderID, &cartJSON, &ord.Quantity, &ord.TotalPrice, &ord.ShippingPrice, &ord.GrandPrice, &ord.Status, &ord.CreatedAt, &ord.UpdatedAt); err != nil {
			return nil, err
		}
		json.Unmarshal(cartJSON, &ord.Cart)
		orders = append(orders, ord)
	}

	return orders, nil
}

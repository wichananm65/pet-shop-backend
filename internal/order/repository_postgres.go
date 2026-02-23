package order

import (
	"database/sql"
	"encoding/json"
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

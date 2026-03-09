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

func (r *PostgresRepository) Create(ord Order, userID int) (Order, error) {
	cartJSON, err := json.Marshal(ord.Cart)
	if err != nil {
		return Order{}, err
	}

	var (
		cartRaw []byte
		status  sql.NullString
	)
	err = r.db.QueryRow(
		`INSERT INTO orders ("userID", cart, quantity, "totalPrice", "shippingPrice", "grandPrice", status, "createdAt", "updatedAt")
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
		 RETURNING "orderID", cart, quantity, "totalPrice", "shippingPrice", "grandPrice", status, "createdAt", "updatedAt"`,
		userID, cartJSON, ord.Quantity, ord.TotalPrice, ord.ShippingPrice, ord.GrandPrice,
		ord.Status, ord.CreatedAt, ord.UpdatedAt,
	).Scan(
		&ord.OrderID, &cartRaw, &ord.Quantity,
		&ord.TotalPrice, &ord.ShippingPrice, &ord.GrandPrice,
		&status, &ord.CreatedAt, &ord.UpdatedAt,
	)
	if err != nil {
		return Order{}, err
	}
	if status.Valid {
		ord.Status = status.String
	}
	if len(cartRaw) > 0 {
		_ = json.Unmarshal(cartRaw, &ord.Cart)
	}
	ord.UserID = userID
	return ord, nil
}

func (r *PostgresRepository) ListByIDs(ids []int) ([]Order, error) {
	if len(ids) == 0 {
		return []Order{}, nil
	}

	rows, err := r.db.Query(
		`SELECT "orderID", "userID", cart, quantity, "totalPrice", "shippingPrice", "grandPrice", status, "createdAt", "updatedAt"
		 FROM orders
		 WHERE "orderID" = ANY($1::int[])
		 ORDER BY array_position($1::int[], "orderID")`,
		pq.Array(ids),
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanOrders(rows)
}

func (r *PostgresRepository) ListByUserID(userID int) ([]Order, error) {
	rows, err := r.db.Query(
		`SELECT "orderID", "userID", cart, quantity, "totalPrice", "shippingPrice", "grandPrice", status, "createdAt", "updatedAt"
		 FROM orders WHERE "userID" = $1 ORDER BY "orderID" DESC`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanOrders(rows)
}

func scanOrders(rows *sql.Rows) ([]Order, error) {
	orders := make([]Order, 0)
	for rows.Next() {
		var ord Order
		var cartRaw []byte
		var status sql.NullString
		if err := rows.Scan(
			&ord.OrderID, &ord.UserID, &cartRaw, &ord.Quantity,
			&ord.TotalPrice, &ord.ShippingPrice, &ord.GrandPrice,
			&status, &ord.CreatedAt, &ord.UpdatedAt,
		); err != nil {
			return nil, err
		}
		if status.Valid {
			ord.Status = status.String
		}
		if len(cartRaw) > 0 {
			_ = json.Unmarshal(cartRaw, &ord.Cart)
		}
		orders = append(orders, ord)
	}
	return orders, nil
}

package cart

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/lib/pq"
	"github.com/wichananm65/pet-shop-backend/internal/user"
)

type PostgresRepository struct {
	db *sql.DB
}

const (
	getCartQuery = `
        SELECT p."productID", p."productName", p."productNameTH", p."productDesc", p."productDescTH", p."productPrice", p."productImg", p."score"
        FROM products p
        WHERE p."productID" = ANY($1::int[])
        ORDER BY array_position($1::int[], p."productID")
    `
)

func NewPostgresRepository(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) AddToCart(userID int, productID int, qty int, updatedAt string) ([]CartItem, error) {
	// load current cart JSON
	var raw sql.NullString
	if err := r.db.QueryRow(`SELECT cart FROM users WHERE "userId" = $1`, userID).Scan(&raw); err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, err
	}

	// parse into map; fallback from legacy array
	m := make(map[string]int)
	if raw.Valid && raw.String != "" {
		if err := json.Unmarshal([]byte(raw.String), &m); err != nil {
			var arr []int
			if err2 := json.Unmarshal([]byte(raw.String), &arr); err2 == nil {
				m = make(map[string]int, len(arr))
				for _, pid := range arr {
					key := strconv.Itoa(pid)
					m[key]++
				}
			} else {
				return nil, err
			}
		}
	}

	key := strconv.Itoa(productID)
	current := m[key]
	newQty := current + qty
	if newQty <= 0 {
		delete(m, key)
	} else {
		m[key] = newQty
	}

	// marshal back
	updatedCart, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}

	// write back to database
	fmt.Printf("[DEBUG] writing cart JSON %s for user %d\n", string(updatedCart), userID)
	if err := r.db.QueryRow(`UPDATE users SET cart = $1, "updateAt" = $2 WHERE "userId" = $3 RETURNING cart`, string(updatedCart), updatedAt, userID).Scan(&raw); err != nil {
		return nil, err
	}
	fmt.Printf("[DEBUG] db returned cart %s\n", raw.String)

	return r.GetCart(userID)
}

func (r *PostgresRepository) GetCart(userID int) ([]CartItem, error) {
	// fetch raw cart json map
	var raw sql.NullString
	if err := r.db.QueryRow(`SELECT cart FROM users WHERE "userId" = $1`, userID).Scan(&raw); err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, err
	}

	if !raw.Valid || raw.String == "" {
		return []CartItem{}, nil
	}

	var m map[string]int
	if err := json.Unmarshal([]byte(raw.String), &m); err != nil {
		// try fallback array
		var arr []int
		if err2 := json.Unmarshal([]byte(raw.String), &arr); err2 == nil {
			m = make(map[string]int, len(arr))
			for _, pid := range arr {
				m[strconv.Itoa(pid)]++
			}
		} else {
			return nil, err
		}
	}

	ids := make([]int, 0, len(m))
	for k := range m {
		if pid, err := strconv.Atoi(k); err == nil {
			ids = append(ids, pid)
		}
	}

	if len(ids) == 0 {
		return []CartItem{}, nil
	}

	rows, err := r.db.Query(getCartQuery, pq.Array(ids))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]CartItem, 0)
	for rows.Next() {
		var f user.FavoriteProduct
		if err := rows.Scan(&f.ProductID, &f.ProductName, &f.ProductNameTH, &f.ProductDesc, &f.ProductDescTH, &f.ProductPrice, &f.ProductImg, &f.Score); err != nil {
			return nil, err
		}
		qty := m[strconv.Itoa(f.ProductID)]
		out = append(out, CartItem{FavoriteProduct: f, Quantity: qty})
	}

	return out, nil
}

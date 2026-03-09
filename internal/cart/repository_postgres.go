package cart

import (
	"database/sql"

	"github.com/wichananm65/pet-shop-backend/internal/user"
)

type PostgresRepository struct {
	db *sql.DB
}

const (
	getCartQuery = `
        SELECT p.productid, p.productname, p.productnameth, p.productdesc, p.productdescth, p.productprice, p.productimg, p.score, c.quantity
        FROM cart c
        JOIN products p ON c.productid = p.productid
        WHERE c.userid = $1
        ORDER BY c.createdat
    `
	insertCartItemQuery = `
        INSERT INTO cart (userid, productid, quantity, createdat, updatedat)
        VALUES ($1, $2, $3, $4, $4)
        ON CONFLICT (userid, productid)
        DO UPDATE SET quantity = cart.quantity + EXCLUDED.quantity, updatedat = EXCLUDED.updatedat
        RETURNING cartid
    `
	updateCartQuantityQuery = `
        UPDATE cart SET quantity = $1, updatedat = $2
        WHERE userid = $3 AND productid = $4
    `
	deleteCartItemQuery = `
        DELETE FROM cart WHERE userid = $1 AND productid = $2
    `
	clearCartQuery = `
        DELETE FROM cart WHERE userid = $1
    `
)

func NewPostgresRepository(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

// ClearCart removes all items from the user's cart.
func (r *PostgresRepository) ClearCart(userID int, updatedAt string) error {
	_, err := r.db.Exec(clearCartQuery, userID)
	return err
}

func (r *PostgresRepository) AddToCart(userID int, productID int, qty int, updatedAt string) ([]CartItem, error) {
	if qty == 0 {
		return r.GetCart(userID)
	}

	if qty > 0 {
		// Add or update quantity
		_, err := r.db.Exec(insertCartItemQuery, userID, productID, qty, updatedAt)
		if err != nil {
			return nil, err
		}
	} else {
		// Reduce quantity
		var currentQty int
		err := r.db.QueryRow(`SELECT quantity FROM cart WHERE userid = $1 AND productid = $2`, userID, productID).Scan(&currentQty)
		if err != nil && err != sql.ErrNoRows {
			return nil, err
		}

		newQty := currentQty + qty
		if newQty <= 0 {
			// Remove item if quantity drops to zero or below
			_, err = r.db.Exec(deleteCartItemQuery, userID, productID)
			if err != nil {
				return nil, err
			}
		} else {
			// Update quantity
			_, err = r.db.Exec(updateCartQuantityQuery, newQty, updatedAt, userID, productID)
			if err != nil {
				return nil, err
			}
		}
	}

	return r.GetCart(userID)
}

func (r *PostgresRepository) GetCart(userID int) ([]CartItem, error) {
	rows, err := r.db.Query(getCartQuery, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]CartItem, 0)
	for rows.Next() {
		var f user.FavoriteProduct
		var quantity int
		if err := rows.Scan(&f.ProductID, &f.ProductName, &f.ProductNameTH, &f.ProductDesc, &f.ProductDescTH, &f.ProductPrice, &f.ProductImg, &f.Score, &quantity); err != nil {
			return nil, err
		}
		out = append(out, CartItem{FavoriteProduct: f, Quantity: quantity})
	}

	return out, nil
}

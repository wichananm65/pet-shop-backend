package shoppingmall

import (
	"database/sql"
)

// Repository provides access to shopping-mall rows.
type Repository interface {
	List(limit int) ([]ShoppingMallItem, error)
	ListLite(limit int) ([]LiteItem, error)
}

// PostgresRepository implements Repository using Postgres.
type PostgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) List(limit int) ([]ShoppingMallItem, error) {
	rows, err := r.db.Query(`SELECT "productID", "productImg", "productPrice", "score", "productName", "productNameTH" FROM products ORDER BY "productID" LIMIT $1`, limit)
	if err != nil {
		return []ShoppingMallItem{}, nil
	}
	defer rows.Close()

	out := make([]ShoppingMallItem, 0)
	for rows.Next() {
		var (
			id     int
			img    sql.NullString
			price  sql.NullInt64
			score  sql.NullInt64
			name   sql.NullString
			nameTH sql.NullString
		)
		if err := rows.Scan(&id, &img, &price, &score, &name, &nameTH); err != nil {
			continue
		}
		it := ShoppingMallItem{ProductID: id}
		if img.Valid {
			s := img.String
			it.ProductImg = &s
		}
		if price.Valid {
			v := int(price.Int64)
			it.Price = &v
		}
		if score.Valid {
			v := int(score.Int64)
			it.Score = &v
		}
		if name.Valid {
			s := name.String
			it.ProductName = &s
		}
		if nameTH.Valid {
			s := nameTH.String
			it.ProductNameTH = &s
		}
		out = append(out, it)
	}
	return out, nil
}

func (r *PostgresRepository) ListLite(limit int) ([]LiteItem, error) {
	rows, err := r.db.Query(`SELECT "productID", "productImg" FROM products ORDER BY "productID" LIMIT $1`, limit)
	if err != nil {
		return []LiteItem{}, nil
	}
	defer rows.Close()

	out := make([]LiteItem, 0)
	for rows.Next() {
		var id int
		var img sql.NullString
		if err := rows.Scan(&id, &img); err != nil {
			continue
		}
		item := LiteItem{ProductId: id}
		if img.Valid {
			s := img.String
			item.ProductPic = &s
		}
		out = append(out, item)
	}
	return out, nil
}

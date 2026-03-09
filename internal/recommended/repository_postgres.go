package recommended

import (
	"database/sql"
)

// Repository provides access to recommended items.
type Repository interface {
	List(limit int, offset int) ([]RecommendedItem, error)
}

// PostgresRepository implements Repository using Postgres.
type PostgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) List(limit int, offset int) ([]RecommendedItem, error) {
	// Use the standardized products table with lowercase column names
	rows, err := r.db.Query(`SELECT productid, productimg, productname, productnameth, productprice, score FROM products ORDER BY score DESC, productid LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return []RecommendedItem{}, nil
	}
	defer rows.Close()

	out := make([]RecommendedItem, 0)
	for rows.Next() {
		var (
			id     int
			img    sql.NullString
			name   sql.NullString
			nameTH sql.NullString
			price  sql.NullInt64
			score  sql.NullInt64
		)
		if err := rows.Scan(&id, &img, &name, &nameTH, &price, &score); err != nil {
			continue
		}
		item := RecommendedItem{ProductID: id}
		if img.Valid {
			item.ProductImg = &img.String
		}
		if name.Valid {
			item.ProductName = &name.String
		}
		if nameTH.Valid {
			item.ProductNameTH = &nameTH.String
		}
		if price.Valid {
			v := int(price.Int64)
			item.ProductPrice = &v
		}
		if score.Valid {
			v := int(score.Int64)
			item.Score = &v
		}
		out = append(out, item)
	}

	return out, nil
}

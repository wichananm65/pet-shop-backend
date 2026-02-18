package recommended

import (
	"database/sql"
	"fmt"
	"strings"
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
	// Primary (preferred) schema: `product` table (newer schema)
	rows, err := r.db.Query(`SELECT product_id, product_pic, product_name_en, product_name, product_price, score FROM product ORDER BY score DESC, product_id LIMIT $1 OFFSET $2`, limit, offset)
	if err == nil {
		defer rows.Close()
		out := make([]RecommendedItem, 0)
		for rows.Next() {
			var (
				id     int
				pic    sql.NullString
				nameEn sql.NullString
				name   sql.NullString
				price  sql.NullInt64
				score  sql.NullInt64
			)
			if err := rows.Scan(&id, &pic, &nameEn, &name, &price, &score); err != nil {
				continue
			}
			item := RecommendedItem{ProductID: id}
			if pic.Valid {
				item.ProductImg = &pic.String
			}
			// prefer English name if available; keep localized in ProductNameTH when present
			if nameEn.Valid {
				item.ProductName = &nameEn.String
			}
			if name.Valid {
				item.ProductNameTH = &name.String
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
		if len(out) > 0 {
			return out, nil
		}
	}

	// Fallback: legacy `products` table (older deployments may only have this table populated).
	// Read id + image, then try to enrich by querying `product` table for matching details (name/price/score) when available.
	rows2, err2 := r.db.Query(`SELECT "productID", "productName", "productPrice", "productImg", "productNameTH", "score" FROM products ORDER BY "productID" LIMIT $1`, limit)
	if err2 != nil {
		// fallback to simple (id,img) query if legacy table doesn't have those columns
		rows2, err2 = r.db.Query(`SELECT "productID", "productImg" FROM products ORDER BY "productID" LIMIT $1`, limit)
		if err2 != nil {
			return []RecommendedItem{}, nil
		}
	}
	defer rows2.Close()

	ids := make([]int, 0)
	out2 := make([]RecommendedItem, 0)
	for rows2.Next() {
		var (
			id     int
			name   sql.NullString
			price  sql.NullInt64
			img    sql.NullString
			nameTH sql.NullString
			score  sql.NullInt64
		)
		// Scan will populate available columns depending on which SELECT was executed
		if err := rows2.Scan(&id, &name, &price, &img, &nameTH, &score); err != nil {
			// try the simpler scan shape (id, img)
			if err2 := rows2.Scan(&id, &img); err2 != nil {
				continue
			}
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
		out2 = append(out2, item)
		ids = append(ids, id)
	}

	if len(ids) == 0 {
		return out2, nil
	}

	// Try to fetch details for these ids from the newer `product` table.
	placeholders := make([]string, len(ids))
	args := make([]interface{}, len(ids))
	for i, v := range ids {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args[i] = v
	}
	query := fmt.Sprintf(`SELECT product_id, product_name, product_price, score FROM product WHERE product_id IN (%s)`, strings.Join(placeholders, ","))
	rows3, err3 := r.db.Query(query, args...)
	if err3 == nil {
		defer rows3.Close()
		details := map[int]struct {
			Name  string
			Price int
			Score int
		}{}
		for rows3.Next() {
			var id int
			var name sql.NullString
			var price sql.NullInt64
			var score sql.NullInt64
			if err := rows3.Scan(&id, &name, &price, &score); err != nil {
				continue
			}
			if name.Valid || price.Valid || score.Valid {
				details[id] = struct {
					Name  string
					Price int
					Score int
				}{
					Name: func() string {
						if name.Valid {
							return name.String
						}
						return ""
					}(),
					Price: func() int {
						if price.Valid {
							return int(price.Int64)
						}
						return 0
					}(),
					Score: func() int {
						if score.Valid {
							return int(score.Int64)
						}
						return 0
					}(),
				}
			}
		}

		for i := range out2 {
			if d, ok := details[out2[i].ProductID]; ok {
				if d.Name != "" {
					out2[i].ProductNameTH = &d.Name
				}
				if d.Price != 0 {
					out2[i].ProductPrice = &d.Price
				}
				if d.Score != 0 {
					out2[i].Score = &d.Score
				}
			}
		}
	}

	return out2, nil
}

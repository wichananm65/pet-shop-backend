package category

import (
	"database/sql"
)

// Repository provides access to category rows.
type Repository interface {
	List(limit int) ([]CategoryItem, error)
}

// PostgresRepository implements Repository using Postgres.
type PostgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

// List returns category rows ordered by `ord` then id.
// If the table/query is not available the function returns an empty slice (caller-friendly).
func (r *PostgresRepository) List(limit int) ([]CategoryItem, error) {
	rows, err := r.db.Query(`SELECT "categoryID", "categoryName", "categoryNameTH", "categoryImg" FROM category ORDER BY COALESCE(ord, 0) DESC, "categoryID" LIMIT $1`, limit)
	if err != nil {
		// table may not exist or be empty — return empty slice to keep API resilient
		return []CategoryItem{}, nil
	}
	defer rows.Close()

	out := make([]CategoryItem, 0)
	for rows.Next() {
		var (
			id     int
			name   string
			nameTH sql.NullString
			img    sql.NullString
		)
		if err := rows.Scan(&id, &name, &nameTH, &img); err != nil {
			continue
		}
		item := CategoryItem{CategoryID: id, CategoryName: name}
		if nameTH.Valid {
			item.CategoryNameTH = &nameTH.String
		}
		if img.Valid {
			item.CategoryImg = &img.String
		}
		out = append(out, item)
	}
	return out, nil
}

package banner

import (
	"database/sql"
)

// Repository provides access to banner items.
type Repository interface {
	List(limit int) ([]BannerItem, error)
}

// PostgresRepository implements Repository using Postgres.
type PostgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

// List returns banner rows from `banner` table ordered by `ord` then id.
// If the table/query is not available the function returns an empty slice (caller-friendly).
func (r *PostgresRepository) List(limit int) ([]BannerItem, error) {
	rows, err := r.db.Query(`SELECT banner_id, banner_img, banner_link, banner_alt FROM banner ORDER BY COALESCE(ord, 0) DESC, banner_id LIMIT $1`, limit)
	if err != nil {
		// table may not exist or be empty — return empty slice to keep API resilient
		return []BannerItem{}, nil
	}
	defer rows.Close()

	out := make([]BannerItem, 0)
	for rows.Next() {
		var (
			id   int
			img  sql.NullString
			link sql.NullString
			alt  sql.NullString
		)
		if err := rows.Scan(&id, &img, &link, &alt); err != nil {
			continue
		}
		item := BannerItem{BannerID: id}
		if img.Valid {
			item.BannerImg = &img.String
		}
		if link.Valid {
			item.Link = &link.String
		}
		if alt.Valid {
			item.Alt = &alt.String
		}
		out = append(out, item)
	}
	return out, nil
}

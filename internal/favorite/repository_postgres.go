package favorite

import (
	"database/sql"
	"strconv"
	"strings"

	"github.com/lib/pq"
	"github.com/wichananm65/pet-shop-backend/internal/user"
)

type PostgresRepository struct {
	db *sql.DB
}

const (
	getFavoritesQuery = `
		SELECT p."productID", p."productName", p."productNameTH", p."productDesc", p."productDescTH", p."productPrice", p."productImg", p."score"
		FROM products p
		WHERE p."productID" = ANY($1::int[])
		ORDER BY array_position($1::int[], p."productID")
	`
	addFavoriteQuery = `
		UPDATE users
		SET "favoriteProductId" = array_append(coalesce("favoriteProductId", ARRAY[]::integer[]), $2),
			"updateAt" = $3
		WHERE "userId" = $1
			AND NOT ($2 = ANY(coalesce("favoriteProductId", ARRAY[]::integer[])))
		RETURNING "favoriteProductId"
	`
	removeFavoriteQuery = `
		UPDATE users
		SET "favoriteProductId" = array_remove(coalesce("favoriteProductId", ARRAY[]::integer[]), $2),
			"updateAt" = $3
		WHERE "userId" = $1
			AND ($2 = ANY(coalesce("favoriteProductId", ARRAY[]::integer[])))
		RETURNING "favoriteProductId"
	`
)

func NewPostgresRepository(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) AddFavorite(userID int, productID int, updatedAt string) ([]int, error) {
	var arr pq.Int64Array
	err := r.db.QueryRow(addFavoriteQuery, userID, productID, updatedAt).Scan(pq.Array(&arr))
	if err != nil {
		if err == sql.ErrNoRows {
			var exists int
			if err2 := r.db.QueryRow(`SELECT 1 FROM users WHERE "userId" = $1`, userID).Scan(&exists); err2 == sql.ErrNoRows {
				return nil, ErrNotFound
			}
			return nil, ErrAlreadyFavorite
		}

		raw := sql.NullString{}
		if err2 := r.db.QueryRow(`SELECT array_to_string("favoriteProductId", ',') FROM users WHERE "userId" = $1`, userID).Scan(&raw); err2 != nil {
			return nil, err
		}

		if !raw.Valid || raw.String == "" {
			return []int{}, nil
		}

		parts := strings.Split(raw.String, ",")
		out := make([]int, 0, len(parts))
		for _, p := range parts {
			p = strings.TrimSpace(p)
			if p == "" {
				continue
			}
			v, convErr := strconv.Atoi(p)
			if convErr != nil {
				return nil, convErr
			}
			out = append(out, v)
		}
		return out, nil
	}

	res := make([]int, len(arr))
	for i, v := range arr {
		res[i] = int(v)
	}
	return res, nil
}

func (r *PostgresRepository) RemoveFavorite(userID int, productID int, updatedAt string) ([]int, error) {
	var arr pq.Int64Array
	err := r.db.QueryRow(removeFavoriteQuery, userID, productID, updatedAt).Scan(pq.Array(&arr))
	if err != nil {
		if err == sql.ErrNoRows {
			var exists int
			if err2 := r.db.QueryRow(`SELECT 1 FROM users WHERE "userId" = $1`, userID).Scan(&exists); err2 == sql.ErrNoRows {
				return nil, ErrNotFound
			}
			return nil, ErrNotFavorite
		}

		raw := sql.NullString{}
		if err2 := r.db.QueryRow(`SELECT array_to_string("favoriteProductId", ',') FROM users WHERE "userId" = $1`, userID).Scan(&raw); err2 != nil {
			return nil, err
		}

		if !raw.Valid || raw.String == "" {
			return []int{}, nil
		}

		parts := strings.Split(raw.String, ",")
		out := make([]int, 0, len(parts))
		for _, p := range parts {
			p = strings.TrimSpace(p)
			if p == "" {
				continue
			}
			v, convErr := strconv.Atoi(p)
			if convErr != nil {
				return nil, convErr
			}
			out = append(out, v)
		}
		return out, nil
	}

	res := make([]int, len(arr))
	for i, v := range arr {
		res[i] = int(v)
	}
	return res, nil
}

func (r *PostgresRepository) GetFavorites(userID int) ([]user.FavoriteProduct, error) {
	// reuse existing user lookup to obtain persisted favorite array
	var favText sql.NullString
	if err := r.db.QueryRow(`SELECT array_to_string("favoriteProductId", ',') FROM users WHERE "userId" = $1`, userID).Scan(&favText); err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, err
	}

	if !favText.Valid || favText.String == "" {
		return []user.FavoriteProduct{}, nil
	}

	parts := strings.Split(favText.String, ",")
	ids := make([]int, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		v, convErr := strconv.Atoi(p)
		if convErr != nil {
			return nil, convErr
		}
		ids = append(ids, v)
	}

	if len(ids) == 0 {
		return []user.FavoriteProduct{}, nil
	}

	rows, err := r.db.Query(getFavoritesQuery, pq.Array(ids))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]user.FavoriteProduct, 0)
	for rows.Next() {
		var f user.FavoriteProduct
		if err := rows.Scan(&f.ProductID, &f.ProductName, &f.ProductNameTH, &f.ProductDesc, &f.ProductDescTH, &f.ProductPrice, &f.ProductImg, &f.Score); err != nil {
			return nil, err
		}
		out = append(out, f)
	}

	return out, nil
}

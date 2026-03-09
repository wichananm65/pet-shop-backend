package favorite

import (
	"database/sql"

	"github.com/wichananm65/pet-shop-backend/internal/user"
)

type PostgresRepository struct {
	db *sql.DB
}

const (
	getFavoritesQuery = `
		SELECT p.productid, p.productname, p.productnameth, p.productdesc, p.productdescth, p.productprice, p.productimg, p.score
		FROM products p
		JOIN "Favorite" f ON p.productid = f.productid
		WHERE f.userid = $1
		ORDER BY p.productid
	`
	addFavoriteQuery = `
		INSERT INTO "Favorite" (userid, productid)
		VALUES ($1, $2)
		ON CONFLICT (userid, productid) DO NOTHING
	`
	removeFavoriteQuery = `
		DELETE FROM "Favorite" WHERE userid = $1 AND productid = $2
	`
)

func NewPostgresRepository(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) AddFavorite(userID int, productID int, updatedAt string) ([]int, error) {
	_, err := r.db.Exec(addFavoriteQuery, userID, productID)
	if err != nil {
		return nil, err
	}
	// Return current favorite IDs
	return r.getFavoriteIDs(userID)
}

func (r *PostgresRepository) RemoveFavorite(userID int, productID int, updatedAt string) ([]int, error) {
	_, err := r.db.Exec(removeFavoriteQuery, userID, productID)
	if err != nil {
		return nil, err
	}
	// Return current favorite IDs
	return r.getFavoriteIDs(userID)
}

func (r *PostgresRepository) getFavoriteIDs(userID int) ([]int, error) {
	rows, err := r.db.Query(`SELECT productid FROM "Favorite" WHERE userid = $1 ORDER BY productid`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []int
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, nil
}

func (r *PostgresRepository) GetFavorites(userID int) ([]user.FavoriteProduct, error) {
	rows, err := r.db.Query(getFavoritesQuery, userID)
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

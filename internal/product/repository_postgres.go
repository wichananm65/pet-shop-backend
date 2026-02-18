package product

import (
	"database/sql"
)

type PostgresRepository struct {
	db *sql.DB
}

const (
	listProductsQuery = `
		SELECT product_id, product_name, product_name_en, category, product_price, score, product_desc, product_desc_en, product_pic, product_pic_second, created_at, updated_at
		FROM product
		ORDER BY product_id
	`
	getProductByIDQuery = `
		SELECT product_id, product_name, product_name_en, category, product_price, score, product_desc, product_desc_en, product_pic, product_pic_second, created_at, updated_at
		FROM product
		WHERE product_id = $1
	`
	insertProductQuery = `
		INSERT INTO product (product_name, product_name_en, category, product_price, score, product_desc, product_desc_en, product_pic, product_pic_second, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
		RETURNING product_id
	`
	updateProductQuery = `
		UPDATE product
		SET product_name = $1,
			product_name_en = $2,
			category = $3,
			product_price = $4,
			score = $5,
			product_desc = $6,
			product_desc_en = $7,
			product_pic = $8,
			product_pic_second = $9,
			updated_at = $10
		WHERE product_id = $11
	`
	deleteProductQuery = `DELETE FROM product WHERE product_id = $1`
)

func NewPostgresRepository(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) List() []Product {
	rows, err := r.db.Query(listProductsQuery)
	if err != nil {
		return []Product{}
	}
	defer rows.Close()

	out := make([]Product, 0)
	for rows.Next() {
		p, err := scanProduct(rows)
		if err != nil {
			continue
		}
		out = append(out, p)
	}
	return out
}

func (r *PostgresRepository) GetByID(id int) (Product, error) {
	row := r.db.QueryRow(getProductByIDQuery, id)
	p, err := scanProduct(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return Product{}, ErrNotFound
		}
		return Product{}, err
	}
	return p, nil
}

func (r *PostgresRepository) Create(p Product) (Product, error) {
	var id int
	err := r.db.QueryRow(
		insertProductQuery,
		p.Name,
		p.NameEn,
		p.Category,
		p.Price,
		p.Score,
		p.Description,
		p.DescriptionEn,
		p.Pic,
		p.PicSecond,
		p.CreatedAt,
		p.UpdatedAt,
	).Scan(&id)
	if err != nil {
		return Product{}, err
	}
	p.ID = id
	return p, nil
}

func (r *PostgresRepository) Update(id int, p Product) (Product, error) {
	result, err := r.db.Exec(
		updateProductQuery,
		p.Name,
		p.NameEn,
		p.Category,
		p.Price,
		p.Score,
		p.Description,
		p.DescriptionEn,
		p.Pic,
		p.PicSecond,
		p.UpdatedAt,
		id,
	)
	if err != nil {
		return Product{}, err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return Product{}, err
	}
	if affected == 0 {
		return Product{}, ErrNotFound
	}
	return r.GetByID(id)
}

func (r *PostgresRepository) Delete(id int) error {
	result, err := r.db.Exec(deleteProductQuery, id)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return ErrNotFound
	}
	return nil
}

// Reset deletes all products and inserts the provided list in a single transaction.
func (r *PostgresRepository) Reset(products []Product) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	if _, err := tx.Exec(`DELETE FROM product`); err != nil {
		return err
	}

	for _, p := range products {
		var id int
		err := tx.QueryRow(insertProductQuery,
			p.Name,
			p.NameEn,
			p.Category,
			p.Price,
			p.Score,
			p.Description,
			p.DescriptionEn,
			p.Pic,
			p.PicSecond,
			p.CreatedAt,
			p.UpdatedAt,
		).Scan(&id)
		if err != nil {
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		return err
	}
	return nil
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanProduct(scanner rowScanner) (Product, error) {
	p := Product{}
	var createdAt sql.NullString
	var updatedAt sql.NullString
	var nameEn sql.NullString
	var category sql.NullString
	var descEn sql.NullString
	var pic sql.NullString
	var picSecond sql.NullString

	if err := scanner.Scan(
		&p.ID,
		&p.Name,
		&nameEn,
		&category,
		&p.Price,
		&p.Score,
		&p.Description,
		&descEn,
		&pic,
		&picSecond,
		&createdAt,
		&updatedAt,
	); err != nil {
		return Product{}, err
	}

	if nameEn.Valid {
		p.NameEn = &nameEn.String
	}
	if category.Valid {
		p.Category = &category.String
	}
	if descEn.Valid {
		p.DescriptionEn = &descEn.String
	}
	if pic.Valid {
		p.Pic = &pic.String
	}
	if picSecond.Valid {
		p.PicSecond = &picSecond.String
	}
	if createdAt.Valid {
		p.CreatedAt = &createdAt.String
	}
	if updatedAt.Valid {
		p.UpdatedAt = &updatedAt.String
	}

	return p, nil
}

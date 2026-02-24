package product

import (
	"database/sql"

	"github.com/lib/pq"
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
		// fallback to legacy `products` table if the modern `product` table is missing
		return r.listLegacy()
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

// ListByCategoryID returns products whose category matches the name stored in
// the category table for the given numeric ID. The join avoids having to perform
// two queries in the service layer.
func (r *PostgresRepository) ListByCategoryID(catID int) []Product {
	q := `
		SELECT p.product_id, p.product_name, p.product_name_en, p.category,
		       p.product_price, p.score, p.product_desc, p.product_desc_en,
		       p.product_pic, p.product_pic_second, p.created_at, p.updated_at
		FROM product p
		JOIN category c ON p.category = c."categoryName"
		WHERE c."categoryID" = $1
		ORDER BY p.product_id
	`
	rows, err := r.db.Query(q, catID)
	if err != nil {
		// try legacy table
		return r.listByCategoryIDLegacy(catID)
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
			// try legacy table as last resort
			return r.getByIDLegacy(id)
		}
		// if the query itself failed (table missing) we get an error from Scan
		return r.getByIDLegacy(id)
	}
	return p, nil
}

// GetV1ByID returns the `products`-style product detail used by the v1 API.
func (r *PostgresRepository) GetV1ByID(id int) (ProductV1, error) {
	q := `SELECT "productID", "productName", "productNameTH", "productPrice", "productImg", "productDesc", "productDescTH", "score", "category" FROM products WHERE "productID" = $1`
	row := r.db.QueryRow(q, id)
	var (
		pid      int
		name     sql.NullString
		nameTH   sql.NullString
		price    sql.NullInt64
		img      sql.NullString
		desc     sql.NullString
		descTH   sql.NullString
		score    sql.NullInt64
		category sql.NullString
	)
	if err := row.Scan(&pid, &name, &nameTH, &price, &img, &desc, &descTH, &score, &category); err != nil {
		if err == sql.ErrNoRows {
			return ProductV1{}, ErrNotFound
		}
		return ProductV1{}, err
	}
	var (
		pName, pNameTH, pImg, pDesc, pDescTH, pCat *string
		pPrice, pScore                             *int
	)
	if name.Valid {
		pName = &name.String
	}
	if nameTH.Valid {
		pNameTH = &nameTH.String
	}
	if img.Valid {
		pImg = &img.String
	}
	if desc.Valid {
		pDesc = &desc.String
	}
	if descTH.Valid {
		pDescTH = &descTH.String
	}
	if category.Valid {
		pCat = &category.String
	}
	if price.Valid {
		v := int(price.Int64)
		pPrice = &v
	}
	if score.Valid {
		v := int(score.Int64)
		pScore = &v
	}

	return ProductV1{
		ProductID:     pid,
		ProductName:   pName,
		ProductNameTH: pNameTH,
		ProductPrice:  pPrice,
		ProductImg:    pImg,
		ProductDesc:   pDesc,
		ProductDescTH: pDescTH,
		Score:         pScore,
		Category:      pCat,
	}, nil
}

// ListV1ByIDs retrieves the v1-style records for all product IDs in the
// provided slice.  Returns empty slice when input is empty.
func (r *PostgresRepository) ListV1ByIDs(ids []int) ([]ProductV1, error) {
	if len(ids) == 0 {
		return []ProductV1{}, nil
	}
	q := `SELECT "productID", "productName", "productNameTH", "productPrice", "productImg", "productDesc", "productDescTH", "score", "category" FROM products WHERE "productID" = ANY($1::int[])`
	rows, err := r.db.Query(q, pq.Array(ids))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]ProductV1, 0)
	for rows.Next() {
		var (
			pid      int
			name     sql.NullString
			nameTH   sql.NullString
			price    sql.NullInt64
			img      sql.NullString
			desc     sql.NullString
			descTH   sql.NullString
			score    sql.NullInt64
			category sql.NullString
		)
		if err := rows.Scan(&pid, &name, &nameTH, &price, &img, &desc, &descTH, &score, &category); err != nil {
			continue
		}
		p := ProductV1{ProductID: pid}
		if name.Valid {
			p.ProductName = &name.String
		}
		if nameTH.Valid {
			p.ProductNameTH = &nameTH.String
		}
		if price.Valid {
			val := int(price.Int64)
			p.ProductPrice = &val
		}
		if img.Valid {
			p.ProductImg = &img.String
		}
		if desc.Valid {
			p.ProductDesc = &desc.String
		}
		if descTH.Valid {
			p.ProductDescTH = &descTH.String
		}
		if score.Valid {
			val := int(score.Int64)
			p.Score = &val
		}
		if category.Valid {
			p.Category = &category.String
		}
		out = append(out, p)
	}
	return out, nil
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
// To support legacy deployments we clear both `product` and `products` tables so
// today’s code can work regardless of which table is actually populated.
func (r *PostgresRepository) Reset(products []Product) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	// clear both tables; ignore errors for non-existent table
	_, _ = tx.Exec(`DELETE FROM product`)
	_, _ = tx.Exec(`DELETE FROM products`)

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

// ---- legacy helpers -------------------------------------------------------

// listLegacy retrieves rows from the older `products` table and converts them
// into the v2 Product struct.
func (r *PostgresRepository) listLegacy() []Product {
	q := `SELECT "productID","productName","productNameTH","productPrice",score,"productDesc","productDescTH","productImg","productImgSec","createdAt","updatedAt",category FROM products ORDER BY "productID"`
	rows, err := r.db.Query(q)
	if err != nil {
		return []Product{}
	}
	defer rows.Close()
	out := make([]Product, 0)
	for rows.Next() {
		p, err := scanProductLegacy(rows)
		if err != nil {
			continue
		}
		out = append(out, p)
	}
	return out
}

func (r *PostgresRepository) listByCategoryIDLegacy(catID int) []Product {
	q := `SELECT p."productID",p."productName",p."productNameTH",p."productPrice",p.score,p."productDesc",p."productDescTH",p."productImg",p."productImgSec",p."createdAt",p."updatedAt",p.category
		FROM products p
		JOIN category c ON p.category = c."categoryName"
		WHERE c."categoryID" = $1
		ORDER BY p."productID"`
	rows, err := r.db.Query(q, catID)
	if err != nil {
		return []Product{}
	}
	defer rows.Close()
	out := make([]Product, 0)
	for rows.Next() {
		p, err := scanProductLegacy(rows)
		if err != nil {
			continue
		}
		out = append(out, p)
	}
	return out
}

func (r *PostgresRepository) getByIDLegacy(id int) (Product, error) {
	q := `SELECT "productID","productName","productNameTH","productPrice",score,"productDesc","productDescTH","productImg","productImgSec","createdAt","updatedAt",category FROM products WHERE "productID" = $1`
	row := r.db.QueryRow(q, id)
	p, err := scanProductLegacy(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return Product{}, ErrNotFound
		}
		return Product{}, err
	}
	return p, nil
}

func scanProductLegacy(scanner rowScanner) (Product, error) {
	p := Product{}
	var (
		nameTH    sql.NullString
		descTH    sql.NullString
		img       sql.NullString
		imgSec    sql.NullString
		category  sql.NullString
		createdAt sql.NullString
		updatedAt sql.NullString
	)
	if err := scanner.Scan(
		&p.ID,
		&p.Name,
		&nameTH,
		&p.Price,
		&p.Score,
		&p.Description,
		&descTH,
		&img,
		&imgSec,
		&createdAt,
		&updatedAt,
		&category,
	); err != nil {
		return Product{}, err
	}
	if nameTH.Valid {
		p.NameEn = &nameTH.String // repurpose Thai name for lack of English field
	}
	if descTH.Valid {
		p.DescriptionEn = &descTH.String
	}
	if img.Valid {
		p.Pic = &img.String
	}
	if imgSec.Valid {
		p.PicSecond = &imgSec.String
	}
	if category.Valid {
		p.Category = &category.String
	}
	if createdAt.Valid {
		p.CreatedAt = &createdAt.String
	}
	if updatedAt.Valid {
		p.UpdatedAt = &updatedAt.String
	}
	return p, nil
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

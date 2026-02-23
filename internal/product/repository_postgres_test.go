package product

import (
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestListByCategoryID_Fallback(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()
	repo := NewPostgresRepository(db)

	// simulate modern query failing (table missing)
	mock.ExpectQuery("SELECT p.product_id").WithArgs(3).WillReturnError(errors.New("no such table"))

	// legacy query should be attempted next and return a single product
	rows := sqlmock.NewRows([]string{"productID", "productName", "productNameTH", "productPrice", "score", "productDesc", "productDescTH", "productImg", "productImgSec", "createdAt", "updatedAt", "category"}).
		AddRow(5, "Foo", "ไทย", 100, 1, "d", "dth", "img", "img2", "t", "u", "cat")
	mock.ExpectQuery("FROM products p").WithArgs(3).WillReturnRows(rows)

	products := repo.ListByCategoryID(3)
	if len(products) != 1 {
		t.Fatalf("expected 1 product, got %d", len(products))
	}
	if products[0].Name != "Foo" {
		t.Fatalf("unexpected product name %q", products[0].Name)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestList_Fallback(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock error: %v", err)
	}
	defer db.Close()
	repo := NewPostgresRepository(db)

	// first query fails
	mock.ExpectQuery("SELECT product_id").WillReturnError(errors.New("no such table"))
	// fallback returns two rows
	rows := sqlmock.NewRows([]string{"productID", "productName", "productNameTH", "productPrice", "score", "productDesc", "productDescTH", "productImg", "productImgSec", "createdAt", "updatedAt", "category"}).
		AddRow(1, "A", "ไทยA", 10, 1, "d", "dth", "img", "img2", "t", "u", "cat").
		AddRow(2, "B", "ไทยB", 20, 2, "d2", "dth2", "imgb", "img2b", "t2", "u2", "cat2")
	mock.ExpectQuery("FROM products").WillReturnRows(rows)

	all := repo.List()
	if len(all) != 2 {
		t.Fatalf("expected 2 products, got %d", len(all))
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestGetByID_Fallback(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock error: %v", err)
	}
	defer db.Close()
	repo := NewPostgresRepository(db)

	// simulate Scan error (table missing)
	mock.ExpectQuery("SELECT .*FROM product").WithArgs(9).WillReturnError(errors.New("no such table"))

	rows := sqlmock.NewRows([]string{"productID", "productName", "productNameTH", "productPrice", "score", "productDesc", "productDescTH", "productImg", "productImgSec", "createdAt", "updatedAt", "category"}).
		AddRow(9, "Z", "ไทยZ", 99, 9, "d", "dth", "img", "img2", "t", "u", "cat")
	mock.ExpectQuery("FROM products").WithArgs(9).WillReturnRows(rows)

	p, err := repo.GetByID(9)
	if err != nil {
		t.Fatalf("expected nil err, got %v", err)
	}
	if p.ID != 9 || p.Name != "Z" {
		t.Fatalf("unexpected product %+v", p)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

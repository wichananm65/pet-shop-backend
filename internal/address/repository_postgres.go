package address

import (
	"database/sql"
)

// Postgres repository stores addresses in a dedicated table with a foreign
// key to users.
// Table layout expected (camelCase column names):
//   "addressID" serial primary key,
//   "userID" int not null,
//   "addressDesc" text,
//   "phone" text,
//   "addressName" text,
//   "createdAt" text,
//   "updatedAt" text

const (
	insertAddressQuery = `INSERT INTO address (userid, addressdesc, phone, addressname, createdat, updatedat)
        VALUES ($1,$2,$3,$4,$5,$6) RETURNING addressid`
	updateAddressQuery = `UPDATE address SET addressdesc=$1, phone=$2, addressname=$3, updatedat=$4
        WHERE userid=$5 AND addressid=$6 RETURNING addressid`
	deleteAddressQuery = `DELETE FROM address WHERE userid=$1 AND addressid=$2`
)

type PostgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) GetAddresses(userID int) ([]Address, error) {
	if userID <= 0 {
		return nil, ErrNotFound
	}
	rows, err := r.db.Query(`SELECT addressid, userid, addressdesc, phone, addressname, createdat, updatedat FROM address WHERE userid = $1`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]Address, 0)
	for rows.Next() {
		var a Address
		if err := rows.Scan(&a.AddressID, &a.UserID, &a.AddressDesc, &a.Phone, &a.AddressName, &a.CreatedAt, &a.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, a)
	}

	return out, nil
}

func (r *PostgresRepository) AddAddress(userID int, desc, phone, name string, updatedAt string) (Address, error) {
	if userID <= 0 {
		return Address{}, ErrNotFound
	}
	var id int
	if err := r.db.QueryRow(insertAddressQuery, userID, desc, phone, name, updatedAt, updatedAt).Scan(&id); err != nil {
		return Address{}, err
	}
	return Address{AddressID: id, UserID: userID, AddressDesc: desc, Phone: phone, AddressName: name, CreatedAt: updatedAt, UpdatedAt: updatedAt}, nil
}

func (r *PostgresRepository) UpdateAddress(userID, addressID int, desc, phone, name string, updatedAt string) (Address, error) {
	if userID <= 0 || addressID <= 0 {
		return Address{}, ErrNotFound
	}
	var id int
	err := r.db.QueryRow(updateAddressQuery, desc, phone, name, updatedAt, userID, addressID).Scan(&id)
	if err != nil {
		if err == sql.ErrNoRows {
			return Address{}, ErrNotFound
		}
		return Address{}, err
	}
	return Address{AddressID: id, UserID: userID, AddressDesc: desc, Phone: phone, AddressName: name, UpdatedAt: updatedAt}, nil
}

func (r *PostgresRepository) DeleteAddress(userID, addressID int) error {
	if userID <= 0 || addressID <= 0 {
		return ErrNotFound
	}
	if _, err := r.db.Exec(deleteAddressQuery, userID, addressID); err != nil {
		return err
	}
	return nil
}

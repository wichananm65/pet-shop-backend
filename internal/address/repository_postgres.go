package address

import (
    "database/sql"
)

// Postgres repository stores addresses in a dedicated table with a foreign
// key to users.
// Table layout expected:
//   address_id serial primary key,
//   user_id int not null,
//   address_desc text,
//   phone text,
//   address_name text,
//   created_at text,
//   updated_at text

type PostgresRepository struct {
    db *sql.DB
}

const (
    insertAddressQuery = `
        INSERT INTO address (user_id, address_desc, phone, address_name, created_at, updated_at)
        VALUES ($1,$2,$3,$4,$5,$6)
        RETURNING address_id, user_id, address_desc, phone, address_name, created_at, updated_at
    `
    updateAddressQuery = `
        UPDATE address
        SET address_desc=$3, phone=$4, address_name=$5, updated_at=$6
        WHERE user_id=$1 AND address_id=$2
        RETURNING address_id, user_id, address_desc, phone, address_name, created_at, updated_at
    `
    deleteAddressQuery = `
        DELETE FROM address WHERE user_id=$1 AND address_id=$2
    `
)

func NewPostgresRepository(db *sql.DB) *PostgresRepository {
    return &PostgresRepository{db: db}
}

func (r *PostgresRepository) GetAddresses(userID int) ([]Address, error) {
    if userID <= 0 {
        return nil, ErrNotFound
    }
    rows, err := r.db.Query(`SELECT address_id, user_id, address_desc, phone, address_name, created_at, updated_at FROM address WHERE user_id = $1`, userID)
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

func (r *PostgresRepository) AddAddress(userID int, addressDesc, phone, addressName string) (Address, error) {
    var a Address
    if err := r.db.QueryRow(insertAddressQuery, userID, addressDesc, phone, addressName, "", "").Scan(&a.AddressID, &a.UserID, &a.AddressDesc, &a.Phone, &a.AddressName, &a.CreatedAt, &a.UpdatedAt); err != nil {
        if err == sql.ErrNoRows {
            return a, ErrNotFound
        }
        return a, err
    }
    return a, nil
}

func (r *PostgresRepository) UpdateAddress(userID int, addressID int, addressDesc, phone, addressName string) (Address, error) {
    var a Address
    if err := r.db.QueryRow(updateAddressQuery, userID, addressID, addressDesc, phone, addressName, "").Scan(&a.AddressID, &a.UserID, &a.AddressDesc, &a.Phone, &a.AddressName, &a.CreatedAt, &a.UpdatedAt); err != nil {
        if err == sql.ErrNoRows {
            return a, ErrNotFound
        }
        return a, err
    }
    return a, nil
}

func (r *PostgresRepository) DeleteAddress(userID int, addressID int) error {
    res, err := r.db.Exec(deleteAddressQuery, userID, addressID)
    if err != nil {
        return err
    }
    cnt, _ := res.RowsAffected()
    if cnt == 0 {
        return ErrNotFound
    }
    return nil
}

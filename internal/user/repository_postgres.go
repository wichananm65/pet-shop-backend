package user

import (
	"database/sql"
)

type PostgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) List() []User {
	rows, err := r.db.Query(`
		SELECT "userId", email, password, "firstName", "lastName", phone, gender, "createAt", "updateAt"
		FROM users
		ORDER BY "userId"
	`)
	if err != nil {
		return []User{}
	}
	defer rows.Close()

	users := make([]User, 0)
	for rows.Next() {
		user := User{}
		var createdAt sql.NullString
		var updatedAt sql.NullString
		if err := rows.Scan(&user.ID, &user.Email, &user.Password, &user.FirstName, &user.LastName, &user.Phone, &user.Gender, &createdAt, &updatedAt); err != nil {
			continue
		}
		if createdAt.Valid {
			user.CreatedAt = createdAt.String
		}
		if updatedAt.Valid {
			user.UpdatedAt = updatedAt.String
		}
		users = append(users, user)
	}

	return users
}

func (r *PostgresRepository) GetByID(id int) (User, error) {
	user := User{}
	var createdAt sql.NullString
	var updatedAt sql.NullString
	row := r.db.QueryRow(`
		SELECT "userId", email, password, "firstName", "lastName", phone, gender, "createAt", "updateAt"
		FROM users
		WHERE "userId" = $1
	`, id)

	if err := row.Scan(&user.ID, &user.Email, &user.Password, &user.FirstName, &user.LastName, &user.Phone, &user.Gender, &createdAt, &updatedAt); err != nil {
		if err == sql.ErrNoRows {
			return User{}, ErrNotFound
		}
		return User{}, err
	}

	if createdAt.Valid {
		user.CreatedAt = createdAt.String
	}
	if updatedAt.Valid {
		user.UpdatedAt = updatedAt.String
	}

	return user, nil
}

func (r *PostgresRepository) GetByEmail(email string) (User, error) {
	user := User{}
	var createdAt sql.NullString
	var updatedAt sql.NullString
	row := r.db.QueryRow(`
		SELECT "userId", email, password, "firstName", "lastName", phone, gender, "createAt", "updateAt"
		FROM users
		WHERE email = $1
	`, email)

	if err := row.Scan(&user.ID, &user.Email, &user.Password, &user.FirstName, &user.LastName, &user.Phone, &user.Gender, &createdAt, &updatedAt); err != nil {
		if err == sql.ErrNoRows {
			return User{}, ErrNotFound
		}
		return User{}, err
	}

	if createdAt.Valid {
		user.CreatedAt = createdAt.String
	}
	if updatedAt.Valid {
		user.UpdatedAt = updatedAt.String
	}

	return user, nil
}

func (r *PostgresRepository) Create(user User) (User, error) {
	var id int
	query := `
		INSERT INTO users (email, password, "firstName", "lastName", phone, gender, "createAt", "updateAt")
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING "userId"
	`

	err := r.db.QueryRow(query, user.Email, user.Password, user.FirstName, user.LastName, user.Phone, user.Gender, user.CreatedAt, user.UpdatedAt).Scan(&id)
	if err != nil {
		return User{}, err
	}

	user.ID = id
	return user, nil
}

func (r *PostgresRepository) Update(id int, userUpdate User) (User, error) {
	query := `
		UPDATE users
		SET email = $1,
			"firstName" = $2,
			"lastName" = $3,
			phone = $4,
			gender = $5,
			"updateAt" = $6
		WHERE "userId" = $7
	`

	result, err := r.db.Exec(query, userUpdate.Email, userUpdate.FirstName, userUpdate.LastName, userUpdate.Phone, userUpdate.Gender, userUpdate.UpdatedAt, id)
	if err != nil {
		return User{}, err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return User{}, err
	}
	if affected == 0 {
		return User{}, ErrNotFound
	}

	return r.GetByID(id)
}

func (r *PostgresRepository) Delete(id int) error {
	result, err := r.db.Exec(`DELETE FROM users WHERE "userId" = $1`, id)
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

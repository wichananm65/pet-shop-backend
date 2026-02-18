package user

import (
	"database/sql"
)

type PostgresRepository struct {
	db *sql.DB
}

const (
	listUsersQuery = `
		SELECT "userId", email, password, "firstName", "lastName", phone, gender, "createAt", "updateAt"
		FROM users
		ORDER BY "userId"
	`
	getUserByIDQuery = `
		SELECT "userId", email, password, "firstName", "lastName", phone, gender, "createAt", "updateAt"
		FROM users
		WHERE "userId" = $1
	`
	getUserByEmailQuery = `
		SELECT "userId", email, password, "firstName", "lastName", phone, gender, "createAt", "updateAt"
		FROM users
		WHERE email = $1
	`
	insertUserQuery = `
		INSERT INTO users (email, password, "firstName", "lastName", phone, gender, "createAt", "updateAt")
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING "userId"
	`
	updateUserQuery = `
		UPDATE users
		SET email = $1,
			"firstName" = $2,
			"lastName" = $3,
			phone = $4,
			gender = $5,
			"updateAt" = $6
		WHERE "userId" = $7
	`
	deleteUserQuery = `DELETE FROM users WHERE "userId" = $1`
)

// Cart table has been removed — CreateCartWithID is now a no-op to keep
// the repository interface stable for older callers.

type rowScanner interface {
	Scan(dest ...any) error
}

func NewPostgresRepository(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) List() []User {
	rows, err := r.db.Query(listUsersQuery)
	if err != nil {
		return []User{}
	}
	defer rows.Close()

	users := make([]User, 0)
	for rows.Next() {
		user, err := scanUser(rows)
		if err != nil {
			continue
		}
		users = append(users, user)
	}

	return users
}

func (r *PostgresRepository) GetByID(id int) (User, error) {
	row := r.db.QueryRow(getUserByIDQuery, id)
	user, err := scanUser(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return User{}, ErrNotFound
		}
		return User{}, err
	}

	return user, nil
}

func (r *PostgresRepository) GetByEmail(email string) (User, error) {
	row := r.db.QueryRow(getUserByEmailQuery, email)
	user, err := scanUser(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return User{}, ErrNotFound
		}
		return User{}, err
	}

	return user, nil
}

func (r *PostgresRepository) Create(user User) (User, error) {
	var id int
	err := r.db.QueryRow(
		insertUserQuery,
		user.Email,
		user.Password,
		user.FirstName,
		user.LastName,
		user.Phone,
		user.Gender,
		user.CreatedAt,
		user.UpdatedAt,
	).Scan(&id)
	if err != nil {
		return User{}, err
	}

	user.ID = id
	return user, nil
}

func (r *PostgresRepository) Update(id int, userUpdate User) (User, error) {
	result, err := r.db.Exec(
		updateUserQuery,
		userUpdate.Email,
		userUpdate.FirstName,
		userUpdate.LastName,
		userUpdate.Phone,
		userUpdate.Gender,
		userUpdate.UpdatedAt,
		id,
	)
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
	result, err := r.db.Exec(deleteUserQuery, id)
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
func (r *PostgresRepository) CreateCartWithID(cartID int) error {
	// cart table removed — no operation required
	return nil
}

func scanUser(scanner rowScanner) (User, error) {
	user := User{}
	var createdAt sql.NullString
	var updatedAt sql.NullString
	if err := scanner.Scan(
		&user.ID,
		&user.Email,
		&user.Password,
		&user.FirstName,
		&user.LastName,
		&user.Phone,
		&user.Gender,
		&createdAt,
		&updatedAt,
	); err != nil {
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

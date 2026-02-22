package user

import (
	"database/sql"
	"encoding/json"
	"strconv"
	"strings"
)

type PostgresRepository struct {
	db *sql.DB
}

type rowScanner interface {
	Scan(dest ...any) error
}

const (
	listUsersQuery = `
		SELECT "userId", email, password, "firstName", "lastName", phone, gender, avatar_pic, array_to_string("favoriteProductId", ',') AS favoriteProductId_text, array_to_string("cartProductId", ',') AS cartProductId_text, "createAt", "updateAt"
		FROM users
		ORDER BY "userId"
	`
	getUserByIDQuery = `
		SELECT "userId", email, password, "firstName", "lastName", phone, gender, avatar_pic, array_to_string("favoriteProductId", ',') AS favoriteProductId_text, array_to_string("cartProductId", ',') AS cartProductId_text, "createAt", "updateAt"
		FROM users
		WHERE "userId" = $1
	`
	getUserByEmailQuery = `
		SELECT "userId", email, password, "firstName", "lastName", phone, gender, avatar_pic, array_to_string("favoriteProductId", ',') AS favoriteProductId_text, array_to_string("cartProductId", ',') AS cartProductId_text, "createAt", "updateAt"
		FROM users
		WHERE email = $1
	`

	insertUserQuery = `
		INSERT INTO users (email, password, "firstName", "lastName", phone, gender, "createAt", "updateAt", avatar_pic)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING "userId"
	`
	updateUserQuery = `
		UPDATE users
		SET email = $1,
			"firstName" = $2,
			"lastName" = $3,
			phone = $4,
			gender = $5,
			avatar_pic = $8,
			"updateAt" = $6
		WHERE "userId" = $7
	`
	deleteUserQuery = `DELETE FROM users WHERE "userId" = $1`
)

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
	// avatarPic may be nil
	avatarVal := sql.NullString{}
	if user.AvatarPic != nil {
		avatarVal = sql.NullString{String: *user.AvatarPic, Valid: true}
	}
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
		avatarVal,
	).Scan(&id)
	if err != nil {
		return User{}, err
	}

	user.ID = id
	return user, nil
}

func (r *PostgresRepository) Update(id int, userUpdate User) (User, error) {
	// if AvatarPic is nil, send a raw nil so that the database column is set to
	// NULL; previous sql.NullString handling could result in an empty string
	// instead of NULL which leads to a non-nil pointer when scanned.
	var avatarArg interface{}
	if userUpdate.AvatarPic != nil {
		avatarArg = *userUpdate.AvatarPic
	} else {
		avatarArg = nil
	}
	result, err := r.db.Exec(
		updateUserQuery,
		userUpdate.Email,
		userUpdate.FirstName,
		userUpdate.LastName,
		userUpdate.Phone,
		userUpdate.Gender,
		userUpdate.UpdatedAt,
		id,
		avatarArg,
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
	var favText sql.NullString
	var cartJSON sql.NullString
	var avatar sql.NullString
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
		&avatar,
		&favText,
		&cartJSON,
		&createdAt,
		&updatedAt,
	); err != nil {
		return User{}, err
	}

	if avatar.Valid {
		user.AvatarPic = &avatar.String
	}

	if favText.Valid && favText.String != "" {
		parts := strings.Split(favText.String, ",")
		user.FavoriteProductIDs = make([]int, 0, len(parts))
		for _, p := range parts {
			p = strings.TrimSpace(p)
			if p == "" {
				continue
			}
			v, err := strconv.Atoi(p)
			if err != nil {
				return User{}, err
			}
			user.FavoriteProductIDs = append(user.FavoriteProductIDs, v)
		}
	}

	if cartJSON.Valid && cartJSON.String != "" {
		// try map first; if legacy integer array stored, handle that too
		var rawMap map[string]int
		if err := json.Unmarshal([]byte(cartJSON.String), &rawMap); err == nil {
			user.Cart = make(map[int]int, len(rawMap))
			user.CartProductIDs = make([]int, 0, len(rawMap))
			for k, qty := range rawMap {
				pid, err := strconv.Atoi(k)
				if err != nil {
					return User{}, err
				}
				user.Cart[pid] = qty
				user.CartProductIDs = append(user.CartProductIDs, pid)
			}
		} else {
			// attempt array unmarshalling fallback
			var arr []int
			if err2 := json.Unmarshal([]byte(cartJSON.String), &arr); err2 == nil {
				user.Cart = make(map[int]int, len(arr))
				user.CartProductIDs = make([]int, 0, len(arr))
				for _, pid := range arr {
					user.Cart[pid]++
					user.CartProductIDs = append(user.CartProductIDs, pid)
				}
			}
		}
	}

	if createdAt.Valid {
		user.CreatedAt = createdAt.String
	}
	if updatedAt.Valid {
		user.UpdatedAt = updatedAt.String
	}

	return user, nil
}

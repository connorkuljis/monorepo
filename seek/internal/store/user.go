package store

import (
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"
)

type User struct {
	ID    int64  `db:"id"`
	Email string `db:"email"`
}

type UserRepository interface {
	Create(user *User) error
	GetByID(id int64) (*User, error)
	Update(user *User) error
	Delete(id int64) error
}

type userRepository struct {
	db *sqlx.DB
}

func NewUserRepository(db *sqlx.DB) UserRepository {
	return &userRepository{db}
}

func (r *userRepository) Create(user *User) error {
	result, err := r.db.NamedExec("INSERT INTO users ( email ) VALUES ( :email )", user)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}

	user.ID = id

	return nil
}

func (r *userRepository) GetByID(id int64) (*User, error) {
	var user User
	err := r.db.Get(&user, "SELECT * FROM users WHERE id = ?", id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) Update(user *User) error {
	_, err := r.db.NamedExec("UPDATE users SET email = :email WHERE id = :id", user)
	return err
}

func (r *userRepository) Delete(id int64) error {
	_, err := r.db.Exec("DELETE FROM users WHERE id = ?", id)
	return err
}

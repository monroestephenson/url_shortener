package repository

import (
	"database/sql"
	"errors"
	"url_shortener/internal/models"
)

var (
	ErrUserNotFound      = errors.New("user not found")
	ErrUserAlreadyExists = errors.New("user already exists")
)

type UserRepository interface {
	Create(user *models.User, password string) error
	GetByUsername(username string) (*models.User, error)
	GetByID(id int) (*models.User, error)
}

type userRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(user *models.User, password string) error {
	query := `
		INSERT INTO users (username, password_hash)
		VALUES (?, ?)
	`

	result, err := r.db.Exec(query, user.Username, password)
	if err != nil {
		if isDuplicateKeyError(err) {
			return ErrUserAlreadyExists
		}
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}

	user.ID = int(id)
	return nil
}

func (r *userRepository) GetByUsername(username string) (*models.User, error) {
	var user models.User
	query := `
		SELECT id, username, password_hash, created_at, updated_at
		FROM users
		WHERE username = ?
	`

	err := r.db.QueryRow(query, username).Scan(
		&user.ID,
		&user.Username,
		&user.PasswordHash,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, ErrUserNotFound
	} else if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *userRepository) GetByID(id int) (*models.User, error) {
	var user models.User
	query := `
		SELECT id, username, password_hash, created_at, updated_at
		FROM users
		WHERE id = ?
	`

	err := r.db.QueryRow(query, id).Scan(
		&user.ID,
		&user.Username,
		&user.PasswordHash,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, ErrUserNotFound
	} else if err != nil {
		return nil, err
	}

	return &user, nil
}

func isDuplicateKeyError(err error) bool {
	return err != nil && err.Error()[0:10] == "Error 1062"
} 
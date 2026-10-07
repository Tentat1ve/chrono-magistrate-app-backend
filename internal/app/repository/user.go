package repository

import (
	"errors"

	"awesomeProject/internal/app/ds"

	"github.com/jackc/pgx/v5/pgconn"
)

// GetUserByID — пользователь по ID
func (r *Repository) GetUserByID(id uint) (ds.User, error) {
	var user ds.User
	err := r.db.First(&user, id).Error
	return user, wrapNotFound(err)
}

// GetUserByLogin — пользователь по логину
func (r *Repository) GetUserByLogin(login string) (ds.User, error) {
	var user ds.User
	err := r.db.Where("login = ?", login).First(&user).Error
	return user, wrapNotFound(err)
}

// CreateUser — регистрация пользователя; ErrAlreadyExists, если логин занят
func (r *Repository) CreateUser(user *ds.User) error {
	err := r.db.Create(user).Error
	if isUniqueViolation(err) {
		return ErrAlreadyExists
	}
	return err
}

// isUniqueViolation — нарушение уникального индекса PostgreSQL (код 23505)
func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}

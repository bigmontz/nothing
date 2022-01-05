package repository

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/bigmontz/nothing/repository/almost_crdb"
	"strconv"
	"time"
)

type userCockRoachRepository struct {
	db *sql.DB
}

func NewUserCockroachRepository(db *sql.DB) UserRepository {
	return &userCockRoachRepository{db: db}
}

func (u *userCockRoachRepository) Create(user *User) (*User, error) {
	now := time.Now().In(time.UTC)
	row := u.db.QueryRow(
		"INSERT INTO users (username, name, age, surname, password, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING *",
		user.Username,
		user.Name,
		user.Age,
		user.Surname,
		user.Password,
		now,
		now)

	if err := row.Err(); err != nil {
		return nil, err
	}
	var result User
	if err := row.Scan(
		&result.Id,
		&result.Username,
		&result.Name,
		&result.Surname,
		&result.Password,
		&result.Age,
		&result.CreatedAt,
		&result.UpdatedAt); err != nil {
		return nil, err
	}
	return &result, nil
}

func (u *userCockRoachRepository) FindById(rawUserId interface{}) (*User, error) {
	userId, err := strconv.Atoi(rawUserId.(string))
	if err != nil {
		return nil, userError{err: fmt.Errorf("invalid user ID: %w", err)}
	}
	row := u.db.QueryRow("SELECT * FROM users WHERE id = $1", userId)
	if err := row.Err(); err != nil {
		return nil, err
	}
	var result User
	if err := row.Scan(
		&result.Id,
		&result.Username,
		&result.Name,
		&result.Surname,
		&result.Password,
		&result.Age,
		&result.CreatedAt,
		&result.UpdatedAt); err != nil {
		return nil, err
	}
	return &result, nil
}

func (u *userCockRoachRepository) UpdatePassword(userId interface{}, passwordUpdate *PasswordUpdate) (*User, error) {
	ctx := context.Background()
	result, err := almost_crdb.SlightlyAdaptedExecuteTx(
		ctx,
		u.db,
		nil,
		func(tx almost_crdb.SlightlyAdaptedTx) (interface{}, error) {
			result, err := tx.SlightlyAdaptedExec(
				ctx,
				"UPDATE users SET password = $3, updated_at = $4 WHERE id = $1 AND password = $2",
				userId,
				passwordUpdate.Current,
				passwordUpdate.New,
				time.Now().In(time.UTC),
			)
			if err != nil {
				return nil, err
			}
			rowsAffected, err := result.RowsAffected()
			if err != nil {
				return nil, err
			}
			if rowsAffected == 0 {
				return nil, userError{err: fmt.Errorf("user not found"), notFound: true}
			}
			return &User{Id: userId}, nil
		},
	)
	if err != nil {
		return nil, err
	}
	return result.(*User), nil
}

func (u *userCockRoachRepository) Close() error {
	return u.db.Close()
}

package repository

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"strconv"
	"time"
)

type userPostgresRepository struct {
	pool *pgxpool.Pool
}

func NewUserPostgresRepository(pool *pgxpool.Pool) UserRepository {
	return &userPostgresRepository{pool: pool}
}

func (u *userPostgresRepository) Create(user *User) (*User, error) {
	now := time.Now().In(time.UTC)
	rows, err := u.pool.Query(
		context.Background(),
		"INSERT INTO users (username, name, age, surname, password, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING *",
		user.Username,
		user.Name,
		user.Age,
		user.Surname,
		user.Password,
		now,
		now,
	)
	if err != nil {
		return nil, err
	}
	return extractUserFromRows(rows)
}

func (u *userPostgresRepository) FindById(rawUserId interface{}) (*User, error) {
	userId, err := strconv.Atoi(rawUserId.(string))
	if err != nil {
		return nil, userError{err: fmt.Errorf("invalid user ID: %w", err)}
	}
	rows, err := u.pool.Query(
		context.Background(),
		"SELECT * FROM users WHERE id = $1",
		userId,
	)
	if err != nil {
		return nil, err
	}
	return extractUserFromRows(rows)
}

func (u *userPostgresRepository) UpdatePassword(rawUserId interface{}, passwordUpdate *PasswordUpdate) (*User, error) {
	userId, err := strconv.Atoi(rawUserId.(string))
	if err != nil {
		return nil, userError{err: fmt.Errorf("invalid user ID: %w", err)}
	}
	result, err := u.pool.Exec(
		context.Background(),
		"UPDATE users SET password = $3, updated_at = $4 WHERE id = $1 AND password = $2",
		userId,
		passwordUpdate.Current,
		passwordUpdate.New,
		time.Now().In(time.UTC),
	)
	if err != nil {
		return nil, err
	}
	if result.RowsAffected() == 0 {
		return nil, userError{err: fmt.Errorf("user not found"), notFound: true}
	}
	return &User{
		Id: userId,
	}, nil
}

func extractUserFromRows(rows pgx.Rows) (*User, error) {
	defer rows.Close()
	for rows.Next() {
		values, err := rows.Values()
		if err != nil {
			return nil, err
		}
		creationTime := values[6].(time.Time)
		updateTime := values[7].(time.Time)
		return &User{
			Id:        int64(values[0].(int32)),
			Username:  values[1].(string),
			Name:      values[2].(string),
			Surname:   values[3].(string),
			Password:  values[4].(string),
			Age:       uint(values[5].(int32)),
			CreatedAt: &creationTime,
			UpdatedAt: &updateTime,
		}, nil
	}
	return nil, fmt.Errorf("no user found")
}

func (u *userPostgresRepository) Close() error {
	u.pool.Close()
	return nil
}

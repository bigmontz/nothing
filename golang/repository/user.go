package repository

import "time"

type User struct {
	Username  string    `json:"username"`
	Name      string    `json:"name"`
	Age       uint      `json:"age"`
	Surname   string    `json:"surname"`
	Password  string    `json:"password"`
	CreatedAt time.Time `json:"created_at,omitempty"`
	UpdatedAt time.Time `json:"updated_at,omitempty"`
	Id        int64     `json:"id,omitempty"`
}

type UserRepository interface {
	Create(user *User) (*User, error)
	FindById(userId int64) (*User, error)
	Close() error
}

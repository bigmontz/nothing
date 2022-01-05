package repository

import "time"

type User struct {
	Username  string      `json:"username,omitempty"`
	Name      string      `json:"name,omitempty"`
	Age       uint        `json:"age,omitempty"`
	Surname   string      `json:"surname,omitempty"`
	Password  string      `json:"password,omitempty"`
	CreatedAt *time.Time  `json:"created_at,omitempty"`
	UpdatedAt *time.Time  `json:"updated_at,omitempty"`
	Id        interface{} `json:"id,omitempty"`
}

type PasswordUpdate struct {
	Current string `json:"password"`
	New     string `json:"newPassword"`
}

type UserRepository interface {
	Create(user *User) (*User, error)
	FindById(userId interface{}) (*User, error)
	UpdatePassword(userId interface{}, passwordUpdate *PasswordUpdate) (*User, error)
	Close() error
}

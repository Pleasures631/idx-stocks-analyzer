package models

import "time"

type RegisterUserRequest struct {
	Name    string `json:"name"`
	Phone   string `json:"phone"`
	Email   string `json:"email"`
	Address string `json:"address"`
}

type User struct {
	ID        uint64    `db:"id" json:"id"`
	Name      string    `db:"name" json:"name"`
	Phone     string    `db:"phone" json:"phone"`
	Email     string    `db:"email" json:"email"`
	Address   string    `db:"address" json:"address"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

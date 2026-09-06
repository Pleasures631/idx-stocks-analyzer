package models

import "time"

type RegisterUserRequest struct {
	Name     string `json:"name"`
	Phone    string `json:"phone"`
	Email    string `json:"email"`
	Address  string `json:"address"`
	Password string `json:"password"`
}

type VerifyRegistrationRequest struct {
	Email string `json:"email"`
	OTP   string `json:"otp"`
}

type PendingRegistration struct {
	ID           uint64    `db:"id"`
	Name         string    `db:"name"`
	Phone        string    `db:"phone"`
	Email        string    `db:"email"`
	Address      string    `db:"address"`
	PasswordHash string    `db:"password_hash"`
	OTPHash      string    `db:"otp_hash"`
	ExpiresAt    time.Time `db:"expires_at"`
	Attempts     int       `db:"attempts"`
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

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type ForgotPasswordRequest struct {
	Email string `json:"email"`
}

type ResetPasswordRequest struct {
	Token    string `json:"token"`
	Password string `json:"password"`
}

type AuthResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	User         *User  `json:"user"`
}

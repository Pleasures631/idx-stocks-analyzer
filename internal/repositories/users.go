package repositories

import (
	"errors"
	"time"

	"indonesia-stocks-api/internal/database"
	"indonesia-stocks-api/internal/models"

	"github.com/go-sql-driver/mysql"
)

var ErrUserConflict = errors.New("user email or phone already registered")

func CreateUser(user models.RegisterUserRequest) (*models.User, error) {
	result, err := database.DB.Exec(`
		INSERT INTO m_users (name, phone, email, address, password_hash)
		VALUES (?, ?, ?, ?, ?)
	`, user.Name, user.Phone, user.Email, user.Address, user.Password)
	if err != nil {
		var mysqlErr *mysql.MySQLError
		if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
			return nil, ErrUserConflict
		}
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}
	var created models.User
	if err := database.DB.Get(&created, `SELECT id, name, phone, email, address, created_at, updated_at FROM m_users WHERE id = ?`, id); err != nil {
		return nil, err
	}
	return &created, nil
}

func UserAlreadyRegistered(email, phone string) (bool, error) {
	var count int
	err := database.DB.Get(&count, `SELECT COUNT(*) FROM m_users WHERE email = ? OR phone = ?`, email, phone)
	return count > 0, err
}

func CreatePendingRegistration(user models.RegisterUserRequest, otpHash string, expiresAt time.Time) error {
	tx, err := database.DB.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err := tx.Exec(`DELETE FROM m_pending_registrations WHERE email = ? OR phone = ?`, user.Email, user.Phone); err != nil {
		return err
	}
	if _, err := tx.Exec(`INSERT INTO m_pending_registrations (name, phone, email, address, password_hash, otp_hash, expires_at) VALUES (?, ?, ?, ?, ?, ?, ?)`, user.Name, user.Phone, user.Email, user.Address, user.Password, otpHash, expiresAt); err != nil {
		return err
	}
	return tx.Commit()
}

func FindPendingRegistration(email string) (*models.PendingRegistration, error) {
	var pending models.PendingRegistration
	err := database.DB.Get(&pending, `SELECT id, name, phone, email, address, password_hash, otp_hash, expires_at, attempts FROM m_pending_registrations WHERE email = ? LIMIT 1`, email)
	if err != nil {
		return nil, err
	}
	return &pending, nil
}

func IncrementPendingRegistrationAttempts(id uint64) error {
	_, err := database.DB.Exec(`UPDATE m_pending_registrations SET attempts = attempts + 1 WHERE id = ?`, id)
	return err
}

func DeletePendingRegistration(id uint64) error {
	_, err := database.DB.Exec(`DELETE FROM m_pending_registrations WHERE id = ?`, id)
	return err
}

func FindUserByEmail(email string) (*models.User, string, error) {
	var row struct {
		models.User
		PasswordHash string `db:"password_hash"`
	}
	if err := database.DB.Get(&row, `SELECT id, name, phone, email, address, COALESCE(password_hash, '') AS password_hash, created_at, updated_at FROM m_users WHERE email = ? LIMIT 1`, email); err != nil {
		return nil, "", err
	}
	return &row.User, row.PasswordHash, nil
}

func CreateSession(userID uint64, accessHash, refreshHash string, accessExpiry, refreshExpiry time.Time) error {
	_, err := database.DB.Exec(`INSERT INTO m_user_sessions (user_id, access_token_hash, refresh_token_hash, access_expires_at, refresh_expires_at) VALUES (?, ?, ?, ?, ?)`, userID, accessHash, refreshHash, accessExpiry, refreshExpiry)
	return err
}

func RotateSession(refreshHash, newAccessHash, newRefreshHash string, accessExpiry, refreshExpiry time.Time) (*models.User, error) {
	tx, err := database.DB.Beginx()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	var userID uint64
	if err := tx.Get(&userID, `SELECT user_id FROM m_user_sessions WHERE refresh_token_hash = ? AND revoked_at IS NULL AND refresh_expires_at > UTC_TIMESTAMP() LIMIT 1 FOR UPDATE`, refreshHash); err != nil {
		return nil, err
	}
	if _, err := tx.Exec(`UPDATE m_user_sessions SET revoked_at = UTC_TIMESTAMP() WHERE refresh_token_hash = ?`, refreshHash); err != nil {
		return nil, err
	}
	if _, err := tx.Exec(`INSERT INTO m_user_sessions (user_id, access_token_hash, refresh_token_hash, access_expires_at, refresh_expires_at) VALUES (?, ?, ?, ?, ?)`, userID, newAccessHash, newRefreshHash, accessExpiry, refreshExpiry); err != nil {
		return nil, err
	}
	var user models.User
	if err := tx.Get(&user, `SELECT id, name, phone, email, address, created_at, updated_at FROM m_users WHERE id = ?`, userID); err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return &user, nil
}

func CreatePasswordResetToken(userID uint64, tokenHash string, expiresAt time.Time) error {
	_, err := database.DB.Exec(`INSERT INTO m_password_reset_tokens (user_id, token_hash, expires_at) VALUES (?, ?, ?)`, userID, tokenHash, expiresAt)
	return err
}

func ResetPassword(tokenHash, passwordHash string) (*models.User, error) {
	tx, err := database.DB.Beginx()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	var userID uint64
	if err := tx.Get(&userID, `SELECT user_id FROM m_password_reset_tokens WHERE token_hash = ? AND used_at IS NULL AND expires_at > UTC_TIMESTAMP() LIMIT 1 FOR UPDATE`, tokenHash); err != nil {
		return nil, err
	}
	if _, err := tx.Exec(`UPDATE m_users SET password_hash = ? WHERE id = ?`, passwordHash, userID); err != nil {
		return nil, err
	}
	if _, err := tx.Exec(`UPDATE m_password_reset_tokens SET used_at = UTC_TIMESTAMP() WHERE token_hash = ?`, tokenHash); err != nil {
		return nil, err
	}
	if _, err := tx.Exec(`UPDATE m_user_sessions SET revoked_at = UTC_TIMESTAMP() WHERE user_id = ? AND revoked_at IS NULL`, userID); err != nil {
		return nil, err
	}
	var user models.User
	if err := tx.Get(&user, `SELECT id, name, phone, email, address, created_at, updated_at FROM m_users WHERE id = ?`, userID); err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return &user, nil
}

package repositories

import (
	"errors"

	"indonesia-stocks-api/internal/database"
	"indonesia-stocks-api/internal/models"

	"github.com/go-sql-driver/mysql"
)

var ErrUserConflict = errors.New("user email or phone already registered")

func CreateUser(user models.RegisterUserRequest) (*models.User, error) {
	result, err := database.DB.Exec(`
		INSERT INTO m_users (name, phone, email, address)
		VALUES (?, ?, ?, ?)
	`, user.Name, user.Phone, user.Email, user.Address)
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

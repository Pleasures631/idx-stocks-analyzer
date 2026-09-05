package models

import "time"

// InterfaceLog merepresentasikan 1 baris tabel t_interface_log.
// Digunakan untuk mencatat setiap panggilan API eksternal.
type InterfaceLog struct {
	ID           uint64    `db:"id" json:"id"`
	FunctionName string    `db:"function_name" json:"function_name"`
	Request      string    `db:"request" json:"request"`
	Response     string    `db:"response" json:"response"`
	HTTPStatus   int       `db:"http_status" json:"http_status"`
	Success      bool      `db:"success" json:"success"`
	ErrorMessage string    `db:"error_message" json:"error_message"`
	CreatedAt    time.Time `db:"created_at" json:"created_at"`
}

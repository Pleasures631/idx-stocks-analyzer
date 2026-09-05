package repositories

import (
	"indonesia-stocks-api/internal/database"
	"indonesia-stocks-api/internal/models"
)

// InsertInterfaceLog menyimpan log panggilan API eksternal.
func InsertInterfaceLog(entry models.InterfaceLog) error {
	query := `
	INSERT INTO t_interface_log (
		function_name, request, response, http_status, success, error_message
	) VALUES (?, ?, ?, ?, ?, ?)
	`

	_, err := database.DB.Exec(query,
		entry.FunctionName,
		entry.Request,
		entry.Response,
		entry.HTTPStatus,
		entry.Success,
		entry.ErrorMessage,
	)
	return err
}

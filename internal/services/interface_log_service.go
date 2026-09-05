package services

import (
	"log"

	"indonesia-stocks-api/internal/models"
	"indonesia-stocks-api/internal/repositories"
)

// LogInterfaceCall mencatat setiap panggilan API eksternal ke t_interface_log.
// Reusable: modul baru yang melakukan HTTP call cukup memanggil fungsi ini
// dengan function_name yang merepresentasikan tujuan hit (mis. FetchIDX,
// FetchExodusMarketDetector). Logging bersifat best-effort; kegagalan insert
// tidak menghentikan alur utama pemanggil.
func LogInterfaceCall(functionName, request, response string, httpStatus int, callErr error) {
	entry := models.InterfaceLog{
		FunctionName: functionName,
		Request:      request,
		Response:     response,
		HTTPStatus:   httpStatus,
		Success:      callErr == nil,
	}
	if callErr != nil {
		entry.ErrorMessage = callErr.Error()
	}

	if err := repositories.InsertInterfaceLog(entry); err != nil {
		log.Printf("⚠️ gagal insert t_interface_log (%s): %v", functionName, err)
	}
}

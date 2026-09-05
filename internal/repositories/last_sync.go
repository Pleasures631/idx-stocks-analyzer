package repositories

import (
	"indonesia-stocks-api/internal/database"
	"time"
)

// GetLastSyncDate mengambil tanggal terakhir sync berdasarkan sync_type.
// Mengembalikan nil jika belum ada record.
func GetLastSyncDate(syncType string) (*time.Time, error) {
	query := `SELECT last_sync_date FROM t_last_sync WHERE sync_type = ?`

	var lastSyncDate time.Time
	err := database.DB.Get(&lastSyncDate, query, syncType)
	if err != nil {
		// Jika tidak ada record, return nil (bukan error)
		return nil, nil
	}
	return &lastSyncDate, nil
}

// UpdateLastSyncDate menyimpan atau update tanggal terakhir sync.
func UpdateLastSyncDate(syncType string, date time.Time) error {
	query := `
	INSERT INTO t_last_sync (sync_type, last_sync_date)
	VALUES (?, ?)
	ON DUPLICATE KEY UPDATE 
		last_sync_date = VALUES(last_sync_date),
		updated_at = NOW()
	`
	_, err := database.DB.Exec(query, syncType, date)
	return err
}

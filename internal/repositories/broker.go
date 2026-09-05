package repositories

import (
	"indonesia-stocks-api/internal/database"
	"indonesia-stocks-api/internal/models"
)

func UpsertBrokers(brokers []models.BrokerList) error {
	query := `
	INSERT INTO m_list_broker (broker_code, broker_name, broker_type, broker_license)
	VALUES (:broker_code, :broker_name, :broker_type, :broker_license)
	ON DUPLICATE KEY UPDATE
		broker_name = VALUES(broker_name),
		broker_license = VALUES(broker_license),
		broker_type = IF(VALUES(broker_type) != '', VALUES(broker_type), broker_type)
	`

	_, err := database.DB.NamedExec(query, brokers)
	return err
}

func GetAllBrokers() ([]models.BrokerList, error) {
	var list []models.BrokerList
	err := database.DB.Select(&list, "SELECT id, broker_code, broker_name, broker_type, broker_license, created_at FROM m_list_broker ORDER BY broker_code ASC")
	return list, err
}

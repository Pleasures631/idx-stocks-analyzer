package repositories

import (
	"indonesia-stocks-api/internal/database"
	"indonesia-stocks-api/internal/models"
)

func UpsertBrokers(brokers []models.BrokerList) error {
	query := `
	INSERT INTO m_list_broker (broker_code, broker_name, broker_license)
	VALUES (:broker_code, :broker_name, :broker_license)
	ON DUPLICATE KEY UPDATE
		broker_name = VALUES(broker_name),
		broker_license = VALUES(broker_license)
	`

	_, err := database.DB.NamedExec(query, brokers)
	return err
}

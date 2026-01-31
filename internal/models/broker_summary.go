package models

import "time"

type BrokerSummaryResponse struct {
	Draw            int             `json:"draw"`
	RecordsTotal    int             `json:"recordsTotal"`
	RecordsFiltered int             `json:"recordsFiltered"`
	Data            []BrokerSummary `json:"data"`
}

type BrokerSummary struct {
	No              int     `json:"No"`
	IDBrokerSummary int64   `json:"IDBrokerSummary"`
	Date            string  `json:"Date"`
	IDFirm          string  `json:"IDFirm"`
	FirmName        string  `json:"FirmName"`
	Volume          float64 `json:"Volume"`
	Value           float64 `json:"Value"`
	Frequency       float64 `json:"Frequency"`
}

type BrokerSummaryDB struct {
	ID uint64 `db:"id"`

	IdxIDBrokerSummary int64     `db:"idx_id_broker_summary"`
	TradeDate          time.Time `db:"trade_date"`

	FirmID   string `db:"firm_id"`
	FirmName string `db:"firm_name"`

	Volume    int64   `db:"volume"`
	Value     float64 `db:"value"`
	Frequency int64   `db:"frequency"`

	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

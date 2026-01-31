package models

import "time"

type IDXBroker struct {
	Code    string `json:"Code"`
	Name    string `json:"Name"`
	License string `json:"License"`
}

type IDXStock struct {
	Code         string  `json:"Code"`
	Name         string  `json:"Name"`
	ListingDate  string  `json:"ListingDate"`
	Shares       float64 `json:"Shares"`
	ListingBoard string  `json:"ListingBoard"`
}

type BrokerList struct {
	ID            uint64    `db:"id" json:"id"`
	BrokerCode    string    `db:"broker_code" json:"broker_code"`
	BrokerName    string    `db:"broker_name" json:"broker_name"`
	BrokerLicense string    `db:"broker_license" json:"broker_license"`
	CreatedAt     time.Time `db:"created_at" json:"created_at"`
}

type StocksList struct {
	ID           uint64    `db:"id" json:"id"`
	StockCode    string    `db:"stock_code" json:"stock_code"`
	StockName    string    `db:"stock_name" json:"stock_name"`
	ListingDate  time.Time `db:"listing_date" json:"listing_date"`
	TotalShares  uint64    `db:"total_shares" json:"total_shares"`
	ListingBoard string    `db:"listing_board" json:"listing_board"`
	IsActive     bool      `db:"is_active" json:"is_active"`
	CreatedAt    time.Time `db:"created_at" json:"created_at"`
}

package models

import "time"

type StockbitIHSGQuote struct {
	Symbol        string `json:"symbol"`
	Name          string `json:"name"`
	Price         string `json:"price"`
	Change        string `json:"change"`
	ChangePct     string `json:"change_pct"`
	AsOf          string `json:"as_of"`
	Volume        string `json:"volume"`
	AverageVolume string `json:"average_volume"`
	SourceURL     string `json:"source_url"`
}

type StockbitIHSGChartPoint struct {
	ID         uint64    `db:"id" json:"-"`
	Symbol     string    `db:"symbol" json:"symbol"`
	TradeDate  time.Time `db:"trade_date" json:"trade_date"`
	Interval   string    `db:"interval" json:"interval"`
	ObservedAt time.Time `db:"observed_at" json:"observed_at"`
	XLabel     string    `db:"xlabel" json:"xlabel"`
	Value      float64   `db:"value" json:"value"`
	Percentage float64   `db:"percentage" json:"percentage"`
	Change     float64   `db:"change_value" json:"change"`
	CreatedAt  time.Time `db:"created_at" json:"-"`
	UpdatedAt  time.Time `db:"updated_at" json:"-"`
}

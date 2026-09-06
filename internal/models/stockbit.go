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

type StockbitForeignDomesticDB struct {
	ID                 uint64    `db:"id" json:"-"`
	Symbol             string    `db:"symbol" json:"symbol"`
	TradeDate          time.Time `db:"trade_date" json:"trade_date"`
	MarketType         string    `db:"market_type" json:"market_type"`
	ForeignBuyValue    int64     `db:"foreign_buy_value" json:"foreign_buy_value"`
	ForeignSellValue   int64     `db:"foreign_sell_value" json:"foreign_sell_value"`
	DomesticBuyValue   int64     `db:"domestic_buy_value" json:"domestic_buy_value"`
	DomesticSellValue  int64     `db:"domestic_sell_value" json:"domestic_sell_value"`
	ForeignNetValue    int64     `db:"foreign_net_value" json:"foreign_net_value"`
	DomesticNetValue   int64     `db:"domestic_net_value" json:"domestic_net_value"`
	ForeignBuyVolume   int64     `db:"foreign_buy_volume" json:"foreign_buy_volume"`
	ForeignSellVolume  int64     `db:"foreign_sell_volume" json:"foreign_sell_volume"`
	DomesticBuyVolume  int64     `db:"domestic_buy_volume" json:"domestic_buy_volume"`
	DomesticSellVolume int64     `db:"domestic_sell_volume" json:"domestic_sell_volume"`
	ForeignBuyFreq     int64     `db:"foreign_buy_frequency" json:"foreign_buy_frequency"`
	ForeignSellFreq    int64     `db:"foreign_sell_frequency" json:"foreign_sell_frequency"`
	DomesticBuyFreq    int64     `db:"domestic_buy_frequency" json:"domestic_buy_frequency"`
	DomesticSellFreq   int64     `db:"domestic_sell_frequency" json:"domestic_sell_frequency"`
	LastUpdated        string    `db:"last_updated" json:"last_updated"`
	CreatedAt          time.Time `db:"created_at" json:"-"`
	UpdatedAt          time.Time `db:"updated_at" json:"-"`
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

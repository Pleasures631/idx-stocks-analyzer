package models

// TickerPriceBar merepresentasikan 1 bar harga harian untuk chart.
type TickerPriceBar struct {
	TradeDate string  `db:"trade_date" json:"trade_date"`
	Open      float64 `db:"open_price" json:"open"`
	High      float64 `db:"high_price" json:"high"`
	Low       float64 `db:"low_price" json:"low"`
	Close     float64 `db:"close_price" json:"close"`
	Volume    float64 `db:"volume" json:"volume"`
	ChangePct float64 `db:"change_pct" json:"change_pct"`
}

// TickerBrokerVolume merepresentasikan agregat trading per broker dalam rentang.
type TickerBrokerVolume struct {
	BrokerCode string  `db:"broker_code" json:"broker_code"`
	BrokerName string  `db:"broker_name" json:"broker_name"`
	BrokerType string  `db:"broker_type" json:"broker_type"`
	BuyLot     float64 `db:"buy_lot" json:"buy_lot"`
	SellLot    float64 `db:"sell_lot" json:"sell_lot"`
	BuyVolume  float64 `db:"buy_volume" json:"buy_volume"`
	SellVolume float64 `db:"sell_volume" json:"sell_volume"`
	BuyValue   float64 `db:"buy_value" json:"buy_value"`
	SellValue  float64 `db:"sell_value" json:"sell_value"`
	NetValue   float64 `db:"net_value" json:"net_value"`
	ActiveDays int     `db:"active_days" json:"active_days"`
}

// TickerBrokerSummary merepresentasikan buy/sell per broker pada 1 hari.
type TickerBrokerSummary struct {
	TradeDate   string  `db:"trade_date" json:"trade_date"`
	BrokerCode  string  `db:"broker_code" json:"broker_code"`
	BrokerName  string  `db:"broker_name" json:"broker_name"`
	BrokerType  string  `db:"broker_type" json:"broker_type"`
	BrokerGroup string  `db:"broker_group" json:"broker_group"`
	BuyLot      float64 `db:"buy_lot" json:"buy_lot"`
	SellLot     float64 `db:"sell_lot" json:"sell_lot"`
	BuyVolume   float64 `db:"buy_volume" json:"buy_volume"`
	SellVolume  float64 `db:"sell_volume" json:"sell_volume"`
	BuyValue    float64 `db:"buy_value" json:"buy_value"`
	SellValue   float64 `db:"sell_value" json:"sell_value"`
	NetValue    float64 `db:"net_value" json:"net_value"`
	Frequency   int64   `db:"frequency" json:"frequency"`
}

// TickerDetailResponse adalah payload lengkap untuk halaman detail sebuah saham.
type TickerDetail struct {
	Symbol      string                `json:"symbol"`
	StockName   string                `json:"stock_name"`
	Range       string                `json:"range"`
	From        string                `json:"from"`
	To          string                `json:"to"`
	PriceChart  []TickerPriceBar      `json:"price_chart"`
	ByBroker    []TickerBrokerVolume  `json:"volume_by_broker"`
	BrokerDaily []TickerBrokerSummary `json:"broker_summary"`
}

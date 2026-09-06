package models

import (
	"encoding/json"
	"time"
)

type ExodusResponse struct {
	Message string             `json:"message"`
	Data    ExodusMarketDetector `json:"data"`
}

type ExodusMarketDetector struct {
	BandarDetector json.RawMessage `json:"bandar_detector"`
	BrokerSummary  ExodusBrokerSummary `json:"broker_summary"`
	From           string             `json:"from"`
	To             string             `json:"to"`
}

type ExodusBrokerSummary struct {
	Symbol     string             `json:"symbol"`
	BrokersBuy  []ExodusBrokerBuy  `json:"brokers_buy"`
	BrokersSell []ExodusBrokerSell `json:"brokers_sell"`
}

type ExodusBrokerBuy struct {
	Blot         string `json:"blot"`
	Blotv        string `json:"blotv"`
	Bval         string `json:"bval"`
	Bvalv        string `json:"bvalv"`
	BrokerCode   string `json:"netbs_broker_code"`
	BuyAvgPrice  string `json:"netbs_buy_avg_price"`
	Date         string `json:"netbs_date"`
	StockCode    string `json:"netbs_stock_code"`
	Type         string `json:"type"`
	Freq         string `json:"freq"`
}

type ExodusBrokerSell struct {
	BrokerCode   string `json:"netbs_broker_code"`
	Date         string `json:"netbs_date"`
	SellAvgPrice string `json:"netbs_sell_avg_price"`
	StockCode    string `json:"netbs_stock_code"`
	Slot         string `json:"slot"`
	Slotv        string `json:"slotv"`
	Sval         string `json:"sval"`
	Svalv        string `json:"svalv"`
	Type         string `json:"type"`
	Freq         string `json:"freq"`
}

type ExodusBrokerSummaryDB struct {
	ID         uint64    `db:"id" json:"id"`
	StockCode  string    `db:"stock_code" json:"stock_code"`
	TradeDate  time.Time `db:"trade_date" json:"trade_date"`
	BrokerCode string    `db:"broker_code" json:"broker_code"`
	Side       string    `db:"side" json:"side"`
	BrokerType string    `db:"broker_type" json:"broker_type"`

	Lot      float64 `db:"lot" json:"lot"`
	Volume   float64 `db:"volume" json:"volume"`
	Value    float64 `db:"value" json:"value"`
	Turnover float64 `db:"turnover" json:"turnover"`
	AvgPrice float64 `db:"avg_price" json:"avg_price"`
	Frequency int64   `db:"frequency" json:"frequency"`

	CreatedAt time.Time `db:"created_at" json:"-"`
	UpdatedAt time.Time `db:"updated_at" json:"-"`
}

// ExodusStockDate merepresentasikan pasangan (stock_code, trade_date) yang
// sudah ada di t_exodus_broker_summary. Dipakai untuk me-skip hit Exodus.
type ExodusStockDate struct {
	StockCode string    `db:"stock_code" json:"stock_code"`
	TradeDate time.Time `db:"trade_date" json:"trade_date"`
}

type ExodusBrokerFlow struct {
	BrokerCode string  `db:"broker_code" json:"broker_code"`
	BrokerType string  `db:"broker_type" json:"broker_type"`
	BuyValue   float64 `db:"buy_value" json:"buy_value"`
	SellValue  float64 `db:"sell_value" json:"sell_value"`
	NetValue   float64 `db:"net_value" json:"net_value"`
	BuyLot     float64 `db:"buy_lot" json:"buy_lot"`
	SellLot    float64 `db:"sell_lot" json:"sell_lot"`
	NetLot     float64 `db:"net_lot" json:"net_lot"`
	ActiveDays int     `db:"active_days" json:"active_days"`

	FormattedNetValue string `json:"formatted_net_value"`
	DisplayStatus     string `json:"display_status"`
}

type ExodusFlowPhase struct {
	Symbol          string             `json:"symbol"`
	StartDate       string             `json:"start_date"`
	EndDate         string             `json:"end_date"`
	TotalDays       int                `json:"total_days"`
	Phase           string             `json:"phase"`
	TotalBuyValue   float64            `json:"total_buy_value"`
	TotalSellValue  float64            `json:"total_sell_value"`
	NetValue        float64            `json:"net_value"`
	ForeignNetValue float64            `json:"foreign_net_value"`
	GovernmentNet   float64            `json:"government_net"`
	LocalNetValue   float64            `json:"local_net_value"`
	TotalBrokers    int                `json:"total_brokers"`

	// Broker Group Breakdown
	RetailNet        float64 `json:"retail_net"`
	InstitutionalNet float64 `json:"institutional_net"`
	LocalMidNet      float64 `json:"local_mid_net"`

	// Metrics
	SmartMoneyRatio   float64 `json:"smart_money_ratio"`  // (Foreign+Inst) / Total
	RetailDominance   float64 `json:"retail_dominance"`   // Retail / Total (if net positive)
	Top1Concentration float64 `json:"top1_concentration"` // Top1 / Top3 net value
	ForeignLeadership bool    `json:"foreign_leadership"` // Top1 is FOREIGN group
	BuyHHI            float64 `json:"buy_hhi"`            // HHI of broker buy-value shares, 0-10000
	SellHHI           float64 `json:"sell_hhi"`           // HHI of broker sell-value shares, 0-10000
	TotalHHI          float64 `json:"total_hhi"`          // HHI of broker buy+sell-value shares, 0-10000

	// --- New Improvement Metrics ---
	SmartMoneyActiveDays  int     `json:"smart_money_active_days"` // hari asing+inst net buy
	SmartMoneyConsistency float64 `json:"smart_money_consistency"` // active/active_days_in_window (%)
	SmartMoneyMomentum    float64 `json:"smart_money_momentum"`    // second half net first-half net (Rp)
	FirstHalfDate         string  `json:"first_half_date"`
	SecondHalfDate        string  `json:"second_half_date"`
	FirstHalfNet          float64 `json:"first_half_net"`
	SecondHalfNet         float64 `json:"second_half_net"`
	MomentumAccelerating  bool    `json:"momentum_accelerating"` // buying makin besar di akhir
	PriceChangePct        float64 `json:"price_change_pct"`      // % perubahan harga selama window
	PriceConfirms         bool    `json:"price_confirms"`        // smart money net & harga searah
	VolumeSpikeRatio      float64 `json:"volume_spike_ratio"`    // avg vol window / avg vol baseline
	HasVolumeSpike        bool    `json:"has_volume_spike"`

	// Anomaly (per-hari, broker yang tiba-tiba beli/jual sangat besar)
	Anomalies []ExodusAnomaly `json:"anomalies"`

	FormattedBuyValue   string `json:"formatted_buy_value"`
	FormattedSellValue  string `json:"formatted_sell_value"`
	FormattedNetValue   string `json:"formatted_net_value"`
	FormattedForeignNet string `json:"formatted_foreign_net"`

	BrokersAccumulation []ExodusBrokerFlow `json:"brokers_accumulation"`
	BrokersDistribution []ExodusBrokerFlow `json:"brokers_distribution"`
	DisplayStatus       string             `json:"display_status"`
}

// DailyGroupNet merepresentasikan net flow harian sekelompok broker
// (dipakai untuk smart money consistency & momentum).
type DailyGroupNet struct {
	TradeDate string  `db:"trade_date" json:"trade_date"`
	NetValue  float64 `db:"net_value" json:"net_value"`
}

// DailyGroupFlow merepresentasikan net flow harian per broker group.
type DailyGroupFlow struct {
	TradeDate   string  `db:"trade_date" json:"trade_date"`
	BrokerGroup string  `db:"broker_group" json:"broker_group"`
	NetValue    float64 `db:"net_value" json:"net_value"`
}

// DailyBrokerNet merepresentasikan net harian per broker untuk anomaly detection.
type DailyBrokerNet struct {
	TradeDate  string  `db:"trade_date" json:"trade_date"`
	BrokerCode string  `db:"broker_code" json:"broker_code"`
	BrokerType string  `db:"broker_type" json:"broker_type"`
	NetValue   float64 `db:"net_value" json:"net_value"`
}

// ExodusAnomaly merepresentasikan 1 hari di mana sebuah broker melakukan
// transaksi sangat besar (anomali) berdasarkan Modified Z-Score (Median+MAD).
type ExodusAnomaly struct {	StockCode    string  `json:"stock_code"`
	BrokerCode   string  `json:"broker_code"`
	BrokerType   string  `json:"broker_type"`
	TradeDate    string  `json:"trade_date"`
	NetValue     float64 `json:"net_value"`
	FormattedNet string  `json:"formatted_net"`
	ZScore       float64 `json:"z_score"`
	IsMarketMaker bool   `json:"is_market_maker"`
}
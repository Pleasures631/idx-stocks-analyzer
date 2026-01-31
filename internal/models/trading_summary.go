package models

import "time"

type TradingSummaryResponse struct {
	Draw            int              `json:"draw"`
	RecordsTotal    int              `json:"recordsTotal"`
	RecordsFiltered int              `json:"recordsFiltered"`
	Data            []TradingSummary `json:"data"`
}

type TradingSummary struct {
	No                  int      `json:"No"`
	IDStockSummary      int64    `json:"IDStockSummary"`
	Date                string   `json:"Date"`
	StockCode           string   `json:"StockCode"`
	StockName           string   `json:"StockName"`
	Previous            float64  `json:"Previous"`
	OpenPrice           float64  `json:"OpenPrice"`
	FirstTrade          float64  `json:"FirstTrade"`
	High                float64  `json:"High"`
	Low                 float64  `json:"Low"`
	Close               float64  `json:"Close"`
	Change              float64  `json:"Change"`
	Volume              float64  `json:"Volume"`
	Value               float64  `json:"Value"`
	Frequency           float64  `json:"Frequency"`
	IndexIndividual     float64  `json:"IndexIndividual"`
	Offer               float64  `json:"Offer"`
	OfferVolume         float64  `json:"OfferVolume"`
	Bid                 float64  `json:"Bid"`
	BidVolume           float64  `json:"BidVolume"`
	ListedShares        float64  `json:"ListedShares"`
	TradebleShares      float64  `json:"TradebleShares"`
	WeightForIndex      float64  `json:"WeightForIndex"`
	ForeignSell         float64  `json:"ForeignSell"`
	ForeignBuy          float64  `json:"ForeignBuy"`
	DelistingDate       string   `json:"DelistingDate"`
	NonRegularVolume    float64  `json:"NonRegularVolume"`
	NonRegularValue     float64  `json:"NonRegularValue"`
	NonRegularFrequency float64  `json:"NonRegularFrequency"`
	Persen              *float64 `json:"persen"`
	Percentage          *float64 `json:"percentage"`
}

type TradingSummaryDB struct {
	ID uint64 `db:"id"`

	IdxIDStockSummary int64     `db:"idx_id_stock_summary"`
	TradeDate         time.Time `db:"trade_date"`

	StockCode string `db:"stock_code"`
	StockName string `db:"stock_name"`

	Previous      float64 `db:"previous_price"`
	OpenPrice     float64 `db:"open_price"`
	FirstTrade    float64 `db:"first_trade"`
	High          float64 `db:"high_price"`
	Low           float64 `db:"low_price"`
	Close         float64 `db:"close_price"`
	Change        float64 `db:"change_price"`
	CloseStrength float64 `db:"close_strength"`

	Volume    int64   `db:"volume"`
	Value     float64 `db:"value"`
	Frequency int64   `db:"frequency"`

	IndexIndividual float64 `db:"index_individual"`

	Offer       float64 `db:"offer_price"`
	OfferVolume int64   `db:"offer_volume"`
	Bid         float64 `db:"bid_price"`
	BidVolume   int64   `db:"bid_volume"`

	ListedShares    int64   `db:"listed_shares"`
	TradeableShares int64   `db:"tradeable_shares"`
	WeightForIndex  float64 `db:"weight_for_index"`

	ForeignSell float64 `db:"foreign_sell"`
	ForeignBuy  float64 `db:"foreign_buy"`

	NonRegularVolume    int64   `db:"non_regular_volume"`
	NonRegularValue     float64 `db:"non_regular_value"`
	NonRegularFrequency int64   `db:"non_regular_frequency"`

	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

type TopAccumulation struct {
	StockCode        string    `db:"stock_code" json:"stock_code"`
	StockName        string    `db:"stock_name" json:"stock_name"`
	AvgCloseStrength float64   `db:"avg_close_strength" json:"avg_close_strength"`
	LastTradeDate    time.Time `db:"last_trade_date" json:"last_trade_date"`
	LastPrice        float64   `db:"last_price" json:"last_price"`
	LastChange       float64   `db:"last_change" json:"last_change"`

	// Data Mentah (Hidden from JSON)
	NetForeign    float64 `db:"net_foreign" json:"-"`
	AvgValue      float64 `db:"avg_value" json:"-"`
	TotalVolume   float64 `db:"total_volume" json:"-"`
	LastVolume    float64 `db:"last_volume" json:"-"`
	AvgVol20      float64 `db:"last_avg_vol20" json:"-"`
	Ma20          float64 `db:"last_ma20" json:"-"`
	Ma50          float64 `db:"last_ma50" json:"-"`
	LastRes20     float64 `db:"last_res_20" json:"-"`
	BreakoutScore float64 `db:"breakout_score" json:"-"`

	// Tampilan Cantik
	FormattedNetForeign string `json:"net_foreign"`
	FormattedAvgValue   string `json:"avg_value"`
	DisplayStatus       string `json:"status"`
}

type TopAccumulationEod struct {
	StockCode        string    `db:"stock_code" json:"stock_code"`
	StockName        string    `db:"stock_name" json:"stock_name"`
	AvgCloseStrength float64   `db:"avg_close_strength" json:"avg_close_strength"`
	NetForeign       float64   `db:"net_foreign" json:"net_foreign"`
	AvgValue         float64   `db:"avg_value" json:"avg_value"`
	LastTradeDate    time.Time `db:"last_trade_date" json:"last_trade_date"`
	LastPrice        float64   `db:"last_price" json:"last_price"`
	LastChange       float64   `db:"last_change" json:"last_change"`
	LastVolume       float64   `db:"last_volume" json:"last_volume"`
	LastAvgVol20     float64   `db:"last_avg_vol20" json:"last_avg_vol20"`
	LastMa20         float64   `db:"last_ma20" json:"last_ma20"`
	LastMa50         float64   `db:"last_ma50" json:"last_ma50"`
	LastRes20        float64   `db:"last_res_20" json:"last_res_20"`
	LastSup20        float64   `db:"last_sup_20" json:"last_sup_20"`
	BreakoutScore    float64   `db:"breakout_score" json:"breakout_score"`

	// Field tambahan untuk formatting dan display logic
	FormattedNetForeign string `json:"formatted_net_foreign"`
	FormattedAvgValue   string `json:"formatted_avg_value"`
	DisplayStatus       string `json:"display_status"`

	LocalParticipation float64 `db:"local_participation" json:"local_participation"` // Persentase lokal
	LastRsi            float64 `db:"last_rsi" json:"last_rsi"`
}

type TopSwinger struct {
	StockCode     string  `db:"stock_code" json:"stock_code"`
	StockName     string  `db:"stock_name" json:"stock_name"`
	TradeDate     string  `db:"trade_date" json:"trade_date"`
	ClosePrice    float64 `db:"close_price" json:"close_price"`
	HighPrice     float64 `db:"high_price" json:"high_price"`
	LowPrice      float64 `db:"low_price" json:"low_price"`
	CloseStr      float64 `db:"close_strength" json:"close_strength"`
	Volume        float64 `db:"volume" json:"volume"`
	Value         float64 `db:"value" json:"value"`
	NetForeign    float64 `db:"net_foreign" json:"net_foreign"`
	AvgStrength5D float64 `db:"avg_strength_5d" json:"avg_strength_5d"`
	VolChangePct  float64 `db:"vol_change_pct" json:"vol_change_pct"`
	SwingScore    float64 `db:"swing_score" json:"swing_score"`
	EntryPrice    float64 `db:"entry_price" json:"entry_price"`
	StopLoss      float64 `db:"stop_loss" json:"stop_loss"`
	TakeProfit    float64 `db:"take_profit" json:"take_profit"`
	VolMultiplier float64 `db:"vol_multiplier" json:"vol_multiplier"`

	// Tambahkan ini untuk menampung alias dari SQL
	PrevCloseVal float64 `db:"prev_close_val" json:"-"`

	DisplayStatus string `json:"display_status"`
}

type BacktestResult struct {
	StockCode string `db:"stock_code" json:"stock_code"`
	StockName string `db:"stock_name" json:"stock_name"`

	// Harga & Indikator di Masa Lalu (Then)
	PriceThen float64 `db:"price_then" json:"price_then"`
	ResThen   float64 `db:"res_20_then" json:"res_then"`
	Ma20Then  float64 `db:"ma20_then" json:"ma20_then"`

	// Harga Sekarang (Now)
	PriceNow float64 `db:"price_now" json:"price_now"`

	// Kalkulasi Backtest (Diisi di loop Go)
	ProfitLossPct float64 `json:"profit_loss_pct"`
	SignalAtThen  string  `json:"signal_at_then"`
	ResultStatus  string  `json:"result_status"`
}

type BacktestResponse struct {
	Details   []BacktestResult `json:"details"`
	WinRate   float64          `json:"win_rate"`
	TotalWin  int              `json:"total_win"`
	TotalLose int              `json:"total_lose"`
	AvgProfit float64          `json:"avg_profit"`
}

type SilentAccumulation struct {
	StockCode        string    `db:"stock_code" json:"stock_code"`
	StockName        string    `db:"stock_name" json:"stock_name"`
	AvgCloseStrength float64   `db:"avg_close_strength" json:"avg_close_strength"`
	NetForeign       float64   `db:"net_foreign" json:"net_foreign"`
	AvgValue         float64   `db:"avg_value" json:"avg_value"`
	LastTradeDate    time.Time `db:"last_trade_date" json:"last_trade_date"`
	LastPrice        float64   `db:"last_price" json:"last_price"`
	LastChange       float64   `db:"last_change" json:"last_change"`
	LastVolume       float64   `db:"last_volume" json:"last_volume"`
	LastAvgVol20     float64   `db:"last_avg_vol20" json:"last_avg_vol20"`
	LastMa20         float64   `db:"last_ma20" json:"last_ma20"`
	LastMa50         float64   `db:"last_ma50" json:"last_ma50"`
	LastRes20        float64   `db:"last_res_20" json:"last_res_20"`
	LastSup20        float64   `db:"last_sup_20" json:"last_sup_20"`
	BreakoutScore    float64   `db:"breakout_score" json:"breakout_score"`

	// Field Kunci untuk Strategi Lu (Silent Accumulation)
	LocalParticipation float64 `db:"local_participation" json:"local_participation"`

	// Field untuk mempermudah tampilan (Display Logic)
	FormattedNetForeign string  `json:"formatted_net_foreign"`
	FormattedAvgValue   string  `json:"formatted_avg_value"`
	RetailSentiment     string  `json:"retail_sentiment"`
	DisplayStatus       string  `json:"display_status"`
	DistToSupport       float64 `db:"dist_to_support" json:"dist_to_support"`
	LastAvgVol100       float64 `db:"last_avg_vol100" json:"last_avg_vol100"`
}

type StatisticSingleStock struct {
	StockCode          string    `db:"code" json:"-"`
	TradeDate          time.Time `db:"date" json:"-"`
	TradeDateFormatted string    `db:"date" json:"date"`
	CloseStrength      string    `db:"strength" json:"strength"`
	Price              float64   `db:"price" json:"price"`
	Volume             float64   `db:"vol" json:"-"`
	VolumeFormatted    string    `db:"vol" json:"vol"`
	ChangePrice        float64   `db:"change_price" json:"change_price"`
	TrendStatus        string    `db:"trend_status" json:"trend_status"`
	VolChangePercent   string    `db:"vol_change_percent" json:"vol_change_percent"`
}

type StatisticSingleStockMapped struct {
	StockCode string                 `db:"code" json:"Stock Code"`
	Details   []StatisticSingleStock `json:details`
}

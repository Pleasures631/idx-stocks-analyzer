package models

import "time"

// BacktestRunRequest adalah payload request POST /api/v1/backtest/run
type BacktestRunRequest struct {
	RunName        string  `json:"run_name"`
	StartDate      string  `json:"start_date"`
	EndDate        string  `json:"end_date"`
	TPPercent      float64 `json:"tp_percent"`
	SLPercent      float64 `json:"sl_percent"`
	MaxHoldingDays int     `json:"max_holding_days"`
}

// BacktestSignal adalah kandidat sinyal hasil query deteksi signal date.
// Digunakan hanya untuk scan dari database.
type BacktestSignal struct {
	StockCode  string    `db:"stock_code"`
	SignalDate time.Time `db:"signal_date"`
}

// BacktestDailyBar adalah baris data OHLC harian dari t_trading_summary.
// Digunakan untuk simulasi entry/exit. Scan dari database.
type BacktestDailyBar struct {
	StockCode  string    `db:"stock_code"`
	TradeDate  time.Time `db:"trade_date"`
	OpenPrice  float64   `db:"open_price"`
	HighPrice  float64   `db:"high_price"`
	LowPrice   float64   `db:"low_price"`
	ClosePrice float64   `db:"close_price"`
}

// BacktestRun merepresentasikan 1 baris tabel t_backtest_run.
type BacktestRun struct {
	ID                uint64  `db:"id" json:"id"`
	RunName           string  `db:"run_name" json:"run_name"`
	StartDate         string  `db:"start_date" json:"start_date"`
	EndDate           string  `db:"end_date" json:"end_date"`
	TPPercent         float64 `db:"tp_percent" json:"tp_percent"`
	SLPercent         float64 `db:"sl_percent" json:"sl_percent"`
	MaxHoldingDays    int     `db:"max_holding_days" json:"max_holding_days"`
	TotalSignals      int     `db:"total_signals" json:"total_signals"`
	TotalTrades       int     `db:"total_trades" json:"total_trades"`
	GrossProfit       float64 `db:"gross_profit" json:"gross_profit"`
	GrossLoss         float64 `db:"gross_loss" json:"gross_loss"`
	WinTrades         int     `db:"win_trades" json:"win_trades"`
	LossTrades        int     `db:"loss_trades" json:"loss_trades"`
	ExpiredTrades     int     `db:"expired_trades" json:"expired_trades"`
	WinRate           float64 `db:"win_rate" json:"win_rate"`
	ProfitFactor      float64 `db:"profit_factor" json:"profit_factor"`
	Expectancy        float64 `db:"expectancy" json:"expectancy"`
	AvgHoldingDays    float64 `db:"avg_holding_days" json:"avg_holding_days"`
	TotalReturnPercent float64 `db:"total_return_percent" json:"total_return_percent"`
	MaxDrawdown       float64 `db:"max_drawdown" json:"max_drawdown"`
	CreatedAt         time.Time `db:"created_at" json:"created_at"`
}

// BacktestDetail merepresentasikan 1 baris tabel t_backtest_detail.
type BacktestDetail struct {
	ID             uint64  `db:"id" json:"id"`
	BacktestRunID  uint64  `db:"backtest_run_id" json:"backtest_run_id"`
	StockCode      string  `db:"stock_code" json:"stock_code"`
	SignalDate     string  `db:"signal_date" json:"signal_date"`
	EntryDate      string  `db:"entry_date" json:"entry_date"`
	EntryPrice     float64 `db:"entry_price" json:"entry_price"`
	TargetTP       float64 `db:"target_tp" json:"target_tp"`
	TargetSL       float64 `db:"target_sl" json:"target_sl"`
	ExitDate       string  `db:"exit_date" json:"exit_date"`
	ExitPrice      float64 `db:"exit_price" json:"exit_price"`
	ExitReason     string  `db:"exit_reason" json:"exit_reason"`
	HoldingDays    int     `db:"holding_days" json:"holding_days"`
	Status         string  `db:"status" json:"status"`
	ReturnPercent  float64 `db:"return_percent" json:"return_percent"`
	CreatedAt      time.Time `db:"created_at" json:"created_at"`
}

// BacktestRunResponse adalah output API setelah run selesai.
type BacktestRunResponse struct {
	Run     BacktestRun      `json:"run"`
	Details []BacktestDetail `json:"details"`
}

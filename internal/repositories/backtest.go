package repositories

import (
	"time"

	"indonesia-stocks-api/internal/database"
	"indonesia-stocks-api/internal/models"

	"github.com/jmoiron/sqlx"
)

// GetBacktestSignals mendeteksi kandidat sinyal pada rentang tanggal.
// Signal Rules (Bandarmologi V3 - Balanced Quality):
//   1. Smart Money: top-1 broker akumulasi harus Asing & foreign net buy > 0
//   2. Broad Accumulation: top-3 broker semuanya net buy (broker_count = 3), strength >= 0.25
//   3. Volume Confirmation: volume spike >2x SMA5 pada candle bullish (close > open)
//   4. Price Structure: close_strength >= 0.5 (close di upper half — buying pressure)
//   5. Trend Alignment: close > MA20 (bullish trend)
func GetBacktestSignals(from, to string) ([]models.BacktestSignal, error) {
	fromTime, err := time.Parse("2006-01-02", from)
	if err != nil {
		return nil, err
	}
	fromSMA := fromTime.AddDate(0, 0, -20).Format("2006-01-02")

	query := `
	WITH vol_window AS (
		SELECT
			stock_code,
			trade_date,
			volume,
			close_price,
			AVG(volume) OVER (
				PARTITION BY stock_code
				ORDER BY trade_date
				ROWS BETWEEN 5 PRECEDING AND 1 PRECEDING
			) AS sma5_vol,
			AVG(close_price) OVER (
				PARTITION BY stock_code
				ORDER BY trade_date
				ROWS BETWEEN 19 PRECEDING AND CURRENT ROW
			) AS ma20_close
		FROM t_trading_summary
		WHERE trade_date >= DATE_SUB(?, INTERVAL 20 DAY)
		  AND trade_date <= ?
	)
	SELECT
		ts.stock_code,
		ts.trade_date AS signal_date
	FROM t_trading_summary ts
	JOIN t_exodus_broker_agg ba
		ON ts.stock_code = ba.stock_code COLLATE utf8mb4_0900_ai_ci
	   AND ts.trade_date = ba.trade_date
	JOIN vol_window v
		ON ts.stock_code = v.stock_code
	   AND ts.trade_date = v.trade_date
	WHERE ts.trade_date BETWEEN ? AND ?
AND ts.value > 0
  AND ba.broker_count = 3
  AND ba.top1_asing = 1
  AND ba.top3_net_buy / ts.value >= 0.25
  AND ba.foreign_net_buy > 0
  AND v.sma5_vol IS NOT NULL
  AND v.ma20_close IS NOT NULL
  AND v.volume > v.sma5_vol * 2.0
  AND ts.close_price > v.ma20_close
  AND ts.close_price > ts.open_price
  AND ts.close_strength >= 0.5
  `

	rows := []models.BacktestSignal{}
	err = database.DB.Select(&rows, query, fromSMA, to, from, to)
	if err != nil {
		return nil, err
	}

	return rows, nil
}

// GetBacktestDailyBars mengambil seluruh baris OHLC harian pada rentang tanggal.
// Dipakai service untuk mensimulasikan entry/exit.
func GetBacktestDailyBars(from, to string) ([]models.BacktestDailyBar, error) {
	query := `
	SELECT
		stock_code,
		trade_date,
		open_price,
		high_price,
		low_price,
		close_price
	FROM t_trading_summary
	WHERE trade_date BETWEEN ? AND ?
	ORDER BY stock_code, trade_date
	`

	rows := []models.BacktestDailyBar{}
	err := database.DB.Select(&rows, query, from, to)
	if err != nil {
		return nil, err
	}

	return rows, nil
}

// InsertBacktestRun menyimpan master run backtest dan mengembalikan ID-nya.
func InsertBacktestRun(tx *sqlx.Tx, run *models.BacktestRun) (uint64, error) {
	query := `
	INSERT INTO t_backtest_run (
		run_name, start_date, end_date, tp_percent, sl_percent,
		max_holding_days, total_signals, total_trades, gross_profit, gross_loss,
		win_trades, loss_trades, expired_trades, win_rate, profit_factor,
		expectancy, avg_holding_days, total_return_percent, max_drawdown
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	result, err := tx.Exec(query,
		run.RunName, run.StartDate, run.EndDate, run.TPPercent, run.SLPercent,
		run.MaxHoldingDays, run.TotalSignals, run.TotalTrades, run.GrossProfit, run.GrossLoss,
		run.WinTrades, run.LossTrades, run.ExpiredTrades, run.WinRate, run.ProfitFactor,
		run.Expectancy, run.AvgHoldingDays, run.TotalReturnPercent, run.MaxDrawdown,
	)
	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return uint64(id), nil
}

// InsertBacktestDetails menyimpan detail trade secara bulk dalam 1 query.
func InsertBacktestDetails(tx *sqlx.Tx, details []models.BacktestDetail) error {
	if len(details) == 0 {
		return nil
	}

	query := `
	INSERT INTO t_backtest_detail (
		backtest_run_id, stock_code, signal_date, entry_date, entry_price,
		target_tp, target_sl, exit_date, exit_price, exit_reason,
		holding_days, status, return_percent
	) VALUES (
		:backtest_run_id, :stock_code, :signal_date, :entry_date, :entry_price,
		:target_tp, :target_sl, :exit_date, :exit_price, :exit_reason,
		:holding_days, :status, :return_percent
	)
	`

	_, err := tx.NamedExec(query, details)
	return err
}

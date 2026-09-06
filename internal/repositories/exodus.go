package repositories

import (
	"time"

	"indonesia-stocks-api/internal/database"
	"indonesia-stocks-api/internal/models"
)

func UpsertExodusBrokerSummary(rows []models.ExodusBrokerSummaryDB) error {
	query := `
	INSERT INTO t_exodus_broker_summary (
		stock_code,
		trade_date,
		broker_code,
		side,
		broker_type,
		lot,
		volume,
		value,
		turnover,
		avg_price,
		frequency,
		created_at,
		updated_at
	)
	VALUES (
		:stock_code,
		:trade_date,
		:broker_code,
		:side,
		:broker_type,
		:lot,
		:volume,
		:value,
		:turnover,
		:avg_price,
		:frequency,
		:created_at,
		:updated_at
	)
	ON DUPLICATE KEY UPDATE
		broker_type = VALUES(broker_type),
		lot = VALUES(lot),
		volume = VALUES(volume),
		value = VALUES(value),
		turnover = VALUES(turnover),
		avg_price = VALUES(avg_price),
		frequency = VALUES(frequency),
		updated_at = VALUES(updated_at)
	`

	_, err := database.DB.NamedExec(query, rows)
	if err != nil {
		return err
	}

	return syncExodusDerived(rows)
}

// syncExodusDerived merekomputasi tabel derived (t_exodus_broker_net &
// t_exodus_broker_agg) untuk setiap (stock_code, trade_date) yang ada di batch
// upsert, sehingga data backtest selalu sinkron dengan tabel master.
func syncExodusDerived(rows []models.ExodusBrokerSummaryDB) error {
	seen := map[string]struct{}{}
	type pair struct{ stock string; date string }
	pairs := []pair{}
	for _, r := range rows {
		key := r.StockCode + "|" + r.TradeDate.Format("2006-01-02")
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		pairs = append(pairs, pair{r.StockCode, r.TradeDate.Format("2006-01-02")})
	}

	for _, p := range pairs {
		// Rebuild t_exodus_broker_net untuk (stock, date) ini.
		if _, err := database.DB.Exec(
			"DELETE FROM t_exodus_broker_net WHERE stock_code = ? AND trade_date = ?",
			p.stock, p.date,
		); err != nil {
			return err
		}
		if _, err := database.DB.Exec(`
		INSERT INTO t_exodus_broker_net (stock_code, trade_date, broker_code, broker_type, net_value)
		SELECT
			stock_code,
			trade_date,
			broker_code,
			MAX(broker_type),
			SUM(IF(side = 'BUY',  ABS(value), 0)) - SUM(IF(side = 'SELL', ABS(value), 0)) AS net_value
		FROM t_exodus_broker_summary
		WHERE stock_code = ? AND trade_date = ?
		GROUP BY stock_code, trade_date, broker_code
		`, p.stock, p.date); err != nil {
			return err
		}

		// Rebuild t_exodus_broker_agg untuk (stock, date) ini.
		if _, err := database.DB.Exec(
			"DELETE FROM t_exodus_broker_agg WHERE stock_code = ? AND trade_date = ?",
			p.stock, p.date,
		); err != nil {
			return err
		}
		if _, err := database.DB.Exec(`
		INSERT INTO t_exodus_broker_agg
			(stock_code, trade_date, top3_net_buy, broker_count, top1_asing, foreign_net_buy)
		WITH broker_ranked AS (
			SELECT
				stock_code,
				trade_date,
				broker_code,
				broker_type,
				net_value,
				ROW_NUMBER() OVER (PARTITION BY stock_code, trade_date ORDER BY net_value DESC) AS rn
			FROM t_exodus_broker_net
			WHERE stock_code = ? AND trade_date = ?
		),
		top_brokers AS (
			SELECT
				stock_code,
				trade_date,
				SUM(net_value) AS top3_net_buy,
				COUNT(*) AS broker_count,
				SUM(CASE WHEN rn = 1 AND broker_type = 'Asing' THEN 1 ELSE 0 END) AS top1_asing
			FROM broker_ranked
			WHERE rn <= 3 AND net_value > 0
			GROUP BY stock_code, trade_date
		),
		foreign_net AS (
			SELECT
				stock_code,
				trade_date,
				SUM(net_value) AS foreign_net_buy
			FROM t_exodus_broker_net
			WHERE broker_type = 'Asing' AND stock_code = ? AND trade_date = ?
			GROUP BY stock_code, trade_date
		)
		SELECT
			t.stock_code,
			t.trade_date,
			t.top3_net_buy,
			t.broker_count,
			t.top1_asing,
			COALESCE(f.foreign_net_buy, 0) AS foreign_net_buy
		FROM top_brokers t
		LEFT JOIN foreign_net f
			ON t.stock_code = f.stock_code COLLATE utf8mb4_unicode_ci
		   AND t.trade_date = f.trade_date
		`, p.stock, p.date, p.stock, p.date); err != nil {
			return err
		}
	}

	return nil
}

func GetExodusBrokerSummary(symbol, from, to string) ([]models.ExodusBrokerSummaryDB, error) {
	query := `
	SELECT 
		id,
		stock_code,
		trade_date,
		broker_code,
		side,
		broker_type,
		lot,
		volume,
		value,
		turnover,
		avg_price,
		frequency,
		created_at,
		updated_at
	FROM t_exodus_broker_summary
	WHERE 1=1
	`

	args := []any{}

	if symbol != "" {
		query += " AND stock_code = ?"
		args = append(args, symbol)
	}

	if from != "" {
		if _, err := time.Parse("2006-01-02", from); err == nil {
			query += " AND trade_date >= ?"
			args = append(args, from)
		}
	}

	if to != "" {
		if _, err := time.Parse("2006-01-02", to); err == nil {
			query += " AND trade_date <= ?"
			args = append(args, to)
		}
	}

	query += " ORDER BY trade_date DESC, value DESC"

	rows := []models.ExodusBrokerSummaryDB{}
	err := database.DB.Select(&rows, query, args...)
	if err != nil {
		return nil, err
	}

	return rows, nil
}

// GetExistingExodusStockDates mengembalikan pasangan (stock_code, trade_date)
// yang sudah tersedia di t_exodus_broker_summary pada rentang tanggal.
// Dipakai bulk fetch untuk me-skip hit ke Exodus jika data sudah ada.
func GetExistingExodusStockDates(from, to string) ([]models.ExodusStockDate, error) {
	query := `
	SELECT DISTINCT
		stock_code,
		trade_date
	FROM t_exodus_broker_summary
	WHERE trade_date >= ?
	  AND trade_date <= ?
	`

	rows := []models.ExodusStockDate{}
	err := database.DB.Select(&rows, query, from, to)
	if err != nil {
		return nil, err
	}

	return rows, nil
}

func GetExodusBrokerFlow(symbol, from, to string) ([]models.ExodusBrokerFlow, error) {
	query := `
	SELECT
		broker_code,
		MAX(broker_type) AS broker_type,
		SUM(IF(side = 'BUY',  ABS(value), 0)) AS buy_value,
		SUM(IF(side = 'SELL', ABS(value), 0)) AS sell_value,
		SUM(IF(side = 'BUY',  ABS(value), 0)) - SUM(IF(side = 'SELL', ABS(value), 0)) AS net_value,
		SUM(IF(side = 'BUY',  lot, 0))     AS buy_lot,
		SUM(IF(side = 'SELL', lot, 0)) * -1 AS sell_lot,
		SUM(lot) AS net_lot,
		COUNT(DISTINCT trade_date) AS active_days
	FROM t_exodus_broker_summary
	WHERE stock_code = ?
	  AND trade_date >= ?
	  AND trade_date <= ?
	GROUP BY broker_code
	ORDER BY net_value DESC
	`

	rows := []models.ExodusBrokerFlow{}
	err := database.DB.Select(&rows, query, symbol, from, to)
	if err != nil {
		return nil, err
	}

	return rows, nil
}

// GetExodusBrokerFlowGrouped mengembalikan breakdown per broker group (RETAIL, FOREIGN, INSTITUTIONAL, LOCAL_MID).
// Digunakan untuk analisis smart money vs retail, detect distribusi masif ke retail.
func GetExodusBrokerFlowGrouped(symbol, from, to string) (map[string]float64, error) {
	query := `
	SELECT
		COALESCE(b.broker_group, 'UNKNOWN') AS grp,
		SUM(IF(s.side = 'BUY',  ABS(s.value), 0)) - SUM(IF(s.side = 'SELL', ABS(s.value), 0)) AS net_value
	FROM t_exodus_broker_summary s
	LEFT JOIN m_list_broker b ON s.broker_code = b.broker_code COLLATE utf8mb4_unicode_ci
	WHERE s.stock_code = ?
	  AND s.trade_date >= ?
	  AND s.trade_date <= ?
	GROUP BY COALESCE(b.broker_group, 'UNKNOWN')
	`

	type GroupNet struct {
		Group string  `db:"grp"`
		Net   float64 `db:"net_value"`
	}
	rows := []GroupNet{}
	err := database.DB.Select(&rows, query, symbol, from, to)
	if err != nil {
		return nil, err
	}

	result := map[string]float64{}
	for _, r := range rows {
		result[r.Group] = r.Net
	}
	return result, nil
}

// DailyGroupNet merepresentasikan net flow harian sekelompok broker.
type DailyGroupNet struct {
	TradeDate string  `db:"trade_date"`
	NetValue  float64 `db:"net_value"`
}

// GetExodusDailySmartMoneyFlow mengembalikan net flow harian broker
// kelompok smart money (Asing + Institutional) dalam rentang tanggal.
func GetExodusDailySmartMoneyFlow(symbol, from, to string) ([]models.DailyGroupNet, error) {
	query := `
	SELECT
		DATE_FORMAT(s.trade_date, '%Y-%m-%d') AS trade_date,
		SUM(IF(s.side = 'BUY',  ABS(s.value), 0)) - SUM(IF(s.side = 'SELL', ABS(s.value), 0)) AS net_value
	FROM t_exodus_broker_summary s
	LEFT JOIN m_list_broker b ON s.broker_code = b.broker_code COLLATE utf8mb4_unicode_ci
	WHERE s.stock_code = ?
	  AND s.trade_date >= ?
	  AND s.trade_date <= ?
	  AND COALESCE(b.broker_group, 'UNKNOWN') IN ('FOREIGN', 'INSTITUTIONAL')
	GROUP BY DATE_FORMAT(s.trade_date, '%Y-%m-%d')
	ORDER BY trade_date
	`
	rows := []models.DailyGroupNet{}
	err := database.DB.Select(&rows, query, symbol, from, to)
	return rows, err
}
func GetExodusDailyGroupFlow(symbol, from, to string) ([]models.DailyGroupFlow, error) {
	query := `
	SELECT
		DATE_FORMAT(s.trade_date, '%Y-%m-%d') AS trade_date,
		COALESCE(b.broker_group, 'UNKNOWN') AS broker_group,
		SUM(IF(s.side = 'BUY',  ABS(s.value), 0)) - SUM(IF(s.side = 'SELL', ABS(s.value), 0)) AS net_value
	FROM t_exodus_broker_summary s
	LEFT JOIN m_list_broker b ON s.broker_code = b.broker_code COLLATE utf8mb4_unicode_ci
	WHERE s.stock_code = ?
	  AND s.trade_date >= ?
	  AND s.trade_date <= ?
	GROUP BY DATE_FORMAT(s.trade_date, '%Y-%m-%d'), COALESCE(b.broker_group, 'UNKNOWN')
	ORDER BY trade_date
	`
	rows := []models.DailyGroupFlow{}
	err := database.DB.Select(&rows, query, symbol, from, to)
	return rows, err
}

// GetExodusDailyBrokerNet mengembalikan net harian per broker untuk perhitungan
// Modified Z-Score (anomaly detection).
func GetExodusDailyBrokerNet(symbol, from, to string) ([]models.DailyBrokerNet, error) {
	query := `
	SELECT
		DATE_FORMAT(s.trade_date, '%Y-%m-%d') AS trade_date,
		s.broker_code,
		MAX(s.broker_type) AS broker_type,
		SUM(IF(s.side = 'BUY',  ABS(s.value), 0)) - SUM(IF(s.side = 'SELL', ABS(s.value), 0)) AS net_value
	FROM t_exodus_broker_summary s
	WHERE s.stock_code = ?
	  AND s.trade_date >= ?
	  AND s.trade_date <= ?
	GROUP BY DATE_FORMAT(s.trade_date, '%Y-%m-%d'), s.broker_code
	ORDER BY trade_date
	`
	rows := []models.DailyBrokerNet{}
	err := database.DB.Select(&rows, query, symbol, from, to)
	return rows, err
}

// GetMarketMakers mengambil daftar broker yang menjadi market maker sebuah saham.
func GetMarketMakers(stockCode string) (map[string]bool, error) {
	query := `SELECT broker_code FROM m_market_maker WHERE stock_code = ?`
	codes := []string{}
	err := database.DB.Select(&codes, query, stockCode)
	if err != nil {
		return nil, err
	}
	set := make(map[string]bool, len(codes))
	for _, c := range codes {
		set[c] = true
	}
	return set, nil
}

// PriceWindow merepresentasikan data harga & volume harian sebuah saham.
type PriceWindow struct {
	TradeDate   string  `db:"trade_date"`
	Open        float64 `db:"open_price"`
	High        float64 `db:"high_price"`
	Low         float64 `db:"low_price"`
	Close       float64 `db:"close_price"`
	Volume      float64 `db:"volume"`
	CloseChange float64 `db:"close_change"`
}

// GetPriceWindow mengambil rangkaian harga & volume sebuah saham dalam rentang.
func GetPriceWindow(symbol, from, to string) ([]PriceWindow, error) {
	query := `
	SELECT
		DATE_FORMAT(trade_date, '%Y-%m-%d') AS trade_date,
		open_price,
		high_price,
		low_price,
		close_price,
		volume
	FROM t_trading_summary
	WHERE stock_code = ?
	  AND trade_date >= ?
	  AND trade_date <= ?
	ORDER BY trade_date
	`
	rows := []PriceWindow{}
	err := database.DB.Select(&rows, query, symbol, from, to)
	return rows, err
}

type StockNameRow struct {
	StockName string `db:"stock_name"`
}

func GetStockName(stockCode string) (string, error) {
	var row StockNameRow
	err := database.DB.Get(&row, `SELECT stock_name FROM m_list_stocks WHERE stock_code = ?`, stockCode)
	if err != nil {
		return "", err
	}
	return row.StockName, nil
}

// GetTickerPriceBars mengambil OHLCV + change_pct harian untuk chart.
func GetTickerPriceBars(symbol, from, to string) ([]models.TickerPriceBar, error) {
	query := `
	SELECT
		DATE_FORMAT(trade_date, '%Y-%m-%d') AS trade_date,
		open_price,
		high_price,
		low_price,
		close_price,
		volume,
		(close_price / NULLIF(previous_price, 0) - 1) * 100 AS change_pct
	FROM t_trading_summary
	WHERE stock_code = ?
	  AND trade_date >= ?
	  AND trade_date <= ?
	ORDER BY trade_date
	`
	rows := []models.TickerPriceBar{}
	err := database.DB.Select(&rows, query, symbol, from, to)
	return rows, err
}

// GetTickerBrokerVolume mengambil agregat trading per broker dalam rentang.
func GetTickerBrokerVolume(symbol, from, to string) ([]models.TickerBrokerVolume, error) {
	query := `
	SELECT
		s.broker_code,
		COALESCE(b.broker_name, s.broker_code) AS broker_name,
		MAX(s.broker_type) AS broker_type,
		SUM(IF(s.side = 'BUY',  s.lot, 0))  AS buy_lot,
		SUM(IF(s.side = 'SELL', s.lot, 0))  AS sell_lot,
		SUM(IF(s.side = 'BUY',  s.volume, 0)) AS buy_volume,
		SUM(IF(s.side = 'SELL', s.volume, 0)) AS sell_volume,
		SUM(IF(s.side = 'BUY',  ABS(s.value), 0)) AS buy_value,
		SUM(IF(s.side = 'SELL', ABS(s.value), 0)) AS sell_value,
		SUM(IF(s.side = 'BUY',  ABS(s.value), 0)) - SUM(IF(s.side = 'SELL', ABS(s.value), 0)) AS net_value,
		COUNT(DISTINCT s.trade_date) AS active_days
	FROM t_exodus_broker_summary s
	LEFT JOIN m_list_broker b
		ON s.broker_code = b.broker_code COLLATE utf8mb4_unicode_ci
	WHERE s.stock_code = ?
	  AND s.trade_date >= ?
	  AND s.trade_date <= ?
	GROUP BY s.broker_code, COALESCE(b.broker_name, s.broker_code)
	ORDER BY net_value DESC
	`
	rows := []models.TickerBrokerVolume{}
	err := database.DB.Select(&rows, query, symbol, from, to)
	return rows, err
}

// GetTickerBrokerSummary mengambil buy/sell per broker per hari.
func GetTickerBrokerSummary(symbol, from, to string) ([]models.TickerBrokerSummary, error) {
	query := `
	SELECT
		DATE_FORMAT(s.trade_date, '%Y-%m-%d') AS trade_date,
		s.broker_code,
		COALESCE(b.broker_name, s.broker_code) AS broker_name,
		MAX(s.broker_type) AS broker_type,
		COALESCE(MAX(b.broker_group), 'UNKNOWN') AS broker_group,
		SUM(IF(s.side = 'BUY',  s.lot, 0))  AS buy_lot,
		SUM(IF(s.side = 'SELL', s.lot, 0))  AS sell_lot,
		SUM(IF(s.side = 'BUY',  s.volume, 0)) AS buy_volume,
		SUM(IF(s.side = 'SELL', s.volume, 0)) AS sell_volume,
		SUM(IF(s.side = 'BUY',  ABS(s.value), 0)) AS buy_value,
		SUM(IF(s.side = 'SELL', ABS(s.value), 0)) AS sell_value,
		SUM(IF(s.side = 'BUY',  ABS(s.value), 0)) - SUM(IF(s.side = 'SELL', ABS(s.value), 0)) AS net_value,
		SUM(s.frequency) AS frequency
	FROM t_exodus_broker_summary s
	LEFT JOIN m_list_broker b
		ON s.broker_code = b.broker_code COLLATE utf8mb4_unicode_ci
	WHERE s.stock_code = ?
	  AND s.trade_date >= ?
	  AND s.trade_date <= ?
	GROUP BY DATE_FORMAT(s.trade_date, '%Y-%m-%d'), s.broker_code, COALESCE(b.broker_name, s.broker_code)
	ORDER BY trade_date, net_value DESC
	`
	rows := []models.TickerBrokerSummary{}
	err := database.DB.Select(&rows, query, symbol, from, to)
	return rows, err
}

package repositories

import (
	"fmt"
	"indonesia-stocks-api/internal/database"
	"indonesia-stocks-api/internal/helpers"
	"indonesia-stocks-api/internal/models"
)

func UpsertStocks(stocks []models.StocksList) error {
	query := `
	INSERT INTO m_list_stocks (
		stock_code,
		stock_name,
		listing_date,
		total_shares,
		listing_board,
		is_active
	)
	VALUES (
		:stock_code,
		:stock_name,
		:listing_date,
		:total_shares,
		:listing_board,
		:is_active
	)
	ON DUPLICATE KEY UPDATE
		stock_name    = VALUES(stock_name),
		total_shares  = VALUES(total_shares),
		listing_board = VALUES(listing_board),
		is_active     = VALUES(is_active)
	`

	_, err := database.DB.NamedExec(query, stocks)
	return err
}

func InsertTradingSummary(summaries []models.TradingSummaryDB) error {
	query := `
	INSERT INTO t_trading_summary (
		idx_id_stock_summary,
		trade_date,
		stock_code,
		stock_name,		
		previous_price,
		open_price,
		first_trade,
		high_price,
		low_price,
		close_price,
		change_price,
		close_strength,
		volume,
		value,
		frequency,
		index_individual,
		offer_price,
		offer_volume,
		bid_price,
		bid_volume,
		listed_shares,
		tradeable_shares,
		weight_for_index,
		foreign_sell,
		foreign_buy,
		non_regular_volume,
		non_regular_value,
		non_regular_frequency,
		created_at,
		updated_at
	)
	VALUES (
		:idx_id_stock_summary,
		:trade_date,
		:stock_code,
		:stock_name,		
		:previous_price,
		:open_price,
		:first_trade,
		:high_price,
		:low_price,
		:close_price,
		:change_price,
		:close_strength,
		:volume,
		:value,
		:frequency,
		:index_individual,
		:offer_price,
		:offer_volume,
		:bid_price,
		:bid_volume,
		:listed_shares,
		:tradeable_shares,
		:weight_for_index,
		:foreign_sell,
		:foreign_buy,
		:non_regular_volume,
		:non_regular_value,
		:non_regular_frequency,
		:created_at,
		:updated_at
	)
	ON DUPLICATE KEY UPDATE
		stock_name = VALUES(stock_name),		
		previous_price = VALUES(previous_price),
		open_price = VALUES(open_price),
		first_trade = VALUES(first_trade),
		high_price = VALUES(high_price),
		low_price = VALUES(low_price),
		close_price = VALUES(close_price),
		close_strength = VALUES(close_strength),
		change_price = VALUES(change_price),
		volume = VALUES(volume),
		value = VALUES(value),
		frequency = VALUES(frequency),
		index_individual = VALUES(index_individual),
		offer_price = VALUES(offer_price),
		offer_volume = VALUES(offer_volume),
		bid_price = VALUES(bid_price),
		bid_volume = VALUES(bid_volume),
		listed_shares = VALUES(listed_shares),
		tradeable_shares = VALUES(tradeable_shares),
		weight_for_index = VALUES(weight_for_index),
		foreign_sell = VALUES(foreign_sell),
		foreign_buy = VALUES(foreign_buy),
		non_regular_volume = VALUES(non_regular_volume),
		non_regular_value = VALUES(non_regular_value),
		non_regular_frequency = VALUES(non_regular_frequency),
		updated_at = NOW()
	`

	_, err := database.DB.NamedExec(query, summaries)
	return err
}

func GetTopAccumulation(days int) ([]models.TopAccumulation, error) {
	query := `
		WITH DailyMetrics AS (
			SELECT 
				stock_code, stock_name, trade_date, close_price, volume, close_strength, value, high_price,
				(foreign_buy - foreign_sell) as daily_net_foreign,
				AVG(close_price) OVER (PARTITION BY stock_code ORDER BY trade_date ROWS BETWEEN 19 PRECEDING AND CURRENT ROW) as ma20,
				AVG(close_price) OVER (PARTITION BY stock_code ORDER BY trade_date ROWS BETWEEN 49 PRECEDING AND CURRENT ROW) as ma50,
				MAX(high_price) OVER (PARTITION BY stock_code ORDER BY trade_date ROWS BETWEEN 20 PRECEDING AND 1 PRECEDING) as resistance_20,
				AVG(volume) OVER (PARTITION BY stock_code ORDER BY trade_date ROWS BETWEEN 19 PRECEDING AND CURRENT ROW) as avg_vol20,
				((close_price - LAG(close_price) OVER (PARTITION BY stock_code ORDER BY trade_date)) / 
				 NULLIF(LAG(close_price) OVER (PARTITION BY stock_code ORDER BY trade_date), 0)) * 100 as change_pct
			FROM t_trading_summary
		),
		Screener AS (
			SELECT 
				stock_code, stock_name,
				AVG(close_strength) AS avg_close_strength,
				SUM(daily_net_foreign) AS net_foreign,
				ROUND(AVG(value), 2) AS avg_value,
				SUM(volume) AS total_volume,
				MAX(trade_date) AS last_trade_date,
				SUBSTRING_INDEX(GROUP_CONCAT(close_price ORDER BY trade_date DESC), ',', 1) + 0 as last_price,
				SUBSTRING_INDEX(GROUP_CONCAT(ma20 ORDER BY trade_date DESC), ',', 1) + 0 as last_ma20,
				SUBSTRING_INDEX(GROUP_CONCAT(ma50 ORDER BY trade_date DESC), ',', 1) + 0 as last_ma50,
				SUBSTRING_INDEX(GROUP_CONCAT(resistance_20 ORDER BY trade_date DESC), ',', 1) + 0 as last_res_20,
				SUBSTRING_INDEX(GROUP_CONCAT(volume ORDER BY trade_date DESC), ',', 1) + 0 as last_volume,
				SUBSTRING_INDEX(GROUP_CONCAT(avg_vol20 ORDER BY trade_date DESC), ',', 1) + 0 as last_avg_vol20,
				SUBSTRING_INDEX(GROUP_CONCAT(change_pct ORDER BY trade_date DESC), ',', 1) + 0 as last_change
			FROM DailyMetrics
			WHERE trade_date >= CURDATE() - INTERVAL ? DAY
			GROUP BY stock_code, stock_name
		)
		SELECT 
			*,
			(last_price / last_res_20) as breakout_score
		FROM Screener
		WHERE 
			net_foreign > 0                             -- Borong Asing
			AND last_price > last_ma20                  -- Tren Naik
			AND avg_value >= 1000000000                 -- Likuid (Min 1M)
			AND last_volume > (last_avg_vol20 * 0.5)    -- Volume Aktif
			AND avg_close_strength >= 60                -- Close Mantap
		ORDER BY 
			breakout_score DESC,                        -- Urutan Breakout Teratas
			net_foreign DESC                            -- Lalu nominal foreign
		LIMIT 50
	`

	rows := []models.TopAccumulation{}
	err := database.DB.Select(&rows, query, days)
	if err != nil {
		return nil, err
	}

	for i := range rows {
		// 1. Format Display Angka
		rows[i].FormattedNetForeign = helpers.FormatBigNumber(rows[i].NetForeign)
		rows[i].FormattedAvgValue = helpers.FormatBigNumber(rows[i].AvgValue)

		// 2. Kalkulasi Jarak & Sinyal
		isSuperBullish := rows[i].LastPrice > rows[i].Ma50
		isBreakout := rows[i].LastPrice > rows[i].LastRes20
		isNearRes := rows[i].LastPrice >= (rows[i].LastRes20 * 0.97)

		diffRes := ((rows[i].LastPrice - rows[i].LastRes20) / rows[i].LastRes20) * 100
		distStr := fmt.Sprintf("(%.1f%% To Res)", diffRes)
		if diffRes >= 0 {
			distStr = fmt.Sprintf("(+%.1f%% Above Res)", diffRes)
		}

		// 3. Tentukan Status & Action
		trendLabel := "BULLISH"
		if isSuperBullish {
			trendLabel = "SUPER BULLISH"
		}

		actionLabel := "HOLD / WATCH"
		if isBreakout {
			actionLabel = "🚀 BREAKOUT! (BUY)"
		} else if isNearRes {
			actionLabel = "⚔️ TESTING RES (SIAP HAKA)"
		}

		changeLabel := fmt.Sprintf("[+%.2f%% Today]", rows[i].LastChange)
		if rows[i].LastChange < 0 {
			changeLabel = fmt.Sprintf("[%.2f%% Today]", rows[i].LastChange)
		}

		// 4. Combine Display Status
		rows[i].DisplayStatus = fmt.Sprintf("%s | %s | %s | %s", trendLabel, actionLabel, distStr, changeLabel)

		// 5. Override Golden Signal
		if isSuperBullish && isBreakout {
			rows[i].DisplayStatus = fmt.Sprintf("🔥 GOLDEN SIGNAL | STRONG BUY | %s | %s", distStr, changeLabel)
		}

		// Warning jika kenaikan harian terlalu ekstrim
		if rows[i].LastChange > 18 {
			rows[i].DisplayStatus += " ⚠️ HIGH VOLATILITY"
		}
	}

	return rows, nil
}

func GetTopAccumulationEOD(days int) ([]models.TopAccumulationEod, error) {
	// Query ini menggunakan teknik "Late Filtering"
	// Supaya Resistance & MA akurat, kita hitung dulu dari histori panjang,
	// baru kita ambil (JOIN) baris terakhirnya saja.
	query := `
WITH BaseData AS (
    -- Ambil histori 100 hari supaya MA50 dan RSI tidak NULL
    SELECT * FROM t_trading_summary
    WHERE trade_date >= (
        SELECT MIN(trade_date) FROM (
            SELECT DISTINCT trade_date FROM t_trading_summary 
            ORDER BY trade_date DESC LIMIT 100
        ) AS t
    )
),
DailyMetrics AS (
    SELECT 
        stock_code, stock_name, trade_date, close_price, volume, close_strength, value, high_price, low_price,
        ((foreign_buy - foreign_sell) * close_price) as daily_net_foreign_val,
        ((foreign_buy + foreign_sell) * close_price) as foreign_turnover_val,
        close_price - LAG(close_price) OVER (PARTITION BY stock_code ORDER BY trade_date) as diff,
        AVG(close_price) OVER (PARTITION BY stock_code ORDER BY trade_date ROWS BETWEEN 19 PRECEDING AND CURRENT ROW) as ma20,
        AVG(close_price) OVER (PARTITION BY stock_code ORDER BY trade_date ROWS BETWEEN 49 PRECEDING AND CURRENT ROW) as ma50,
        -- Mencari harga tertinggi (Resistance) murni 20 hari sebelum hari H
        MAX(high_price) OVER (PARTITION BY stock_code ORDER BY trade_date ROWS BETWEEN 20 PRECEDING AND 1 PRECEDING) as resistance_20,
        MIN(low_price) OVER (PARTITION BY stock_code ORDER BY trade_date ROWS BETWEEN 20 PRECEDING AND 1 PRECEDING) as support_20,
        AVG(volume) OVER (PARTITION BY stock_code ORDER BY trade_date ROWS BETWEEN 19 PRECEDING AND CURRENT ROW) as avg_vol20
    FROM BaseData
),
RSICalculation AS (
    SELECT *,
        AVG(IF(diff > 0, diff, 0)) OVER (PARTITION BY stock_code ORDER BY trade_date ROWS BETWEEN 13 PRECEDING AND CURRENT ROW) as avg_gain,
        AVG(IF(diff < 0, ABS(diff), 0)) OVER (PARTITION BY stock_code ORDER BY trade_date ROWS BETWEEN 13 PRECEDING AND CURRENT ROW) as avg_loss
    FROM DailyMetrics
),
RSIDone AS (
    SELECT *,
        100 - (100 / (1 + (avg_gain / NULLIF(avg_loss, 0)))) as rsi_14
    FROM RSICalculation
),
FilteredStats AS (
    -- Bagian ini menghitung AKUMULASI selama 'days' (misal 20 hari)
    SELECT 
        stock_code, 
        stock_name,
        SUM(daily_net_foreign_val) AS net_foreign,
        AVG(close_strength) AS avg_close_strength,
        AVG((value - foreign_turnover_val) / NULLIF(value, 0)) * 100 AS local_participation,
        AVG(value) AS avg_value,
        MAX(trade_date) as last_date
    FROM RSIDone
    WHERE trade_date >= (
        SELECT MIN(trade_date) FROM (
            SELECT DISTINCT trade_date FROM t_trading_summary 
            ORDER BY trade_date DESC LIMIT ?
        ) AS t
    )
    GROUP BY stock_code, stock_name
),
FinalData AS (
    -- Ambil baris murni dari tanggal terakhir untuk menghindari data ngaco
    SELECT 
        f.stock_code, f.stock_name, f.net_foreign, f.avg_close_strength, f.local_participation, f.avg_value,
        r.trade_date as last_trade_date,
        r.close_price as last_price,
        r.ma20 as last_ma20,
        r.ma50 as last_ma50,
        r.rsi_14 as last_rsi,
        r.resistance_20 as last_res_20,
        r.support_20 as last_sup_20,
        r.volume as last_volume,
        r.avg_vol20 as last_avg_vol20,
        r.diff as last_change
    FROM FilteredStats f
    JOIN RSIDone r ON f.stock_code = r.stock_code AND f.last_date = r.trade_date
)
SELECT *, (last_price / NULLIF(last_res_20, 0)) as breakout_score 
FROM FinalData
WHERE net_foreign > 0 
  AND last_price > last_ma20 
  AND avg_value >= 1000000000 
  AND last_rsi BETWEEN 30 AND 70
ORDER BY net_foreign DESC
LIMIT 50`

	rows := []models.TopAccumulationEod{}
	err := database.DB.Select(&rows, query, days)
	if err != nil {
		return nil, err
	}

	for i := range rows {
		// 0. Formatting Numbers
		rows[i].FormattedNetForeign = helpers.FormatBigNumber(rows[i].NetForeign)
		rows[i].FormattedAvgValue = helpers.FormatBigNumber(rows[i].AvgValue)

		// 1. Sentimen Lokal
		retailLabel := "💎 INST"
		if rows[i].LocalParticipation > 80 {
			retailLabel = "🤡 FOMO"
		} else if rows[i].LocalParticipation > 60 {
			retailLabel = "👥 MIX"
		}

		// 2. Volume & Trend
		volRatio := 0.0
		if rows[i].LastAvgVol20 > 0 {
			volRatio = rows[i].LastVolume / rows[i].LastAvgVol20
		}

		volEmoji := "⚪ Normal"
		if volRatio >= 2.0 {
			volEmoji = "💎 GIANT"
		} else if volRatio >= 1.2 {
			volEmoji = "🔊 HIGH"
		}

		trendEmoji := "📈 BULL"
		if rows[i].LastMa50 > 0 && rows[i].LastPrice > rows[i].LastMa50 {
			trendEmoji = "🔥 SUPER"
		}

		// 3. Smart Money Signal
		isSmartMoney := rows[i].NetForeign > (rows[i].AvgValue*0.15) && volRatio >= 1.5 && rows[i].AvgCloseStrength > 0.7
		smLabel := ""
		if isSmartMoney {
			smLabel = "🐋 SMART MONEY | "
		}

		// 4. Action & Strategy
		// breakout_score 1.0 = tepat di resistance
		diffRes := ((rows[i].LastPrice - rows[i].LastRes20) / NULLIF_FLOAT(rows[i].LastRes20)) * 100
		distToSup := ((rows[i].LastPrice - rows[i].LastSup20) / NULLIF_FLOAT(rows[i].LastSup20)) * 100

		action := "👀 WATCH"
		entryPrice := rows[i].LastPrice

		if rows[i].LastChange < 0 && distToSup <= 3 {
			action = "🛡️ BOW"
		} else if diffRes >= 0 && rows[i].LastChange < 10 {
			action = "🎯 HAKA!"
		} else if diffRes >= 0 && rows[i].LastChange >= 10 {
			action = "⌛ RETRACE"
			entryPrice = rows[i].LastRes20
		} else if diffRes < 0 && diffRes >= -2 {
			action = "🚀 BREAKOUT"
			entryPrice = rows[i].LastRes20 + 2
		}

		// 5. Risk Calculation
		stopLoss := rows[i].LastMa20
		if rows[i].LastSup20 > 0 && rows[i].LastSup20 < stopLoss {
			stopLoss = rows[i].LastSup20 * 0.99
		}

		riskPct := 0.0
		if entryPrice > 0 {
			riskPct = ((entryPrice - stopLoss) / entryPrice) * 100
		}

		riskEmoji := "🟢"
		if riskPct > 7 {
			riskEmoji = "🔴"
		}

		// 6. FINAL OUTPUT
		rows[i].DisplayStatus = fmt.Sprintf("%s%s | %s (Lokal: %.0f%%) | %s | %s | Entry: %.0f | SL: %.0f (Risk: %.1f%%) %s",
			smLabel, trendEmoji, retailLabel, rows[i].LocalParticipation, volEmoji, action, entryPrice, stopLoss, riskPct, riskEmoji)
	}

	return rows, nil
}

// Helper sederhana untuk menghindari divide by zero di Go
func NULLIF_FLOAT(val float64) float64 {
	if val == 0 {
		return 1
	}
	return val
}

func RunBacktestEOD(targetDate string) ([]models.BacktestResult, error) {
	query := `
		WITH DailyMetrics AS (
			SELECT 
				stock_code, stock_name, trade_date, close_price, high_price, low_price,
				MAX(high_price) OVER (PARTITION BY stock_code ORDER BY trade_date ROWS BETWEEN 20 PRECEDING AND 1 PRECEDING) as res_20_then,
				AVG(close_price) OVER (PARTITION BY stock_code ORDER BY trade_date ROWS BETWEEN 19 PRECEDING AND CURRENT ROW) as ma20_then,
				AVG(volume) OVER (PARTITION BY stock_code ORDER BY trade_date ROWS BETWEEN 19 PRECEDING AND CURRENT ROW) as avg_vol_then,
                volume as vol_then,
                (foreign_buy - foreign_sell) as net_foreign_then
			FROM t_trading_summary
		),
		ScreenerAtDate AS (
			SELECT * FROM DailyMetrics WHERE trade_date = ?
		),
		CurrentPrice AS (
			SELECT stock_code, close_price as price_now 
			FROM t_trading_summary 
			WHERE trade_date = (SELECT MAX(trade_date) FROM t_trading_summary)
		)
		SELECT 
			s.stock_code, s.stock_name, s.close_price as price_then, 
            c.price_now, s.res_20_then, s.ma20_then
		FROM ScreenerAtDate s
		JOIN CurrentPrice c ON s.stock_code = c.stock_code
		WHERE s.net_foreign_then > 0 
		  AND s.close_price > s.ma20_then 
          AND s.vol_then > (s.avg_vol_then * 0.5)
	`

	rows := []models.BacktestResult{}
	err := database.DB.Select(&rows, query, targetDate)
	if err != nil {
		return nil, err
	}

	var totalWin, totalLose int

	for i := range rows {
		// Hitung Performance
		rows[i].ProfitLossPct = ((rows[i].PriceNow - rows[i].PriceThen) / rows[i].PriceThen) * 100

		// Status awal saat itu (Pura-pura masa lalu)
		if rows[i].PriceThen >= rows[i].ResThen {
			rows[i].SignalAtThen = "🎯 SIKAT (Breakout)"
		} else {
			rows[i].SignalAtThen = "👀 WATCH"
		}

		// Kesimpulan Akhir
		if rows[i].ProfitLossPct > 0.5 { // Anggap win kalau naik di atas 0.5% (cover fee)
			rows[i].ResultStatus = "✅ WIN"
			totalWin++
		} else if rows[i].ProfitLossPct < -0.5 {
			rows[i].ResultStatus = "❌ LOSE"
			totalLose++
		} else {
			rows[i].ResultStatus = "🟡 FLAT" // Status baru biar jelas
		}
	}

	return rows, nil
}
func GetTopSwinger(tradeDate string) ([]models.TopSwinger, error) {
	query := `
WITH BaseData AS (
    SELECT * FROM t_trading_summary
    WHERE trade_date <= ?
    ORDER BY trade_date DESC
    LIMIT 20000 
),
History AS (
    SELECT 
        stock_code, stock_name, trade_date, close_price, high_price, low_price, 
        volume, value, close_strength,
        (foreign_buy - foreign_sell) * close_price AS net_foreign_val,
        LAG(close_price) OVER (PARTITION BY stock_code ORDER BY trade_date) as prev_close,
        LAG(volume) OVER (PARTITION BY stock_code ORDER BY trade_date) as prev_vol,
        AVG(close_strength) OVER (PARTITION BY stock_code ORDER BY trade_date ROWS BETWEEN 4 PRECEDING AND CURRENT ROW) as avg_strength_5d
    FROM BaseData
),
Calculated AS (
    SELECT *,
        COALESCE(((volume - prev_vol) / NULLIF(prev_vol, 0)) * 100, 0) as vol_change_pct,
        COALESCE(volume / NULLIF(prev_vol, 0), 1) as vol_multiplier
    FROM History
),
FinalData AS (
    SELECT 
        stock_code, stock_name, trade_date, close_price, high_price, low_price, 
        close_strength, volume, value, 
        net_foreign_val as net_foreign,
        avg_strength_5d, vol_change_pct, vol_multiplier,
        ROUND(
            (avg_strength_5d * 0.3) + 
            (IF(close_price >= prev_close, 10, 0)) + 
            (CASE 
                WHEN vol_multiplier >= 3 THEN 60 
                WHEN vol_multiplier >= 2 THEN 40
                WHEN vol_multiplier >= 1.5 THEN 20
                ELSE 0 
            END), 
            2
        ) AS swing_score,
        close_price as entry_price,
        ROUND(low_price * 0.96, 0) AS stop_loss,
        ROUND(close_price * 1.10, 0) AS take_profit,
        COALESCE(prev_close, close_price) as prev_close_val
    FROM Calculated
)
SELECT * FROM FinalData
WHERE trade_date = ?
  AND close_price < 500
  AND value >= 2000000000
  AND avg_strength_5d >= 40
  AND net_foreign >= 0 
  AND (
      close_price >= prev_close_val 
      OR 
      (vol_multiplier >= 2 AND close_strength > 50)
  )
ORDER BY swing_score DESC, value DESC
LIMIT 50;`

	rows := []models.TopSwinger{}
	err := database.DB.Select(&rows, query, tradeDate, tradeDate)
	if err != nil {
		return nil, err
	}

	for i := range rows {
		status := "🧘 SIDEWAYS"
		if rows[i].VolMultiplier >= 3 {
			status = "🚀 BOOM VOLUME"
		} else if rows[i].ClosePrice > rows[i].PrevCloseVal {
			status = "📈 UPTREND"
		}

		rows[i].DisplayStatus = fmt.Sprintf("%s | Score: %.2f | Multi: %.1fx",
			status, rows[i].SwingScore, rows[i].VolMultiplier)
	}

	return rows, nil
}

func GetSilentAccumulation(days int) ([]models.SilentAccumulation, error) {
	query := `
		WITH DailyMetrics AS (
			SELECT 
				stock_code, stock_name, trade_date, close_price, volume, value,
				((foreign_buy - foreign_sell) * close_price) as daily_net_foreign_val,
				((foreign_buy + foreign_sell) * close_price) as foreign_turnover_val,
				-- Atap (Resistance 20 hari) untuk filter harga belum lari
				MAX(high_price) OVER (PARTITION BY stock_code ORDER BY trade_date ROWS BETWEEN 20 PRECEDING AND 1 PRECEDING) as resistance_20,
				-- Volume Pendek (20 hari) vs Volume Panjang (100 hari) buat cek durasi tidur
				AVG(volume) OVER (PARTITION BY stock_code ORDER BY trade_date ROWS BETWEEN 19 PRECEDING AND CURRENT ROW) as avg_vol20,
				AVG(volume) OVER (PARTITION BY stock_code ORDER BY trade_date ROWS BETWEEN 99 PRECEDING AND CURRENT ROW) as avg_vol100,
                ((close_price - LAG(close_price) OVER (PARTITION BY stock_code ORDER BY trade_date)) / 
				 NULLIF(LAG(close_price) OVER (PARTITION BY stock_code ORDER BY trade_date), 0)) * 100 as change_pct
			FROM t_trading_summary
		),
		SilentScreener AS (
			SELECT 
				stock_code, stock_name,
				SUM(daily_net_foreign_val) AS net_foreign,
				AVG((value - foreign_turnover_val) / NULLIF(value, 0)) * 100 AS local_participation,
				ROUND(AVG(value), 2) AS avg_value,
				SUBSTRING_INDEX(GROUP_CONCAT(close_price ORDER BY trade_date DESC), ',', 1) + 0 as last_price,
				SUBSTRING_INDEX(GROUP_CONCAT(resistance_20 ORDER BY trade_date DESC), ',', 1) + 0 as last_res_20,
				SUBSTRING_INDEX(GROUP_CONCAT(volume ORDER BY trade_date DESC), ',', 1) + 0 as last_volume,
				SUBSTRING_INDEX(GROUP_CONCAT(avg_vol20 ORDER BY trade_date DESC), ',', 1) + 0 as last_avg_vol20,
				SUBSTRING_INDEX(GROUP_CONCAT(avg_vol100 ORDER BY trade_date DESC), ',', 1) + 0 as last_avg_vol100,
                SUBSTRING_INDEX(GROUP_CONCAT(change_pct ORDER BY trade_date DESC), ',', 1) + 0 as last_change
			FROM DailyMetrics
			WHERE trade_date >= CURDATE() - INTERVAL ? DAY
			GROUP BY stock_code, stock_name
		),
		FinalFilter AS (
			SELECT *, 
			(last_price / NULLIF(last_res_20, 0)) as breakout_score
			FROM SilentScreener
		)
		SELECT * FROM FinalFilter
		WHERE net_foreign > 0 
		  AND avg_value >= 500000000 
		  AND last_volume > (last_avg_vol20 * 2)  -- Ledakan Volume harian
		  -- AND last_avg_vol20 < last_avg_vol100   -- VALIDASI TIDUR: Sebulan terakhir lebih sepi dr rata-rata 5 bulan
		  -- AND breakout_score <= 1.03             -- Belum lari jauh dr atap (Maks +3%)
		  AND local_participation < 50           -- INSTITUSI DOMINAN
		ORDER BY 
			local_participation ASC,              -- 1. Cari yang retailnya paling dikit (Utama)
			(last_volume / last_avg_vol20) DESC   -- 2. Cari yang lonjakannya paling anomali
		LIMIT 50`

	rows := []models.SilentAccumulation{}

	err := database.DB.Select(&rows, query, days)
	if err != nil {
		return nil, err
	}

	for i := range rows {
		rows[i].FormattedNetForeign = helpers.FormatBigNumber(rows[i].NetForeign)
		rows[i].FormattedAvgValue = helpers.FormatBigNumber(rows[i].AvgValue)

		volRatio := rows[i].LastVolume / rows[i].LastAvgVol20

		// Status Labeling
		action := "🏹 COLLECTIONS"
		if rows[i].LocalParticipation < 25 {
			action = "🐋 WHALE ONLY" // Retail hampir nggak ada, murni mainan institusi
		}

		rows[i].DisplayStatus = fmt.Sprintf("🤫 SILENT | Lokal: %.0f%% | Vol Spike: %.1fx | %s | Price: %v",
			rows[i].LocalParticipation, volRatio, action, rows[i].LastPrice)
	}

	return rows, nil
}

func StatisticSingleStock(stockCode string) ([]models.StatisticSingleStockMapped, error) {
	query := `WITH TradingData AS (
    SELECT 
        tts.stock_code,
        tts.trade_date,
        tts.close_strength,
        tts.close_price,
        tts.change_price,
        tts.volume AS current_volume,
        LAG(tts.close_price) OVER (PARTITION BY tts.stock_code ORDER BY tts.trade_date) AS prev_close,
        LAG(tts.volume) OVER (PARTITION BY tts.stock_code ORDER BY tts.trade_date) AS prev_volume,
        AVG(tts.volume) OVER (PARTITION BY tts.stock_code ORDER BY tts.trade_date ROWS BETWEEN 19 PRECEDING AND CURRENT ROW) AS avg_vol_20d
    FROM t_trading_summary AS tts
)
SELECT 
    td.stock_code AS code,
    td.trade_date AS date,
    td.close_strength AS strength,
    td.close_price AS price,
    td.current_volume AS vol,
    td.change_price,
    case
        when td.prev_volume > 0
        then round(((td.current_volume - td.prev_volume) / td.prev_volume) * 100, 2)
        else 0
    end as vol_change_percent,
    CASE 
        -- 1. VOLUME TRAP (Kasus kamu di 462!)
        -- Harga naik, Volume meledak tinggi, tapi Strength ampas (< 35)
        -- Ini tanda distribusi terselubung.
        WHEN td.close_price > td.prev_close 
             AND td.current_volume > td.prev_volume * 1.5
             AND td.close_strength < 35
             THEN '⚠️ VOLUME TRAP (Distribusi)'

        -- 2. PRICE REJECTION / BUYING CLIMAX
        -- Sempat naik tinggi tapi dibanting closingnya
        WHEN td.close_strength < 30 AND td.change_price > 0
             THEN 'Markup -> Guyur (Ekor Atas Panjang)'

        -- 3. STRONG UPTREND (Konfirmasi Real Accumulation)
        -- Harga naik, Vol naik, dan Strength harus kuat (> 60)
        WHEN td.close_price > td.prev_close 
             AND td.current_volume > td.avg_vol_20d 
             AND td.close_strength >= 60
             THEN 'Strong Uptrend (Valid Accum)'
             
        -- 4. STRONG DOWNTREND (Panic Selling)
        WHEN td.close_price < td.prev_close 
             AND td.current_volume > td.avg_vol_20d 
             THEN 'Strong Downtrend (High Pressure)'
             
        -- 5. KOREKSI SEHAT (Vol Kering)
        WHEN td.close_price < td.prev_close 
             AND td.current_volume < td.prev_volume 
             AND td.current_volume < td.avg_vol_20d
             THEN 'Healthy Correction (Wait & See)'

        -- 6. EARLY UPTREND
        WHEN td.close_price > td.prev_close 
             AND td.current_volume > td.prev_volume 
             AND td.close_strength >= 50
             THEN 'Early Uptrend (Accumulation)'

        -- 7. DISTRIBUTION (Jualan pelan-pelan)
        WHEN td.close_price < td.prev_close 
             AND td.current_volume > td.prev_volume 
             THEN 'Early Downtrend (Distribution)'

        ELSE 'Sideways/Consolidation'
    END AS trend_status
FROM TradingData AS td
WHERE td.stock_code = ?
AND td.trade_date >= '2025-12-01'
ORDER BY td.trade_date DESC`

	var flatRows []models.StatisticSingleStock

	err := database.DB.Select(&flatRows, query, stockCode)
	if err != nil {
		return nil, err
	}

	if len(flatRows) == 0 {
		return []models.StatisticSingleStockMapped{}, nil
	}

	for i := range flatRows {
		flatRows[i].VolChangePercent = flatRows[i].VolChangePercent + "%"
		flatRows[i].VolumeFormatted = helpers.FormatBigNumber(flatRows[i].Volume)
		flatRows[i].TradeDateFormatted = flatRows[i].TradeDate.Format("2006-01-02")
	}

	result := models.StatisticSingleStockMapped{
		StockCode: flatRows[0].StockCode,
		Details:   flatRows,
	}

	return []models.StatisticSingleStockMapped{result}, nil

}

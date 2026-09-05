package services

import (
	"fmt"
	"math"
	"sort"
	"time"

	"indonesia-stocks-api/internal/database"
	"indonesia-stocks-api/internal/models"
	"indonesia-stocks-api/internal/repositories"
)

// maxGapUpPercent menolak sinyal jika open entry day gap-up lebih besar
// dari persentase ini terhadap close signal day (menghindari beli di puncak).
const maxGapUpPercent = 5.0

// RunBacktestV1 menjalankan Backtest Strategy Engine V1 (Bandarmologi).
// Signal date (T) terpisah dari entry date (hari trading berikutnya) untuk
// menghindari lookahead bias. Transaction dikelola di layer ini agar atomic.
func RunBacktestV1(req models.BacktestRunRequest) (*models.BacktestRunResponse, error) {
	if req.RunName == "" {
		return nil, fmt.Errorf("run_name is required")
	}
	// Default dari hasil tuning (best win rate & profit factor).
	if req.TPPercent <= 0 {
		req.TPPercent = 10
	}
	if req.SLPercent <= 0 {
		req.SLPercent = 10
	}
	if req.MaxHoldingDays <= 0 {
		req.MaxHoldingDays = 15
	}

	startDate, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		return nil, fmt.Errorf("invalid start_date, format: YYYY-MM-DD")
	}
	endDate, err := time.Parse("2006-01-02", req.EndDate)
	if err != nil {
		return nil, fmt.Errorf("invalid end_date, format: YYYY-MM-DD")
	}
	if startDate.After(endDate) {
		return nil, fmt.Errorf("start_date cannot be after end_date")
	}

	from := startDate.Format("2006-01-02")
	to := endDate.Format("2006-01-02")

	signals, err := repositories.GetBacktestSignals(from, to)
	if err != nil {
		return nil, err
	}

	bars, err := repositories.GetBacktestDailyBars(from, to)
	if err != nil {
		return nil, err
	}

	stockBars := buildStockBars(bars)

	details := simulateTrades(signals, stockBars, req)

	run := buildBacktestRun(req, startDate, endDate, len(signals), details)

	tx, err := database.DB.Beginx()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	runID, err := repositories.InsertBacktestRun(tx, &run)
	if err != nil {
		return nil, err
	}

	run.ID = runID
	for i := range details {
		details[i].BacktestRunID = runID
	}

	if err := repositories.InsertBacktestDetails(tx, details); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return &models.BacktestRunResponse{
		Run:     run,
		Details: details,
	}, nil
}

// buildStockBars mengelompokkan daily bars per stock dan mengurutkan per tanggal.
func buildStockBars(bars []models.BacktestDailyBar) map[string][]models.BacktestDailyBar {
	m := make(map[string][]models.BacktestDailyBar)
	for _, b := range bars {
		m[b.StockCode] = append(m[b.StockCode], b)
	}
	for code := range m {
		sort.SliceStable(m[code], func(i, j int) bool {
			return m[code][i].TradeDate.Before(m[code][j].TradeDate)
		})
	}
	return m
}

// simulateTrades menjalankan logika entry/exit untuk setiap sinyal.
// Aturan:
//   - Entry = trading day berikutnya setelah signal date, harga = open.
//   - TP & SL tersentuh hari sama => LOSS (prioritas SL).
//   - max_holding_days tercapai => EXPIRED (exit di close).
//   - Tidak boleh ada posisi aktif yang overlap utk stock yang sama.
func simulateTrades(signals []models.BacktestSignal, stockBars map[string][]models.BacktestDailyBar, req models.BacktestRunRequest) []models.BacktestDetail {
	stockSignals := make(map[string][]models.BacktestSignal)
	for _, s := range signals {
		stockSignals[s.StockCode] = append(stockSignals[s.StockCode], s)
	}

	details := []models.BacktestDetail{}
	lastExit := make(map[string]time.Time)

	for stock, sigs := range stockSignals {
		sort.SliceStable(sigs, func(i, j int) bool {
			return sigs[i].SignalDate.Before(sigs[j].SignalDate)
		})

		for _, sig := range sigs {
			d := simulateSingleTrade(stock, sig.SignalDate, stockBars[stock], req)
			if d == nil {
				continue
			}

			// No overlapping position utk stock yang sama:
			// posisi aktif dari entry s/d exit, sinyal berikutnya harus
			// entry setelah posisi sebelumnya exit.
			entryTime, err := time.Parse("2006-01-02", d.EntryDate)
			if err != nil {
				continue
			}
			if !entryTime.After(lastExit[stock]) {
				continue
			}

			exitTime, err := time.Parse("2006-01-02", d.ExitDate)
			if err != nil {
				continue
			}
			lastExit[stock] = exitTime
			details = append(details, *d)
		}
	}

	return details
}

func simulateSingleTrade(stock string, signalDate time.Time, bars []models.BacktestDailyBar, req models.BacktestRunRequest) *models.BacktestDetail {
	// Cari entry day = hari trading pertama setelah signal date.
	entryIdx := -1
	for i, b := range bars {
		if b.TradeDate.After(signalDate) {
			entryIdx = i
			break
		}
	}
	if entryIdx == -1 {
		return nil
	}

	entryBar := bars[entryIdx]
	if entryBar.OpenPrice <= 0 {
		return nil
	}

	// Filter gap-up berlebihan: open entry day vs close signal day.
	// Gap-up besar biasanya tanda beli di harga tertinggi (overbought).
	if entryIdx > 0 && bars[entryIdx-1].TradeDate.Equal(signalDate) {
		signalClose := bars[entryIdx-1].ClosePrice
		if signalClose > 0 {
			gapPct := (entryBar.OpenPrice - signalClose) / signalClose * 100
			if gapPct > maxGapUpPercent {
				return nil
			}
		}
	}

	entryPrice := entryBar.OpenPrice
	tpPrice := entryPrice * (1 + req.TPPercent/100)
	slPrice := entryPrice * (1 - req.SLPercent/100)

	result := ""
	exitReason := ""
	exitPrice := 0.0
	exitDate := entryBar.TradeDate
	holdingDays := 0

	for i := entryIdx; i < len(bars); i++ {
		b := bars[i]
		holdingDays = i - entryIdx + 1
		exitDate = b.TradeDate

		// TP & SL tersentuh hari sama => LOSS (prioritas SL).
		if b.HighPrice >= tpPrice && b.LowPrice <= slPrice {
			result = "LOSS"
			exitReason = "TP+SL SAME DAY"
			exitPrice = slPrice
			break
		}
		if b.HighPrice >= tpPrice {
			result = "WIN"
			exitReason = "TP HIT"
			exitPrice = tpPrice
			break
		}
		if b.LowPrice <= slPrice {
			result = "LOSS"
			exitReason = "SL HIT"
			exitPrice = slPrice
			break
		}
		if holdingDays >= req.MaxHoldingDays {
			result = "EXPIRED"
			exitReason = "MAX HOLDING DAYS"
			exitPrice = b.ClosePrice
			break
		}
	}

	returnPercent := 0.0
	switch result {
	case "WIN":
		returnPercent = req.TPPercent
	case "LOSS":
		returnPercent = -req.SLPercent
	case "EXPIRED":
		if entryPrice > 0 {
			returnPercent = (exitPrice - entryPrice) / entryPrice * 100
		}
	default:
		return nil
	}

	return &models.BacktestDetail{
		StockCode:     stock,
		SignalDate:    signalDate.Format("2006-01-02"),
		EntryDate:     entryBar.TradeDate.Format("2006-01-02"),
		EntryPrice:    entryPrice,
		TargetTP:      tpPrice,
		TargetSL:      slPrice,
		ExitDate:      exitDate.Format("2006-01-02"),
		ExitPrice:     exitPrice,
		ExitReason:    exitReason,
		HoldingDays:   holdingDays,
		Status:        result,
		ReturnPercent: returnPercent,
	}
}

// buildBacktestRun menghitung metrik agregat dan mengisi struct run.
func buildBacktestRun(req models.BacktestRunRequest, startDate, endDate time.Time, totalSignals int, details []models.BacktestDetail) models.BacktestRun {
	totalTrades := len(details)
	winCount, lossCount, expiredCount := 0, 0, 0
	sumHoldingDays := 0
	var grossProfit, grossLoss, totalReturn float64

	for _, d := range details {
		sumHoldingDays += d.HoldingDays
		totalReturn += d.ReturnPercent
		if d.ReturnPercent > 0 {
			grossProfit += d.ReturnPercent
		} else if d.ReturnPercent < 0 {
			grossLoss += -d.ReturnPercent
		}

		switch d.Status {
		case "WIN":
			winCount++
		case "LOSS":
			lossCount++
		case "EXPIRED":
			expiredCount++
		}
	}

	winRate := 0.0
	if totalTrades > 0 {
		winRate = float64(winCount) / float64(totalTrades) * 100
	}

	profitFactor := 0.0
	if grossLoss > 0 {
		profitFactor = grossProfit / grossLoss
	}

	expectancy := 0.0
	avgHoldingDays := 0.0
	if totalTrades > 0 {
		expectancy = totalReturn / float64(totalTrades)
		avgHoldingDays = float64(sumHoldingDays) / float64(totalTrades)
	}

	return models.BacktestRun{
		RunName:            req.RunName,
		StartDate:          startDate.Format("2006-01-02"),
		EndDate:            endDate.Format("2006-01-02"),
		TPPercent:          req.TPPercent,
		SLPercent:          req.SLPercent,
		MaxHoldingDays:     req.MaxHoldingDays,
		TotalSignals:       totalSignals,
		TotalTrades:        totalTrades,
		GrossProfit:        round2(grossProfit),
		GrossLoss:          round2(grossLoss),
		WinTrades:          winCount,
		LossTrades:         lossCount,
		ExpiredTrades:      expiredCount,
		WinRate:            round2(winRate),
		ProfitFactor:       round2(profitFactor),
		Expectancy:         round2(expectancy),
		AvgHoldingDays:     round2(avgHoldingDays),
		TotalReturnPercent: round2(totalReturn),
		MaxDrawdown:        round2(computeMaxDrawdown(details)),
	}
}

// computeMaxDrawdown menghitung penurunan terbesar dari kurva ekuitas kumulatif.
func computeMaxDrawdown(details []models.BacktestDetail) float64 {
	var equity, peak, maxDD float64
	for _, d := range details {
		equity += d.ReturnPercent
		if equity > peak {
			peak = equity
		}
		if dd := peak - equity; dd > maxDD {
			maxDD = dd
		}
	}
	return maxDD
}

func round2(v float64) float64 {
	return math.Round(v*100) / 100
}

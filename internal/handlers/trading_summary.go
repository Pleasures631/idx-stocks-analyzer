package handlers

import (
	"fmt"
	"indonesia-stocks-api/internal/constants"
	"indonesia-stocks-api/internal/helpers"
	"indonesia-stocks-api/internal/models"
	"indonesia-stocks-api/internal/repositories"
	"indonesia-stocks-api/internal/services"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type SyncRequest struct {
	StartDate string `json:"start_date" binding:"required"`
	EndDate   string `json:"end_date"`
}

func InsertTradingSummary(c *gin.Context) {
	start := time.Now()

	var req SyncRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "invalid request body"})
		return
	}

	dates, err := helpers.GenerateDateRange(req.StartDate, req.EndDate)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	success := 0
	failed := []string{}
	totalRows := 0

	for _, date := range dates {
		data, err := services.FetchIDX[models.TradingSummary](
			constants.IDXBaseURL,
			constants.ModuleTradingSummary,
			constants.ServiceStockSummary,
			date,
		)

		if err != nil {
			failed = append(failed, date)
			continue
		}

		tradingSummary := make([]models.TradingSummaryDB, 0, len(data))
		for _, d := range data {
			tradingSummary = append(tradingSummary, MapIDXTradingSummaryToModel(d))
		}

		if err := repositories.InsertTradingSummary(tradingSummary); err != nil {
			failed = append(failed, date)
			continue
		}

		success++
		totalRows += len(tradingSummary)

		time.Sleep(300 * time.Millisecond)
	}

	duration := time.Since(start)

	c.JSON(http.StatusOK, gin.H{
		"message": "Trading summary sync completed",
		"mode": func() string {
			if req.EndDate == "" {
				return "single-day"
			}
			return "range"
		}(),
		"start_date":   req.StartDate,
		"end_date":     req.EndDate,
		"success_days": success,
		"failed_days":  failed,
		"total_rows":   totalRows,
		"process_time": duration.String(),
		"process_ms":   duration.Milliseconds(),
		"execute_date": time.Now().Format("2006-01-02"),
	})
}

func GetTopAccumulation(c *gin.Context) {
	days := 7

	data, err := repositories.GetTopAccumulation(days)
	if err != nil {
		c.JSON(500, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(200, gin.H{
		"mode":        "top_accumulation",
		"period_days": days,
		"total":       len(data),
		"data":        data,
	})
}

func GetTopAccumulationEod(c *gin.Context) {
	days := 60

	data, err := repositories.GetTopAccumulationEOD(days)
	if err != nil {
		c.JSON(500, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(200, gin.H{
		"mode":        "top_accumulation_end_of_day",
		"period_days": days,
		"total":       len(data),
		"data":        data,
	})
}

func RunBacktestEOD(c *gin.Context) {
	// Ambil tanggal target dari query param, default ke 7 hari lalu jika kosong
	targetDate := c.Query("date")
	if targetDate == "" {
		// Otomatis hitung tanggal 7 hari kalender ke belakang
		targetDate = time.Now().AddDate(0, 0, -7).Format("2006-01-02")
	}

	data, err := repositories.RunBacktestEOD(targetDate)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	// --- LOGIC STATISTIK SEDERHANA ---
	var totalWin, totalLose int
	var sumProfit float64

	for _, res := range data {
		if res.ProfitLossPct > 0 {
			totalWin++
		} else {
			totalLose++
		}
		sumProfit += res.ProfitLossPct
	}

	winRate := 0.0
	avgProfit := 0.0
	if len(data) > 0 {
		winRate = (float64(totalWin) / float64(len(data))) * 100
		avgProfit = sumProfit / float64(len(data))
	}

	c.JSON(200, gin.H{
		"mode":        "backtest_eod",
		"target_date": targetDate,
		"summary": gin.H{
			"total_signals": len(data),
			"win_rate":      fmt.Sprintf("%.2f%%", winRate),
			"avg_profit":    fmt.Sprintf("%.2f%%", avgProfit),
			"win_count":     totalWin,
			"lose_count":    totalLose,
		},
		"data": data,
	})
}

func GetTopScalping(c *gin.Context) {
	tradeDate := c.Query("date")

	if tradeDate == "" {
		tradeDate = time.Now().Format("2006-01-02")
	}

	data, err := repositories.GetTopSwinger(tradeDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"date":  tradeDate,
		"total": len(data),
		"data":  data,
	})
}

func GetSilentAccumulation(c *gin.Context) {
	days := 7

	data, err := repositories.GetSilentAccumulation(days)
	if err != nil {
		c.JSON(500, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(200, gin.H{
		"mode":        "silent_accumulation_end_of_day",
		"period_days": days,
		"total":       len(data),
		"data":        data,
	})
}

func StatisticSingleStock(c *gin.Context) {
	code := c.Query("stock_code")

	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Stock Code is required",
		})
		return
	}

	data, err := repositories.StatisticSingleStock(code)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{
		"mode":        "Single Stock Statistic",
		"target_date": time.Now(),
		"data":        data,
	})
}

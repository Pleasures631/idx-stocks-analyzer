package handlers

import (
	"fmt"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"indonesia-stocks-api/internal/repositories"
	"indonesia-stocks-api/internal/services"

	"github.com/gin-gonic/gin"
)

type SyncExodusRequest struct {
	Symbol    string `json:"symbol" binding:"required"`
	StartDate string `json:"start_date" binding:"required"`
	EndDate   string `json:"end_date"`
}

type SyncExodusAllRequest struct {
	StartDate string `json:"start_date" binding:"required"`
	EndDate   string `json:"end_date"`
}

type ExodusJob struct {
	Symbol    string
	TradeDate string
}

func FetchExodusBrokerSummaryAll(c *gin.Context) {
	start := time.Now()

	var req SyncExodusAllRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	startDate, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid start_date"})
		return
	}

	endDate := startDate
	if req.EndDate != "" {
		endDate, err = time.Parse("2006-01-02", req.EndDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid end_date"})
			return
		}
	}

	stocks, err := repositories.GetActiveStockCodes()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	existingPairs, err := repositories.GetExistingExodusStockDates(req.StartDate, req.EndDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	existing := make(map[string]map[string]struct{})
	for _, p := range existingPairs {
		dateKey := p.TradeDate.Format("2006-01-02")
		if existing[p.StockCode] == nil {
			existing[p.StockCode] = make(map[string]struct{})
		}
		existing[p.StockCode][dateKey] = struct{}{}
	}

	var jobList []ExodusJob
	var atomicSkippedDays atomic.Int64

	for _, symbol := range stocks {
		for d := startDate; !d.After(endDate); d = d.AddDate(0, 0, 1) {
			dateStr := d.Format("2006-01-02")
			if _, ok := existing[symbol][dateStr]; ok {
				atomicSkippedDays.Add(1)
				continue
			}
			jobList = append(jobList, ExodusJob{Symbol: symbol, TradeDate: dateStr})
		}
	}

	if len(jobList) == 0 {
		c.JSON(http.StatusOK, gin.H{"message": "all data already exists"})
		return
	}

	// 1. Inisialisasi Channel Sejumlah Total Jobs
	jobs := make(chan ExodusJob, len(jobList))
	for _, job := range jobList {
		jobs <- job
	}
	close(jobs) // Langsung tutup channel setelah diisi agar worker tahu kapan selesai

	// 2. Jalankan 3 Worker Paralel
	const workerCount = 2
	var wg sync.WaitGroup

	var (
		atomicSuccessDays atomic.Int64
		atomicTotalRows   atomic.Int64
		mu                sync.Mutex
		failedStocks      []string
	)

	for w := 1; w <= workerCount; w++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			for job := range jobs {
				// Log sederhana untuk memastikan worker aktif di terminal
				fmt.Printf("[Worker %d] Fetching %s (%s)...\n", workerID, job.Symbol, job.TradeDate)

				data, err := services.FetchExodusMarketDetector(job.Symbol, job.TradeDate, job.TradeDate)
				if err != nil {
					mu.Lock()
					failedStocks = append(failedStocks, job.Symbol)
					mu.Unlock()
					continue
				}

				if len(data.BrokerSummary.BrokersBuy) == 0 && len(data.BrokerSummary.BrokersSell) == 0 {
					continue
				}

				rows := MapExodusBrokerSummaryToModel(data.BrokerSummary)
				if err := repositories.UpsertExodusBrokerSummary(rows); err != nil {
					mu.Lock()
					failedStocks = append(failedStocks, job.Symbol)
					mu.Unlock()
					continue
				}

				atomicSuccessDays.Add(1)
				atomicTotalRows.Add(int64(len(rows)))
			}
		}(w)
	}

	// Tunggu semua worker selesai
	wg.Wait()

	duration := time.Since(start)

	c.JSON(http.StatusOK, gin.H{
		"message":      "bulk sync completed",
		"total_jobs":   len(jobList),
		"success_days": atomicSuccessDays.Load(),
		"skipped_days": atomicSkippedDays.Load(),
		"total_rows":   atomicTotalRows.Load(),
		"failed_count": len(failedStocks),
		"process_time": duration.String(),
	})
}

func FetchExodusBrokerSummary(c *gin.Context) {
	start := time.Now()

	var req SyncExodusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request body: symbol, start_date (YYYY-MM-DD) required",
		})
		return
	}

	startDate, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid start_date, format: YYYY-MM-DD"})
		return
	}

	endDate := startDate
	if req.EndDate != "" {
		endDate, err = time.Parse("2006-01-02", req.EndDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid end_date, format: YYYY-MM-DD"})
			return
		}
	}

	if startDate.After(endDate) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "start_date cannot be after end_date"})
		return
	}

	successDays := 0
	failedDays := []string{}
	skippedDays := []string{}
	totalRows := 0
	totalBuy := 0
	totalSell := 0

	for d := startDate; !d.After(endDate); d = d.AddDate(0, 0, 1) {
		dateStr := d.Format("2006-01-02")

		data, err := services.FetchExodusMarketDetector(req.Symbol, dateStr, dateStr)
		if err != nil {
			failedDays = append(failedDays, dateStr)
			continue
		}

		if len(data.BrokerSummary.BrokersBuy) == 0 && len(data.BrokerSummary.BrokersSell) == 0 {
			skippedDays = append(skippedDays, dateStr)
			continue
		}

		rows := MapExodusBrokerSummaryToModel(data.BrokerSummary)

		if err := repositories.UpsertExodusBrokerSummary(rows); err != nil {
			failedDays = append(failedDays, dateStr)
			continue
		}

		successDays++
		totalRows += len(rows)
		totalBuy += len(data.BrokerSummary.BrokersBuy)
		totalSell += len(data.BrokerSummary.BrokersSell)
	}

	duration := time.Since(start)

	c.JSON(http.StatusOK, gin.H{
		"message":      "exodus broker summary sync completed",
		"symbol":       req.Symbol,
		"start_date":   req.StartDate,
		"end_date":     req.EndDate,
		"success_days": successDays,
		"failed_days":  failedDays,
		"skipped_days": skippedDays,
		"total_buy":    totalBuy,
		"total_sell":   totalSell,
		"total_rows":   totalRows,
		"process_time": duration.String(),
		"process_ms":   duration.Milliseconds(),
		"execute_date": time.Now().Format("2006-01-02"),
	})
}

func GetExodusBrokerSummary(c *gin.Context) {
	symbol := c.Query("symbol")
	from := c.Query("from")
	to := c.Query("to")

	rows, err := repositories.GetExodusBrokerSummary(symbol, from, to)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"symbol": symbol,
		"from":   from,
		"to":     to,
		"total":  len(rows),
		"data":   rows,
	})
}

// AnalyzeExodusBrokerFlow menganalisis broker flow sebuah saham (sama dengan
// perintah /analyze di Telegram) untuk dikonsumsi frontend.
func AnalyzeExodusBrokerFlow(c *gin.Context) {
	symbol := c.Query("symbol")
	if symbol == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "symbol is required"})
		return
	}

	from := c.Query("from")
	to := c.Query("to")

	result, err := services.AnalyzeExodusBrokerFlowService(symbol, from, to)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    result,
	})
}

// GetAnalyzeBySymbol menganalisis broker flow sebuah saham dari path param
// `/stocks/:symbol/analyze`, konsisten dengan `/stocks/:symbol`.
func GetAnalyzeBySymbol(c *gin.Context) {
	symbol := c.Param("symbol")
	if symbol == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "symbol is required"})
		return
	}

	from := c.Query("from")
	to := c.Query("to")

	result, err := services.AnalyzeExodusBrokerFlowService(symbol, from, to)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    result,
	})
}

// GetTickerDetail mengembalikan data detail saham (chart harga, volume per
// broker, dan summary buy/sell per broker) untuk halaman ticker frontend.
func GetTickerDetail(c *gin.Context) {
	symbol := c.Param("symbol")
	if symbol == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "symbol is required"})
		return
	}

	from := c.Query("from")
	to := c.Query("to")
	rangeStr := c.Query("range")

	result, err := services.GetStockDetail(symbol, from, to, rangeStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    result,
	})
}

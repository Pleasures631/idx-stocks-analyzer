package routes

import (
	"indonesia-stocks-api/internal/handlers"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.Engine) {
	r.GET("/health", handlers.HealthCheck)
	r.GET("/idx/brokersummary", handlers.FetchBrokerSummary)
	r.GET("/idx/brokersummary/analyze", handlers.AnalyzeBrokerSummary)
	r.POST("/tradingsummary/insert", handlers.InsertTradingSummary)
	r.POST("/idx/syncbroker", handlers.SyncBrokerFromIDX)
	r.POST("/idx/syncstocks", handlers.SyncStocksFromIDX)
	r.GET("/analyze/single-stocks", handlers.StatisticSingleStock)
	r.GET("/analyze/top-accumulation", handlers.GetTopAccumulation)
	r.GET("/analyze/top-accumulation-eod", handlers.GetTopAccumulationEod)
	r.GET("/analyze/silent-accumulation", handlers.GetSilentAccumulation)
	r.GET("/backtest/top-accumulation-eod", handlers.RunBacktestEOD)
	r.GET("/analyze/top-scalping-daily", handlers.GetTopScalping)
}

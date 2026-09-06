package routes

import (
	"indonesia-stocks-api/internal/handlers"
	"net/http"

	"github.com/gin-gonic/gin"
)

// CORS middleware membolehkan request dari origin browser lain (mis. frontend
// Next.js di localhost:3000) dan menangani preflight OPTIONS.
func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if origin == "" {
			c.Next()
			return
		}

		c.Header("Access-Control-Allow-Origin", origin)
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Authorization, Content-Type, Accept, X-Requested-With")
		c.Header("Vary", "Origin")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

func RegisterRoutes(r *gin.Engine) {
	r.Use(CORS())

	r.GET("/health", handlers.HealthCheck)
	r.POST("/users/register", handlers.RegisterUser)
	r.POST("/users/register/verify", handlers.VerifyRegistration)
	r.POST("/auth/login", handlers.Login)
	r.POST("/auth/refresh", handlers.RefreshToken)
	r.POST("/auth/forgot-password", handlers.ForgotPassword)
	r.POST("/auth/reset-password", handlers.ResetPassword)
	r.GET("/education/articles", handlers.GetEducationArticles)
	r.GET("/education/articles/:slug", handlers.GetEducationArticle)
	r.GET("/stocks/list", handlers.GetStocksList)
	r.GET("/market/ihsg", handlers.GetStockbitIHSG)
	r.GET("/market/ihsg/foreign-domestic", handlers.GetStockbitForeignDomestic)
	r.POST("/market/ihsg/foreign-domestic/sync", handlers.SyncStockbitForeignDomestic)
	r.GET("/market/ihsg/chart", handlers.GetStockbitIHSGChart)
	r.POST("/market/ihsg/chart/sync", handlers.SyncStockbitIHSGChart)
	r.GET("/brokers/list", handlers.GetBrokersList)
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
	r.GET("/exodus/broker-summary", handlers.GetExodusBrokerSummary)
	r.GET("/exodus/broker-summary/analyze-flow", handlers.AnalyzeExodusBrokerFlow)
	r.GET("/stocks/:symbol", handlers.GetTickerDetail)
	r.GET("/stocks/:symbol/analyze", handlers.GetAnalyzeBySymbol)
	r.POST("/exodus/broker-summary/fetch", handlers.FetchExodusBrokerSummary)
	r.POST("/exodus/broker-summary/bulkfetch", handlers.FetchExodusBrokerSummaryAll)
	r.POST("/api/v1/backtest/run", handlers.RunBacktestV1)
}

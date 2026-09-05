package handlers

import (
	"net/http"

	"indonesia-stocks-api/internal/models"
	"indonesia-stocks-api/internal/services"

	"github.com/gin-gonic/gin"
)

// RunBacktestV1 menerima konfigurasi backtest dan menjalankan strategi.
func RunBacktestV1(c *gin.Context) {
	var req models.BacktestRunRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	result, err := services.RunBacktestV1(req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    result,
	})
}

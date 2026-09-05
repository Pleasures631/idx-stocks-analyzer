package handlers

import (
	"indonesia-stocks-api/internal/repositories"
	"net/http"

	"github.com/gin-gonic/gin"
)

func GetStocksList(c *gin.Context) {
	stocks, err := repositories.GetAllStocks()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "failed fetch stock list",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"total":   len(stocks),
		"data":    stocks,
	})
}

func GetBrokersList(c *gin.Context) {
	brokers, err := repositories.GetAllBrokers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "failed fetch broker list",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"total":   len(brokers),
		"data":    brokers,
	})
}

package handlers

import (
	"indonesia-stocks-api/internal/constants"
	"indonesia-stocks-api/internal/models"
	"indonesia-stocks-api/internal/repositories"
	"indonesia-stocks-api/internal/services"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func SyncStocksFromIDX(c *gin.Context) {

	start := time.Now()

	data, err := services.FetchIDX[models.IDXStock](constants.IDXBaseURL, constants.ModuleStockData, constants.ServiceStocksList)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
	}

	//debug mode
	// for i, d := range data {
	// 	log.Printf("[%d] %+v\n", i, d)
	// }

	stocks := make([]models.StocksList, 0, len(data))
	for _, b := range data {
		stocks = append(stocks, MapIDXStockToModel(b))
	}

	if err := repositories.UpsertStocks(stocks); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":  "failed insert stocks",
			"detail": err.Error(),
		})
		return
	}

	duration := time.Since(start)

	c.JSON(http.StatusOK, gin.H{
		"message":      "stocks synced",
		"total":        len(stocks),
		"process_time": duration.String(),
		"process_ms":   duration.Milliseconds(),
	})
}

func SyncBrokerFromIDX(c *gin.Context) {
	start := time.Now()

	data, err := services.FetchIDX[models.IDXBroker](constants.IDXBaseURL, constants.ModuleExchangeMember, constants.ServiceBrokerList)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
	}

	brokers := make([]models.BrokerList, 0, len(data))
	for _, b := range data {
		brokers = append(brokers, MapIDXBrokerToModel(b))
	}

	if err := repositories.UpsertBrokers(brokers); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":  "failed insert brokers",
			"detail": err.Error(),
		})
		return
	}

	duration := time.Since(start)

	c.JSON(http.StatusOK, gin.H{
		"message":      "brokers synced",
		"total":        len(brokers),
		"process_time": duration.String(),
		"process_ms":   duration.Milliseconds(),
	})
}

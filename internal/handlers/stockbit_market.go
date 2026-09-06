package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"indonesia-stocks-api/internal/repositories"
	"indonesia-stocks-api/internal/services"
)

func GetStockbitIHSG(c *gin.Context) {
	quote, err := services.GetStockbitIHSG(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"success": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": quote})
}

func GetStockbitForeignDomestic(c *gin.Context) {
	symbol := c.DefaultQuery("symbol", "IHSG")
	from, to := c.Query("from"), c.Query("to")
	if from != "" {
		if _, err := time.Parse("2006-01-02", from); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "from must use YYYY-MM-DD"})
			return
		}
	}
	if to != "" {
		if _, err := time.Parse("2006-01-02", to); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "to must use YYYY-MM-DD"})
			return
		}
	}
	if from != "" && to != "" && from > to {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "from cannot be after to"})
		return
	}
	rows, err := repositories.GetStockbitForeignDomestic(symbol, from, to)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "symbol": symbol, "total": len(rows), "data": rows})
}

func SyncStockbitForeignDomestic(c *gin.Context) {
	symbol := c.DefaultQuery("symbol", "IHSG")
	row, err := services.SyncStockbitForeignDomestic(c.Request.Context(), symbol)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"success": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": row})
}

func GetStockbitIHSGChart(c *gin.Context) {
	symbol := c.DefaultQuery("symbol", "IHSG")
	from, to := c.Query("from"), c.Query("to")
	if from == "" || to == "" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "from and to are required in YYYY-MM-DD format"})
		return
	}
	if _, err := time.Parse("2006-01-02", from); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "from must use YYYY-MM-DD"})
		return
	}
	if _, err := time.Parse("2006-01-02", to); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "to must use YYYY-MM-DD"})
		return
	}
	if from > to {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "from cannot be after to"})
		return
	}
	rows, err := repositories.GetStockbitIHSGChart(symbol, from, to)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "symbol": symbol, "total": len(rows), "data": rows})
}

func SyncStockbitIHSGChart(c *gin.Context) {
	symbol := c.DefaultQuery("symbol", "IHSG")
	from, to := c.Query("from"), c.Query("to")
	if from == "" || to == "" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "from and to are required in YYYY-MM-DD format"})
		return
	}
	count, err := services.SyncStockbitIHSGChart(c.Request.Context(), symbol, from, to)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"success": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "symbol": symbol, "from": from, "to": to, "inserted_or_updated": count})
}

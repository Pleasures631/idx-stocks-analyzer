package handlers

import (
	"net/http"
	"time"

	"indonesia-stocks-api/internal/constants"
	"indonesia-stocks-api/internal/models"
	"indonesia-stocks-api/internal/services"

	"github.com/gin-gonic/gin"
)

const (
	maxRetry   = 3
	retryDelay = 2 * time.Second
)

//hit broksum to idx
//if 403 forbiden by cloudflare hit one more again

func FetchBrokerSummary(c *gin.Context) {
	date := c.Query("date")
	if date == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "date is required, format: YYYYMMDD",
		})
		return
	}

	//kalo ga mau ada variabel yang di pake ganti pake _
	_, err := time.Parse("20060102", date)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid start_date"})
		return
	}

	data, err := services.FetchIDX[models.BrokerSummary](constants.IDXBaseURL, constants.ModuleTradingSummary, constants.ServiceBrokerSummary, date)
	if err != nil {
		time.Sleep(2 * time.Second)

		data, err = services.FetchIDX[models.BrokerSummary](constants.IDXBaseURL, constants.ModuleTradingSummary, constants.ServiceBrokerSummary, date)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"data": data,
	})
}

func AnalyzeBrokerSummary(c *gin.Context) {
	start_date := c.Query("start_date")
	end_date := c.Query("end_date")

	if start_date == "" || end_date == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "start date and end date is required, format: YYYYMMDD",
		})
		return
	}

	if start_date > end_date {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "start date tidak boleh melebihi end date",
		})
		return
	}

	if start_date > end_date {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "start date tidak boleh melebihi end date",
		})
		return
	}

	start, err := time.Parse("20060102", start_date)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid start_date"})
		return
	}

	end, err := time.Parse("20060102", end_date)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid end_date"})
		return
	}

	days := int(end.Sub(start).Hours()/24) + 1
	if days > 7 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "range tanggal maksimal 7 hari",
		})
		return
	}

	var brokers []models.BrokerSummary

	for d := start; !d.After(end); d = d.AddDate(0, 0, 1) {
		dateStr := d.Format("20060102")

		data, err := services.FetchIDX[models.BrokerSummary](constants.IDXBaseURL, constants.ModuleTradingSummary, constants.ServiceBrokerSummary, dateStr)
		if err != nil {
			time.Sleep(2 * time.Second)

			data, err = services.FetchIDX[models.BrokerSummary](constants.IDXBaseURL, constants.ModuleTradingSummary, constants.ServiceBrokerSummary, dateStr)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": err.Error(),
				})
				return
			}
		}

		brokers = append(brokers, data...)
		time.Sleep(1 * time.Second)
	}

	c.JSON(http.StatusOK, gin.H{
		"data": brokers,
	})
}

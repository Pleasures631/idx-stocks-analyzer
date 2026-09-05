package handlers

import (
	"indonesia-stocks-api/internal/helpers"
	"indonesia-stocks-api/internal/models"
	"time"
)

func MapIDXBrokerToModel(b models.IDXBroker) models.BrokerList {
	return models.BrokerList{
		BrokerCode:    b.Code,
		BrokerName:    b.Name,
		BrokerLicense: b.License,
	}
}

func MapIDXStockToModel(s models.IDXStock) models.StocksList {
	var listingDate time.Time

	if s.ListingDate != "" {
		t, err := time.Parse("2006-01-02T15:04:05", s.ListingDate)
		if err == nil {
			listingDate = t
		}
	}

	return models.StocksList{
		StockCode:    s.Code,
		StockName:    s.Name,
		ListingDate:  listingDate,
		TotalShares:  uint64(s.Shares),
		ListingBoard: s.ListingBoard,
		IsActive:     true,
	}
}

func MapIDXTradingSummaryToModel(s models.TradingSummary) models.TradingSummaryDB {
	var tradeDate time.Time
	if s.Date != "" {
		t, err := time.Parse("2006-01-02T15:04:05", s.Date)
		if err == nil {
			tradeDate = t
		}
	}

	high := s.High
	low := s.Low
	closePrice := s.Close

	closeStrength := float64(0)
	if high > low {
		closeStrength = ((closePrice - low) / (high - low)) * 100
	}

	return models.TradingSummaryDB{
		IdxIDStockSummary: s.IDStockSummary,
		TradeDate:         tradeDate,

		StockCode: s.StockCode,
		StockName: s.StockName,

		Previous:      s.Previous,
		OpenPrice:     s.OpenPrice,
		FirstTrade:    s.FirstTrade,
		High:          s.High,
		Low:           s.Low,
		Close:         s.Close,
		Change:        s.Change,
		CloseStrength: closeStrength,

		Volume:    int64(s.Volume),
		Value:     s.Value,
		Frequency: int64(s.Frequency),

		IndexIndividual: s.IndexIndividual,

		Offer:       s.Offer,
		OfferVolume: int64(s.OfferVolume),
		Bid:         s.Bid,
		BidVolume:   int64(s.BidVolume),

		ListedShares:    int64(s.ListedShares),
		TradeableShares: int64(s.TradebleShares),
		WeightForIndex:  s.WeightForIndex,

		ForeignSell: s.ForeignSell,
		ForeignBuy:  s.ForeignBuy,

		NonRegularVolume:    int64(s.NonRegularVolume),
		NonRegularValue:     s.NonRegularValue,
		NonRegularFrequency: int64(s.NonRegularFrequency),

		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

func MapExodusBrokerSummaryToModel(summary models.ExodusBrokerSummary) []models.ExodusBrokerSummaryDB {
	rows := make([]models.ExodusBrokerSummaryDB, 0, len(summary.BrokersBuy)+len(summary.BrokersSell))

	for _, b := range summary.BrokersBuy {
		tradeDate := parseExodusDate(b.Date)

		rows = append(rows, models.ExodusBrokerSummaryDB{
			StockCode:   b.StockCode,
			TradeDate:   tradeDate,
			BrokerCode:  b.BrokerCode,
			Side:        "BUY",
			BrokerType:  b.Type,
			Lot:         helpers.ParseFloat(b.Blot),
			Volume:      helpers.ParseFloat(b.Blotv),
			Value:       helpers.ParseFloat(b.Bval),
			Turnover:    helpers.ParseFloat(b.Bvalv),
			AvgPrice:    helpers.ParseFloat(b.BuyAvgPrice),
			Frequency:   int64(helpers.ParseFloat(b.Freq)),
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		})
	}

	for _, s := range summary.BrokersSell {
		tradeDate := parseExodusDate(s.Date)
		rows = append(rows, models.ExodusBrokerSummaryDB{
			StockCode:   s.StockCode,
			TradeDate:   tradeDate,
			BrokerCode:  s.BrokerCode,
			Side:        "SELL",
			BrokerType:  s.Type,
			Lot:         helpers.ParseFloat(s.Slot),
			Volume:      helpers.ParseFloat(s.Slotv),
			Value:       helpers.ParseFloat(s.Sval),
			Turnover:    helpers.ParseFloat(s.Svalv),
			AvgPrice:    helpers.ParseFloat(s.SellAvgPrice),
			Frequency:   int64(helpers.ParseFloat(s.Freq)),
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		})
	}

	return rows
}

func parseExodusDate(dateStr string) time.Time {
	if dateStr == "" {
		return time.Time{}
	}

	t, err := time.Parse("20060102", dateStr)
	if err != nil {
		return time.Time{}
	}

	return t
}

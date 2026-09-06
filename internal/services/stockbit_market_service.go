package services

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"indonesia-stocks-api/internal/constants"
	"indonesia-stocks-api/internal/models"
	"indonesia-stocks-api/internal/repositories"
)

type stockbitIHSGCrawlerResult struct {
	Symbol        string `json:"symbol"`
	Name          string `json:"name"`
	Price         string `json:"price"`
	Change        string `json:"change"`
	ChangePct     string `json:"change_pct"`
	AsOf          string `json:"as_of"`
	Volume        string `json:"volume"`
	AverageVolume string `json:"average_volume"`
	SourceURL     string `json:"source_url"`
}

var stockbitIHSGCache struct {
	sync.Mutex
	quote     models.StockbitIHSGQuote
	fetchedAt time.Time
}

func CrawlStockbitIHSG(ctx context.Context) (models.StockbitIHSGQuote, error) {
	script := strings.TrimSpace(os.Getenv("STOCKBIT_IHSG_CRAWLER_SCRIPT"))
	if script == "" {
		script = filepath.Join("..", "tools", "stockbit-auth", "crawl-ihsg.mjs")
	}
	request := "Playwright browser: https://stockbit.com/symbol/IHSG"
	output, err := exec.CommandContext(ctx, "node", script).CombinedOutput()
	if err != nil {
		LogInterfaceCall("CrawlStockbitIHSG", request, string(output), 0, err)
		return models.StockbitIHSGQuote{}, fmt.Errorf("crawl Stockbit IHSG gagal: %w", err)
	}

	var result stockbitIHSGCrawlerResult
	if err := json.Unmarshal(output, &result); err != nil {
		LogInterfaceCall("CrawlStockbitIHSG", request, string(output), 200, err)
		return models.StockbitIHSGQuote{}, fmt.Errorf("response crawler IHSG tidak valid: %w", err)
	}
	if result.Symbol == "" || result.Price == "" || result.Change == "" || result.Volume == "" {
		err := fmt.Errorf("response crawler IHSG tidak lengkap")
		LogInterfaceCall("CrawlStockbitIHSG", request, string(output), 200, err)
		return models.StockbitIHSGQuote{}, err
	}

	LogInterfaceCall("CrawlStockbitIHSG", request, string(output), 200, nil)
	return models.StockbitIHSGQuote{
		Symbol: result.Symbol, Name: result.Name, Price: result.Price,
		Change: result.Change, ChangePct: result.ChangePct, AsOf: result.AsOf,
		Volume: result.Volume, AverageVolume: result.AverageVolume, SourceURL: result.SourceURL,
	}, nil
}

func GetStockbitIHSG(ctx context.Context) (models.StockbitIHSGQuote, error) {
	stockbitIHSGCache.Lock()
	if stockbitIHSGCache.quote.Price != "" && time.Since(stockbitIHSGCache.fetchedAt) < 5*time.Minute {
		quote := stockbitIHSGCache.quote
		stockbitIHSGCache.Unlock()
		return quote, nil
	}
	stockbitIHSGCache.Unlock()

	quote, err := CrawlStockbitIHSG(ctx)
	if err != nil {
		return models.StockbitIHSGQuote{}, err
	}
	stockbitIHSGCache.Lock()
	stockbitIHSGCache.quote = quote
	stockbitIHSGCache.fetchedAt = time.Now()
	stockbitIHSGCache.Unlock()
	return quote, nil
}

type stockbitFlowMetric struct {
	Value struct {
		Raw int64 `json:"raw"`
	} `json:"value"`
}

type stockbitFlowPayload struct {
	Data struct {
		Value struct {
			ForeignBuy   stockbitFlowMetric `json:"foreign_buy"`
			ForeignSell  stockbitFlowMetric `json:"foreign_sell"`
			DomesticBuy  stockbitFlowMetric `json:"domestic_buy"`
			DomesticSell stockbitFlowMetric `json:"domestic_sell"`
		} `json:"value"`
		Volume struct {
			ForeignBuy   stockbitFlowMetric `json:"foreign_buy"`
			ForeignSell  stockbitFlowMetric `json:"foreign_sell"`
			DomesticBuy  stockbitFlowMetric `json:"domestic_buy"`
			DomesticSell stockbitFlowMetric `json:"domestic_sell"`
		} `json:"volume"`
		Frequency struct {
			ForeignBuy   stockbitFlowMetric `json:"foreign_buy"`
			ForeignSell  stockbitFlowMetric `json:"foreign_sell"`
			DomesticBuy  stockbitFlowMetric `json:"domestic_buy"`
			DomesticSell stockbitFlowMetric `json:"domestic_sell"`
		} `json:"frequency"`
		LastUpdated string `json:"last_updated"`
		From        string `json:"from"`
		To          string `json:"to"`
	} `json:"data"`
}

func FetchStockbitForeignDomestic(ctx context.Context, symbol string) (models.StockbitForeignDomesticDB, error) {
	var result models.StockbitForeignDomesticDB
	token := normalizeExodusToken(os.Getenv("EXODUS_TOKEN"))
	if token == "" {
		return result, fmt.Errorf("EXODUS_TOKEN not set in .env")
	}
	if symbol == "" {
		return result, fmt.Errorf("symbol is required")
	}

	endpoint := fmt.Sprintf("%s/findata-view/foreign-domestic/v1/chart-data/%s?market_type=MARKET_TYPE_REGULAR&period=PERIOD_RANGE_1D", constants.ExodusBaseURL, url.PathEscape(strings.ToUpper(symbol)))
	var lastErr error
	refreshAttempted := false
	for attempt := 1; attempt <= 3; attempt++ {
		request, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
		if err != nil {
			return result, err
		}
		request.Header.Set("Authorization", "Bearer "+token)
		request.Header.Set("User-Agent", NextUserAgent())
		request.Header.Set("Accept", "application/json, text/plain, */*")

		response, err := exodusHTTPClient.Do(request)
		if err != nil {
			lastErr = err
			LogInterfaceCall("FetchStockbitForeignDomestic", endpoint, "", 0, err)
			time.Sleep(time.Duration(attempt) * 500 * time.Millisecond)
			continue
		}
		body, readErr := io.ReadAll(response.Body)
		response.Body.Close()
		if readErr != nil {
			lastErr = readErr
			LogInterfaceCall("FetchStockbitForeignDomestic", endpoint, "", response.StatusCode, readErr)
			continue
		}
		LogInterfaceCall("FetchStockbitForeignDomestic", endpoint, string(body), response.StatusCode, nil)

		if response.StatusCode == http.StatusUnauthorized && !refreshAttempted {
			refreshAttempted = true
			if err := refreshExodusToken(ctx, token); err != nil {
				return result, fmt.Errorf("refresh Exodus token: %w", err)
			}
			token = normalizeExodusToken(os.Getenv("EXODUS_TOKEN"))
			continue
		}
		if response.StatusCode != http.StatusOK {
			return result, fmt.Errorf("Stockbit foreign-domestic request failed: %s", response.Status)
		}

		var payload stockbitFlowPayload
		if err := json.Unmarshal(body, &payload); err != nil {
			return result, fmt.Errorf("invalid Stockbit foreign-domestic response: %w", err)
		}
		tradeDate, err := time.Parse("2006-01-02", payload.Data.From)
		if err != nil || payload.Data.From != payload.Data.To {
			return result, fmt.Errorf("invalid 1D Stockbit flow date: from=%q to=%q", payload.Data.From, payload.Data.To)
		}
		result = models.StockbitForeignDomesticDB{
			Symbol: symbol, TradeDate: tradeDate, MarketType: "REGULAR",
			ForeignBuyValue: payload.Data.Value.ForeignBuy.Value.Raw, ForeignSellValue: payload.Data.Value.ForeignSell.Value.Raw,
			DomesticBuyValue: payload.Data.Value.DomesticBuy.Value.Raw, DomesticSellValue: payload.Data.Value.DomesticSell.Value.Raw,
			ForeignNetValue:  payload.Data.Value.ForeignBuy.Value.Raw - payload.Data.Value.ForeignSell.Value.Raw,
			DomesticNetValue: payload.Data.Value.DomesticBuy.Value.Raw - payload.Data.Value.DomesticSell.Value.Raw,
			ForeignBuyVolume: payload.Data.Volume.ForeignBuy.Value.Raw, ForeignSellVolume: payload.Data.Volume.ForeignSell.Value.Raw,
			DomesticBuyVolume: payload.Data.Volume.DomesticBuy.Value.Raw, DomesticSellVolume: payload.Data.Volume.DomesticSell.Value.Raw,
			ForeignBuyFreq: payload.Data.Frequency.ForeignBuy.Value.Raw, ForeignSellFreq: payload.Data.Frequency.ForeignSell.Value.Raw,
			DomesticBuyFreq: payload.Data.Frequency.DomesticBuy.Value.Raw, DomesticSellFreq: payload.Data.Frequency.DomesticSell.Value.Raw,
			LastUpdated: payload.Data.LastUpdated,
		}
		if result.TradeDate.IsZero() || result.Symbol == "" {
			return result, fmt.Errorf("Stockbit foreign-domestic response missing required data")
		}
		return result, nil
	}
	return result, fmt.Errorf("Stockbit foreign-domestic request failed after retries: %v", lastErr)
}

func SyncStockbitForeignDomestic(ctx context.Context, symbol string) (models.StockbitForeignDomesticDB, error) {
	row, err := FetchStockbitForeignDomestic(ctx, symbol)
	if err != nil {
		return row, err
	}
	if err := repositories.UpsertStockbitForeignDomestic(row); err != nil {
		return row, err
	}
	return row, nil
}

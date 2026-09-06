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
	"strconv"
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

type stockbitChartPrice struct {
	FormattedDate string  `json:"formatted_date"`
	XLabel        string  `json:"xlabel"`
	Value         string  `json:"value"`
	Percentage    string  `json:"percentage"`
	Change        float64 `json:"change"`
}

type stockbitChartPayload struct {
	Data struct {
		Prices []stockbitChartPrice `json:"prices"`
	} `json:"data"`
}

func FetchStockbitIHSGChart(ctx context.Context, symbol, from, to string) ([]models.StockbitIHSGChartPoint, error) {
	token := normalizeExodusToken(os.Getenv("EXODUS_TOKEN"))
	if token == "" {
		return nil, fmt.Errorf("EXODUS_TOKEN not set in .env")
	}
	if symbol == "" || from == "" || to == "" {
		return nil, fmt.Errorf("symbol, from, and to are required")
	}
	endpoint := fmt.Sprintf("%s/charts/%s/daily?from=%s&to=%s&interval=INTERVAL_CHART_MINUTELY", constants.ExodusBaseURL, url.PathEscape(strings.ToUpper(symbol)), url.QueryEscape(from), url.QueryEscape(to))
	refreshAttempted := false
	for attempt := 1; attempt <= 3; attempt++ {
		request, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
		if err != nil {
			return nil, err
		}
		request.Header.Set("Authorization", "Bearer "+token)
		request.Header.Set("User-Agent", NextUserAgent())
		request.Header.Set("Accept", "application/json, text/plain, */*")
		response, err := exodusHTTPClient.Do(request)
		if err != nil {
			LogInterfaceCall("FetchStockbitIHSGChart", endpoint, "", 0, err)
			time.Sleep(time.Duration(attempt) * 500 * time.Millisecond)
			continue
		}
		body, readErr := io.ReadAll(response.Body)
		response.Body.Close()
		if readErr != nil {
			LogInterfaceCall("FetchStockbitIHSGChart", endpoint, "", response.StatusCode, readErr)
			return nil, readErr
		}
		LogInterfaceCall("FetchStockbitIHSGChart", endpoint, string(body), response.StatusCode, nil)
		if response.StatusCode == http.StatusUnauthorized && !refreshAttempted {
			refreshAttempted = true
			if err := refreshExodusToken(ctx, token); err != nil {
				return nil, fmt.Errorf("refresh Exodus token: %w", err)
			}
			token = normalizeExodusToken(os.Getenv("EXODUS_TOKEN"))
			continue
		}
		if response.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("Stockbit chart request failed: %s", response.Status)
		}
		var payload stockbitChartPayload
		if err := json.Unmarshal(body, &payload); err != nil {
			return nil, fmt.Errorf("invalid Stockbit chart response: %w", err)
		}
		location, err := time.LoadLocation("Asia/Jakarta")
		if err != nil {
			return nil, err
		}
		points := make([]models.StockbitIHSGChartPoint, 0, len(payload.Data.Prices))
		for _, price := range payload.Data.Prices {
			observedAt, parseErr := time.ParseInLocation("2006-01-02 15:04:05", price.FormattedDate, location)
			if parseErr != nil {
				return nil, fmt.Errorf("invalid Stockbit chart date %q: %w", price.FormattedDate, parseErr)
			}
			value, parseErr := strconv.ParseFloat(price.Value, 64)
			if parseErr != nil {
				return nil, fmt.Errorf("invalid Stockbit chart value %q: %w", price.Value, parseErr)
			}
			percentage, parseErr := strconv.ParseFloat(price.Percentage, 64)
			if parseErr != nil {
				return nil, fmt.Errorf("invalid Stockbit chart percentage %q: %w", price.Percentage, parseErr)
			}
			tradeDate := time.Date(observedAt.Year(), observedAt.Month(), observedAt.Day(), 0, 0, 0, 0, location)
			points = append(points, models.StockbitIHSGChartPoint{Symbol: strings.ToUpper(symbol), TradeDate: tradeDate, Interval: "INTERVAL_CHART_MINUTELY", ObservedAt: observedAt, XLabel: price.XLabel, Value: value, Percentage: percentage, Change: price.Change})
		}
		return points, nil
	}
	return nil, fmt.Errorf("Stockbit chart request failed after retries")
}

func SyncStockbitIHSGChart(ctx context.Context, symbol, from, to string) (int, error) {
	points, err := FetchStockbitIHSGChart(ctx, symbol, from, to)
	if err != nil {
		return 0, err
	}
	latestByDate := make(map[string]models.StockbitIHSGChartPoint)
	for _, point := range points {
		key := point.TradeDate.Format("2006-01-02")
		current, exists := latestByDate[key]
		if !exists || point.ObservedAt.After(current.ObservedAt) {
			latestByDate[key] = point
		}
	}
	dailyClose := make([]models.StockbitIHSGChartPoint, 0, len(latestByDate))
	for _, point := range latestByDate {
		dailyClose = append(dailyClose, point)
	}
	if err := repositories.ReplaceStockbitIHSGDailyClose(dailyClose); err != nil {
		return 0, err
	}
	return len(dailyClose), nil
}

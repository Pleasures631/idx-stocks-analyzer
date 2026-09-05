package services

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"indonesia-stocks-api/internal/models"
)

type insiderCrawlerOutput struct {
	Records      []models.InsiderBrokerMapping `json:"records"`
	Pages        int                           `json:"pages"`
	RawResponses []insiderRawResponse          `json:"raw_responses"`
}

type insiderRawResponse struct {
	URL    string `json:"url"`
	Status int    `json:"status"`
	Body   string `json:"body"`
}

// CrawlKSEIInsider runs the browser-based crawler using the saved Stockbit session.
// It deliberately does not make a direct HTTP request from Go.
func CrawlKSEIInsider(ctx context.Context) ([]models.InsiderBrokerMapping, int, error) {
	script := strings.TrimSpace(os.Getenv("STOCKBIT_INSIDER_CRAWLER_SCRIPT"))
	if script == "" {
		script = filepath.Join("..", "tools", "stockbit-auth", "crawl-insider-ksei.mjs")
	}
	request := "Playwright browser: https://stockbit.com/insider (KSEI only)"
	cmd := exec.CommandContext(ctx, "node", script)
	output, err := cmd.CombinedOutput()
	if err != nil {
		LogInterfaceCall("CrawlStockbitKSEIInsider", request, string(output), 0, err)
		return nil, 0, fmt.Errorf("crawl Stockbit Insider gagal: %w", err)
	}

	var result insiderCrawlerOutput
	if err := json.Unmarshal(output, &result); err != nil {
		LogInterfaceCall("CrawlStockbitKSEIInsider", request, string(output), 200, err)
		return nil, 0, fmt.Errorf("response crawler Insider tidak valid: %w", err)
	}
	if len(result.RawResponses) == 0 {
		err := fmt.Errorf("crawler tidak mengembalikan response Insider")
		LogInterfaceCall("CrawlStockbitKSEIInsider", request, string(output), 0, err)
		return nil, 0, err
	}

	status := result.RawResponses[len(result.RawResponses)-1].Status
	rawBody := make([]string, 0, len(result.RawResponses))
	for _, response := range result.RawResponses {
		rawBody = append(rawBody, response.Body)
	}
	LogInterfaceCall("CrawlStockbitKSEIInsider", request, strings.Join(rawBody, "\n"), status, nil)
	return result.Records, result.Pages, nil
}

// CrawlKSEIInsiderWithTimeout is the fetchDaily-friendly entry point.
func CrawlKSEIInsiderWithTimeout() ([]models.InsiderBrokerMapping, int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	return CrawlKSEIInsider(ctx)
}

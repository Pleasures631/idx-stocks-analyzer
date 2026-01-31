package services

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const (
	maxRetry   = 3
	retryDelay = 2 * time.Second
)

func FetchIDX[T any](
	idx_url string,
	module string,
	service string,
	dates ...string,
) ([]T, error) {

	var date string
	if len(dates) > 0 {
		date = dates[0]
	}

	url := fmt.Sprintf(
		"%s/%s/%s?length=9999&start=0",
		idx_url,
		module,
		service,
	)

	if date != "" {
		url += "&date=" + date
	}

	client := &http.Client{
		Timeout: 15 * time.Second,
	}

	var lastErr error

	for attempt := 1; attempt <= maxRetry; attempt++ {
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return nil, err
		}

		req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
		req.Header.Set("Accept", "application/json, text/plain, */*")
		req.Header.Set("Accept-Language", "en-US,en;q=0.9")
		req.Header.Set("Referer", "https://www.idx.co.id")
		req.Header.Set("Origin", "https://www.idx.co.id")

		resp, err := client.Do(req)
		if err != nil {
			lastErr = err
			time.Sleep(retryDelay)
			continue
		}

		defer resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			var result struct {
				Data []T `json:"data"`
			}

			if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
				return nil, err
			}

			return result.Data, nil
		}

		if resp.StatusCode >= 500 && resp.StatusCode <= 599 {
			lastErr = fmt.Errorf("IDX server error: %s (attempt %d)", resp.Status, attempt)
			time.Sleep(retryDelay)
			continue
		}

		return nil, fmt.Errorf("IDX request failed: %s", resp.Status)
	}

	return nil, fmt.Errorf("retry failed after %d attempts: %v", maxRetry, lastErr)
}

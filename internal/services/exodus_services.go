package services

import (
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"sync"
	"time"

	"indonesia-stocks-api/internal/constants"
	"indonesia-stocks-api/internal/models"
)

const (
	// exodusMinInterval memastikan jeda minimal antar request ke Exodus
	// untuk menghindari rate limit (SERVICE_BUSY / Too Many Requests).
	exodusMinInterval = 2 * time.Second
	// exodusMaxRetry lebih banyak dari default karena retry juga menunggu
	// lebih lama. User rela lebih lambat asalkan tidak kena SERVICE_BUSY.
	exodusMaxRetry = 10
)

var (
	exodusMu       sync.Mutex
	lastExodusCall time.Time
)

var exodusHTTPClient = &http.Client{
	Timeout: 30 * time.Second,
	Transport: &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 50,
		MaxConnsPerHost:     50,
		IdleConnTimeout:     90 * time.Second,
	},
}

// throttleExodus meng-serialize request ke Exodus dan menjamin jeda minimal
// antar request, berlaku global lintas goroutine (worker pool bulk fetch).
func throttleExodus() {
	exodusMu.Lock()
	defer exodusMu.Unlock()

	wait := exodusMinInterval - time.Since(lastExodusCall)
	if wait > 0 {
		time.Sleep(wait)
	}
	lastExodusCall = time.Now()
}

// exodusBackoff menghitung durasi tunggu antar retry (exponential + jitter).
func exodusBackoff(attempt int) time.Duration {
	base := time.Duration(1<<(attempt-1)) * time.Second // 1s, 2s, 4s, 8s, ...
	if base > 30*time.Second {
		base = 30 * time.Second
	}
	jitter := time.Duration(rand.Int63n(1000)) * time.Millisecond
	return base + jitter
}

// isExodusBusy mengecek apakah response mengindikasikan rate limit:
//   - status HTTP 429 (Too Many Requests)
//   - body berisi error_type = "SERVICE_BUSY" (kadang dikirim dengan status 200)
func isExodusBusy(status int, body []byte) bool {
	if status == http.StatusTooManyRequests {
		return true
	}

	var probe struct {
		ErrorType string `json:"error_type"`
	}
	if err := json.Unmarshal(body, &probe); err == nil {
		return probe.ErrorType == "SERVICE_BUSY"
	}

	return false
}

func FetchExodusMarketDetector(symbol, from, to string) (models.ExodusMarketDetector, error) {
	var result models.ExodusMarketDetector

	token := os.Getenv("EXODUS_TOKEN")
	if token == "" {
		return result, fmt.Errorf("EXODUS_TOKEN not set in .env")
	}

	url := fmt.Sprintf(
		"%s/%s/%s?transaction_type=TRANSACTION_TYPE_NET&market_board=MARKET_BOARD_REGULER&investor_type=INVESTOR_TYPE_ALL&limit=25&from=%s&to=%s",
		constants.ExodusBaseURL,
		constants.ModuleExodusMarketDetector,
		symbol,
		from,
		to,
	)

	var lastErr error

	for attempt := 1; attempt <= exodusMaxRetry; attempt++ {
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return result, err
		}

		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("User-Agent", NextUserAgent())
		req.Header.Set("Accept", "application/json, text/plain, */*")
		req.Header.Set("Accept-Language", "en-US,en;q=0.9")

		resp, err := exodusHTTPClient.Do(req)
		if err != nil {
			lastErr = err
			// LOG ERROR NETWORK / CONNECTION
			LogInterfaceCall("FetchExodusMarketDetector", url, "", 0, err)
			time.Sleep(exodusBackoff(attempt))
			continue
		}

		bodyBytes, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			lastErr = err
			// LOG ERROR READ BODY
			LogInterfaceCall("FetchExodusMarketDetector", url, "", resp.StatusCode, err)
			time.Sleep(exodusBackoff(attempt))
			continue
		}

		// LOG SETIAP RESPONS DARI EXODUS (SUKSES MAUPUN ERROR STATUS)
		LogInterfaceCall("FetchExodusMarketDetector", url, string(bodyBytes), resp.StatusCode, nil)

		// Cek jika kena 429 / SERVICE_BUSY
		if isExodusBusy(resp.StatusCode, bodyBytes) || resp.StatusCode == http.StatusTooManyRequests {
			lastErr = fmt.Errorf("exodus service busy (attempt %d): %s", attempt, resp.Status)
			time.Sleep(exodusBackoff(attempt))
			continue
		}

		if resp.StatusCode == http.StatusOK {
			var payload struct {
				Message string                      `json:"message"`
				Data    models.ExodusMarketDetector `json:"data"`
			}

			if err := json.Unmarshal(bodyBytes, &payload); err != nil {
				return result, err
			}

			// Jeda halus antar worker
			time.Sleep(150 * time.Millisecond)
			return payload.Data, nil
		}

		if resp.StatusCode == http.StatusUnauthorized {
			return result, fmt.Errorf("exodus unauthorized: invalid/expired EXODUS_TOKEN")
		}

		if resp.StatusCode >= 500 && resp.StatusCode <= 599 {
			lastErr = fmt.Errorf("exodus server error: %s (attempt %d)", resp.Status, attempt)
			time.Sleep(exodusBackoff(attempt))
			continue
		}

		return result, fmt.Errorf("exodus request failed: %s", resp.Status)
	}

	return result, fmt.Errorf("retry failed after %d attempts: %v", exodusMaxRetry, lastErr)
}

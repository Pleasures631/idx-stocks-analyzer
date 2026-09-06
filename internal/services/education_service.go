package services

import (
	"bytes"
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

func CrawlEducationArticles(ctx context.Context) ([]models.EducationArticle, error) {
	script := strings.TrimSpace(os.Getenv("EDUCATION_CRAWLER_SCRIPT"))
	if script == "" {
		script = filepath.Join("..", "tools", "stockbit-auth", "crawl-education.mjs")
	}
	cmd := exec.CommandContext(ctx, "node", script)
	var output, errorOutput bytes.Buffer
	cmd.Stdout = &output
	cmd.Stderr = &errorOutput
	err := cmd.Run()
	if err != nil {
		LogInterfaceCall("CrawlEducationArticles", "Playwright browser: https://saham.jamet.id/akademi/", errorOutput.String(), 0, err)
		return nil, fmt.Errorf("crawl edukasi gagal: %w", err)
	}
	var articles []models.EducationArticle
	if err := json.Unmarshal(output.Bytes(), &articles); err != nil {
		LogInterfaceCall("CrawlEducationArticles", "Playwright browser: https://saham.jamet.id/akademi/", output.String(), 200, err)
		return nil, fmt.Errorf("hasil crawl edukasi tidak valid: %w", err)
	}
	LogInterfaceCall("CrawlEducationArticles", "Playwright browser: https://saham.jamet.id/akademi/", output.String(), 200, nil)
	return articles, nil
}

func CrawlEducationArticlesWithTimeout() ([]models.EducationArticle, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	return CrawlEducationArticles(ctx)
}

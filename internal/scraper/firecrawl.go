package scraper

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/dantezy/cold-send0r-bot/internal/config"
	"github.com/dantezy/cold-send0r-bot/internal/models"
)

type FirecrawlScraper struct {
	cfg         config.ScraperConfig
	client      *http.Client
	rateLimiter *time.Ticker
}

func NewFirecrawlScraper(cfg config.ScraperConfig) *FirecrawlScraper {
	return &FirecrawlScraper{
		cfg: cfg,
		client: &http.Client{
			Timeout: time.Duration(cfg.TimeoutMs) * time.Millisecond,
		},
		rateLimiter: time.NewTicker(time.Duration(cfg.RateLimitMs) * time.Millisecond),
	}
}

type firecrawlRequest struct {
	URL     string   `json:"url"`
	Formats []string `json:"formats"`
}

type firecrawlResponse struct {
	Success bool `json:"success"`
	Data    struct {
		Markdown string `json:"markdown"`
	} `json:"data"`
}

func (s *FirecrawlScraper) Scrape(url string) (*models.ScrapeResult, error) {
	<-s.rateLimiter.C

	result := &models.ScrapeResult{URL: url}

	body, err := json.Marshal(firecrawlRequest{
		URL:     url,
		Formats: []string{"markdown"},
	})
	if err != nil {
		return nil, fmt.Errorf("marshaling firecrawl request: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, "https://api.firecrawl.dev/v1/scrape", strings.NewReader(string(body)))
	if err != nil {
		return nil, fmt.Errorf("creating firecrawl request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.cfg.FirecrawlAPIKey)

	resp, err := s.client.Do(req)
	if err != nil {
		result.Error = err.Error()
		return result, nil
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		result.Error = err.Error()
		return result, nil
	}

	if resp.StatusCode != http.StatusOK {
		result.Error = fmt.Sprintf("firecrawl returned %d: %s", resp.StatusCode, string(respBody))
		return result, nil
	}

	var fcResp firecrawlResponse
	if err := json.Unmarshal(respBody, &fcResp); err != nil {
		result.Error = fmt.Sprintf("parsing firecrawl response: %v", err)
		return result, nil
	}

	md := fcResp.Data.Markdown
	if md == "" {
		result.Error = "firecrawl returned empty markdown"
		return result, nil
	}

	result.Markdown = truncate(md, s.cfg.MaxContentLength)
	return result, nil
}

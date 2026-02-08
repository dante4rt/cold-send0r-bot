package scraper

import (
	"fmt"
	"net/url"
	"strings"
	"time"

	htmltomarkdown "github.com/JohannesKaufmann/html-to-markdown/v2"
	readability "github.com/go-shiori/go-readability"
	"github.com/gocolly/colly/v2"
	"github.com/rs/zerolog/log"

	"github.com/dantezy/cold-send0r-bot/internal/config"
	"github.com/dantezy/cold-send0r-bot/internal/models"
)

type CollyRodScraper struct {
	cfg         config.ScraperConfig
	rateLimiter *time.Ticker
}

func NewCollyRodScraper(cfg config.ScraperConfig) *CollyRodScraper {
	return &CollyRodScraper{
		cfg:         cfg,
		rateLimiter: time.NewTicker(time.Duration(cfg.RateLimitMs) * time.Millisecond),
	}
}

func (s *CollyRodScraper) Scrape(url string) (*models.ScrapeResult, error) {
	<-s.rateLimiter.C

	result := &models.ScrapeResult{URL: url}

	md, err := s.scrapeWithColly(url)
	if err != nil {
		log.Warn().Str("url", url).Err(err).Msg("colly scrape failed")
	}

	if len(md) < 200 && s.cfg.RodFallback {
		log.Info().Str("url", url).Int("colly_len", len(md)).Msg("thin content, falling back to rod")
		rodMd, rodErr := s.scrapeWithRod(url)
		if rodErr != nil {
			log.Warn().Str("url", url).Err(rodErr).Msg("rod scrape failed")
		}
		if len(rodMd) > len(md) {
			md = rodMd
		}
	}

	if md == "" {
		result.Error = "no content extracted"
		return result, nil
	}

	result.Markdown = truncate(md, s.cfg.MaxContentLength)
	return result, nil
}

func (s *CollyRodScraper) scrapeWithColly(url string) (string, error) {
	var htmlContent string

	c := colly.NewCollector(
		colly.AllowURLRevisit(),
	)
	c.SetRequestTimeout(time.Duration(s.cfg.TimeoutMs) * time.Millisecond)

	c.OnResponse(func(r *colly.Response) {
		htmlContent = string(r.Body)
	})

	if err := c.Visit(url); err != nil {
		return "", fmt.Errorf("visiting %s: %w", url, err)
	}

	if htmlContent == "" {
		return "", fmt.Errorf("empty response from %s", url)
	}

	return htmlToMarkdown(htmlContent, url)
}

func htmlToMarkdown(html, sourceURL string) (string, error) {
	parsedURL, _ := url.Parse(sourceURL)
	article, err := readability.FromReader(strings.NewReader(html), parsedURL)
	if err != nil {
		log.Warn().Err(err).Msg("readability extraction failed, using raw html")
	} else if article.Content != "" {
		html = article.Content
	}

	md, err := htmltomarkdown.ConvertString(html)
	if err != nil {
		return "", fmt.Errorf("converting html to markdown: %w", err)
	}

	return strings.TrimSpace(md), nil
}

func truncate(s string, maxLen int) string {
	if maxLen <= 0 || len(s) <= maxLen {
		return s
	}
	return s[:maxLen]
}

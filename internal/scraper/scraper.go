package scraper

import (
	"github.com/dantezy/cold-send0r-bot/internal/config"
	"github.com/dantezy/cold-send0r-bot/internal/models"
)

type Scraper interface {
	Scrape(url string) (*models.ScrapeResult, error)
}

func NewScraper(cfg config.ScraperConfig) Scraper {
	if cfg.Provider == "firecrawl" {
		return NewFirecrawlScraper(cfg)
	}
	return NewCollyRodScraper(cfg)
}

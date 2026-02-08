package scraper

import (
	"fmt"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/proto"
	"github.com/rs/zerolog/log"
)

func (s *CollyRodScraper) scrapeWithRod(url string) (string, error) {
	path, hasChrome := launcher.LookPath()
	if !hasChrome {
		log.Warn().Msg("chrome not found, skipping rod fallback")
		return "", fmt.Errorf("chrome not found at any standard path")
	}

	timeout := time.Duration(s.cfg.TimeoutMs) * time.Millisecond

	u := launcher.New().Bin(path).Headless(true).MustLaunch()
	browser := rod.New().ControlURL(u).Timeout(timeout).MustConnect()
	defer browser.MustClose()

	page, err := browser.Page(proto.TargetCreateTarget{URL: url})
	if err != nil {
		return "", fmt.Errorf("navigating to %s: %w", url, err)
	}

	err = page.Timeout(timeout / 3).WaitStable(timeout / 3)
	if err != nil {
		log.Warn().Str("url", url).Err(err).Msg("page did not stabilize, proceeding anyway")
	}

	html, err := page.Timeout(timeout).HTML()
	if err != nil {
		return "", fmt.Errorf("getting page HTML: %w", err)
	}

	md, err := htmlToMarkdown(html, url)
	if err != nil {
		return "", fmt.Errorf("converting rod html to markdown: %w", err)
	}

	return md, nil
}

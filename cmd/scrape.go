package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"github.com/dantezy/cold-send0r-bot/internal/contacts"
	"github.com/dantezy/cold-send0r-bot/internal/models"
	"github.com/dantezy/cold-send0r-bot/internal/scraper"
)

var scrapeOutput string

var scrapeCmd = &cobra.Command{
	Use:   "scrape",
	Short: "Scrape company websites from contacts list",
	RunE: func(cmd *cobra.Command, args []string) error {
		contactList, err := contacts.Load(cfg.Contacts.Path)
		if err != nil {
			return err
		}

		s := scraper.NewScraper(cfg.Scraper)
		var results []models.ScrapeResult

		for i, c := range contactList {
			log.Info().Int("index", i+1).Int("total", len(contactList)).Str("url", c.URL).Msg("scraping")
			result, err := s.Scrape(c.URL)
			if err != nil {
				log.Error().Str("url", c.URL).Err(err).Msg("scrape failed")
				results = append(results, models.ScrapeResult{URL: c.URL, Error: err.Error()})
				continue
			}
			results = append(results, *result)
			log.Info().Str("url", c.URL).Int("length", len(result.Markdown)).Msg("scraped")
		}

		data, err := json.MarshalIndent(results, "", "  ")
		if err != nil {
			return fmt.Errorf("marshaling results: %w", err)
		}

		if err := os.MkdirAll(filepath.Dir(scrapeOutput), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(scrapeOutput, data, 0o644); err != nil {
			return fmt.Errorf("writing scrape results: %w", err)
		}

		log.Info().Str("path", scrapeOutput).Int("count", len(results)).Msg("scrape results written")
		return nil
	},
}

func init() {
	scrapeCmd.Flags().StringVarP(&scrapeOutput, "output", "o", "output/scrape_results.json", "output path for scrape results")
	rootCmd.AddCommand(scrapeCmd)
}

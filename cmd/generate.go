package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"github.com/dantezy/cold-send0r-bot/internal/contacts"
	"github.com/dantezy/cold-send0r-bot/internal/generator"
	"github.com/dantezy/cold-send0r-bot/internal/models"
	"github.com/dantezy/cold-send0r-bot/internal/output"
	"github.com/dantezy/cold-send0r-bot/internal/resume"
)

var (
	generateScrapeInput string
)

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate personalized emails from scraped data",
	RunE: func(cmd *cobra.Command, args []string) error {
		contactList, err := contacts.Load(cfg.Contacts.Path)
		if err != nil {
			return err
		}

		resumeText, err := resume.ReadText(cfg.Resume.TextPath)
		if err != nil {
			log.Warn().Err(err).Msg("could not read resume, proceeding without it")
			resumeText = ""
		}

		scrapeMap := make(map[string]*models.ScrapeResult)
		if generateScrapeInput != "" {
			data, err := os.ReadFile(generateScrapeInput)
			if err != nil {
				return fmt.Errorf("reading scrape results: %w", err)
			}
			var results []models.ScrapeResult
			if err := json.Unmarshal(data, &results); err != nil {
				return fmt.Errorf("parsing scrape results: %w", err)
			}
			for i := range results {
				scrapeMap[results[i].URL] = &results[i]
			}
		}

		gen := generator.NewGenerator(cfg.LLM, cfg.Sender.Name)
		var emails []models.Email

		for i, c := range contactList {
			log.Info().Int("index", i+1).Int("total", len(contactList)).Str("contact", c.Name).Str("company", c.Company).Msg("generating email")

			scrapeResult := scrapeMap[c.URL]
			email, err := gen.Generate(c, scrapeResult, resumeText, cfg.Sender.Links)
			if err != nil {
				log.Error().Str("contact", c.Name).Err(err).Msg("generation failed")
				continue
			}
			emails = append(emails, *email)
			log.Info().Str("contact", c.Name).Str("subject", email.Subject).Msg("email generated")
		}

		if err := output.WriteEmails(cfg.Output.Path, emails); err != nil {
			return err
		}

		log.Info().Str("path", cfg.Output.Path).Int("count", len(emails)).Msg("emails written")
		return nil
	},
}

func init() {
	generateCmd.Flags().StringVar(&generateScrapeInput, "scrape-input", "", "path to pre-scraped results JSON")
	rootCmd.AddCommand(generateCmd)
}

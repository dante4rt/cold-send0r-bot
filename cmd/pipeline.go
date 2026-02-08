package cmd

import (
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"github.com/dantezy/cold-send0r-bot/internal/contacts"
	"github.com/dantezy/cold-send0r-bot/internal/generator"
	"github.com/dantezy/cold-send0r-bot/internal/models"
	"github.com/dantezy/cold-send0r-bot/internal/output"
	"github.com/dantezy/cold-send0r-bot/internal/resume"
	"github.com/dantezy/cold-send0r-bot/internal/scraper"
	"github.com/dantezy/cold-send0r-bot/internal/sender"
)

var (
	pipelineDryRun bool
	pipelineOutput string
)

var pipelineCmd = &cobra.Command{
	Use:   "pipeline",
	Short: "Full pipeline: scrape -> generate -> optionally send",
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

		// Scrape
		s := scraper.NewScraper(cfg.Scraper)
		scrapeResults := make(map[string]*models.ScrapeResult)

		log.Info().Int("count", len(contactList)).Msg("scraping company websites")
		for i, c := range contactList {
			log.Info().Int("index", i+1).Int("total", len(contactList)).Str("url", c.URL).Msg("scraping")
			result, err := s.Scrape(c.URL)
			if err != nil {
				log.Error().Str("url", c.URL).Err(err).Msg("scrape failed")
				continue
			}
			scrapeResults[c.URL] = result
			if result.Error != "" {
				log.Warn().Str("url", c.URL).Str("error", result.Error).Msg("scrape had issues")
			} else {
				log.Info().Str("url", c.URL).Int("length", len(result.Markdown)).Msg("scraped")
			}
		}

		// Generate
		gen := generator.NewGenerator(cfg.LLM, cfg.Sender.Name)
		var emails []models.Email

		log.Info().Int("count", len(contactList)).Str("model", cfg.LLM.Model).Msg("generating personalized emails")
		for i, c := range contactList {
			log.Info().Int("index", i+1).Int("total", len(contactList)).Str("contact", c.Name).Str("company", c.Company).Msg("generating email")
			email, err := gen.Generate(c, scrapeResults[c.URL], resumeText, cfg.Sender.Links)
			if err != nil {
				log.Error().Str("contact", c.Name).Err(err).Msg("generation failed")
				continue
			}
			emails = append(emails, *email)
			log.Info().Str("contact", c.Name).Str("subject", email.Subject).Msg("email generated")
		}

		outPath := pipelineOutput
		if outPath == "" {
			outPath = cfg.Output.Path
		}

		if err := output.WriteEmails(outPath, emails); err != nil {
			return err
		}
		log.Info().Str("path", outPath).Int("count", len(emails)).Msg("emails written")

		// Send (unless dry run)
		if pipelineDryRun {
			log.Info().Str("path", outPath).Msg("DRY RUN -- review output, then run: send0r send --input <path> --confirm")
			return nil
		}

		smtpSender := sender.NewSMTPSender(cfg.SMTP, cfg.Sender.Email, cfg.Sender.Name, cfg.Resume.Attachments)
		var sent, failed int
		for i := range emails {
			log.Info().Int("index", i+1).Int("total", len(emails)).Str("to", emails[i].Contact.Email).Msg("sending")
			if err := smtpSender.Send(&emails[i]); err != nil {
				failed++
				continue
			}
			sent++
		}

		if err := output.WriteEmails(outPath, emails); err != nil {
			log.Error().Err(err).Msg("failed to update email statuses")
		}

		log.Info().Int("sent", sent).Int("failed", failed).Msg("pipeline complete")
		return nil
	},
}

func init() {
	pipelineCmd.Flags().BoolVar(&pipelineDryRun, "dry-run", true, "generate emails without sending (default: true)")
	pipelineCmd.Flags().StringVarP(&pipelineOutput, "output", "o", "", "output path (default: from config)")
	rootCmd.AddCommand(pipelineCmd)
}

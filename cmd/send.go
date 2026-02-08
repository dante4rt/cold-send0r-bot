package cmd

import (
	"fmt"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"github.com/dantezy/cold-send0r-bot/internal/output"
	"github.com/dantezy/cold-send0r-bot/internal/sender"
)

var (
	sendInput   string
	sendConfirm bool
)

var sendCmd = &cobra.Command{
	Use:   "send",
	Short: "Send emails from a generated emails JSON file",
	RunE: func(cmd *cobra.Command, args []string) error {
		if !sendConfirm {
			return fmt.Errorf("you must pass --confirm to actually send emails. Review %s first", sendInput)
		}

		emails, err := output.ReadEmails(sendInput)
		if err != nil {
			return err
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

		if err := output.WriteEmails(sendInput, emails); err != nil {
			log.Error().Err(err).Msg("failed to update email statuses")
		}

		log.Info().Int("sent", sent).Int("failed", failed).Msg("send complete")
		return nil
	},
}

func init() {
	sendCmd.Flags().StringVar(&sendInput, "input", "output/emails.json", "path to emails JSON file")
	sendCmd.Flags().BoolVar(&sendConfirm, "confirm", false, "confirm sending (required)")
	rootCmd.AddCommand(sendCmd)
}

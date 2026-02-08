package cmd

import (
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"github.com/dantezy/cold-send0r-bot/internal/config"
)

var (
	cfgFile string
	verbose bool
	cfg     *config.Config
)

var rootCmd = &cobra.Command{
	Use:   "send0r",
	Short: "Cold email outreach bot with LLM personalization",
	Long:  "Automates personalized cold outreach: scrape company websites, generate emails via LLM, send via SMTP.",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
		if verbose {
			zerolog.SetGlobalLevel(zerolog.DebugLevel)
		}
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.Kitchen})

		// Skip config loading for commands that don't need it
		if cmd.Name() == "init" || cmd.Name() == "otter" {
			return nil
		}

		var err error
		cfg, err = config.Load(cfgFile)
		if err != nil {
			return err
		}
		return nil
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default: ./config.yaml)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "enable debug logging")
}

package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize config files from examples",
	Long:  "Copies example files (config, resume, contacts, env) and creates the attachments directory.",
	RunE: func(cmd *cobra.Command, args []string) error {
		copies := []struct {
			src  string
			dest string
		}{
			{"config.example.yaml", "config.yaml"},
			{"resume.example.txt", "resume.txt"},
			{"contacts.example.json", "contacts.json"},
			{".env.example", ".env"},
		}

		for _, c := range copies {
			if _, err := os.Stat(c.dest); err == nil {
				fmt.Printf("  skip     %s (already exists)\n", c.dest)
				continue
			}

			data, err := os.ReadFile(c.src)
			if err != nil {
				return fmt.Errorf("reading %s: %w", c.src, err)
			}

			if err := os.WriteFile(c.dest, data, 0644); err != nil {
				return fmt.Errorf("writing %s: %w", c.dest, err)
			}
			fmt.Printf("  created  %s\n", c.dest)
		}

		// Create attachments directory
		if _, err := os.Stat("attachments"); err == nil {
			fmt.Println("  skip     attachments/ (already exists)")
		} else {
			if err := os.MkdirAll("attachments", 0755); err != nil {
				return fmt.Errorf("creating attachments dir: %w", err)
			}
			fmt.Println("  created  attachments/")
		}

		fmt.Println()
		fmt.Println("Setup complete! Next steps:")
		fmt.Println("  1. Edit config.yaml      — set your name, email, and links")
		fmt.Println("  2. Edit resume.txt        — paste your resume summary")
		fmt.Println("  3. Edit contacts.json     — add your target contacts")
		fmt.Println("  4. Edit .env              — add OPENROUTER_API_KEY, SMTP_USERNAME, SMTP_PASSWORD")
		fmt.Println("  5. Drop your CV PDF into  attachments/ and update the path in config.yaml")
		fmt.Println("  6. Run: send0r pipeline --dry-run")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}

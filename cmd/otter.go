package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/dantezy/cold-send0r-bot/internal/models"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

const otterSupabaseURL = "https://uitzqqugzhhvgvrgxjaw.supabase.co/rest/v1/startups?select=*%2Cstartup_employees%28*%29%2Cstartup_tags%28tag%29&order=created_at.desc"

type otterStartup struct {
	Name             string          `json:"name"`
	Website          *string         `json:"website"`
	Sector           string          `json:"sector"`
	StartupEmployees []otterEmployee `json:"startup_employees"`
	StartupTags      []otterTag      `json:"startup_tags"`
}

type otterEmployee struct {
	Name  string `json:"name"`
	Role  string `json:"role"`
	Email string `json:"email"`
}

type otterTag struct {
	Tag string `json:"tag"`
}

var (
	otterOutput string
	otterApiKey string
)

var otterCmd = &cobra.Command{
	Use:   "otter",
	Short: "Import contacts from useotter.app",
	Long:  "Fetches startup data from Otter and converts to contacts.json format.",
	RunE: func(cmd *cobra.Command, args []string) error {
		if otterApiKey == "" {
			otterApiKey = os.Getenv("OTTER_API_KEY")
		}
		if otterApiKey == "" {
			return fmt.Errorf("apikey required: use --apikey or set OTTER_API_KEY env var\n  (copy from browser DevTools > Network > apikey header)")
		}

		client := &http.Client{Timeout: 30 * time.Second}
		req, err := http.NewRequest(http.MethodGet, otterSupabaseURL, nil)
		if err != nil {
			return fmt.Errorf("creating request: %w", err)
		}
		req.Header.Set("apikey", otterApiKey)
		req.Header.Set("Authorization", "Bearer "+otterApiKey)
		req.Header.Set("Accept-Profile", "public")

		log.Info().Msg("fetching startups from otter")
		resp, err := client.Do(req)
		if err != nil {
			return fmt.Errorf("fetching otter data: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			return fmt.Errorf("otter API returned %d: %s", resp.StatusCode, string(body))
		}

		var startups []otterStartup
		if err := json.NewDecoder(resp.Body).Decode(&startups); err != nil {
			return fmt.Errorf("parsing response: %w", err)
		}

		var contacts []models.Contact
		var skipped int
		for _, s := range startups {
			url := ""
			if s.Website != nil {
				url = *s.Website
			}
			for _, emp := range s.StartupEmployees {
				if emp.Email == "" {
					skipped++
					continue
				}
				contacts = append(contacts, models.Contact{
					Email:   emp.Email,
					Name:    emp.Name,
					Company: s.Name,
					Role:    emp.Role,
					URL:     url,
				})
			}
		}

		data, err := json.MarshalIndent(contacts, "", "  ")
		if err != nil {
			return fmt.Errorf("marshaling contacts: %w", err)
		}

		if err := os.WriteFile(otterOutput, data, 0644); err != nil {
			return fmt.Errorf("writing %s: %w", otterOutput, err)
		}

		log.Info().
			Int("startups", len(startups)).
			Int("contacts", len(contacts)).
			Int("skipped_no_email", skipped).
			Str("output", otterOutput).
			Msg("import complete")

		return nil
	},
}

func init() {
	otterCmd.Flags().StringVarP(&otterOutput, "output", "o", "contacts.json", "output file path")
	otterCmd.Flags().StringVar(&otterApiKey, "apikey", "", "otter supabase apikey (or set OTTER_API_KEY env)")
	rootCmd.AddCommand(otterCmd)
}

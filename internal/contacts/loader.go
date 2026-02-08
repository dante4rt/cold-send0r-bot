package contacts

import (
	"encoding/json"
	"fmt"
	"net/mail"
	"os"
	"strings"

	"github.com/dantezy/cold-send0r-bot/internal/models"
	"github.com/rs/zerolog/log"
)

func Load(path string) ([]models.Contact, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading contacts file: %w", err)
	}

	var raw []models.Contact
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("parsing contacts JSON: %w", err)
	}

	var valid []models.Contact
	var skipped int

	for i, c := range raw {
		if err := validateContact(c); err != nil {
			log.Warn().Int("index", i).Str("email", c.Email).Err(err).Msg("skipping invalid contact")
			skipped++
			continue
		}
		valid = append(valid, c)
	}

	log.Info().Int("valid", len(valid)).Int("skipped", skipped).Msg("contacts loaded")
	return valid, nil
}

func validateContact(c models.Contact) error {
	if _, err := mail.ParseAddress(c.Email); err != nil {
		return fmt.Errorf("invalid email %q: %w", c.Email, err)
	}
	if strings.TrimSpace(c.Name) == "" {
		return fmt.Errorf("name is empty")
	}
	if strings.TrimSpace(c.Company) == "" {
		return fmt.Errorf("company is empty")
	}
	if strings.TrimSpace(c.URL) == "" {
		return fmt.Errorf("url is empty")
	}
	return nil
}

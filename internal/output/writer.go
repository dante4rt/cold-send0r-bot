package output

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/dantezy/cold-send0r-bot/internal/models"
)

func WriteEmails(path string, emails []models.Email) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("creating output directory: %w", err)
	}

	data, err := json.MarshalIndent(emails, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling emails: %w", err)
	}

	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("writing output file: %w", err)
	}

	return nil
}

func ReadEmails(path string) ([]models.Email, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading emails file: %w", err)
	}

	var emails []models.Email
	if err := json.Unmarshal(data, &emails); err != nil {
		return nil, fmt.Errorf("parsing emails JSON: %w", err)
	}

	return emails, nil
}

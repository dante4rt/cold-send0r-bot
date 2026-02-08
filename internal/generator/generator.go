package generator

import (
	"github.com/dantezy/cold-send0r-bot/internal/config"
	"github.com/dantezy/cold-send0r-bot/internal/models"
)

type Generator interface {
	Generate(contact models.Contact, scrapeResult *models.ScrapeResult, resumeText string, links map[string]string) (*models.Email, error)
}

func NewGenerator(cfg config.LLMConfig, senderName string) Generator {
	return NewOpenRouterGenerator(cfg, senderName)
}

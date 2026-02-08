package generator

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/dantezy/cold-send0r-bot/internal/config"
	"github.com/dantezy/cold-send0r-bot/internal/models"
)

type OpenRouterGenerator struct {
	cfg         config.LLMConfig
	senderName  string
	client      *http.Client
	rateLimiter *time.Ticker
}

func NewOpenRouterGenerator(cfg config.LLMConfig, senderName string) *OpenRouterGenerator {
	return &OpenRouterGenerator{
		cfg:         cfg,
		senderName:  senderName,
		client:      &http.Client{Timeout: 60 * time.Second},
		rateLimiter: time.NewTicker(time.Duration(cfg.RateLimitMs) * time.Millisecond),
	}
}

type openRouterRequest struct {
	Model       string    `json:"model"`
	Messages    []message `json:"messages"`
	Temperature float64   `json:"temperature"`
	MaxTokens   int       `json:"max_tokens"`
}

type message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type openRouterResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
}

func (g *OpenRouterGenerator) Generate(contact models.Contact, scrapeResult *models.ScrapeResult, resumeText string, links map[string]string) (*models.Email, error) {
	<-g.rateLimiter.C

	scrapeMarkdown := ""
	if scrapeResult != nil {
		scrapeMarkdown = scrapeResult.Markdown
	}

	prompt := BuildPrompt(contact, scrapeMarkdown, resumeText, g.senderName, links)

	reqBody := openRouterRequest{
		Model: g.cfg.Model,
		Messages: []message{
			{Role: "user", Content: prompt},
		},
		Temperature: g.cfg.Temperature,
		MaxTokens:   g.cfg.MaxTokens,
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshaling request: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, "https://openrouter.ai/api/v1/chat/completions", strings.NewReader(string(bodyBytes)))
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+g.cfg.APIKey)

	resp, err := g.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("calling openrouter: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	var orResp openRouterResponse
	if err := json.Unmarshal(respBody, &orResp); err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}

	if orResp.Error != nil {
		return nil, fmt.Errorf("openrouter error: %s", orResp.Error.Message)
	}

	if len(orResp.Choices) == 0 {
		return nil, fmt.Errorf("no choices in response")
	}

	content := orResp.Choices[0].Message.Content
	subject, body, err := parseEmailResponse(content)
	if err != nil {
		return nil, fmt.Errorf("parsing LLM output: %w", err)
	}

	return &models.Email{
		Contact:     contact,
		Subject:     subject,
		Body:        body,
		Status:      "draft",
		GeneratedAt: time.Now(),
	}, nil
}

func parseEmailResponse(content string) (subject, body string, err error) {
	lines := strings.Split(content, "\n")

	subjectIdx := -1
	bodyIdx := -1

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		upper := strings.ToUpper(trimmed)
		if strings.HasPrefix(upper, "SUBJECT:") {
			subjectIdx = i
			subject = strings.TrimSpace(trimmed[len("SUBJECT:"):])
		}
		if strings.HasPrefix(upper, "BODY:") {
			bodyIdx = i
			// Some LLMs put body content on the same line as BODY:
			inline := strings.TrimSpace(trimmed[len("BODY:"):])
			if inline != "" {
				bodyIdx = -2 // signal: body starts on this line
				body = inline + "\n"
			}
		}
	}

	if subjectIdx == -1 {
		return "", "", fmt.Errorf("could not find SUBJECT: in response:\n%s", content)
	}

	switch {
	case bodyIdx == -2:
		// body was already captured inline, but also grab remaining lines
		foundBody := false
		for _, line := range lines {
			upper := strings.ToUpper(strings.TrimSpace(line))
			if strings.HasPrefix(upper, "BODY:") {
				foundBody = true
				continue
			}
			if foundBody {
				body += line + "\n"
			}
		}
		body = strings.TrimSpace(body)
	case bodyIdx >= 0:
		bodyLines := lines[bodyIdx+1:]
		body = strings.TrimSpace(strings.Join(bodyLines, "\n"))
	default:
		// No BODY: marker — treat everything after SUBJECT line as body
		bodyLines := lines[subjectIdx+1:]
		body = strings.TrimSpace(strings.Join(bodyLines, "\n"))
	}

	if body == "" {
		return "", "", fmt.Errorf("empty body in response:\n%s", content)
	}

	return subject, body, nil
}

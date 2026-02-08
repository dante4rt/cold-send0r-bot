package models

import "time"

type Contact struct {
	Email   string `json:"email"`
	Name    string `json:"name"`
	Company string `json:"company"`
	Role    string `json:"role"`
	URL     string `json:"url"`
}

type ScrapeResult struct {
	URL      string `json:"url"`
	Markdown string `json:"markdown"`
	Error    string `json:"error,omitempty"`
}

type Email struct {
	Contact     Contact   `json:"contact"`
	Subject     string    `json:"subject"`
	Body        string    `json:"body"`
	Status      string    `json:"status"`
	GeneratedAt time.Time `json:"generated_at"`
}

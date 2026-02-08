package generator

import (
	"fmt"
	"strings"

	"github.com/dantezy/cold-send0r-bot/internal/models"
)

func BuildPrompt(contact models.Contact, scrapeMarkdown, resumeText, senderName string, links map[string]string) string {
	firstName := strings.Split(contact.Name, " ")[0]

	companyContext := scrapeMarkdown
	if companyContext == "" {
		companyContext = fmt.Sprintf("(No website content available. Use the company name '%s' and role '%s' as context.)", contact.Company, contact.Role)
	}

	// Decide greeting style: use first name if we have a person, team name if generic
	greeting := fmt.Sprintf("Hi %s,", firstName)
	if contact.Role == "" || strings.Contains(strings.ToLower(contact.Name), "team") {
		greeting = fmt.Sprintf("Dear %s Hiring Team,", contact.Company)
	}

	// Build links section dynamically from config
	var linksSection string
	if len(links) > 0 {
		var linkLines []string
		for label, url := range links {
			linkLines = append(linkLines, fmt.Sprintf("- %s: %s", label, url))
		}
		linksSection = strings.Join(linkLines, "\n")
	} else {
		linksSection = "(No links provided)"
	}

	// Build a list of link labels for the prompt instruction
	var linkLabels []string
	for label := range links {
		linkLabels = append(linkLabels, label)
	}
	linkMention := strings.Join(linkLabels, "/")
	if linkMention == "" {
		linkMention = "relevant"
	}

	return fmt.Sprintf(`Write a cold outreach email from %s to %s (%s) at %s.

Company website content:
%s

Sender's background:
%s

Sender's links:
%s

Style rules (FOLLOW STRICTLY — match this exact tone and structure):
- Professional but warm, NOT overly casual
- Opening: "%s" (already provided, use as-is)
- Body: 2-3 short paragraphs, max 5 sentences total
- First paragraph: State interest in contributing to the company, reference something SPECIFIC from their website that caught your attention
- Second paragraph: Briefly connect sender's relevant experience to what the company does
- Third paragraph (short): Mention CV is attached, include the sender's %s links, and say you're available to discuss further
- Sign off with exactly: "Regards,\n%s"
- Do NOT use "hey" — keep it professional
- Do NOT use "I hope this finds you well" or generic corporate openers
- Do NOT use markdown formatting like ** or ## in the email body
- Keep it concise and direct, like a real person writing a real email

Subject line rules:
- Format: "<Role/Position> Application – %s"
- Examples: "Senior Backend Developer Application – %s", "Full Stack Developer Application – %s"
- Infer an appropriate role/position from the company website content and sender's background
- Always end with " – %s" (en-dash, then sender full name)

Output exactly:
SUBJECT: <subject line>
BODY:
<email body>`,
		senderName, contact.Name, contact.Role, contact.Company,
		companyContext,
		resumeText,
		linksSection,
		greeting,
		linkMention,
		senderName,
		senderName, senderName, senderName,
	)
}

package sender

import (
	"crypto/rand"
	"fmt"
	"strings"
	"time"

	"github.com/dantezy/cold-send0r-bot/internal/config"
	"github.com/dantezy/cold-send0r-bot/internal/models"
	"github.com/rs/zerolog/log"
	"gopkg.in/gomail.v2"
)

type SMTPSender struct {
	cfg         config.SMTPConfig
	senderEmail string
	senderName  string
	attachments []string
	rateLimiter *time.Ticker
}

func NewSMTPSender(smtpCfg config.SMTPConfig, senderEmail, senderName string, attachments []string) *SMTPSender {
	return &SMTPSender{
		cfg:         smtpCfg,
		senderEmail: senderEmail,
		senderName:  senderName,
		attachments: attachments,
		rateLimiter: time.NewTicker(time.Duration(smtpCfg.RateLimitMs) * time.Millisecond),
	}
}

func (s *SMTPSender) Send(email *models.Email) error {
	<-s.rateLimiter.C

	if s.cfg.Username == "" || s.cfg.Password == "" {
		return fmt.Errorf("SMTP credentials not configured (set %s and %s env vars)", s.cfg.UsernameEnv, s.cfg.PasswordEnv)
	}

	m := gomail.NewMessage()
	m.SetAddressHeader("From", s.senderEmail, s.senderName)
	m.SetHeader("To", email.Contact.Email)
	m.SetHeader("Subject", email.Subject)
	m.SetHeader("Message-ID", generateMessageID(s.senderEmail))
	m.SetBody("text/plain", email.Body)

	for _, attachment := range s.attachments {
		m.Attach(attachment)
	}

	d := gomail.NewDialer(s.cfg.Host, s.cfg.Port, s.cfg.Username, s.cfg.Password)

	if err := d.DialAndSend(m); err != nil {
		email.Status = "failed"
		log.Error().
			Str("to", email.Contact.Email).
			Err(err).
			Msg("failed to send email")
		return fmt.Errorf("sending email to %s: %w", email.Contact.Email, err)
	}

	email.Status = "sent"
	log.Info().
		Str("to", email.Contact.Email).
		Str("subject", email.Subject).
		Msg("email sent")

	return nil
}

func generateMessageID(senderEmail string) string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	domain := "localhost"
	if parts := strings.SplitN(senderEmail, "@", 2); len(parts) == 2 {
		domain = parts[1]
	}
	return fmt.Sprintf("<%x.%d@%s>", b, time.Now().UnixNano(), domain)
}

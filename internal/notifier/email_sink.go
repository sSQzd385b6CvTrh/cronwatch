package notifier

import (
	"fmt"
	"net/smtp"
	"strings"
	"time"

	"github.com/cronwatch/cronwatch/internal/tracker"
)

// EmailConfig holds configuration for the SMTP email sink.
type EmailConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
	To       []string
}

// emailSink delivers alert notifications via SMTP email.
type emailSink struct {
	cfg  EmailConfig
	auth smtp.Auth
}

// NewEmailSink creates a new email sink using the provided SMTP configuration.
func NewEmailSink(cfg EmailConfig) (*emailSink, error) {
	if cfg.Host == "" {
		return nil, fmt.Errorf("email sink: SMTP host is required")
	}
	if len(cfg.To) == 0 {
		return nil, fmt.Errorf("email sink: at least one recipient is required")
	}
	if cfg.Port == 0 {
		cfg.Port = 587
	}
	auth := smtp.PlainAuth("", cfg.Username, cfg.Password, cfg.Host)
	return &emailSink{cfg: cfg, auth: auth}, nil
}

// Send formats and delivers an alert email for the given alert.
func (e *emailSink) Send(a tracker.Alert) error {
	subject := fmt.Sprintf("[cronwatch] %s – %s", a.Kind, a.JobName)
	body := fmt.Sprintf(
		"Job:       %s\nKind:      %s\nMessage:   %s\nAt:        %s\n",
		a.JobName,
		a.Kind,
		a.Message,
		a.At.Format(time.RFC3339),
	)
	msg := []byte(fmt.Sprintf(
		"From: %s\r\nTo: %s\r\nSubject: %s\r\n\r\n%s",
		e.cfg.From,
		strings.Join(e.cfg.To, ", "),
		subject,
		body,
	))
	addr := fmt.Sprintf("%s:%d", e.cfg.Host, e.cfg.Port)
	if err := smtp.SendMail(addr, e.auth, e.cfg.From, e.cfg.To, msg); err != nil {
		return fmt.Errorf("email sink: send failed: %w", err)
	}
	return nil
}

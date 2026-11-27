package notificator

import (
	"errors"
	"net/smtp"
	"sync"
)

// SMTP is the outgoing mail configuration edited on the Administration page.
type SMTP struct {
	Host     string
	Port     string
	User     string
	Password string
	From     string
}

var mailState = struct {
	sync.RWMutex
	cfg SMTP
}{}

// Configure replaces the mail settings used by SendEmail.
func Configure(cfg SMTP) {
	if cfg.Port == "" {
		cfg.Port = "587"
	}
	mailState.Lock()
	mailState.cfg = cfg
	mailState.Unlock()
}

// Configured reports whether a host has been saved.
func Configured() bool {
	mailState.RLock()
	defer mailState.RUnlock()
	return mailState.cfg.Host != ""
}

type Email struct {
	To      string
	Subject string
	Body    string
}

// SendEmail delivers one message with the Administration SMTP settings.
func SendEmail(email Email) error {
	mailState.RLock()
	cfg := mailState.cfg
	mailState.RUnlock()
	if cfg.Host == "" {
		return errors.New("smtp not configured")
	}
	if cfg.Port == "" {
		cfg.Port = "587"
	}
	from := cfg.From
	if from == "" {
		from = cfg.User
	}
	msg := "From: " + from + "\n" +
		"To: " + email.To + "\n" +
		"Subject: " + email.Subject + "\n\n" +
		email.Body
	var auth smtp.Auth
	if cfg.User != "" {
		auth = smtp.PlainAuth("", cfg.User, cfg.Password, cfg.Host)
	}
	return smtp.SendMail(cfg.Host+":"+cfg.Port, auth, from, []string{email.To}, []byte(msg))
}

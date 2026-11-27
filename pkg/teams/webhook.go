package teams

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// ValidateIncomingWebhook allows only Microsoft Teams connector URLs.
func ValidateIncomingWebhook(raw string) error {
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	parsed, err := url.Parse(raw)
	if err != nil || parsed.Host == "" {
		return errors.New("invalid webhook URL")
	}
	if parsed.Scheme != "https" {
		return errors.New("webhook URL must be https")
	}
	host := strings.ToLower(parsed.Hostname())
	if ip := net.ParseIP(host); ip != nil {
		return errors.New("webhook host must be a Teams domain")
	}
	if host != "outlook.office.com" && host != "webhook.office.com" && !strings.HasSuffix(host, ".webhook.office.com") {
		return errors.New("webhook host must be outlook.office.com or webhook.office.com")
	}
	return nil
}

// Post sends a MessageCard. Failures are returned to the caller; CI paths should ignore them.
func Post(parent context.Context, webhookURL, title, text string) error {
	if err := ValidateIncomingWebhook(webhookURL); err != nil {
		return err
	}
	if webhookURL == "" {
		return nil
	}
	body, _ := json.Marshal(map[string]string{
		"@type":      "MessageCard",
		"@context":   "https://schema.org/extensions",
		"summary":    title,
		"themeColor": "5b21b6",
		"title":      title,
		"text":       text,
	})
	ctx, cancel := context.WithTimeout(parent, 5*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, webhookURL, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return errors.New("teams webhook returned " + resp.Status)
	}
	return nil
}

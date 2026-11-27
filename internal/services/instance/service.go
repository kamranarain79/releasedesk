package instance

import (
	"crypto/subtle"
	"log"
	"os"

	"github.com/brewbits-co/releasedesk/internal/domains/instance"
	"github.com/brewbits-co/releasedesk/pkg/notificator"
	"github.com/brewbits-co/releasedesk/pkg/tokenhash"
)

type Service interface {
	Current() (instance.Instance, error)
	Matches(raw string) bool
	RegenerateToken() (string, error)
	DismissToken() error
	SaveSMTP(host, port, user, password, from string, clearPassword bool) error
}

func NewService(repo instance.Repository) Service {
	svc := &service{repo: repo}
	svc.ensure()
	return svc
}

type service struct {
	repo instance.Repository
}

func (s *service) ensure() {
	row, err := s.repo.Get()
	if err != nil {
		log.Printf("instance settings: %v", err)
		return
	}
	changed := row.ID == 0
	if row.APITokenHash == "" {
		raw, prefix, hash, err := tokenhash.Mint()
		if err != nil {
			log.Printf("instance token: %v", err)
			return
		}
		row.APITokenHash = hash
		row.APITokenPrefix = prefix
		row.APITokenPending = raw
		changed = true
		log.Printf("generated instance API token %s; copy it from Administration", prefix)
	}
	if row.SMTPHost == "" {
		if host := os.Getenv("SMTP_HOST"); host != "" {
			row.SMTPHost = host
			row.SMTPPort = os.Getenv("SMTP_PORT")
			row.SMTPUser = os.Getenv("SMTP_USER")
			row.SMTPPassword = os.Getenv("SMTP_PASSWORD")
			row.SMTPFrom = os.Getenv("SMTP_FROM")
			changed = true
		}
	}
	if changed {
		row.ID = 1
		if err = s.repo.Save(&row); err != nil {
			log.Printf("instance settings save: %v", err)
		}
	}
	s.applyMail(row)
}

func (s *service) applyMail(row instance.Instance) {
	notificator.Configure(notificator.SMTP{
		Host:     row.SMTPHost,
		Port:     row.SMTPPort,
		User:     row.SMTPUser,
		Password: row.SMTPPassword,
		From:     row.SMTPFrom,
	})
}

func (s *service) Current() (instance.Instance, error) {
	return s.repo.Get()
}

func (s *service) Matches(raw string) bool {
	if raw == "" {
		return false
	}
	row, err := s.repo.Get()
	if err != nil || row.APITokenHash == "" {
		return false
	}
	got := tokenhash.Hash(raw)
	if len(got) != len(row.APITokenHash) {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(got), []byte(row.APITokenHash)) == 1
}

func (s *service) RegenerateToken() (string, error) {
	raw, prefix, hash, err := tokenhash.Mint()
	if err != nil {
		return "", err
	}
	row, err := s.repo.Get()
	if err != nil {
		return "", err
	}
	row.ID = 1
	row.APITokenHash = hash
	row.APITokenPrefix = prefix
	row.APITokenPending = raw
	if err = s.repo.Save(&row); err != nil {
		return "", err
	}
	log.Printf("regenerated instance API token %s", prefix)
	return raw, nil
}

func (s *service) DismissToken() error {
	row, err := s.repo.Get()
	if err != nil {
		return err
	}
	row.ID = 1
	row.APITokenPending = ""
	return s.repo.Save(&row)
}

func (s *service) SaveSMTP(host, port, user, password, from string, clearPassword bool) error {
	row, err := s.repo.Get()
	if err != nil {
		return err
	}
	row.ID = 1
	row.SMTPHost = host
	if port == "" {
		port = "587"
	}
	row.SMTPPort = port
	row.SMTPUser = user
	row.SMTPFrom = from
	if clearPassword {
		row.SMTPPassword = ""
	} else if password != "" {
		row.SMTPPassword = password
	}
	if err = s.repo.Save(&row); err != nil {
		return err
	}
	s.applyMail(row)
	return nil
}

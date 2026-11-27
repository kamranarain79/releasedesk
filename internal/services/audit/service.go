package audit

import (
	"strings"
	"time"

	"github.com/brewbits-co/releasedesk/internal/domains/audit"
	"github.com/brewbits-co/releasedesk/internal/domains/user"
)

type Service interface {
	Record(appID int, actor, action, subject, summary string)
	List(appID int, subject string) ([]audit.Event, error)
	Directory() user.Directory
}

func NewService(repo audit.Repository, users user.UserRepository) Service {
	return &service{repo: repo, users: users}
}

type service struct {
	repo  audit.Repository
	users user.UserRepository
}

func (s *service) Record(appID int, actor, action, subject, summary string) {
	actor = strings.TrimSpace(actor)
	if actor == "" {
		actor = "unknown"
	}
	_ = s.repo.Save(&audit.Event{
		AppID:     appID,
		Actor:     actor,
		Action:    action,
		Subject:   subject,
		Summary:   summary,
		CreatedAt: time.Now().UTC(),
	})
}

func (s *service) List(appID int, subject string) ([]audit.Event, error) {
	return s.repo.List(appID, subject)
}

func (s *service) Directory() user.Directory {
	accounts, err := s.users.Find()
	if err != nil {
		return user.Directory{}
	}
	return user.NewDirectory(accounts)
}

// Annotate fills each event with the person's name and avatar id.
func Annotate(events []audit.Event, directory user.Directory) []audit.Event {
	if events == nil {
		return []audit.Event{}
	}
	for i := range events {
		person := directory.Person(events[i].Actor)
		events[i].UserID = person.UserID
		events[i].Name = person.Name
	}
	return events
}

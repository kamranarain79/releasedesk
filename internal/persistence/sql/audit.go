package sql

import (
	"github.com/brewbits-co/releasedesk/internal/domains/audit"
	"github.com/brewbits-co/releasedesk/pkg/utils"
	"xorm.io/xorm"
)

func NewAuditRepository(engine *xorm.Engine) audit.Repository {
	return &auditRepository{engine: engine}
}

type auditRepository struct {
	engine *xorm.Engine
}

func (r *auditRepository) Save(event *audit.Event) error {
	_, err := r.engine.Insert(event)
	return err
}

func (r *auditRepository) List(appID int, subject string) ([]audit.Event, error) {
	var events []audit.Event
	err := r.engine.Where("app_id = ? AND subject = ?", appID, subject).
		Desc("created_at").Desc("id").
		Find(&events)
	if err != nil {
		return nil, err
	}
	for i := range events {
		events[i].CreatedAtFormatted = utils.FormatTime(events[i].CreatedAt)
	}
	return events, nil
}

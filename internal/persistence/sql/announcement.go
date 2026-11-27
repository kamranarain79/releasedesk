package sql

import (
	"github.com/brewbits-co/releasedesk/internal/domains/release"
	"xorm.io/xorm"
)

func NewAnnouncementRepository(engine *xorm.Engine) release.AnnouncementRepository {
	return &announcementRepository{engine: engine}
}

type announcementRepository struct {
	engine *xorm.Engine
}

func (r *announcementRepository) Save(item *release.Announcement) error {
	_, err := r.engine.Insert(item)
	return err
}

func (r *announcementRepository) ListByApp(appID int) ([]release.Announcement, error) {
	var items []release.Announcement
	err := r.engine.Where("app_id = ?", appID).Desc("id").Find(&items)
	return items, err
}

package release

import "time"

// Announcement is a message published for an app release.
type Announcement struct {
	ID        int       `xorm:"pk autoincr" json:"id"`
	AppID     int       `xorm:"not null index" json:"appId"`
	Version   string    `xorm:"not null" json:"version"`
	Message   string    `xorm:"text not null" json:"message"`
	CreatedAt time.Time `json:"createdAt"`
}

type AnnouncementRepository interface {
	Save(item *Announcement) error
	ListByApp(appID int) ([]Announcement, error)
}

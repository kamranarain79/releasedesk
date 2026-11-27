package audit

import "time"

// Event is one recorded action on an application.
type Event struct {
	ID                 int    `xorm:"pk autoincr"`
	AppID              int    `xorm:"not null index"`
	Actor              string `xorm:"varchar(100) not null"`
	Action             string `xorm:"varchar(40) not null"`
	Subject            string `xorm:"varchar(200) not null index"`
	Summary            string `xorm:"varchar(300) not null"`
	CreatedAt          time.Time
	CreatedAtFormatted string `xorm:"-"`
	// UserID and Name are filled in for display. They are not stored.
	UserID int    `xorm:"-"`
	Name   string `xorm:"-"`
}

func (Event) TableName() string { return "audit_event" }

func ReleaseSubject(version string) string {
	return "release:" + version
}

func BuildSubject(osName, number string) string {
	return "build:" + osName + ":" + number
}

type Repository interface {
	Save(event *Event) error
	List(appID int, subject string) ([]Event, error)
}

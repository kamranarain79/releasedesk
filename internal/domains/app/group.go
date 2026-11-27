package app

// DistributionGroup is an optional extra list of users for portal gating.
type DistributionGroup struct {
	ID             int    `xorm:"pk autoincr" json:"id"`
	AppID          int    `xorm:"not null index" json:"appId"`
	Name           string `xorm:"not null" json:"name"`
	RestrictPortal bool   `xorm:"not null default false" json:"restrictPortal"`
}

type GroupMember struct {
	GroupID int `xorm:"pk not null"`
	UserID  int `xorm:"pk not null"`
}

type GroupRepository interface {
	Save(group *DistributionGroup) error
	ListByApp(appID int) ([]DistributionGroup, error)
	AddMember(groupID, userID int) error
}

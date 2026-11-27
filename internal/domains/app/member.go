package app

// MemberRole is access to one application, like a Teams channel role.
type MemberRole string

const (
	MemberAdmin MemberRole = "admin"
	MemberUser  MemberRole = "user"
)

// Member links a console user to one application.
type Member struct {
	ID     int        `xorm:"pk autoincr" json:"id"`
	AppID  int        `xorm:"not null unique(member_uidx)" json:"appId"`
	UserID int        `xorm:"not null unique(member_uidx)" json:"userId"`
	Role   MemberRole `xorm:"not null" json:"role"`
}

type MemberRepository interface {
	Save(member *Member) error
	Get(appID, userID int) (Member, error)
	ListByApp(appID int) ([]Member, error)
	ListByUser(userID int) ([]Member, error)
	Delete(appID, userID int) error
}

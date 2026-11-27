package app

// CIToken is a per-app credential for GitLab and the factory Python tool.
// The raw token is shown once; only Prefix and Hash are stored.
type CIToken struct {
	ID      int    `xorm:"pk autoincr" json:"id"`
	AppID   int    `xorm:"not null index" json:"appId"`
	Name    string `xorm:"not null" json:"name"`
	Prefix  string `xorm:"varchar(16) not null" json:"prefix"`
	Hash    string `xorm:"varchar(64) not null unique" json:"-"`
	Scope   string `xorm:"varchar(16) not null default 'write'" json:"scope"`
	Revoked bool   `xorm:"not null default false" json:"revoked"`
}

const (
	TokenScopeRead  = "read"
	TokenScopeWrite = "write"
)

type CITokenRepository interface {
	Save(token *CIToken) error
	GetByHash(hash string) (CIToken, error)
	ListByApp(appID int) ([]CIToken, error)
	Revoke(appID, id int) error
}

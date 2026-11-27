package instance

// Instance is the single row of server-wide settings.
// The API token is generated on first boot. Pending holds the raw value until an admin dismisses it.
type Instance struct {
	ID              int    `xorm:"pk"`
	APITokenHash    string `xorm:"varchar(64)"`
	APITokenPrefix  string `xorm:"varchar(16)"`
	APITokenPending string `xorm:"text"`
	SMTPHost        string `xorm:"varchar(255)"`
	SMTPPort        string `xorm:"varchar(8)"`
	SMTPUser        string `xorm:"varchar(255)"`
	SMTPPassword    string `xorm:"varchar(255)"`
	SMTPFrom        string `xorm:"varchar(255)"`
}

type Repository interface {
	Get() (Instance, error)
	Save(row *Instance) error
}

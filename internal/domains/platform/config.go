package platform

// Config is a key/value flash hint for a platform (esptool args, RAUC slot, notes).
type Config struct {
	PlatformID int    `xorm:"pk not null"`
	Key        string `xorm:"pk 'cfg_key' not null"`
	Value      string `xorm:"text"`
}

type ConfigRepository interface {
	Save(item Config) error
	List(platformID int) ([]Config, error)
}

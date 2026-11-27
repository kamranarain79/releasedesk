package sql

import (
	"github.com/brewbits-co/releasedesk/internal/domains/platform"
	"xorm.io/xorm"
)

func NewPlatformConfigRepository(engine *xorm.Engine) platform.ConfigRepository {
	return &platformConfigRepository{engine: engine}
}

type platformConfigRepository struct {
	engine *xorm.Engine
}

func (r *platformConfigRepository) Save(item platform.Config) error {
	existing := platform.Config{}
	has, err := r.engine.Where("platform_id = ? AND cfg_key = ?", item.PlatformID, item.Key).Get(&existing)
	if err != nil {
		return err
	}
	if has {
		_, err = r.engine.Where("platform_id = ? AND cfg_key = ?", item.PlatformID, item.Key).Cols("value").Update(&platform.Config{Value: item.Value})
		return err
	}
	_, err = r.engine.Insert(&item)
	return err
}

func (r *platformConfigRepository) List(platformID int) ([]platform.Config, error) {
	var items []platform.Config
	err := r.engine.Where("platform_id = ?", platformID).Find(&items)
	return items, err
}

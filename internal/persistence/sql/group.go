package sql

import (
	"github.com/brewbits-co/releasedesk/internal/domains/app"
	"xorm.io/xorm"
)

func NewGroupRepository(engine *xorm.Engine) app.GroupRepository {
	return &groupRepository{engine: engine}
}

type groupRepository struct {
	engine *xorm.Engine
}

func (r *groupRepository) Save(group *app.DistributionGroup) error {
	_, err := r.engine.Insert(group)
	return err
}

func (r *groupRepository) ListByApp(appID int) ([]app.DistributionGroup, error) {
	var groups []app.DistributionGroup
	err := r.engine.Where("app_id = ?", appID).Find(&groups)
	return groups, err
}

func (r *groupRepository) AddMember(groupID, userID int) error {
	_, err := r.engine.Insert(&app.GroupMember{GroupID: groupID, UserID: userID})
	return err
}

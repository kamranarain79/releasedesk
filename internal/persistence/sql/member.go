package sql

import (
	"database/sql"

	"github.com/brewbits-co/releasedesk/internal/domains/app"
	"xorm.io/xorm"
)

func NewMemberRepository(engine *xorm.Engine) app.MemberRepository {
	return &memberRepository{engine: engine}
}

type memberRepository struct {
	engine *xorm.Engine
}

func (r *memberRepository) Save(member *app.Member) error {
	_, err := r.engine.Insert(member)
	return err
}

func (r *memberRepository) Get(appID, userID int) (app.Member, error) {
	var member app.Member
	has, err := r.engine.Where("app_id = ? AND user_id = ?", appID, userID).Get(&member)
	if err != nil {
		return app.Member{}, err
	}
	if !has {
		return app.Member{}, sql.ErrNoRows
	}
	return member, nil
}

func (r *memberRepository) ListByApp(appID int) ([]app.Member, error) {
	var members []app.Member
	err := r.engine.Where("app_id = ?", appID).Find(&members)
	return members, err
}

func (r *memberRepository) ListByUser(userID int) ([]app.Member, error) {
	var members []app.Member
	err := r.engine.Where("user_id = ?", userID).Find(&members)
	return members, err
}

func (r *memberRepository) Delete(appID, userID int) error {
	_, err := r.engine.Exec("DELETE FROM member WHERE app_id = ? AND user_id = ?", appID, userID)
	return err
}

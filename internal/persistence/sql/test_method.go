package sql

import (
	"github.com/brewbits-co/releasedesk/internal/domains/release"
	"time"
	"xorm.io/xorm"
)

func NewTestMethodRepository(engine *xorm.Engine) release.TestMethodRepository {
	return &testMethodRepository{engine: engine}
}

type testMethodRepository struct {
	engine *xorm.Engine
}

func (r *testMethodRepository) SavePlan(releaseID int, overview string, methods []release.TestMethod) error {
	session := r.engine.NewSession()
	defer session.Close()
	if err := session.Begin(); err != nil {
		return err
	}

	if _, err := session.Table("release").Where("id = ?", releaseID).Update(map[string]interface{}{
		"qa_notes":   overview,
		"updated_at": time.Now(),
	}); err != nil {
		_ = session.Rollback()
		return err
	}

	if _, err := session.Where("release_id = ?", releaseID).Delete(&release.TestMethod{}); err != nil {
		_ = session.Rollback()
		return err
	}

	for i := range methods {
		methods[i].ID = 0
		methods[i].ReleaseID = releaseID
		methods[i].SortOrder = i
		if methods[i].Title == "" {
			continue
		}
		if _, err := session.Insert(&methods[i]); err != nil {
			_ = session.Rollback()
			return err
		}
	}

	return session.Commit()
}

func (r *testMethodRepository) GetPlan(releaseID int) (release.TestPlan, error) {
	var info release.BasicInfo
	has, err := r.engine.Table("release").ID(releaseID).Get(&info)
	if err != nil {
		return release.TestPlan{}, err
	}
	plan := release.TestPlan{ReleaseID: releaseID, Methods: []release.TestMethod{}}
	if has {
		plan.Overview = info.QANotes
	}

	var methods []release.TestMethod
	if err := r.engine.Where("release_id = ?", releaseID).Asc("sort_order").Find(&methods); err != nil {
		return release.TestPlan{}, err
	}
	plan.Methods = methods
	return plan, nil
}

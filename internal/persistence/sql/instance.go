package sql

import (
	"github.com/brewbits-co/releasedesk/internal/domains/instance"
	"xorm.io/xorm"
)

func NewInstanceRepository(engine *xorm.Engine) instance.Repository {
	return &instanceRepository{engine: engine}
}

type instanceRepository struct {
	engine *xorm.Engine
}

func (r *instanceRepository) Get() (instance.Instance, error) {
	var row instance.Instance
	has, err := r.engine.ID(1).Get(&row)
	if err != nil {
		return instance.Instance{}, err
	}
	if !has {
		return instance.Instance{}, nil
	}
	return row, nil
}

func (r *instanceRepository) Save(row *instance.Instance) error {
	row.ID = 1
	has, err := r.engine.ID(1).Exist(&instance.Instance{})
	if err != nil {
		return err
	}
	if !has {
		_, err = r.engine.Insert(row)
		return err
	}
	_, err = r.engine.ID(1).AllCols().Update(row)
	return err
}

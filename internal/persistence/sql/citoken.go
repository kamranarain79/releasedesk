package sql

import (
	"database/sql"

	"github.com/brewbits-co/releasedesk/internal/domains/app"
	"xorm.io/xorm"
)

func NewCITokenRepository(engine *xorm.Engine) app.CITokenRepository {
	return &ciTokenRepository{engine: engine}
}

type ciTokenRepository struct {
	engine *xorm.Engine
}

func (r *ciTokenRepository) Save(token *app.CIToken) error {
	_, err := r.engine.Insert(token)
	return err
}

func (r *ciTokenRepository) GetByHash(hash string) (app.CIToken, error) {
	var token app.CIToken
	has, err := r.engine.Where("hash = ? AND revoked = ?", hash, false).Get(&token)
	if err != nil {
		return app.CIToken{}, err
	}
	if !has {
		return app.CIToken{}, sql.ErrNoRows
	}
	return token, nil
}

func (r *ciTokenRepository) ListByApp(appID int) ([]app.CIToken, error) {
	var tokens []app.CIToken
	err := r.engine.Where("app_id = ?", appID).Desc("id").Find(&tokens)
	return tokens, err
}

func (r *ciTokenRepository) Revoke(appID, id int) error {
	result, err := r.engine.Exec("UPDATE ci_token SET revoked = 1 WHERE app_id = ? AND id = ?", appID, id)
	if err != nil {
		return err
	}
	n, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

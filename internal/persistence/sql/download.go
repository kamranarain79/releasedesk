package sql

import (
	"github.com/brewbits-co/releasedesk/internal/domains/build"
	"xorm.io/xorm"
)

func NewDownloadRepository(engine *xorm.Engine) build.DownloadRepository {
	return &downloadRepository{engine: engine}
}

type downloadRepository struct {
	engine *xorm.Engine
}

func (r *downloadRepository) Increment(sha256 string) error {
	res, err := r.engine.Exec("UPDATE download_stat SET count = count + 1 WHERE sha256 = ?", sha256)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n > 0 {
		return nil
	}
	_, err = r.engine.Insert(&build.DownloadStat{Sha256: sha256, Count: 1})
	return err
}

func (r *downloadRepository) Sum(checksums []string) (int64, error) {
	if len(checksums) == 0 {
		return 0, nil
	}
	var total int64
	_, err := r.engine.Table("download_stat").In("sha256", checksums).Select("sum(count)").Get(&total)
	return total, err
}

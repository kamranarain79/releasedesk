package sql

import (
	"strconv"
	"strings"

	"github.com/brewbits-co/releasedesk/internal/domains/app"
	"github.com/brewbits-co/releasedesk/internal/values"
	"xorm.io/xorm"
)

// NewApplicationRepository is the constructor for appRepository
func NewApplicationRepository(engine *xorm.Engine) app.AppRepository {
	return &appRepository{engine: engine}
}

// appRepository is the implementation of app.AppRepository
type appRepository struct {
	engine *xorm.Engine
}

func (r *appRepository) Save(app *app.App) error {
	_, err := r.engine.Insert(app)
	if err != nil {
		return err
	}

	return nil
}

func (r *appRepository) Find() ([]app.App, error) {
	var apps []app.App
	err := r.engine.Find(&apps)
	if err != nil {
		return nil, err
	}

	for i := range apps {
		apps[i].FormatAuditable()
	}

	return apps, nil
}

func (r *appRepository) GetByID(id int) (app.App, error) {
	var entity app.App
	has, err := r.engine.ID(id).Get(&entity)
	if err != nil {
		return app.App{}, err
	}
	if !has {
		return app.App{}, app.ErrAppNotFound
	}
	return entity, nil
}

func (r *appRepository) UpdateSettings(id int, description string, private bool, teamsWebhook string) error {
	_, err := r.engine.Table("app").ID(id).Update(map[string]interface{}{
		"description":       description,
		"private":           private,
		"teams_webhook_url": teamsWebhook,
	})
	return err
}

func (r *appRepository) GetBySlug(slug values.Slug) (app.App, error) {
	var p app.App
	p.Slug = slug
	has, err := r.engine.Get(&p)
	if err != nil {
		return app.App{}, err
	}
	if !has {
		return app.App{}, app.ErrAppNotFound
	}

	return p, nil
}

func (r *appRepository) Update(app app.App) error {
	_, err := r.engine.ID(app.ID).Update(&app)
	if err != nil {
		return err
	}

	return nil
}

func (r *appRepository) Delete(application app.App) ([]string, error) {
	session := r.engine.NewSession()
	defer session.Close()
	if err := session.Begin(); err != nil {
		return nil, err
	}
	fail := func(err error) ([]string, error) {
		_ = session.Rollback()
		return nil, err
	}

	platformIDs, err := queryIDs(session, "SELECT id FROM platform WHERE app_id = ?", application.ID)
	if err != nil {
		return fail(err)
	}
	var buildIDs []int
	if len(platformIDs) > 0 {
		buildIDs, err = queryIDs(session, "SELECT id FROM build WHERE platform_id IN ("+placeholders(len(platformIDs))+")", ints(platformIDs)...)
		if err != nil {
			return fail(err)
		}
	}
	var checksums []string
	if len(buildIDs) > 0 {
		rows, err := session.QueryString(append([]interface{}{"SELECT sha256 FROM artifact WHERE build_id IN (" + placeholders(len(buildIDs)) + ")"}, ints(buildIDs)...)...)
		if err != nil {
			return fail(err)
		}
		for _, row := range rows {
			if row["sha256"] != "" {
				checksums = append(checksums, row["sha256"])
			}
		}
	}
	releaseIDs, err := queryIDs(session, "SELECT id FROM `release` WHERE app_id = ?", application.ID)
	if err != nil {
		return fail(err)
	}
	groupIDs, err := queryIDs(session, "SELECT id FROM distribution_group WHERE app_id = ?", application.ID)
	if err != nil {
		return fail(err)
	}

	statements := make([]sqlExec, 0, 16)
	if len(releaseIDs) > 0 {
		in := placeholders(len(releaseIDs))
		args := ints(releaseIDs)
		statements = append(statements,
			sqlExec{"DELETE FROM changelog WHERE release_id IN (" + in + ")", args},
			sqlExec{"DELETE FROM test_method WHERE release_id IN (" + in + ")", args},
			sqlExec{"DELETE FROM linked_builds WHERE release_id IN (" + in + ")", args},
			sqlExec{"DELETE FROM `release` WHERE id IN (" + in + ")", args},
		)
	}
	statements = append(statements, sqlExec{"DELETE FROM announcement WHERE app_id = ?", []interface{}{application.ID}})
	if len(buildIDs) > 0 {
		in := placeholders(len(buildIDs))
		args := ints(buildIDs)
		statements = append(statements,
			sqlExec{"DELETE FROM linked_builds WHERE build_id IN (" + in + ")", args},
			sqlExec{"DELETE FROM artifact WHERE build_id IN (" + in + ")", args},
			sqlExec{"DELETE FROM build_metadata WHERE build_id IN (" + in + ")", args},
			sqlExec{"DELETE FROM build WHERE id IN (" + in + ")", args},
		)
	}
	if len(platformIDs) > 0 {
		in := placeholders(len(platformIDs))
		args := ints(platformIDs)
		statements = append(statements, sqlExec{"DELETE FROM config WHERE platform_id IN (" + in + ")", args})
		statements = append(statements, sqlExec{"DELETE FROM platform WHERE id IN (" + in + ")", args})
	}
	if len(groupIDs) > 0 {
		in := placeholders(len(groupIDs))
		statements = append(statements, sqlExec{"DELETE FROM group_member WHERE group_id IN (" + in + ")", ints(groupIDs)})
	}
	statements = append(statements,
		sqlExec{"DELETE FROM distribution_group WHERE app_id = ?", []interface{}{application.ID}},
		sqlExec{"DELETE FROM member WHERE app_id = ?", []interface{}{application.ID}},
		sqlExec{"DELETE FROM ci_token WHERE app_id = ?", []interface{}{application.ID}},
		sqlExec{"DELETE FROM channel WHERE app_id = ?", []interface{}{application.ID}},
		sqlExec{"DELETE FROM audit_event WHERE app_id = ?", []interface{}{application.ID}},
		sqlExec{"DELETE FROM app WHERE id = ?", []interface{}{application.ID}},
	)
	for _, statement := range statements {
		if _, err = session.Exec(append([]interface{}{statement.query}, statement.args...)...); err != nil {
			return fail(err)
		}
	}
	if err = session.Commit(); err != nil {
		return fail(err)
	}

	var unused []string
	seen := map[string]bool{}
	for _, checksum := range checksums {
		if seen[checksum] {
			continue
		}
		seen[checksum] = true
		rows, err := r.engine.QueryString("SELECT COUNT(*) AS n FROM artifact WHERE sha256 = ?", checksum)
		if err != nil || len(rows) == 0 || rows[0]["n"] != "0" {
			continue
		}
		unused = append(unused, checksum)
	}
	return unused, nil
}

type sqlExec struct {
	query string
	args  []interface{}
}

func queryIDs(session *xorm.Session, query string, args ...interface{}) ([]int, error) {
	rows, err := session.QueryString(append([]interface{}{query}, args...)...)
	if err != nil {
		return nil, err
	}
	ids := make([]int, 0, len(rows))
	for _, row := range rows {
		value := row["id"]
		if value == "" {
			value = row["ID"]
		}
		id, err := strconv.Atoi(value)
		if err != nil {
			continue
		}
		ids = append(ids, id)
	}
	return ids, nil
}

func placeholders(n int) string {
	return strings.TrimRight(strings.Repeat("?,", n), ",")
}

func ints(ids []int) []interface{} {
	args := make([]interface{}, len(ids))
	for i, id := range ids {
		args[i] = id
	}
	return args
}

func (r *appRepository) SaveSetupGuide(guide app.SetupGuide) error {
	session := r.engine.NewSession()
	defer session.Close()

	if err := session.Begin(); err != nil {
		return err
	}

	_, err := session.Table("app").
		Where("id = ?", guide.AppID).
		Update(map[string]interface{}{
			"version_format":        guide.VersionFormat,
			"setup_guide_completed": true,
		})
	if err != nil {
		session.Rollback()
		return err
	}

	for _, channel := range guide.Channels {
		_, err := session.Insert(&channel)
		if err != nil {
			session.Rollback()
			return err
		}
	}

	return session.Commit()
}

func (r *appRepository) GetPlatformAvailability(app *app.App) error {
	platforms := make([]string, 0)
	err := r.engine.Table("platform").
		Where("app_id = ?", app.ID).
		Cols("os").
		Find(&platforms)
	if err != nil {
		return err
	}

	for _, platform := range platforms {
		switch platform {
		case "Android":
			app.HasAndroid = true
		case "iOS":
			app.HasIOS = true
		case "Windows":
			app.HasWindows = true
		case "Linux":
			app.HasLinux = true
		case "macOS":
			app.HasMacOS = true
		case "ESP32":
			app.HasESP32 = true
		case "STM32":
			app.HasSTM32 = true
		}
	}

	return nil
}

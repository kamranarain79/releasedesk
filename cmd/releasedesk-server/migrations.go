package main

import (
	"database/sql"
	"strconv"
	"time"

	"github.com/brewbits-co/releasedesk/internal/domains/app"
	"github.com/brewbits-co/releasedesk/internal/domains/audit"
	"github.com/brewbits-co/releasedesk/internal/domains/build"
	"github.com/brewbits-co/releasedesk/internal/domains/instance"
	"github.com/brewbits-co/releasedesk/internal/domains/platform"
	"github.com/brewbits-co/releasedesk/internal/domains/release"
	"github.com/brewbits-co/releasedesk/internal/domains/user"
	"github.com/brewbits-co/releasedesk/internal/values"
	"github.com/brewbits-co/releasedesk/pkg/fields"
	"github.com/brewbits-co/releasedesk/pkg/validator"
	"xorm.io/xorm"
	"xorm.io/xorm/migrate"
)

func applyMigrations(engine *xorm.Engine) error {
	var migrations = []*migrate.Migration{
		{
			ID: "202506142200",
			Migrate: func(tx *xorm.Engine) error {
				return tx.Sync2(&app.App{})
			},
			Rollback: func(tx *xorm.Engine) error {
				return tx.DropTables(&app.App{})
			},
		},
		{
			ID: "202506142201",
			Migrate: func(tx *xorm.Engine) error {
				return tx.Sync2(&release.Channel{})
			},
			Rollback: func(tx *xorm.Engine) error {
				return tx.DropTables(&release.Channel{})
			},
		},
		{
			ID: "202506142202",
			Migrate: func(tx *xorm.Engine) error {
				return tx.Sync2(&build.Build{}, &build.BuildMetadata{})
			},
			Rollback: func(tx *xorm.Engine) error {
				return tx.DropTables(&build.Build{}, &build.BuildMetadata{})
			},
		},
		{
			ID: "202506142203",
			Migrate: func(tx *xorm.Engine) error {
				return tx.Sync2(&build.Artifact{})
			},
			Rollback: func(tx *xorm.Engine) error {
				return tx.DropTables(&build.Build{})
			},
		},
		{
			ID: "202506142204",
			Migrate: func(tx *xorm.Engine) error {
				return tx.Sync2(&platform.Platform{})
			},
			Rollback: func(tx *xorm.Engine) error {
				return tx.DropTables(&platform.Platform{})
			},
		},
		{
			ID: "202506142205",
			Migrate: func(tx *xorm.Engine) error {
				err := tx.Sync2(&user.User{})
				if err != nil {
					return err
				}
				_, err = tx.Insert(user.User{
					BaseValidator: validator.BaseValidator{},
					Auditable:     fields.Auditable{},
					ID:            0,
					Username:      "admin",
					Email:         sql.NullString{},
					Password:      "$2a$10$Z13RQlu6HdKSW41rJsz7Ju5NZ0VMyUdm6YZMr0wjJqW955qd2pzx2", // admin
					FirstName:     sql.NullString{},
					LastName:      sql.NullString{},
					Role:          values.Admin,
				})
				if err != nil {
					return err
				}
				return nil
			},
			Rollback: func(tx *xorm.Engine) error {
				return tx.DropTables(&user.User{})
			},
		},
		{
			ID: "202506142206",
			Migrate: func(tx *xorm.Engine) error {
				return tx.Sync2(&release.Release{})
			},
			Rollback: func(tx *xorm.Engine) error {
				return tx.DropTables(&release.Release{})
			},
		},
		{
			ID: "202506142207",
			Migrate: func(tx *xorm.Engine) error {
				type Release struct {
					ReleaseNotes string `xorm:"longtext"`
				}

				return tx.Sync2(&release.Changelog{}, &Release{})
			},
			Rollback: func(tx *xorm.Engine) error {
				return tx.DropTables(&release.Changelog{})
			},
		},
		{
			ID: "202506162330",
			Migrate: func(tx *xorm.Engine) error {
				type Release struct {
					BuildSelection string `xorm:"varchar(10) not null default 'Last'"`
				}

				// Create LinkedBuilds table
				err := tx.Sync2(&release.LinkedBuilds{}, &Release{})
				if err != nil {
					return err
				}

				return nil
			},
			Rollback: func(tx *xorm.Engine) error {
				// Drop LinkedBuilds table
				err := tx.DropTables(&release.LinkedBuilds{})
				if err != nil {
					return err
				}

				return nil
			},
		},
		{
			ID: "202611251800",
			Migrate: func(tx *xorm.Engine) error {
				type Release struct {
					QANotes string `xorm:"longtext qa_notes"`
				}
				if err := tx.Sync2(&Release{}, &release.TestMethod{}, &build.Artifact{}); err != nil {
					return err
				}
				_, _ = tx.Exec("UPDATE artifact SET kind = 'firmware' WHERE kind IS NULL OR kind = ''")
				_, _ = tx.Exec("DROP INDEX IF EXISTS UQE_artifact_artifact_uidx")
				_, _ = tx.Exec("CREATE UNIQUE INDEX IF NOT EXISTS UQE_artifact_artifact_uidx ON artifact (build_id, architecture, kind)")
				return nil
			},
			Rollback: func(tx *xorm.Engine) error {
				return tx.DropTables(&release.TestMethod{})
			},
		},
		{
			ID: "202611271700",
			Migrate: func(tx *xorm.Engine) error {
				if err := tx.Sync2(
					&app.App{},
					&app.Member{},
					&app.CIToken{},
					&app.DistributionGroup{},
					&app.GroupMember{},
					&user.User{},
					&release.Channel{},
					&release.Release{},
					&release.Announcement{},
					&build.Build{},
					&build.DownloadStat{},
					&platform.Platform{},
					&platform.Config{},
				); err != nil {
					return err
				}
				_, _ = tx.Table("user").Where("username = ?", "admin").Update(map[string]interface{}{"must_change_password": true})
				var apps []app.App
				if err := tx.Find(&apps); err != nil {
					return err
				}
				for _, application := range apps {
					_, _ = tx.Insert(&app.Member{AppID: application.ID, UserID: 1, Role: app.MemberAdmin})
				}
				return nil
			},
			Rollback: func(tx *xorm.Engine) error {
				return tx.DropTables(&app.Member{}, &app.CIToken{}, &app.DistributionGroup{}, &app.GroupMember{}, &release.Announcement{}, &build.DownloadStat{}, &platform.Config{})
			},
		},
		{
			ID: "202611271900",
			Migrate: func(tx *xorm.Engine) error {
				return tx.Sync2(&instance.Instance{})
			},
			Rollback: func(tx *xorm.Engine) error {
				return tx.DropTables(&instance.Instance{})
			},
		},
		{
			ID: "202611272100",
			Migrate: func(tx *xorm.Engine) error {
				if err := tx.Sync2(&app.CIToken{}); err != nil {
					return err
				}
				_, err := tx.Exec("UPDATE ci_token SET scope = 'write' WHERE scope IS NULL OR scope = ''")
				return err
			},
			Rollback: func(tx *xorm.Engine) error {
				return nil
			},
		},
		{
			ID: "202611272200",
			Migrate: func(tx *xorm.Engine) error {
				if err := tx.Sync2(&build.Build{}); err != nil {
					return err
				}
				_, err := tx.Exec("UPDATE build SET status = 'Unpublished' WHERE status IS NULL OR status = ''")
				return err
			},
			Rollback: func(tx *xorm.Engine) error {
				return nil
			},
		},
		{
			ID: "202611272300",
			Migrate: func(tx *xorm.Engine) error {
				if err := tx.Sync2(&audit.Event{}); err != nil {
					return err
				}
				return backfillAudit(tx)
			},
			Rollback: func(tx *xorm.Engine) error {
				return tx.DropTables(&audit.Event{})
			},
		},
	}

	m := migrate.New(engine, &migrate.Options{
		TableName:    "migrations",
		IDColumnName: "id",
	}, migrations)

	return m.Migrate()
}

func backfillAudit(tx *xorm.Engine) error {
	releases, err := tx.QueryString("SELECT app_id, version, created_by, created_at FROM `release`")
	if err != nil {
		return err
	}
	for _, row := range releases {
		appID, _ := strconv.Atoi(row["app_id"])
		actor := row["created_by"]
		if actor == "" {
			actor = "unknown"
		}
		if _, err = tx.Insert(&audit.Event{
			AppID:     appID,
			Actor:     actor,
			Action:    "created",
			Subject:   audit.ReleaseSubject(row["version"]),
			Summary:   "Created release " + row["version"],
			CreatedAt: parseSQLTime(row["created_at"]),
		}); err != nil {
			return err
		}
	}
	builds, err := tx.QueryString("SELECT p.app_id AS app_id, p.os AS os, b.number AS number, b.created_by AS created_by, b.created_at AS created_at FROM build b JOIN platform p ON p.id = b.platform_id")
	if err != nil {
		return err
	}
	for _, row := range builds {
		appID, _ := strconv.Atoi(row["app_id"])
		actor := row["created_by"]
		if actor == "" {
			actor = "unknown"
		}
		label := row["os"] + " build #" + row["number"]
		if _, err = tx.Insert(&audit.Event{
			AppID:     appID,
			Actor:     actor,
			Action:    "uploaded",
			Subject:   audit.BuildSubject(row["os"], row["number"]),
			Summary:   "Uploaded " + label,
			CreatedAt: parseSQLTime(row["created_at"]),
		}); err != nil {
			return err
		}
	}
	return nil
}

func parseSQLTime(value string) time.Time {
	for _, layout := range []string{time.RFC3339Nano, time.RFC3339, "2006-01-02 15:04:05", "2006-01-02 15:04:05.999999999-07:00"} {
		if parsed, err := time.Parse(layout, value); err == nil {
			return parsed
		}
	}
	return time.Now().UTC()
}

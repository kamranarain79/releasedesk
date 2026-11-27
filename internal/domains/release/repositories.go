package release

import (
	"github.com/brewbits-co/releasedesk/internal/values"
)

type ReleaseRepository interface {
	Save(release *Release) error
	Update(release *Release) error
	FindByAppIDAndChannel(appID int, channelID int) ([]BasicInfo, error)
	FindRecentByAppID(appID int, limit int) ([]BasicInfo, error)
	FindChannelsByAppID(appID int) ([]Channel, error)
	GetByAppIDAndVersion(appID int, version string) (Release, error)
	GetByID(id int) (Release, error)
	SetStatus(id int, status ReleaseStatus) error
	Delete(id int) error
	LinkBuild(releaseID int, buildID int, os values.OS) error
	UnlinkBuild(releaseID int, buildID int, os values.OS) error
	ReplaceLinks(releaseID int, links []LinkedBuilds) error
}

type ReleaseNotesRepository interface {
	Save(releaseNotes *ReleaseNotes) error
	GetByReleaseID(releaseID int) (ReleaseNotes, error)
	FindChangelogsByReleaseID(releaseID int) ([]Changelog, error)
}

type TestMethodRepository interface {
	SavePlan(releaseID int, overview string, methods []TestMethod) error
	GetPlan(releaseID int) (TestPlan, error)
}

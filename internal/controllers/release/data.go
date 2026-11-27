package release

import (
	"github.com/brewbits-co/releasedesk/internal/domains/audit"
	"github.com/brewbits-co/releasedesk/internal/domains/release"
	"github.com/brewbits-co/releasedesk/internal/domains/user"
	"github.com/brewbits-co/releasedesk/pkg/session"
)

type ReleaseListData struct {
	CurrentPage string
	session.SessionData
	session.CurrentAppData
	Releases         []release.BasicInfo
	Channels         []release.Channel
	CurrentChannelID int
	AvailableBuilds  []ReleaseBinary
	CanEdit          bool
	Directory        user.Directory
}

type ReleaseBinary struct {
	BuildID  int
	OS       string
	Number   string
	Version  string
	Files    string
	Checksum string
	Included bool
}

type ReleaseSummaryData struct {
	CurrentPage string
	session.SessionData
	session.CurrentAppData
	release.Release
	Channels      []release.Channel
	DownloadCount int64
	CanEdit       bool
	Binaries      []ReleaseBinary
	Audit         []audit.Event
	Creator       user.Person
	Editor        user.Person
	HasEditor     bool
}

type ReleaseNotesData struct {
	CurrentPage string
	session.SessionData
	session.CurrentAppData
	Release      release.Release
	ReleaseNotes release.ReleaseNotes
	CanEdit      bool
	Editing      bool
}

type TestMethodsData struct {
	CurrentPage string
	session.SessionData
	session.CurrentAppData
	Release  release.Release
	TestPlan release.TestPlan
	CanEdit  bool
	Editing  bool
}

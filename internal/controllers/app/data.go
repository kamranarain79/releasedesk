package app

import (
	domainapp "github.com/brewbits-co/releasedesk/internal/domains/app"
	"github.com/brewbits-co/releasedesk/internal/domains/build"
	"github.com/brewbits-co/releasedesk/internal/domains/platform"
	"github.com/brewbits-co/releasedesk/internal/domains/release"
	"github.com/brewbits-co/releasedesk/internal/domains/user"
	"github.com/brewbits-co/releasedesk/internal/values"
	"github.com/brewbits-co/releasedesk/pkg/session"
)

type DashboardData struct {
	CurrentPage string
	session.SessionData
	session.CurrentAppData
	SetupGuideCompleted bool
	LatestReleases      []DashboardRelease
	LatestBuilds        []DashboardBuild
}

type DashboardRelease struct {
	release.BasicInfo
	Creator   user.Person
	Editor    user.Person
	HasEditor bool
}

type DashboardBuild struct {
	OS values.OS
	build.BasicInfo
	Uploader user.Person
}

type StubPageData struct {
	CurrentPage string
	session.SessionData
	session.CurrentAppData
}

type settingsMember struct {
	UserID   int
	Username string
	Role     domainapp.MemberRole
}

type settingsPage struct {
	CurrentPage string
	session.SessionData
	session.CurrentAppData
	App      domainapp.App
	Members  []settingsMember
	Tokens   []domainapp.CIToken
	Users    []user.User
	CanAdmin bool
	NewToken string
	Error    string
}

type configPage struct {
	CurrentPage string
	session.SessionData
	session.CurrentAppData
	OS    string
	Items []platform.Config
}

type groupPage struct {
	CurrentPage string
	session.SessionData
	session.CurrentAppData
	Groups []domainapp.DistributionGroup
}

type announcePage struct {
	CurrentPage string
	session.SessionData
	session.CurrentAppData
	Items []release.Announcement
}

type feedPage struct {
	CurrentPage string
	session.SessionData
	session.CurrentAppData
}

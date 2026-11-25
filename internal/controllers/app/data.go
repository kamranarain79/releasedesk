package app

import (
	"github.com/brewbits-co/releasedesk/internal/domains/build"
	"github.com/brewbits-co/releasedesk/internal/domains/release"
	"github.com/brewbits-co/releasedesk/internal/values"
	"github.com/brewbits-co/releasedesk/pkg/session"
)

type DashboardData struct {
	CurrentPage string
	session.SessionData
	session.CurrentAppData
	SetupGuideCompleted bool
	LatestReleases      []release.BasicInfo
	LatestBuilds        []DashboardBuild
}

type DashboardBuild struct {
	OS values.OS
	build.BasicInfo
}

type StubPageData struct {
	CurrentPage string
	session.SessionData
	session.CurrentAppData
}

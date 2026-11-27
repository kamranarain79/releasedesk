package build

import (
	"github.com/brewbits-co/releasedesk/internal/domains/audit"
	"github.com/brewbits-co/releasedesk/internal/domains/build"
	"github.com/brewbits-co/releasedesk/internal/domains/user"
	"github.com/brewbits-co/releasedesk/pkg/session"
)

type BuildListData struct {
	CurrentPage string
	session.SessionData
	session.CurrentAppData
	CurrentPlatform session.CurrentPlatformData
	Builds          []build.BasicInfo
	CanEdit         bool
	Directory       user.Directory
}

type BuildDetailsData struct {
	CurrentPage string
	session.SessionData
	session.CurrentAppData
	CurrentPlatform session.CurrentPlatformData
	Build           build.Build
	CanEdit         bool
	Audit           []audit.Event
	Uploader        user.Person
}

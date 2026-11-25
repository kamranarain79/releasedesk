package misc

import (
	"github.com/brewbits-co/releasedesk/internal/domains/app"
	"github.com/brewbits-co/releasedesk/internal/domains/user"
	"github.com/brewbits-co/releasedesk/pkg/session"
)

type HomepageData struct {
	session.SessionData
	CurrentPage string
	Apps        []app.App
}

type DownloadPortalData struct {
	session.SessionData
	CurrentPage string
	PortalURL   string
	PublicApps  []app.App
}

type PeoplePageData struct {
	session.SessionData
	CurrentPage string
	Users       []user.User
}

type AdministrationData struct {
	session.SessionData
	CurrentPage string
}

type AccountSettingsData struct {
	session.SessionData
	CurrentPage string
	User        user.User
}

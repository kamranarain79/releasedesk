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
	IsAdmin     bool
	Users       []user.User
	Apps        []app.App
}

type AdministrationData struct {
	session.SessionData
	CurrentPage     string
	IsAdmin         bool
	DataDir         string
	Database        string
	DatabaseNote    string
	SMTPHost        string
	SMTPPort        string
	SMTPUser        string
	SMTPFrom        string
	SMTPPasswordSet bool
	TokenPrefix     string
	TokenPending    string
	EnvToken        bool
	Section         string
	HasLogo         bool
	HasFavicon      bool
}

type AccountSettingsData struct {
	session.SessionData
	CurrentPage string
	User        user.User
}

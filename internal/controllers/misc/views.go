package misc

import (
	"log"
	"net/http"
	"os"

	"github.com/brewbits-co/releasedesk/internal/views"
	"github.com/brewbits-co/releasedesk/pkg/branding"
	"github.com/brewbits-co/releasedesk/pkg/session"
	"github.com/go-chi/chi/v5"
)

func (c *miscController) RenderHomepage(w http.ResponseWriter, r *http.Request) {
	userID, _ := r.Context().Value("userID").(int)
	apps, err := c.appService.GetUserAccessibleApps(userID)
	if err != nil {
		log.Println(err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	data := HomepageData{
		SessionData: session.NewSessionData(r.Context()),
		CurrentPage: "Applications",
		Apps:        apps,
	}

	tmpl, err := views.ParseTemplate(views.NavbarLayout, "templates/console/homepage.gohtml")
	if err != nil {
		log.Println(err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	views.Execute(w, tmpl, data)
}

func (c *miscController) RenderDownloadPortal(w http.ResponseWriter, r *http.Request) {
	publicApps, err := c.appService.GetPublicApps()
	if err != nil {
		log.Println(err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	data := DownloadPortalData{
		SessionData: session.NewSessionData(r.Context()),
		CurrentPage: "Download Portal",
		PortalURL:   portalPublicURL(),
		PublicApps:  publicApps,
	}

	tmpl, err := views.ParseTemplate(views.NavbarLayout, "templates/console/download_portal.gohtml")
	if err != nil {
		log.Println(err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	views.Execute(w, tmpl, data)
}

func (c *miscController) RenderPeople(w http.ResponseWriter, r *http.Request) {
	users, err := c.authService.ListUsers()
	if err != nil {
		log.Println(err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	userID, _ := r.Context().Value("userID").(int)
	apps, _ := c.appService.GetUserAccessibleApps(userID)
	data := PeoplePageData{
		SessionData: session.NewSessionData(r.Context()),
		CurrentPage: "People",
		IsAdmin:     c.appService.IsInstanceAdmin(userID),
		Users:       users,
		Apps:        apps,
	}

	tmpl, err := views.ParseTemplate(views.NavbarLayout, "templates/console/people.gohtml")
	if err != nil {
		log.Println(err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	views.Execute(w, tmpl, data)
}

func (c *miscController) RenderAdministration(w http.ResponseWriter, r *http.Request) {
	section := chi.URLParam(r, "section")
	if section == "" {
		section = "general"
	}
	switch section {
	case "general", "branding", "email", "token":
	default:
		http.NotFound(w, r)
		return
	}
	userID, _ := r.Context().Value("userID").(int)
	settings, err := c.instanceService.Current()
	if err != nil {
		log.Println(err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	database := "SQLite file"
	note := "This process is using a SQLite file for metadata. Docker Compose starts PostgreSQL with the app, so metadata is not a file there. Firmware and other artifacts stay as files in the data directory either way."
	if os.Getenv("RELEASEDESK_DATABASE_URL") != "" {
		database = "PostgreSQL"
		note = "PostgreSQL is running with the app. Firmware and other artifacts stay as files in the data directory."
	}
	dataDir := os.Getenv("RELEASEDESK_DATA_DIR")
	if dataDir == "" {
		dataDir = "_data"
	}
	data := AdministrationData{
		SessionData:     session.NewSessionData(r.Context()),
		CurrentPage:     "Administration",
		IsAdmin:         c.appService.IsInstanceAdmin(userID),
		DataDir:         dataDir,
		Database:        database,
		DatabaseNote:    note,
		SMTPHost:        settings.SMTPHost,
		SMTPPort:        settings.SMTPPort,
		SMTPUser:        settings.SMTPUser,
		SMTPFrom:        settings.SMTPFrom,
		SMTPPasswordSet: settings.SMTPPassword != "",
		TokenPrefix:     settings.APITokenPrefix,
		TokenPending:    settings.APITokenPending,
		EnvToken:        os.Getenv("RELEASEDESK_API_TOKEN") != "",
		Section:         section,
		HasLogo:         brandExists("logo"),
		HasFavicon:      brandExists("favicon"),
	}

	tmpl, err := views.ParseTemplate(views.NavbarLayout, "templates/console/administration.gohtml")
	if err != nil {
		log.Println(err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	views.Execute(w, tmpl, data)
}

func brandExists(kind string) bool {
	file, _, err := branding.OpenBrand(kind)
	if err != nil {
		return false
	}
	file.Close()
	return true
}

func (c *miscController) RenderAccountSettings(w http.ResponseWriter, r *http.Request) {
	userID, _ := r.Context().Value("userID").(int)
	currentUser, err := c.authService.GetUser(userID)
	if err != nil {
		log.Println(err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	data := AccountSettingsData{
		SessionData: session.NewSessionData(r.Context()),
		CurrentPage: "Account Settings",
		User:        currentUser,
	}

	tmpl, err := views.ParseTemplate(views.NavbarLayout, "templates/console/account_settings.gohtml")
	if err != nil {
		log.Println(err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	views.Execute(w, tmpl, data)
}

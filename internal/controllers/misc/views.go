package misc

import (
	"log"
	"net/http"

	"github.com/brewbits-co/releasedesk/internal/views"
	"github.com/brewbits-co/releasedesk/pkg/session"
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

	data := PeoplePageData{
		SessionData: session.NewSessionData(r.Context()),
		CurrentPage: "People",
		Users:       users,
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
	data := AdministrationData{
		SessionData: session.NewSessionData(r.Context()),
		CurrentPage: "Administration",
	}

	tmpl, err := views.ParseTemplate(views.NavbarLayout, "templates/console/administration.gohtml")
	if err != nil {
		log.Println(err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	views.Execute(w, tmpl, data)
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

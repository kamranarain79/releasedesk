package build

import (
	"log"
	"net/http"
	"strconv"

	domainaudit "github.com/brewbits-co/releasedesk/internal/domains/audit"
	"github.com/brewbits-co/releasedesk/internal/domains/user"
	auditsvc "github.com/brewbits-co/releasedesk/internal/services/audit"
	"github.com/brewbits-co/releasedesk/internal/values"
	"github.com/brewbits-co/releasedesk/internal/views"
	"github.com/brewbits-co/releasedesk/pkg/session"
	"github.com/go-chi/chi/v5"
)

func (c *buildController) auditFor(appID int, subject string, directory user.Directory) []domainaudit.Event {
	events, err := c.audits.List(appID, subject)
	if err != nil {
		return []domainaudit.Event{}
	}
	return auditsvc.Annotate(events, directory)
}

func (c *buildController) userID(r *http.Request) int {
	id, _ := r.Context().Value("userID").(int)
	return id
}

func (c *buildController) canEdit(r *http.Request, slug string) bool {
	ok, err := c.appService.CanEditContent(c.userID(r), values.Slug(slug))
	return err == nil && ok
}

func (c *buildController) RenderBuildList(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	currentApp, err := c.appService.GetCurrentAppData(values.Slug(slug))
	if err != nil {
		http.NotFound(w, r)
		return
	}

	platform := chi.URLParam(r, "platform")
	currentPlatform := session.NewCurrentPlatformData(currentApp, values.OS(platform))
	if currentPlatform.PlatformID == 0 {
		http.NotFound(w, r)
		return
	}

	tmpl, err := views.ParseTemplate(views.SidebarLayout, "templates/console/build_list.gohtml")
	if err != nil {
		log.Println(err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	builds, err := c.service.GetPlatformBuilds(currentPlatform.PlatformID)
	if err != nil {
		log.Println(err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	data := BuildListData{
		CurrentPage:     "Builds",
		SessionData:     session.NewSessionData(r.Context()),
		CurrentAppData:  currentApp,
		CurrentPlatform: currentPlatform,
		Builds:          builds,
		CanEdit:         c.canEdit(r, slug),
		Directory:       c.audits.Directory(),
	}

	views.Execute(w, tmpl, data)
}

func (c *buildController) RenderBuildDetails(w http.ResponseWriter, r *http.Request) {
	c.renderBuildPage(w, r, "templates/console/build_details.gohtml")
}

func (c *buildController) RenderBuildMetadata(w http.ResponseWriter, r *http.Request) {
	c.renderBuildPage(w, r, "templates/console/build_metadata.gohtml")
}

func (c *buildController) renderBuildPage(w http.ResponseWriter, r *http.Request, templateName string) {
	slug := chi.URLParam(r, "slug")
	currentApp, err := c.appService.GetCurrentAppData(values.Slug(slug))
	if err != nil {
		http.NotFound(w, r)
		return
	}

	platform := chi.URLParam(r, "platform")
	currentPlatform := session.NewCurrentPlatformData(currentApp, values.OS(platform))
	if currentPlatform.PlatformID == 0 {
		http.NotFound(w, r)
		return
	}

	number, err := strconv.Atoi(chi.URLParam(r, "number"))
	if err != nil {
		http.NotFound(w, r)
		return
	}

	buildDetails, err := c.service.GetBuildDetails(currentPlatform.PlatformID, number)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	tmpl, err := views.ParseTemplate(views.SidebarLayout, templateName)
	if err != nil {
		log.Println(err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	directory := c.audits.Directory()
	data := BuildDetailsData{
		CurrentPage:     "Builds",
		SessionData:     session.NewSessionData(r.Context()),
		CurrentAppData:  currentApp,
		CurrentPlatform: currentPlatform,
		Build:           buildDetails,
		CanEdit:         c.canEdit(r, slug),
		Audit:           c.auditFor(currentApp.AppID, domainaudit.BuildSubject(string(currentPlatform.OS), buildDetails.Number), directory),
		Uploader:        directory.Person(buildDetails.CreatedBy),
	}

	views.Execute(w, tmpl, data)
}

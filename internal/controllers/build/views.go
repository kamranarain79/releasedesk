package build

import (
	"log"
	"net/http"
	"strconv"

	"github.com/brewbits-co/releasedesk/internal/values"
	"github.com/brewbits-co/releasedesk/internal/views"
	"github.com/brewbits-co/releasedesk/pkg/session"
	"github.com/go-chi/chi/v5"
)

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

	data := BuildDetailsData{
		CurrentPage:     "Builds",
		SessionData:     session.NewSessionData(r.Context()),
		CurrentAppData:  currentApp,
		CurrentPlatform: currentPlatform,
		Build:           buildDetails,
	}

	views.Execute(w, tmpl, data)
}

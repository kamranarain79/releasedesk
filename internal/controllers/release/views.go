package release

import (
	"log"
	"net/http"
	"strconv"

	"github.com/brewbits-co/releasedesk/internal/domains/release"
	"github.com/brewbits-co/releasedesk/internal/values"
	"github.com/brewbits-co/releasedesk/internal/views"
	"github.com/brewbits-co/releasedesk/pkg/session"
	"github.com/go-chi/chi/v5"
)

func (c *releaseController) RenderReleaseList(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	currentApp, err := c.appService.GetCurrentAppData(values.Slug(slug))
	if err != nil {
		http.NotFound(w, r)
		return
	}

	channels, err := c.service.GetReleaseChannels(currentApp.AppID)
	if err != nil {
		log.Println(err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	channel := r.URL.Query().Get("channel")
	var currentChannelID int
	if len(channels) > 0 {
		if channel == "" {
			currentChannelID = channels[0].ID
		} else {
			currentChannelID, _ = strconv.Atoi(channel)
		}
	}

	var releases []release.BasicInfo
	if currentChannelID != 0 {
		releases, err = c.service.ListReleasesByChannel(currentApp.AppID, currentChannelID)
		if err != nil {
			log.Println(err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	}

	data := ReleaseListData{
		CurrentPage:      "Releases",
		SessionData:      session.NewSessionData(r.Context()),
		CurrentAppData:   currentApp,
		Releases:         releases,
		Channels:         channels,
		CurrentChannelID: currentChannelID,
	}

	tmpl, err := views.ParseTemplate(views.SidebarLayout, "templates/console/release_list.gohtml")
	if err != nil {
		log.Println(err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	views.Execute(w, tmpl, data)
}

func (c *releaseController) RenderReleaseSummary(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	currentApp, err := c.appService.GetCurrentAppData(values.Slug(slug))
	if err != nil {
		http.NotFound(w, r)
		return
	}

	channels, err := c.service.GetReleaseChannels(currentApp.AppID)
	if err != nil {
		log.Println(err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	version := chi.URLParam(r, "version")
	releaseSummary, err := c.service.GetReleaseSummary(currentApp.AppID, version)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	tmpl, err := views.ParseTemplate(views.SidebarLayout, "templates/console/release_summary.gohtml")
	if err != nil {
		log.Println(err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	data := ReleaseSummaryData{
		CurrentPage:    "Releases",
		SessionData:    session.NewSessionData(r.Context()),
		CurrentAppData: currentApp,
		Channels:       channels,
		Release:        releaseSummary,
	}

	views.Execute(w, tmpl, data)
}

func (c *releaseController) RenderReleaseNotes(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	currentApp, err := c.appService.GetCurrentAppData(values.Slug(slug))
	if err != nil {
		http.NotFound(w, r)
		return
	}

	version := chi.URLParam(r, "version")
	releaseSummary, err := c.service.GetReleaseSummary(currentApp.AppID, version)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	releaseNotes, err := c.service.GetReleaseNotes(releaseSummary.ID)
	if err != nil {
		releaseNotes = release.NewReleaseNotes(releaseSummary.ID, "")
	}

	tmpl, err := views.ParseTemplate(views.SidebarLayout, "templates/console/release_notes.gohtml")
	if err != nil {
		log.Println(err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	data := ReleaseNotesData{
		CurrentPage:    "Releases",
		SessionData:    session.NewSessionData(r.Context()),
		CurrentAppData: currentApp,
		Release:        releaseSummary,
		ReleaseNotes:   releaseNotes,
	}

	views.Execute(w, tmpl, data)
}

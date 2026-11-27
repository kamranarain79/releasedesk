package release

import (
	"log"
	"net/http"
	"strconv"

	domainaudit "github.com/brewbits-co/releasedesk/internal/domains/audit"
	"github.com/brewbits-co/releasedesk/internal/domains/release"
	"github.com/brewbits-co/releasedesk/internal/domains/user"
	auditsvc "github.com/brewbits-co/releasedesk/internal/services/audit"
	"github.com/brewbits-co/releasedesk/internal/values"
	"github.com/brewbits-co/releasedesk/internal/views"
	"github.com/brewbits-co/releasedesk/pkg/session"
	"github.com/go-chi/chi/v5"
)

func (c *releaseController) auditFor(appID int, subject string, directory user.Directory) []domainaudit.Event {
	events, err := c.audits.List(appID, subject)
	if err != nil {
		return []domainaudit.Event{}
	}
	return auditsvc.Annotate(events, directory)
}

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

	available := c.binaries(currentApp, release.Release{BasicInfo: release.BasicInfo{BuildSelection: release.Manual}})
	if available == nil {
		available = []ReleaseBinary{}
	}

	data := ReleaseListData{
		CurrentPage:      "Releases",
		SessionData:      session.NewSessionData(r.Context()),
		CurrentAppData:   currentApp,
		Releases:         releases,
		Channels:         channels,
		CurrentChannelID: currentChannelID,
		AvailableBuilds:  available,
		CanEdit:          c.canEdit(r, slug),
		Directory:        c.audits.Directory(),
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

	var checksums []string
	for _, platformData := range currentApp.Platforms {
		matches, err := c.buildService.FindByPlatformAndVersion(platformData.PlatformID, version)
		if err != nil {
			continue
		}
		for _, match := range matches {
			full, err := c.buildService.GetBuildByID(match.ID)
			if err != nil {
				continue
			}
			for _, artifact := range full.Artifacts {
				checksums = append(checksums, artifact.Sha256)
			}
		}
	}
	downloads, _ := c.buildService.SumDownloads(checksums)
	directory := c.audits.Directory()
	events := c.auditFor(currentApp.AppID, domainaudit.ReleaseSubject(releaseSummary.Version), directory)
	creator := directory.Person(releaseSummary.CreatedBy)
	if creator.Name == "" || creator.Name == "Unknown" {
		for _, event := range events {
			if event.Action == "created" {
				creator = directory.Person(event.Actor)
				break
			}
		}
	}
	var editor user.Person
	hasEditor := false
	for _, event := range events {
		if event.Action == "edited" {
			editor = directory.Person(event.Actor)
			hasEditor = true
			break
		}
	}
	data := ReleaseSummaryData{
		CurrentPage:    "Releases",
		SessionData:    session.NewSessionData(r.Context()),
		CurrentAppData: currentApp,
		Channels:       channels,
		Release:        releaseSummary,
		DownloadCount:  downloads,
		CanEdit:        c.canEdit(r, slug),
		Binaries:       c.binaries(currentApp, releaseSummary),
		Audit:          events,
		Creator:        creator,
		Editor:         editor,
		HasEditor:      hasEditor,
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

	canEdit := c.canEdit(r, slug)
	data := ReleaseNotesData{
		CurrentPage:    "Releases",
		SessionData:    session.NewSessionData(r.Context()),
		CurrentAppData: currentApp,
		Release:        releaseSummary,
		ReleaseNotes:   releaseNotes,
		CanEdit:        canEdit,
		Editing:        c.editing(r, canEdit),
	}

	views.Execute(w, tmpl, data)
}

func (c *releaseController) RenderTestMethods(w http.ResponseWriter, r *http.Request) {
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

	plan, err := c.service.GetTestPlan(releaseSummary.ID)
	if err != nil {
		plan = release.TestPlan{ReleaseID: releaseSummary.ID}
	}

	tmpl, err := views.ParseTemplate(views.SidebarLayout, "templates/console/release_test_methods.gohtml")
	if err != nil {
		log.Println(err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	canEdit := c.canEdit(r, slug)
	data := TestMethodsData{
		CurrentPage:    "Releases",
		SessionData:    session.NewSessionData(r.Context()),
		CurrentAppData: currentApp,
		Release:        releaseSummary,
		TestPlan:       plan,
		CanEdit:        canEdit,
		Editing:        c.editing(r, canEdit),
	}

	views.Execute(w, tmpl, data)
}

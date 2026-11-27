package app

import (
	"log"
	"net/http"

	domainaudit "github.com/brewbits-co/releasedesk/internal/domains/audit"
	auditsvc "github.com/brewbits-co/releasedesk/internal/services/audit"
	"github.com/brewbits-co/releasedesk/internal/values"
	"github.com/brewbits-co/releasedesk/internal/views"
	"github.com/brewbits-co/releasedesk/pkg/session"
	"github.com/go-chi/chi/v5"
)

func (c *appController) RenderDashboard(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	currentApp, err := c.service.GetCurrentAppData(values.Slug(slug))
	if err != nil {
		http.NotFound(w, r)
		return
	}

	overview, err := c.service.GetAppOverview(values.Slug(slug))
	if err != nil {
		http.NotFound(w, r)
		return
	}

	data := DashboardData{
		CurrentPage:         "Overview",
		SessionData:         session.NewSessionData(r.Context()),
		CurrentAppData:      currentApp,
		SetupGuideCompleted: overview.SetupGuideCompleted,
	}

	if overview.SetupGuideCompleted {
		releases, listErr := c.releaseService.ListRecentReleases(currentApp.AppID, 5)
		if listErr != nil {
			log.Println(listErr)
		} else {
			directory := c.audits.Directory()
			rows := make([]DashboardRelease, 0, len(releases))
			for _, item := range releases {
				events, _ := c.audits.List(currentApp.AppID, domainaudit.ReleaseSubject(item.Version))
				events = auditsvc.Annotate(events, directory)
				row := DashboardRelease{BasicInfo: item, Creator: directory.Person(item.CreatedBy)}
				if row.Creator.Name == "" || row.Creator.Name == "Unknown" {
					for _, event := range events {
						if event.Action == "created" {
							row.Creator = directory.Person(event.Actor)
							break
						}
					}
				}
				for _, event := range events {
					if event.Action == "edited" {
						row.Editor = directory.Person(event.Actor)
						row.HasEditor = true
						break
					}
				}
				rows = append(rows, row)
			}
			data.LatestReleases = rows
		}

		directory := c.audits.Directory()
		var latestBuilds []DashboardBuild
		for _, platformData := range currentApp.Platforms {
			builds, buildErr := c.buildService.GetPlatformBuilds(platformData.PlatformID)
			if buildErr != nil {
				log.Println(buildErr)
				continue
			}
			limit := 3
			if len(builds) < limit {
				limit = len(builds)
			}
			for i := 0; i < limit; i++ {
				latestBuilds = append(latestBuilds, DashboardBuild{
					OS:        platformData.OS,
					BasicInfo: builds[i],
					Uploader:  directory.Person(builds[i].CreatedBy),
				})
			}
		}
		data.LatestBuilds = latestBuilds
	}

	tmpl, err := views.ParseTemplate(views.SidebarLayout, "templates/console/dashboard.gohtml")
	if err != nil {
		log.Println(err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	views.Execute(w, tmpl, data)
}

func (c *appController) RenderStub(w http.ResponseWriter, r *http.Request, currentPage string) {
	slug := chi.URLParam(r, "slug")
	currentApp, err := c.service.GetCurrentAppData(values.Slug(slug))
	if err != nil {
		http.NotFound(w, r)
		return
	}

	data := StubPageData{
		CurrentPage:    currentPage,
		SessionData:    session.NewSessionData(r.Context()),
		CurrentAppData: currentApp,
	}

	tmpl, err := views.ParseTemplate(views.SidebarLayout, "templates/console/stub_page.gohtml")
	if err != nil {
		log.Println(err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	views.Execute(w, tmpl, data)
}

package release

import (
	"net/http"

	"github.com/brewbits-co/releasedesk/internal/services/app"
	auditsvc "github.com/brewbits-co/releasedesk/internal/services/audit"
	"github.com/brewbits-co/releasedesk/internal/services/build"
	"github.com/brewbits-co/releasedesk/internal/services/release"
)

type ReleaseController interface {
	RenderReleaseList(w http.ResponseWriter, r *http.Request)
	RenderReleaseSummary(w http.ResponseWriter, r *http.Request)
	RenderReleaseNotes(w http.ResponseWriter, r *http.Request)
	RenderTestMethods(w http.ResponseWriter, r *http.Request)
	HandleCreateRelease(w http.ResponseWriter, r *http.Request)
	HandleSaveReleaseNotes(w http.ResponseWriter, r *http.Request)
	HandleSaveTestPlan(w http.ResponseWriter, r *http.Request)
	HandleUpdateBasicInfo(w http.ResponseWriter, r *http.Request)
	HandleLinkBuilds(w http.ResponseWriter, r *http.Request)
	HandlePublishRelease(w http.ResponseWriter, r *http.Request)
	HandleDeleteRelease(w http.ResponseWriter, r *http.Request)
}

func NewReleaseController(service release.Service, appService app.Service, buildService build.Service, audits auditsvc.Service) ReleaseController {
	return &releaseController{
		service:      service,
		appService:   appService,
		buildService: buildService,
		audits:       audits,
	}
}

type releaseController struct {
	service      release.Service
	appService   app.Service
	buildService build.Service
	audits       auditsvc.Service
}

package build

import (
	"net/http"

	"github.com/brewbits-co/releasedesk/internal/services/app"
	auditsvc "github.com/brewbits-co/releasedesk/internal/services/audit"
	"github.com/brewbits-co/releasedesk/internal/services/build"
)

// BuildController defines the interface for managing build operations and rendering appropriate views.
type BuildController interface {
	HandleBuildUpload(w http.ResponseWriter, r *http.Request)
	HandleArtifactDownload(w http.ResponseWriter, r *http.Request)
	RenderBuildList(w http.ResponseWriter, r *http.Request)
	RenderBuildDetails(w http.ResponseWriter, r *http.Request)
	RenderBuildMetadata(w http.ResponseWriter, r *http.Request)
	HandlePublishBuild(w http.ResponseWriter, r *http.Request)
	HandleDeleteBuild(w http.ResponseWriter, r *http.Request)
}

func NewBuildController(service build.Service, appService app.Service, audits auditsvc.Service) BuildController {
	return &buildController{
		service:    service,
		appService: appService,
		audits:     audits,
	}
}

type buildController struct {
	service    build.Service
	appService app.Service
	audits     auditsvc.Service
}

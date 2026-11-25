package app

import (
	"net/http"

	"github.com/brewbits-co/releasedesk/internal/services/app"
	"github.com/brewbits-co/releasedesk/internal/services/build"
	"github.com/brewbits-co/releasedesk/internal/services/platform"
	"github.com/brewbits-co/releasedesk/internal/services/release"
)

type AppController interface {
	HandleCreateApp(w http.ResponseWriter, r *http.Request)
	HandleAppSetupGuide(w http.ResponseWriter, r *http.Request)
	RenderDashboard(w http.ResponseWriter, r *http.Request)
	RenderStub(w http.ResponseWriter, r *http.Request, currentPage string)
}

func NewAppController(service app.Service, platformService platform.Service, releaseService release.Service, buildService build.Service) AppController {
	return &appController{
		service:         service,
		platformService: platformService,
		releaseService:  releaseService,
		buildService:    buildService,
	}
}

type appController struct {
	service         app.Service
	platformService platform.Service
	releaseService  release.Service
	buildService    build.Service
}

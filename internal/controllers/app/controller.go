package app

import (
	"net/http"

	"github.com/brewbits-co/releasedesk/internal/services/app"
	auditsvc "github.com/brewbits-co/releasedesk/internal/services/audit"
	"github.com/brewbits-co/releasedesk/internal/services/auth"
	"github.com/brewbits-co/releasedesk/internal/services/build"
	"github.com/brewbits-co/releasedesk/internal/services/platform"
	"github.com/brewbits-co/releasedesk/internal/services/release"
)

type AppController interface {
	HandleCreateApp(w http.ResponseWriter, r *http.Request)
	HandleAppSetupGuide(w http.ResponseWriter, r *http.Request)
	RenderDashboard(w http.ResponseWriter, r *http.Request)
	RenderStub(w http.ResponseWriter, r *http.Request, currentPage string)
	RenderSettings(w http.ResponseWriter, r *http.Request)
	HandleSettings(w http.ResponseWriter, r *http.Request)
	HandleTestWebhook(w http.ResponseWriter, r *http.Request)
	HandleMintToken(w http.ResponseWriter, r *http.Request)
	HandleRevokeToken(w http.ResponseWriter, r *http.Request)
	HandleAddMember(w http.ResponseWriter, r *http.Request)
	HandleRemoveMember(w http.ResponseWriter, r *http.Request)
	HandleDeleteApp(w http.ResponseWriter, r *http.Request)
	RenderPlatformConfig(w http.ResponseWriter, r *http.Request)
	HandlePlatformConfig(w http.ResponseWriter, r *http.Request)
	RenderGroups(w http.ResponseWriter, r *http.Request)
	HandleGroup(w http.ResponseWriter, r *http.Request)
	RenderAnnouncements(w http.ResponseWriter, r *http.Request)
	HandleAnnouncement(w http.ResponseWriter, r *http.Request)
	RenderFeeds(w http.ResponseWriter, r *http.Request)
	APIListMembers(w http.ResponseWriter, r *http.Request)
	APIPutMember(w http.ResponseWriter, r *http.Request)
	APIMintToken(w http.ResponseWriter, r *http.Request)
	APISettings(w http.ResponseWriter, r *http.Request)
	APITestWebhook(w http.ResponseWriter, r *http.Request)
}

func NewAppController(service app.Service, platformService platform.Service, releaseService release.Service, buildService build.Service, authService auth.Service, audits auditsvc.Service) AppController {
	return &appController{
		service:         service,
		platformService: platformService,
		releaseService:  releaseService,
		buildService:    buildService,
		authService:     authService,
		audits:          audits,
	}
}

type appController struct {
	service         app.Service
	platformService platform.Service
	releaseService  release.Service
	buildService    build.Service
	authService     auth.Service
	audits          auditsvc.Service
}

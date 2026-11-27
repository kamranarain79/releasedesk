package misc

import (
	"net/http"
	"os"

	"github.com/brewbits-co/releasedesk/internal/services/app"
	"github.com/brewbits-co/releasedesk/internal/services/auth"
	"github.com/brewbits-co/releasedesk/internal/services/instance"
)

// MiscController renders organization-level console pages (navbar layout).
type MiscController interface {
	RenderHomepage(w http.ResponseWriter, r *http.Request)
	RenderDownloadPortal(w http.ResponseWriter, r *http.Request)
	RenderPeople(w http.ResponseWriter, r *http.Request)
	RenderAdministration(w http.ResponseWriter, r *http.Request)
	RenderAccountSettings(w http.ResponseWriter, r *http.Request)
	HandleCreateUser(w http.ResponseWriter, r *http.Request)
	HandleChangePassword(w http.ResponseWriter, r *http.Request)
	HandleSMTP(w http.ResponseWriter, r *http.Request)
	HandleRegenerateToken(w http.ResponseWriter, r *http.Request)
	HandleDismissToken(w http.ResponseWriter, r *http.Request)
	HandleBranding(w http.ResponseWriter, r *http.Request)
	HandleProfile(w http.ResponseWriter, r *http.Request)
	HandleAvatar(w http.ResponseWriter, r *http.Request)
}

func NewMiscController(appService app.Service, authService auth.Service, instanceService instance.Service) MiscController {
	return &miscController{
		appService:      appService,
		authService:     authService,
		instanceService: instanceService,
	}
}

type miscController struct {
	appService      app.Service
	authService     auth.Service
	instanceService instance.Service
}

func portalPublicURL() string {
	if url := os.Getenv("RELEASEDESK_PORTAL_URL"); url != "" {
		return url
	}
	return "http://localhost:9090"
}

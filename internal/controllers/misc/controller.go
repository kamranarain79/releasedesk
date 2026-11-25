package misc

import (
	"net/http"
	"os"

	"github.com/brewbits-co/releasedesk/internal/services/app"
	"github.com/brewbits-co/releasedesk/internal/services/auth"
)

// MiscController renders organization-level console pages (navbar layout).
type MiscController interface {
	RenderHomepage(w http.ResponseWriter, r *http.Request)
	RenderDownloadPortal(w http.ResponseWriter, r *http.Request)
	RenderPeople(w http.ResponseWriter, r *http.Request)
	RenderAdministration(w http.ResponseWriter, r *http.Request)
	RenderAccountSettings(w http.ResponseWriter, r *http.Request)
}

func NewMiscController(appService app.Service, authService auth.Service) MiscController {
	return &miscController{
		appService:  appService,
		authService: authService,
	}
}

type miscController struct {
	appService  app.Service
	authService auth.Service
}

func portalPublicURL() string {
	if url := os.Getenv("RELEASEDESK_PORTAL_URL"); url != "" {
		return url
	}
	return "http://localhost:9090"
}

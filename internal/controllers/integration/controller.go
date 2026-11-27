package integration

import (
	"net/http"

	"github.com/brewbits-co/releasedesk/internal/services/app"
	"github.com/brewbits-co/releasedesk/internal/services/build"
	"github.com/brewbits-co/releasedesk/internal/services/release"
)

// Controller is the token-auth API used by the production Python tool
// (Filebrowser-compatible listing + release notes + QA test methods).
type Controller interface {
	ListApps(w http.ResponseWriter, r *http.Request)
	GetRelease(w http.ResponseWriter, r *http.Request)
	ListReleaseFiles(w http.ResponseWriter, r *http.Request)
	DownloadReleaseFile(w http.ResponseWriter, r *http.Request)
	GetReleaseNotes(w http.ResponseWriter, r *http.Request)
	GetTestMethods(w http.ResponseWriter, r *http.Request)
	LatestRelease(w http.ResponseWriter, r *http.Request)
	DownloadByChecksum(w http.ResponseWriter, r *http.Request)
}

func NewController(appService app.Service, releaseService release.Service, buildService build.Service) Controller {
	return &controller{
		apps:     appService,
		releases: releaseService,
		builds:   buildService,
	}
}

type controller struct {
	apps     app.Service
	releases release.Service
	builds   build.Service
}

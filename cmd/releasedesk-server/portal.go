package main

import (
	"io"
	"log"
	"net/http"
	"os"

	appSrv "github.com/brewbits-co/releasedesk/internal/services/app"
	buildSrv "github.com/brewbits-co/releasedesk/internal/services/build"
	"github.com/brewbits-co/releasedesk/internal/values"
	"github.com/brewbits-co/releasedesk/internal/views"
	"github.com/brewbits-co/releasedesk/pkg/storage"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/dig"
)

type portalHomeData struct {
	PortalURL  string
	PublicApps []portalApp
}

type portalApp struct {
	Name        string
	Slug        string
	Description string
	HasAndroid  bool
	HasIOS      bool
	HasWindows  bool
	HasLinux    bool
	HasMacOS    bool
	HasESP32    bool
	HasSTM32    bool
}

func buildPortal(container *dig.Container) *http.Server {
	router := chi.NewRouter()
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(middleware.NoCache)

	err := container.Invoke(func(appService appSrv.Service, buildService buildSrv.Service) {
		router.Get("/branding/logo", serveBrandLogo)
		router.Get("/branding/favicon", serveBrandFavicon)
		router.Get("/", func(w http.ResponseWriter, r *http.Request) {
			renderPortalHome(w, r, appService)
		})
		router.Get("/apps/{slug}", func(w http.ResponseWriter, r *http.Request) {
			renderPortalApp(w, r, appService, buildService)
		})
		router.Get("/apps/{slug}/files/{checksum}", func(w http.ResponseWriter, r *http.Request) {
			servePortalFile(w, r, appService, buildService)
		})
	})
	if err != nil {
		log.Fatalf("Failed to invoke portal container: %v", err)
	}

	router.Handle("/assets/*", http.FileServer(http.FS(assetsFS)))

	server := &http.Server{
		Addr:    ":9090",
		Handler: router,
	}
	return server
}

func renderPortalHome(w http.ResponseWriter, r *http.Request, appService appSrv.Service) {
	apps, err := appService.GetPublicApps()
	if err != nil {
		log.Println(err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	publicApps := make([]portalApp, 0, len(apps))
	for _, application := range apps {
		description := ""
		if application.Description.Valid {
			description = application.Description.String
		}
		publicApps = append(publicApps, portalApp{
			Name:        application.Name,
			Slug:        string(application.Slug),
			Description: description,
			HasAndroid:  application.HasAndroid,
			HasIOS:      application.HasIOS,
			HasWindows:  application.HasWindows,
			HasLinux:    application.HasLinux,
			HasMacOS:    application.HasMacOS,
			HasESP32:    application.HasESP32,
			HasSTM32:    application.HasSTM32,
		})
	}

	tmpl, err := views.ParseTemplate(views.NoLayout, "templates/portal/home.gohtml")
	if err != nil {
		log.Println(err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	views.Execute(w, tmpl, portalHomeData{
		PortalURL:  portalPublicURL(),
		PublicApps: publicApps,
	})
}

func renderPortalApp(w http.ResponseWriter, r *http.Request, appService appSrv.Service, buildService buildSrv.Service) {
	slug := chi.URLParam(r, "slug")
	entity, err := appService.GetBySlug(values.Slug(slug))
	if err != nil || entity.Private {
		http.NotFound(w, r)
		return
	}
	data, err := appService.GetCurrentAppData(values.Slug(slug))
	if err != nil {
		http.NotFound(w, r)
		return
	}
	type fileRow struct {
		Name     string
		Platform string
		Version  string
		Checksum string
	}
	var files []fileRow
	for _, platformData := range data.Platforms {
		builds, err := buildService.GetPlatformBuilds(platformData.PlatformID)
		if err != nil || len(builds) == 0 {
			continue
		}
		var publishedID int
		for _, item := range builds {
			if item.Status == "Published" {
				publishedID = item.ID
				break
			}
		}
		if publishedID == 0 {
			continue
		}
		full, err := buildService.GetBuildByID(publishedID)
		if err != nil {
			continue
		}
		for _, artifact := range full.Artifacts {
			files = append(files, fileRow{Name: artifact.Filename, Platform: string(platformData.OS), Version: full.Version, Checksum: artifact.Sha256})
		}
	}
	tmpl, err := views.ParseTemplate(views.NoLayout, "templates/portal/app.gohtml")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	views.Execute(w, tmpl, struct {
		Name  string
		Slug  string
		Files []fileRow
	}{Name: entity.Name, Slug: slug, Files: files})
}

func servePortalFile(w http.ResponseWriter, r *http.Request, appService appSrv.Service, buildService buildSrv.Service) {
	slug := chi.URLParam(r, "slug")
	checksum := chi.URLParam(r, "checksum")
	entity, err := appService.GetBySlug(values.Slug(slug))
	if err != nil || entity.Private || !storage.ValidChecksum(checksum) {
		http.NotFound(w, r)
		return
	}
	artifact, err := buildService.GetArtifactByChecksum(checksum)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	full, err := buildService.GetBuildByID(artifact.BuildID)
	if err != nil || full.Status != "Published" {
		http.NotFound(w, r)
		return
	}
	file, err := os.Open(storage.ConvertChecksumToPath(checksum))
	if err != nil {
		http.NotFound(w, r)
		return
	}
	defer file.Close()
	_ = buildService.IncrementDownload(checksum)
	w.Header().Set("Content-Disposition", `attachment; filename="`+artifact.Filename+`"`)
	w.Header().Set("Content-Type", "application/octet-stream")
	_, _ = io.Copy(w, file)
}

package main

import (
	"log"
	"net/http"

	appSrv "github.com/brewbits-co/releasedesk/internal/services/app"
	"github.com/brewbits-co/releasedesk/internal/views"
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
}

func buildPortal(container *dig.Container) *http.Server {
	router := chi.NewRouter()
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(middleware.NoCache)

	err := container.Invoke(func(appService appSrv.Service) {
		router.Get("/", func(w http.ResponseWriter, r *http.Request) {
			renderPortalHome(w, r, appService)
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

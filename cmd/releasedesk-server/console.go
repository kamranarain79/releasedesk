package main

import (
	"github.com/brewbits-co/releasedesk/internal/controllers/app"
	"github.com/brewbits-co/releasedesk/internal/controllers/auth"
	"github.com/brewbits-co/releasedesk/internal/controllers/build"
	"github.com/brewbits-co/releasedesk/internal/controllers/integration"
	"github.com/brewbits-co/releasedesk/internal/controllers/misc"
	"github.com/brewbits-co/releasedesk/internal/controllers/platform"
	"github.com/brewbits-co/releasedesk/internal/controllers/release"
	appSrv "github.com/brewbits-co/releasedesk/internal/services/app"
	instanceSrv "github.com/brewbits-co/releasedesk/internal/services/instance"
	"github.com/brewbits-co/releasedesk/internal/values"
	"github.com/brewbits-co/releasedesk/pkg/middlewares"
	"github.com/brewbits-co/releasedesk/pkg/session"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/jwtauth/v5"
	"go.uber.org/dig"
	"log"
	"net/http"
)

// buildConsole sets up the Chi router with necessary middleware, resolves dependencies, configures routes, and serves static files.
func buildConsole(container *dig.Container) *http.Server {
	router := chi.NewRouter()
	router.Use(middlewares.SafeLogger)
	router.Use(middleware.Recoverer)
	router.Use(middleware.NoCache)

	router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/homepage", http.StatusTemporaryRedirect)
	})
	router.Get("/branding/logo", serveBrandLogo)
	router.Get("/branding/favicon", serveBrandFavicon)
	router.Get("/branding/avatars/{id}", serveAvatar)
	router.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	// Resolve dependencies and set up endpoints and views
	err := container.Invoke(func(
		appService appSrv.Service,
		instanceService instanceSrv.Service,
		authCtrl auth.AuthController,
		miscCtrl misc.MiscController,
		appCtrl app.AppController,
		platformCtrl platform.PlatformController,
		releaseCtrl release.ReleaseController,
		buildCtrl build.BuildController,
		integrationCtrl integration.Controller,
	) {
		// Public Routes
		router.Group(func(r chi.Router) {
			// Public Views
			r.Get("/login", authCtrl.RenderLogin)
			// Public Internal API
			r.Post("/internal/login", authCtrl.HandleLogin)
		})

		// Private Routes
		router.Group(func(r chi.Router) {
			r.Use(jwtauth.Verifier(session.TokenAuth))
			r.Use(middlewares.RedirectOnUnauthorized(session.TokenAuth))
			r.Use(func(next http.Handler) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
					slug := chi.URLParam(req, "slug")
					if slug == "" {
						next.ServeHTTP(w, req)
						return
					}
					userID, _ := req.Context().Value("userID").(int)
					ok, err := appService.CanView(userID, values.Slug(slug))
					if err != nil || !ok {
						http.NotFound(w, req)
						return
					}
					next.ServeHTTP(w, req)
				})
			})

			// Private Views
			r.Get("/homepage", miscCtrl.RenderHomepage)
			r.Get("/portal", miscCtrl.RenderDownloadPortal)
			r.Get("/people", miscCtrl.RenderPeople)
			r.Get("/administration", miscCtrl.RenderAdministration)
			r.Get("/administration/{section}", miscCtrl.RenderAdministration)
			r.Get("/account/settings", miscCtrl.RenderAccountSettings)
			r.Get("/dashboard/{slug}", appCtrl.RenderDashboard)
			r.Get("/dashboard/{slug}/distribution-groups", appCtrl.RenderGroups)
			r.Get("/dashboard/{slug}/announcements", appCtrl.RenderAnnouncements)
			r.Get("/dashboard/{slug}/feeds", appCtrl.RenderFeeds)
			r.Get("/dashboard/{slug}/analytics", func(w http.ResponseWriter, r *http.Request) {
				appCtrl.RenderStub(w, r, "Analytics")
			})
			r.Get("/dashboard/{slug}/settings", appCtrl.RenderSettings)
			r.Get("/dashboard/{slug}/platforms/{platform}/configs", appCtrl.RenderPlatformConfig)
			r.Get("/api/docs", serveAPIDocs)
			r.Get("/api/openapi.yaml", serveOpenAPI)
			r.Get("/api/v1/apps/{slug}/members", appCtrl.APIListMembers)
			r.Put("/api/v1/apps/{slug}/members", appCtrl.APIPutMember)
			r.Post("/api/v1/apps/{slug}/tokens", appCtrl.APIMintToken)
			r.Put("/api/v1/apps/{slug}/settings", appCtrl.APISettings)
			r.Post("/api/v1/apps/{slug}/webhooks/test", appCtrl.APITestWebhook)
			r.Get("/dashboard/{slug}/releases", releaseCtrl.RenderReleaseList)
			r.Get("/dashboard/{slug}/releases/{version}", releaseCtrl.RenderReleaseSummary)
			r.Get("/dashboard/{slug}/releases/{version}/release-notes", releaseCtrl.RenderReleaseNotes)
			r.Get("/dashboard/{slug}/releases/{version}/test-methods", releaseCtrl.RenderTestMethods)
			r.Get("/dashboard/{slug}/platforms/{platform}/builds", buildCtrl.RenderBuildList)
			r.Get("/dashboard/{slug}/platforms/{platform}/builds/{number}", buildCtrl.RenderBuildDetails)
			r.Get("/dashboard/{slug}/platforms/{platform}/builds/{number}/metadata", buildCtrl.RenderBuildMetadata)
			// Private Internal API
			r.Post("/internal/logout", authCtrl.HandleLogout)
			r.Post("/internal/apps", appCtrl.HandleCreateApp)
			r.Post("/internal/apps/{slug}/setup", appCtrl.HandleAppSetupGuide)
			r.Post("/internal/apps/{slug}/platforms", platformCtrl.HandleAddPlatform)
			r.Post("/internal/apps/{slug}/settings", appCtrl.HandleSettings)
			r.Post("/internal/apps/{slug}/delete", appCtrl.HandleDeleteApp)
			r.Post("/internal/apps/{slug}/webhooks/test", appCtrl.HandleTestWebhook)
			r.Post("/internal/apps/{slug}/tokens", appCtrl.HandleMintToken)
			r.Post("/internal/apps/{slug}/tokens/{id}/revoke", appCtrl.HandleRevokeToken)
			r.Post("/internal/apps/{slug}/members", appCtrl.HandleAddMember)
			r.Post("/internal/apps/{slug}/members/{userID}/remove", appCtrl.HandleRemoveMember)
			r.Post("/internal/apps/{slug}/platforms/{platform}/configs", appCtrl.HandlePlatformConfig)
			r.Post("/internal/apps/{slug}/distribution-groups", appCtrl.HandleGroup)
			r.Post("/internal/apps/{slug}/announcements", appCtrl.HandleAnnouncement)
			r.Post("/internal/people", miscCtrl.HandleCreateUser)
			r.Post("/internal/account/password", miscCtrl.HandleChangePassword)
			r.Post("/internal/administration/smtp", miscCtrl.HandleSMTP)
			r.Post("/internal/administration/token", miscCtrl.HandleRegenerateToken)
			r.Post("/internal/administration/token/dismiss", miscCtrl.HandleDismissToken)
			r.Post("/internal/administration/branding", miscCtrl.HandleBranding)
			r.Post("/internal/account/profile", miscCtrl.HandleProfile)
			r.Post("/internal/account/avatar", miscCtrl.HandleAvatar)
			r.Post("/internal/releases/{id}/builds", releaseCtrl.HandleLinkBuilds)
			r.Post("/internal/releases/{id}/publish", releaseCtrl.HandlePublishRelease)
			r.Post("/internal/releases/{id}/delete", releaseCtrl.HandleDeleteRelease)
			r.Post("/internal/apps/{slug}/platforms/{platform}/builds/{number}/publish", buildCtrl.HandlePublishBuild)
			r.Post("/internal/apps/{slug}/platforms/{platform}/builds/{number}/delete", buildCtrl.HandleDeleteBuild)
			r.Post("/internal/apps/{slug}/releases", releaseCtrl.HandleCreateRelease)
			r.Put("/internal/releases/{id}", releaseCtrl.HandleUpdateBasicInfo)
			r.Post("/internal/releases/{id}/release-notes", releaseCtrl.HandleSaveReleaseNotes)
			r.Post("/internal/releases/{id}/test-methods", releaseCtrl.HandleSaveTestPlan)
			r.Post("/internal/apps/{slug}/platforms/{platform}/builds", buildCtrl.HandleBuildUpload)
			r.Get("/internal/artifacts/{checksum}", buildCtrl.HandleArtifactDownload)
		})

		// Public API Routes
		router.Group(func(r chi.Router) {
			r.Use(middlewares.APITokenAuthorization(appService, instanceService))

			r.Post("/api/apps/{slug}/platforms/{platform}/builds", buildCtrl.HandleBuildUpload)
			r.Get("/api/v1/apps", integrationCtrl.ListApps)
			r.Post("/api/v1/apps/{slug}/releases", integrationCtrl.CreateRelease)
			r.Get("/api/v1/apps/{slug}/releases/{version}", integrationCtrl.GetRelease)
			r.Put("/api/v1/apps/{slug}/releases/{version}/notes", integrationCtrl.PutReleaseNotes)
			r.Put("/api/v1/apps/{slug}/releases/{version}/test-methods", integrationCtrl.PutTestMethods)
			r.Get("/api/v1/apps/{slug}/releases/{version}/bundle.zip", integrationCtrl.DownloadBundle)
			r.Get("/api/v1/apps/{slug}/feed.json", integrationCtrl.Feed)
			r.Get("/api/v1/apps/{slug}/releases/{version}/files", integrationCtrl.ListReleaseFiles)
			r.Get("/api/v1/apps/{slug}/releases/{version}/files/{name}", integrationCtrl.DownloadReleaseFile)
			r.Get("/api/v1/apps/{slug}/releases/{version}/notes", integrationCtrl.GetReleaseNotes)
			r.Get("/api/v1/apps/{slug}/releases/{version}/test-methods", integrationCtrl.GetTestMethods)
			r.Get("/api/v1/apps/{slug}/channels/{channel}/latest", integrationCtrl.LatestRelease)
			r.Get("/api/v1/files/{checksum}", integrationCtrl.DownloadByChecksum)
		})
	})
	if err != nil {
		log.Fatalf("Failed to invoke container: %v", err)
	}

	// Handle static files from the assets folder
	router.Handle("/assets/*", http.FileServer(http.FS(assetsFS)))

	// Create HTTP server
	server := &http.Server{
		Addr:    ":8080",
		Handler: router,
	}
	return server
}

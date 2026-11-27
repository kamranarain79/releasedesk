package integration

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strings"

	domainbuild "github.com/brewbits-co/releasedesk/internal/domains/build"
	"github.com/brewbits-co/releasedesk/internal/domains/release"
	"github.com/brewbits-co/releasedesk/internal/values"
	"github.com/brewbits-co/releasedesk/pkg/storage"
	"github.com/go-chi/chi/v5"
)

type fileItem struct {
	Name      string `json:"name"`
	Path      string `json:"path"`
	Size      int64  `json:"size"`
	IsDir     bool   `json:"isDir"`
	Kind      string `json:"kind"`
	Platform  string `json:"platform"`
	Arch      string `json:"architecture,omitempty"`
	Sha256    string `json:"sha256"`
	URL       string `json:"url"`
	FlashHint string `json:"flashHint,omitempty"`
}

type resourceListing struct {
	Name  string     `json:"name"`
	Path  string     `json:"path"`
	IsDir bool       `json:"isDir"`
	Items []fileItem `json:"items"`
}

func (c *controller) authorize(r *http.Request, slug string) (bool, int) {
	entity, err := c.apps.GetBySlug(values.Slug(slug))
	if err != nil {
		return false, 0
	}
	if r.Context().Value("tokenGlobal") == true {
		return true, entity.ID
	}
	if id, ok := r.Context().Value("tokenAppID").(int); ok && id == entity.ID {
		return true, entity.ID
	}
	return false, 0
}

func (c *controller) ListApps(w http.ResponseWriter, r *http.Request) {
	apps, err := c.apps.GetUserAccessibleApps(0)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	type appRow struct {
		Name     string `json:"name"`
		Slug     string `json:"slug"`
		Private  bool   `json:"private"`
		HasESP32 bool   `json:"hasEsp32"`
		HasSTM32 bool   `json:"hasStm32"`
		HasLinux bool   `json:"hasLinux"`
	}
	out := make([]appRow, 0, len(apps))
	for _, application := range apps {
		if id, ok := r.Context().Value("tokenAppID").(int); ok && application.ID != id {
			continue
		}
		out = append(out, appRow{
			Name:     application.Name,
			Slug:     string(application.Slug),
			Private:  application.Private,
			HasESP32: application.HasESP32,
			HasSTM32: application.HasSTM32,
			HasLinux: application.HasLinux,
		})
	}
	writeJSON(w, http.StatusOK, out)
}

func (c *controller) GetRelease(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	version := chi.URLParam(r, "version")
	ok, appID := c.authorize(r, slug)
	if !ok {
		http.NotFound(w, r)
		return
	}
	rel, err := c.releases.GetReleaseSummary(appID, version)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	notes, _ := c.releases.GetReleaseNotes(rel.ID)
	plan, _ := c.releases.GetTestPlan(rel.ID)
	files, _ := c.collectFiles(slug, version, appID)

	writeJSON(w, http.StatusOK, map[string]any{
		"app":         slug,
		"version":     rel.Version,
		"status":      rel.Status,
		"releaseId":   rel.ID,
		"notes":       notes.Text,
		"changelogs":  notes.Changelogs,
		"qaOverview":  plan.Overview,
		"testMethods": plan.Methods,
		"files":       files,
	})
}

func (c *controller) ListReleaseFiles(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	version := chi.URLParam(r, "version")
	ok, appID := c.authorize(r, slug)
	if !ok {
		http.NotFound(w, r)
		return
	}
	files, err := c.collectFiles(slug, version, appID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, resourceListing{
		Name:  version,
		Path:  "/api/v1/apps/" + slug + "/releases/" + version + "/files",
		IsDir: true,
		Items: files,
	})
}

func (c *controller) DownloadReleaseFile(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	version := chi.URLParam(r, "version")
	name := chi.URLParam(r, "name")
	ok, appID := c.authorize(r, slug)
	if !ok {
		http.NotFound(w, r)
		return
	}
	files, err := c.collectFiles(slug, version, appID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	for _, item := range files {
		if item.Name == name {
			c.serveStoredFile(w, item.Sha256, item.Name)
			return
		}
	}
	http.NotFound(w, r)
}

func (c *controller) GetReleaseNotes(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	version := chi.URLParam(r, "version")
	ok, appID := c.authorize(r, slug)
	if !ok {
		http.NotFound(w, r)
		return
	}
	rel, err := c.releases.GetReleaseSummary(appID, version)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	notes, err := c.releases.GetReleaseNotes(rel.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, notes)
}

func (c *controller) GetTestMethods(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	version := chi.URLParam(r, "version")
	ok, appID := c.authorize(r, slug)
	if !ok {
		http.NotFound(w, r)
		return
	}
	rel, err := c.releases.GetReleaseSummary(appID, version)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	plan, err := c.releases.GetTestPlan(rel.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, plan)
}

func (c *controller) LatestRelease(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	channelName := chi.URLParam(r, "channel")
	ok, appID := c.authorize(r, slug)
	if !ok {
		http.NotFound(w, r)
		return
	}
	channels, err := c.releases.GetReleaseChannels(appID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	var channelID int
	for _, ch := range channels {
		if strings.EqualFold(ch.Name, channelName) {
			channelID = ch.ID
			break
		}
	}
	if channelID == 0 {
		http.NotFound(w, r)
		return
	}
	releases, err := c.releases.ListReleasesByChannel(appID, channelID)
	if err != nil || len(releases) == 0 {
		http.NotFound(w, r)
		return
	}
	version := ""
	for _, item := range releases {
		if item.Status == release.Published {
			version = item.Version
			break
		}
	}
	if version == "" {
		http.NotFound(w, r)
		return
	}
	http.Redirect(w, r, "/api/v1/apps/"+slug+"/releases/"+version, http.StatusTemporaryRedirect)
}

func (c *controller) DownloadByChecksum(w http.ResponseWriter, r *http.Request) {
	checksum := chi.URLParam(r, "checksum")
	if !storage.ValidChecksum(checksum) {
		http.Error(w, "Invalid checksum", http.StatusBadRequest)
		return
	}
	filename := checksum
	artifact, err := c.builds.GetArtifactByChecksum(checksum)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	filename = artifact.Filename
	if id, ok := r.Context().Value("tokenAppID").(int); ok {
		entity, err := c.apps.GetByID(id)
		if err != nil {
			http.NotFound(w, r)
			return
		}
		data, err := c.apps.GetCurrentAppData(entity.Slug)
		if err != nil {
			http.NotFound(w, r)
			return
		}
		full, err := c.builds.GetBuildByID(artifact.BuildID)
		if err != nil {
			http.NotFound(w, r)
			return
		}
		allowed := false
		for _, platformData := range data.Platforms {
			if platformData.PlatformID == full.PlatformID {
				allowed = true
			}
		}
		if !allowed {
			http.NotFound(w, r)
			return
		}
	}
	c.serveStoredFile(w, checksum, filename)
}

func (c *controller) collectFiles(slug, version string, appID int) ([]fileItem, error) {
	seen := map[string]bool{}
	usedNames := map[string]bool{}
	var items []fileItem

	addBuild := func(buildID int, os values.OS) {
		if buildID == 0 {
			return
		}
		full, err := c.builds.GetBuildByID(buildID)
		if err != nil {
			return
		}
		for _, artifact := range full.Artifacts {
			if seen[artifact.Sha256] {
				continue
			}
			seen[artifact.Sha256] = true
			item := fileFromArtifact(slug, version, os, artifact)
			if usedNames[item.Name] {
				item.Name = string(os) + "-" + item.Kind + "-" + artifact.Filename
				item.Path = "/api/v1/apps/" + slug + "/releases/" + version + "/files/" + item.Name
			}
			usedNames[item.Name] = true
			items = append(items, item)
		}
	}

	if rel, err := c.releases.GetReleaseSummary(appID, version); err == nil {
		for osName, linked := range rel.Builds {
			addBuild(linked.ID, osName)
		}
	}

	appData, err := c.apps.GetCurrentAppData(values.Slug(slug))
	if err != nil {
		return items, err
	}
	for _, platformData := range appData.Platforms {
		matches, err := c.builds.FindByPlatformAndVersion(platformData.PlatformID, version)
		if err != nil {
			continue
		}
		for _, match := range matches {
			addBuild(match.ID, platformData.OS)
		}
	}
	hints := map[string]string{}
	for _, platformData := range appData.Platforms {
		if note := c.platforms.FlashNotes(platformData.PlatformID); note != "" {
			hints[string(platformData.OS)] = note
		}
	}
	for i := range items {
		items[i].FlashHint = hints[items[i].Platform]
	}
	return items, nil
}

func fileFromArtifact(slug, version string, os values.OS, artifact domainbuild.Artifact) fileItem {
	return fileItem{
		Name:     artifact.Filename,
		Path:     "/api/v1/apps/" + slug + "/releases/" + version + "/files/" + artifact.Filename,
		Size:     int64(artifact.Size),
		IsDir:    false,
		Kind:     string(artifact.Kind),
		Platform: string(os),
		Arch:     string(artifact.Architecture),
		Sha256:   artifact.Sha256,
		URL:      "/api/v1/files/" + artifact.Sha256,
	}
}

func (c *controller) serveStoredFile(w http.ResponseWriter, checksum, filename string) {
	if !storage.ValidChecksum(checksum) {
		http.Error(w, "Invalid checksum", http.StatusBadRequest)
		return
	}
	path := storage.ConvertChecksumToPath(checksum)
	file, err := os.Open(path)
	if err != nil {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}
	defer file.Close()
	_ = c.builds.IncrementDownload(checksum)
	safe := filename
	if safe == "" || strings.ContainsAny(safe, "\"\\/") {
		safe = "download"
	}
	w.Header().Set("Content-Disposition", `attachment; filename="`+safe+`"`)
	w.Header().Set("Content-Type", "application/octet-stream")
	_, _ = io.Copy(w, file)
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

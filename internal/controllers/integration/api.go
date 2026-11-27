package integration

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	domainbuild "github.com/brewbits-co/releasedesk/internal/domains/build"
	"github.com/brewbits-co/releasedesk/internal/values"
	"github.com/brewbits-co/releasedesk/pkg/storage"
	"github.com/go-chi/chi/v5"
)

type fileItem struct {
	Name     string `json:"name"`
	Path     string `json:"path"`
	Size     int64  `json:"size"`
	IsDir    bool   `json:"isDir"`
	Kind     string `json:"kind"`
	Platform string `json:"platform"`
	Arch     string `json:"architecture,omitempty"`
	Sha256   string `json:"sha256"`
	URL      string `json:"url"`
}

type resourceListing struct {
	Name  string     `json:"name"`
	Path  string     `json:"path"`
	IsDir bool       `json:"isDir"`
	Items []fileItem `json:"items"`
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
	appData, err := c.apps.GetCurrentAppData(values.Slug(slug))
	if err != nil {
		http.NotFound(w, r)
		return
	}
	rel, err := c.releases.GetReleaseSummary(appData.AppID, version)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	notes, _ := c.releases.GetReleaseNotes(rel.ID)
	plan, _ := c.releases.GetTestPlan(rel.ID)
	files, _ := c.collectFiles(slug, version, appData.AppID)

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
	appData, err := c.apps.GetCurrentAppData(values.Slug(slug))
	if err != nil {
		http.NotFound(w, r)
		return
	}
	files, err := c.collectFiles(slug, version, appData.AppID)
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
	appData, err := c.apps.GetCurrentAppData(values.Slug(slug))
	if err != nil {
		http.NotFound(w, r)
		return
	}
	files, err := c.collectFiles(slug, version, appData.AppID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	for _, item := range files {
		if item.Name == name {
			serveStoredFile(w, item.Sha256, item.Name)
			return
		}
	}
	http.NotFound(w, r)
}

func (c *controller) GetReleaseNotes(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	version := chi.URLParam(r, "version")
	appData, err := c.apps.GetCurrentAppData(values.Slug(slug))
	if err != nil {
		http.NotFound(w, r)
		return
	}
	rel, err := c.releases.GetReleaseSummary(appData.AppID, version)
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
	appData, err := c.apps.GetCurrentAppData(values.Slug(slug))
	if err != nil {
		http.NotFound(w, r)
		return
	}
	rel, err := c.releases.GetReleaseSummary(appData.AppID, version)
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
	appData, err := c.apps.GetCurrentAppData(values.Slug(slug))
	if err != nil {
		http.NotFound(w, r)
		return
	}
	channels, err := c.releases.GetReleaseChannels(appData.AppID)
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
	releases, err := c.releases.ListReleasesByChannel(appData.AppID, channelID)
	if err != nil || len(releases) == 0 {
		http.NotFound(w, r)
		return
	}
	http.Redirect(w, r, "/api/v1/apps/"+slug+"/releases/"+releases[0].Version, http.StatusTemporaryRedirect)
}

func (c *controller) DownloadByChecksum(w http.ResponseWriter, r *http.Request) {
	checksum := chi.URLParam(r, "checksum")
	filename := checksum
	if artifact, err := c.builds.GetArtifactByChecksum(checksum); err == nil {
		filename = artifact.Filename
	}
	serveStoredFile(w, checksum, filename)
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

func serveStoredFile(w http.ResponseWriter, checksum, filename string) {
	if len(checksum) < 2 {
		http.Error(w, "Invalid checksum", http.StatusBadRequest)
		return
	}
	path := filepath.Join(storage.StorageDir, checksum[:2], checksum)
	file, err := os.Open(path)
	if err != nil {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}
	defer file.Close()
	w.Header().Set("Content-Disposition", `attachment; filename="`+filename+`"`)
	w.Header().Set("Content-Type", "application/octet-stream")
	_, _ = io.Copy(w, file)
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

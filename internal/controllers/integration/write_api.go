package integration

import (
	"archive/zip"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/brewbits-co/releasedesk/internal/domains/release"
	"github.com/brewbits-co/releasedesk/internal/values"
	"github.com/brewbits-co/releasedesk/pkg/storage"
	"github.com/brewbits-co/releasedesk/pkg/teams"
	"github.com/go-chi/chi/v5"
)

func (c *controller) CreateRelease(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	ok, appID := c.authorize(r, slug)
	if !ok {
		http.NotFound(w, r)
		return
	}
	version, channelName := readVersionChannel(r)
	if version == "" {
		http.Error(w, "version is required", http.StatusBadRequest)
		return
	}
	channels, err := c.releases.GetReleaseChannels(appID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	channelID, webhook := pickChannel(channels, channelName)
	if channelID == 0 {
		http.Error(w, "channel not found", http.StatusBadRequest)
		return
	}
	actorName, _ := r.Context().Value("actor").(string)
	created, err := c.releases.CreateRelease(values.Slug(slug), release.BasicInfo{
		Version:       version,
		TargetChannel: channelID,
		CreatedBy:     actorName,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusConflict)
		return
	}
	c.notify(slug, webhook, "Release "+version, slug+" "+version+" created by "+actorName)
	writeJSON(w, http.StatusCreated, map[string]any{
		"id":        created.ID,
		"version":   created.Version,
		"channelId": channelID,
		"createdBy": actorName,
	})
}

func (c *controller) PutReleaseNotes(w http.ResponseWriter, r *http.Request) {
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
	var body struct {
		Text       string              `json:"text"`
		Changelogs []release.Changelog `json:"changelogs"`
	}
	if err = json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	actorName, _ := r.Context().Value("actor").(string)
	notes, err := c.releases.SaveReleaseNotes(rel.ID, body.Text, body.Changelogs, actorName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, notes)
}

func (c *controller) PutTestMethods(w http.ResponseWriter, r *http.Request) {
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
	var body struct {
		Overview string               `json:"overview"`
		Methods  []release.TestMethod `json:"methods"`
	}
	if err = json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	actorName, _ := r.Context().Value("actor").(string)
	plan, err := c.releases.SaveTestPlan(rel.ID, body.Overview, body.Methods, actorName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, plan)
}

func (c *controller) DownloadBundle(w http.ResponseWriter, r *http.Request) {
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
	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", `attachment; filename="`+slug+"-"+version+`.zip"`)
	zw := zip.NewWriter(w)
	defer zw.Close()
	var sums strings.Builder
	for _, item := range files {
		if !storage.ValidChecksum(item.Sha256) {
			continue
		}
		src, err := os.Open(storage.ConvertChecksumToPath(item.Sha256))
		if err != nil {
			continue
		}
		entry, err := zw.Create(item.Name)
		if err != nil {
			src.Close()
			continue
		}
		_, _ = io.Copy(entry, src)
		src.Close()
		sums.WriteString(item.Sha256 + "  " + item.Name + "\n")
	}
	entry, err := zw.Create("sha256sums.txt")
	if err == nil {
		_, _ = io.WriteString(entry, sums.String())
	}
}

func (c *controller) Feed(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	ok, _ := c.authorize(r, slug)
	if !ok {
		http.NotFound(w, r)
		return
	}
	appData, err := c.apps.GetCurrentAppData(values.Slug(slug))
	if err != nil {
		http.NotFound(w, r)
		return
	}
	type row struct {
		Platform string `json:"platform"`
		Version  string `json:"version"`
		Number   string `json:"number"`
	}
	var latest []row
	for _, platformData := range appData.Platforms {
		builds, err := c.builds.GetPlatformBuilds(platformData.PlatformID)
		if err != nil || len(builds) == 0 {
			continue
		}
		for _, item := range builds {
			if item.Status != "Published" {
				continue
			}
			latest = append(latest, row{Platform: string(platformData.OS), Version: item.Version, Number: item.Number})
			break
		}
	}
	writeJSON(w, http.StatusOK, map[string]any{"app": slug, "platforms": latest})
}

func (c *controller) notify(slug, channelWebhook, title, text string) {
	entity, err := c.apps.GetBySlug(values.Slug(slug))
	if err != nil {
		return
	}
	webhook := channelWebhook
	if webhook == "" {
		webhook = entity.TeamsWebhookURL
	}
	if webhook == "" {
		return
	}
	go func() {
		_ = teams.Post(context.Background(), webhook, title, text)
	}()
}

func readVersionChannel(r *http.Request) (string, string) {
	if strings.Contains(r.Header.Get("Content-Type"), "json") {
		var body struct {
			Version string `json:"version"`
			Channel string `json:"channel"`
		}
		_ = json.NewDecoder(r.Body).Decode(&body)
		return body.Version, body.Channel
	}
	_ = r.ParseForm()
	version := r.Form.Get("Version")
	if version == "" {
		version = r.Form.Get("version")
	}
	channel := r.Form.Get("Channel")
	if channel == "" {
		channel = r.Form.Get("channel")
	}
	return version, channel
}

func pickChannel(channels []release.Channel, name string) (int, string) {
	for _, ch := range channels {
		if name == "" || strings.EqualFold(ch.Name, name) {
			return ch.ID, ch.TeamsWebhookURL
		}
	}
	return 0, ""
}

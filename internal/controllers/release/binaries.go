package release

import (
	"net/http"
	"strconv"
	"strings"

	domainbuild "github.com/brewbits-co/releasedesk/internal/domains/build"
	domainrelease "github.com/brewbits-co/releasedesk/internal/domains/release"
	"github.com/brewbits-co/releasedesk/internal/values"
	"github.com/brewbits-co/releasedesk/pkg/session"
)

func (c *releaseController) userID(r *http.Request) int {
	id, _ := r.Context().Value("userID").(int)
	return id
}

func (c *releaseController) actor(r *http.Request) string {
	if value, ok := r.Context().Value("actor").(string); ok && value != "" {
		return value
	}
	return "unknown"
}

func (c *releaseController) canEdit(r *http.Request, slug string) bool {
	ok, err := c.appService.CanEditContent(c.userID(r), values.Slug(slug))
	return err == nil && ok
}

func (c *releaseController) allowContentEdit(w http.ResponseWriter, r *http.Request, releaseID int) bool {
	rel, err := c.service.GetReleaseByID(releaseID)
	if err != nil {
		http.NotFound(w, r)
		return false
	}
	application, err := c.appService.GetByID(rel.AppID)
	if err != nil || !c.canEdit(r, string(application.Slug)) {
		http.Error(w, "forbidden", http.StatusForbidden)
		return false
	}
	return true
}

func (c *releaseController) redirectToRelease(w http.ResponseWriter, r *http.Request, releaseID int, suffix string) {
	rel, err := c.service.GetReleaseByID(releaseID)
	if err != nil {
		http.Redirect(w, r, "/homepage", http.StatusSeeOther)
		return
	}
	application, err := c.appService.GetByID(rel.AppID)
	if err != nil {
		http.Redirect(w, r, "/homepage", http.StatusSeeOther)
		return
	}
	http.Redirect(w, r, "/dashboard/"+string(application.Slug)+"/releases/"+rel.Version+suffix, http.StatusSeeOther)
}

func (c *releaseController) editing(r *http.Request, canEdit bool) bool {
	return canEdit && r.URL.Query().Get("edit") == "1"
}

func (c *releaseController) binaries(app session.CurrentAppData, rel domainrelease.Release) []ReleaseBinary {
	linked := map[int]bool{}
	for _, info := range rel.Builds {
		linked[info.ID] = true
	}
	var rows []ReleaseBinary
	for _, platformData := range app.Platforms {
		builds, err := c.buildService.GetPlatformBuilds(platformData.PlatformID)
		if err != nil {
			continue
		}
		for _, info := range builds {
			full, err := c.buildService.GetBuildByID(info.ID)
			if err != nil {
				continue
			}
			included := linked[info.ID]
			if rel.BuildSelection == domainrelease.Last && info.Version == rel.Version {
				included = true
			}
			rows = append(rows, ReleaseBinary{
				BuildID:  info.ID,
				OS:       string(platformData.OS),
				Number:   info.Number,
				Version:  info.Version,
				Files:    artifactNames(full.Artifacts),
				Checksum: firstChecksum(full.Artifacts),
				Included: included,
			})
		}
	}
	return rows
}

func artifactNames(artifacts []domainbuild.Artifact) string {
	names := make([]string, 0, len(artifacts))
	for _, artifact := range artifacts {
		if artifact.Filename != "" {
			names = append(names, artifact.Filename)
		}
	}
	return strings.Join(names, ", ")
}

func firstChecksum(artifacts []domainbuild.Artifact) string {
	for _, artifact := range artifacts {
		if artifact.Sha256 != "" {
			return artifact.Sha256
		}
	}
	return ""
}

func parseBuildLinks(rawValues []string) []domainrelease.LinkedBuilds {
	var links []domainrelease.LinkedBuilds
	for _, raw := range rawValues {
		parts := strings.SplitN(raw, "|", 2)
		if len(parts) != 2 {
			continue
		}
		buildID, err := strconv.Atoi(parts[0])
		if err != nil || buildID == 0 || parts[1] == "" {
			continue
		}
		links = append(links, domainrelease.LinkedBuilds{BuildID: buildID, OS: values.OS(parts[1])})
	}
	return links
}

func (c *releaseController) HandleLinkBuilds(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "binaries are chosen when the release is created", http.StatusConflict)
}

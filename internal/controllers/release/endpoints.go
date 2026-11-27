package release

import (
	"github.com/brewbits-co/releasedesk/internal/domains/release"
	"github.com/brewbits-co/releasedesk/internal/values"
	"github.com/brewbits-co/releasedesk/pkg/schemas"
	"github.com/brewbits-co/releasedesk/pkg/utils"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"log"
	"net/http"
)

func (c *releaseController) HandleCreateRelease(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		render.Status(r, http.StatusInternalServerError)
		render.PlainText(w, r, err.Error())
		return
	}

	slug := chi.URLParam(r, "slug")

	var releaseCreationRequest release.BasicInfo

	err = utils.NewDecoder().Decode(&releaseCreationRequest, r.Form)
	if err != nil {
		render.Status(r, http.StatusInternalServerError)
		render.PlainText(w, r, err.Error())
		return
	}
	releaseCreationRequest.CreatedBy = c.actor(r)

	newRelease, err := c.service.CreateRelease(values.Slug(slug), releaseCreationRequest)
	if err != nil {
		render.Status(r, http.StatusConflict)
		render.JSON(w, r, schemas.NewErrorResponse(err.Error(), []string{}))
		return
	}

	links := parseBuildLinks(r.Form["build"])
	if len(links) > 0 {
		if err = c.service.ReplaceLinks(newRelease.ID, links); err != nil {
			render.Status(r, http.StatusConflict)
			render.JSON(w, r, schemas.NewErrorResponse(err.Error(), []string{}))
			return
		}
		if _, err = c.service.UpdateReleaseBasicInfo(release.BasicInfo{
			ID:             newRelease.ID,
			TargetChannel:  newRelease.TargetChannel,
			Status:         newRelease.Status,
			BuildSelection: release.Manual,
		}); err != nil {
			render.Status(r, http.StatusConflict)
			render.JSON(w, r, schemas.NewErrorResponse(err.Error(), []string{}))
			return
		}
	}

	log.Println(newRelease.ID)

	render.JSON(w, r, newRelease)
}

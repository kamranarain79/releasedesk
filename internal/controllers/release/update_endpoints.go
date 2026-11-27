package release

import (
	"net/http"
	"strconv"

	"github.com/brewbits-co/releasedesk/pkg/schemas"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
)

// HandleUpdateBasicInfo rejects changes to channel and status.
// Those are chosen when the release is created.
func (c *releaseController) HandleUpdateBasicInfo(w http.ResponseWriter, r *http.Request) {
	render.Status(r, http.StatusConflict)
	render.JSON(w, r, schemas.NewErrorResponse("The channel and status are fixed when the release is created", []string{}))
}

func (c *releaseController) HandlePublishRelease(w http.ResponseWriter, r *http.Request) {
	releaseID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		http.NotFound(w, r)
		return
	}
	rel, err := c.service.GetReleaseByID(releaseID)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	application, err := c.appService.GetByID(rel.AppID)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	if !c.canEdit(r, string(application.Slug)) {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}
	if err = c.service.PublishRelease(releaseID, c.actor(r)); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	http.Redirect(w, r, "/dashboard/"+string(application.Slug)+"/releases/"+rel.Version, http.StatusSeeOther)
}

func (c *releaseController) HandleDeleteRelease(w http.ResponseWriter, r *http.Request) {
	releaseID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		http.NotFound(w, r)
		return
	}
	rel, err := c.service.GetReleaseByID(releaseID)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	application, err := c.appService.GetByID(rel.AppID)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	if !c.canEdit(r, string(application.Slug)) {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}
	if err = c.service.DeleteRelease(releaseID, c.actor(r)); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	http.Redirect(w, r, "/dashboard/"+string(application.Slug)+"/releases", http.StatusSeeOther)
}

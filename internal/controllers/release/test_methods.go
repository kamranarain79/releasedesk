package release

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/brewbits-co/releasedesk/internal/domains/release"
	"github.com/brewbits-co/releasedesk/pkg/schemas"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
)

func (c *releaseController) HandleSaveTestPlan(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		render.Status(r, http.StatusBadRequest)
		return
	}

	releaseID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, schemas.NewErrorResponse("Invalid release ID", nil))
		return
	}

	overview := r.Form.Get("qa-overview")
	titles := r.Form["Title"]
	steps := r.Form["Steps"]
	expected := r.Form["ExpectedResult"]

	methods := make([]release.TestMethod, 0, len(titles))
	for i, title := range titles {
		title = strings.TrimSpace(title)
		if title == "" {
			continue
		}
		method := release.TestMethod{Title: title}
		if i < len(steps) {
			method.Steps = steps[i]
		}
		if i < len(expected) {
			method.ExpectedResult = expected[i]
		}
		methods = append(methods, method)
	}

	if _, err := c.service.SaveTestPlan(releaseID, overview, methods); err != nil {
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, schemas.NewErrorResponse("Failed to save test methods", []string{err.Error()}))
		return
	}

	http.Redirect(w, r, r.Header.Get("Referer"), http.StatusSeeOther)
}

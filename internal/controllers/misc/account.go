package misc

import (
	"net/http"

	domainapp "github.com/brewbits-co/releasedesk/internal/domains/app"
	"github.com/brewbits-co/releasedesk/internal/values"
	"github.com/brewbits-co/releasedesk/pkg/branding"
)

func (c *miscController) HandleCreateUser(w http.ResponseWriter, r *http.Request) {
	actorID, _ := r.Context().Value("userID").(int)
	if !c.appService.IsInstanceAdmin(actorID) {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	role := values.Viewer
	switch r.Form.Get("role") {
	case "manager":
		role = values.Manager
	case "developer":
		role = values.Developer
	case "tester":
		role = values.Tester
	}
	account, err := c.authService.CreateUser(
		r.Form.Get("username"),
		r.Form.Get("firstName"),
		r.Form.Get("lastName"),
		r.Form.Get("email"),
		r.Form.Get("password"),
		role,
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if slug := r.Form.Get("app"); slug != "" {
		role := domainapp.MemberUser
		if r.Form.Get("appRole") == "admin" {
			role = domainapp.MemberAdmin
		}
		_ = c.appService.AddMember(actorID, values.Slug(slug), account.ID, role)
	}
	http.Redirect(w, r, "/people", http.StatusSeeOther)
}

func (c *miscController) HandleChangePassword(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	userID, _ := r.Context().Value("userID").(int)
	if err := c.authService.ChangePassword(userID, r.Form.Get("current"), r.Form.Get("next")); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	http.Redirect(w, r, "/account/settings", http.StatusSeeOther)
}

func (c *miscController) HandleProfile(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	userID, _ := r.Context().Value("userID").(int)
	if err := c.authService.UpdateProfile(userID, r.Form.Get("firstName"), r.Form.Get("lastName"), r.Form.Get("email")); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	http.Redirect(w, r, "/account/settings", http.StatusSeeOther)
}

func (c *miscController) HandleAvatar(w http.ResponseWriter, r *http.Request) {
	userID, _ := r.Context().Value("userID").(int)
	if err := r.ParseMultipartForm(2 << 20); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if r.Form.Get("clear") == "1" {
		branding.RemoveAvatar(userID)
		http.Redirect(w, r, "/account/settings", http.StatusSeeOther)
		return
	}
	file, header, err := r.FormFile("avatar")
	if err != nil {
		http.Error(w, "choose a photo", http.StatusBadRequest)
		return
	}
	defer file.Close()
	if err = branding.SaveAvatar(userID, file, header.Header.Get("Content-Type")); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	http.Redirect(w, r, "/account/settings", http.StatusSeeOther)
}

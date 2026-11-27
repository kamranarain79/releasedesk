package misc

import (
	"net/http"

	"github.com/brewbits-co/releasedesk/pkg/branding"
	"github.com/brewbits-co/releasedesk/pkg/notificator"
)

func (c *miscController) requireAdmin(w http.ResponseWriter, r *http.Request) bool {
	userID, _ := r.Context().Value("userID").(int)
	if c.appService.IsInstanceAdmin(userID) {
		return true
	}
	http.Error(w, "forbidden", http.StatusForbidden)
	return false
}

func (c *miscController) HandleSMTP(w http.ResponseWriter, r *http.Request) {
	if !c.requireAdmin(w, r) {
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := c.instanceService.SaveSMTP(
		r.Form.Get("host"),
		r.Form.Get("port"),
		r.Form.Get("user"),
		r.Form.Get("password"),
		r.Form.Get("from"),
		r.Form.Get("clearPassword") == "1",
	); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if to := r.Form.Get("testTo"); to != "" {
		if err := notificator.SendEmail(notificator.Email{
			To:      to,
			Subject: "ReleaseDesk test",
			Body:    "Outgoing mail from Administration is working.",
		}); err != nil {
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		}
	}
	http.Redirect(w, r, "/administration/email", http.StatusSeeOther)
}

func (c *miscController) HandleRegenerateToken(w http.ResponseWriter, r *http.Request) {
	if !c.requireAdmin(w, r) {
		return
	}
	if _, err := c.instanceService.RegenerateToken(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/administration/token", http.StatusSeeOther)
}

func (c *miscController) HandleBranding(w http.ResponseWriter, r *http.Request) {
	if !c.requireAdmin(w, r) {
		return
	}
	if err := r.ParseMultipartForm(2 << 20); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	save := func(kind, field string) error {
		if r.Form.Get("clear"+kind) == "1" {
			branding.RemoveBrand(kind)
			return nil
		}
		file, header, err := r.FormFile(field)
		if err != nil {
			return nil
		}
		defer file.Close()
		return branding.SaveImage(kind, file, header.Header.Get("Content-Type"))
	}
	if err := save("logo", "logo"); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := save("favicon", "favicon"); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	http.Redirect(w, r, "/administration/branding", http.StatusSeeOther)
}

func (c *miscController) HandleDismissToken(w http.ResponseWriter, r *http.Request) {
	if !c.requireAdmin(w, r) {
		return
	}
	if err := c.instanceService.DismissToken(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/administration/token", http.StatusSeeOther)
}

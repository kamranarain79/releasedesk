package app

import (
	"context"
	"encoding/json"
	"net/http"

	domainapp "github.com/brewbits-co/releasedesk/internal/domains/app"
	"github.com/brewbits-co/releasedesk/internal/values"
	"github.com/brewbits-co/releasedesk/pkg/teams"
	"github.com/go-chi/chi/v5"
)

func writeAPI(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func (c *appController) APIListMembers(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	members, err := c.service.ListMembers(c.userID(r), values.Slug(slug))
	if err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}
	writeAPI(w, http.StatusOK, members)
}

func (c *appController) APIPutMember(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	var body struct {
		UserID int    `json:"userId"`
		Role   string `json:"role"`
		Remove bool   `json:"remove"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	var err error
	if body.Remove {
		err = c.service.RemoveMember(c.userID(r), values.Slug(slug), body.UserID)
	} else {
		err = c.service.AddMember(c.userID(r), values.Slug(slug), body.UserID, domainapp.MemberRole(body.Role))
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	writeAPI(w, http.StatusOK, map[string]bool{"ok": true})
}

func (c *appController) APIMintToken(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	var body struct {
		Name  string `json:"name"`
		Scope string `json:"scope"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)
	raw, token, err := c.service.MintToken(c.userID(r), values.Slug(slug), body.Name, body.Scope)
	if err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}
	writeAPI(w, http.StatusCreated, map[string]any{
		"id":     token.ID,
		"prefix": token.Prefix,
		"scope":  token.Scope,
		"token":  raw,
	})
}

func (c *appController) APISettings(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	var body struct {
		Description     string `json:"description"`
		Private         bool   `json:"private"`
		TeamsWebhookURL string `json:"teamsWebhookUrl"`
		ClearWebhook    bool   `json:"clearWebhook"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	webhook := body.TeamsWebhookURL
	if body.ClearWebhook {
		webhook = "-"
	}
	entity, err := c.service.UpdateSettings(c.userID(r), values.Slug(slug), body.Description, body.Private, webhook)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	writeAPI(w, http.StatusOK, map[string]any{
		"slug":        entity.Slug,
		"private":     entity.Private,
		"description": entity.Description.String,
		"webhookSet":  entity.TeamsWebhookURL != "",
	})
}

func (c *appController) APITestWebhook(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	if ok, _ := c.service.CanAdmin(c.userID(r), values.Slug(slug)); !ok {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}
	entity, err := c.service.GetBySlug(values.Slug(slug))
	if err != nil {
		http.NotFound(w, r)
		return
	}
	if err = teams.Post(context.Background(), entity.TeamsWebhookURL, "ReleaseDesk test", "Webhook for "+slug+" is working."); err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	writeAPI(w, http.StatusOK, map[string]bool{"ok": true})
}

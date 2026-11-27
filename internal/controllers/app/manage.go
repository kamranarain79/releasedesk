package app

import (
	"context"
	"log"
	"net/http"
	"net/url"
	"strconv"

	"github.com/brewbits-co/releasedesk/pkg/notificator"

	domainapp "github.com/brewbits-co/releasedesk/internal/domains/app"
	"github.com/brewbits-co/releasedesk/internal/domains/release"
	"github.com/brewbits-co/releasedesk/internal/values"
	"github.com/brewbits-co/releasedesk/internal/views"
	"github.com/brewbits-co/releasedesk/pkg/session"
	"github.com/brewbits-co/releasedesk/pkg/teams"
	"github.com/go-chi/chi/v5"
)

func (c *appController) userID(r *http.Request) int {
	id, _ := r.Context().Value("userID").(int)
	return id
}

func (c *appController) RenderSettings(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	currentApp, err := c.service.GetCurrentAppData(values.Slug(slug))
	if err != nil {
		http.NotFound(w, r)
		return
	}
	entity, err := c.service.GetBySlug(values.Slug(slug))
	if err != nil {
		http.NotFound(w, r)
		return
	}
	members, _ := c.service.ListMembers(c.userID(r), values.Slug(slug))
	tokens, _ := c.service.ListTokens(c.userID(r), values.Slug(slug))
	users, _ := c.authService.ListUsers()
	canAdmin, _ := c.service.CanAdmin(c.userID(r), values.Slug(slug))
	names := map[int]string{}
	for _, account := range users {
		names[account.ID] = account.Username
	}
	rows := make([]settingsMember, 0, len(members))
	for _, member := range members {
		name := names[member.UserID]
		if name == "" {
			name = "user " + strconv.Itoa(member.UserID)
		}
		rows = append(rows, settingsMember{UserID: member.UserID, Username: name, Role: member.Role})
	}
	tmpl, err := views.ParseTemplate(views.SidebarLayout, "templates/console/app_settings.gohtml")
	if err != nil {
		log.Println(err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	views.Execute(w, tmpl, settingsPage{
		CurrentPage:    "Settings",
		SessionData:    session.NewSessionData(r.Context()),
		CurrentAppData: currentApp,
		App:            entity,
		Members:        rows,
		Tokens:         tokens,
		Users:          users,
		CanAdmin:       canAdmin,
		NewToken:       r.URL.Query().Get("token"),
		Error:          r.URL.Query().Get("error"),
	})
}

func (c *appController) HandleSettings(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	webhook := r.Form.Get("teamsWebhookUrl")
	if r.Form.Get("clearWebhook") == "1" {
		webhook = "-"
	}
	description := r.Form.Get("description")
	private := r.Form.Get("private") == "1"
	if _, err := c.service.UpdateSettings(c.userID(r), values.Slug(slug), description, private, webhook); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	http.Redirect(w, r, "/dashboard/"+slug+"/settings", http.StatusSeeOther)
}

func (c *appController) HandleTestWebhook(w http.ResponseWriter, r *http.Request) {
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
	http.Redirect(w, r, "/dashboard/"+slug+"/settings", http.StatusSeeOther)
}

func (c *appController) HandleMintToken(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	_ = r.ParseForm()
	name := r.Form.Get("name")
	if name == "" {
		http.Redirect(w, r, "/dashboard/"+slug+"/settings?error="+url.QueryEscape("name is required"), http.StatusSeeOther)
		return
	}
	raw, _, err := c.service.MintToken(c.userID(r), values.Slug(slug), name, r.Form.Get("scope"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}
	http.Redirect(w, r, "/dashboard/"+slug+"/settings?token="+raw, http.StatusSeeOther)
}

func (c *appController) HandleRevokeToken(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))
	if err := c.service.RevokeToken(c.userID(r), values.Slug(slug), id); err != nil {
		http.Redirect(w, r, "/dashboard/"+slug+"/settings?error="+url.QueryEscape(err.Error()), http.StatusSeeOther)
		return
	}
	http.Redirect(w, r, "/dashboard/"+slug+"/settings", http.StatusSeeOther)
}

func (c *appController) HandleAddMember(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	_ = r.ParseForm()
	userID, _ := strconv.Atoi(r.Form.Get("userId"))
	if err := c.service.AddMember(c.userID(r), values.Slug(slug), userID, domainapp.MemberRole(r.Form.Get("role"))); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	http.Redirect(w, r, "/dashboard/"+slug+"/settings", http.StatusSeeOther)
}

func (c *appController) HandleRemoveMember(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	userID, _ := strconv.Atoi(chi.URLParam(r, "userID"))
	if err := c.service.RemoveMember(c.userID(r), values.Slug(slug), userID); err != nil {
		http.Redirect(w, r, "/dashboard/"+slug+"/settings?error="+url.QueryEscape(err.Error()), http.StatusSeeOther)
		return
	}
	http.Redirect(w, r, "/dashboard/"+slug+"/settings", http.StatusSeeOther)
}

func (c *appController) HandleDeleteApp(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	if err := c.service.DeleteApp(c.userID(r), values.Slug(slug)); err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}
	http.Redirect(w, r, "/homepage", http.StatusSeeOther)
}

func (c *appController) RenderPlatformConfig(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	osName := chi.URLParam(r, "platform")
	currentApp, err := c.service.GetCurrentAppData(values.Slug(slug))
	if err != nil {
		http.NotFound(w, r)
		return
	}
	items, _ := c.platformService.ListConfig(values.Slug(slug), values.OS(osName))
	tmpl, err := views.ParseTemplate(views.SidebarLayout, "templates/console/platform_config.gohtml")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	views.Execute(w, tmpl, configPage{
		CurrentPage:    "Configurations",
		SessionData:    session.NewSessionData(r.Context()),
		CurrentAppData: currentApp,
		OS:             osName,
		Items:          items,
	})
}

func (c *appController) HandlePlatformConfig(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	osName := chi.URLParam(r, "platform")
	_ = r.ParseForm()
	if err := c.platformService.SaveConfig(values.Slug(slug), values.OS(osName), "flash_notes", r.Form.Get("flash_notes")); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	http.Redirect(w, r, "/dashboard/"+slug+"/platforms/"+osName+"/configs", http.StatusSeeOther)
}

func (c *appController) RenderGroups(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	currentApp, err := c.service.GetCurrentAppData(values.Slug(slug))
	if err != nil {
		http.NotFound(w, r)
		return
	}
	groups, _ := c.service.ListGroups(c.userID(r), values.Slug(slug))
	tmpl, err := views.ParseTemplate(views.SidebarLayout, "templates/console/distribution_groups.gohtml")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	views.Execute(w, tmpl, groupPage{CurrentPage: "Distribution Groups", SessionData: session.NewSessionData(r.Context()), CurrentAppData: currentApp, Groups: groups})
}

func (c *appController) HandleGroup(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	_ = r.ParseForm()
	if err := c.service.SaveGroup(c.userID(r), values.Slug(slug), r.Form.Get("name"), r.Form.Get("restrict") == "1"); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	http.Redirect(w, r, "/dashboard/"+slug+"/distribution-groups", http.StatusSeeOther)
}

func (c *appController) RenderAnnouncements(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	currentApp, err := c.service.GetCurrentAppData(values.Slug(slug))
	if err != nil {
		http.NotFound(w, r)
		return
	}
	items, _ := c.releaseService.ListAnnouncements(currentApp.AppID)
	tmpl, err := views.ParseTemplate(views.SidebarLayout, "templates/console/announcements.gohtml")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	views.Execute(w, tmpl, announcePage{CurrentPage: "Announcements", SessionData: session.NewSessionData(r.Context()), CurrentAppData: currentApp, Items: items})
}

func (c *appController) HandleAnnouncement(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	currentApp, err := c.service.GetCurrentAppData(values.Slug(slug))
	if err != nil {
		http.NotFound(w, r)
		return
	}
	if ok, _ := c.service.CanAdmin(c.userID(r), values.Slug(slug)); !ok {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}
	_ = r.ParseForm()
	item, err := c.releaseService.SaveAnnouncement(release.Announcement{
		AppID: currentApp.AppID, Version: r.Form.Get("version"), Message: r.Form.Get("message"),
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	entity, _ := c.service.GetBySlug(values.Slug(slug))
	members, _ := c.service.ListMembers(c.userID(r), values.Slug(slug))
	users, _ := c.authService.ListUsers()
	go func() {
		if entity.TeamsWebhookURL != "" {
			_ = teams.Post(context.Background(), entity.TeamsWebhookURL, "Announcement "+item.Version, item.Message)
		}
		if !notificator.Configured() {
			return
		}
		allowed := map[int]bool{}
		for _, member := range members {
			allowed[member.UserID] = true
		}
		for _, account := range users {
			if !allowed[account.ID] || !account.Email.Valid || account.Email.String == "" {
				continue
			}
			_ = notificator.SendEmail(notificator.Email{
				To:      account.Email.String,
				Subject: "Announcement " + item.Version,
				Body:    item.Message,
			})
		}
	}()
	http.Redirect(w, r, "/dashboard/"+slug+"/announcements", http.StatusSeeOther)
}

func (c *appController) RenderFeeds(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	currentApp, err := c.service.GetCurrentAppData(values.Slug(slug))
	if err != nil {
		http.NotFound(w, r)
		return
	}
	tmpl, err := views.ParseTemplate(views.SidebarLayout, "templates/console/feeds.gohtml")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	views.Execute(w, tmpl, feedPage{CurrentPage: "Feeds", SessionData: session.NewSessionData(r.Context()), CurrentAppData: currentApp})
}

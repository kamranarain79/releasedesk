package auth

import (
	"encoding/json"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/brewbits-co/releasedesk/pkg/session"
	"github.com/go-chi/render"
)

var loginHits = struct {
	sync.Mutex
	hits map[string][]time.Time
}{hits: map[string][]time.Time{}}

func allowLogin(remote string) bool {
	host, _, err := net.SplitHostPort(remote)
	if err != nil {
		host = remote
	}
	now := time.Now()
	loginHits.Lock()
	defer loginHits.Unlock()
	prev := loginHits.hits[host]
	recent := make([]time.Time, 0, len(prev)+1)
	for _, ts := range prev {
		if now.Sub(ts) < time.Minute {
			recent = append(recent, ts)
		}
	}
	if len(recent) >= 5 {
		loginHits.hits[host] = recent
		return false
	}
	loginHits.hits[host] = append(recent, now)
	return true
}

func (c *authController) HandleLogin(w http.ResponseWriter, r *http.Request) {
	if !allowLogin(r.RemoteAddr) {
		http.Error(w, "too many login attempts", http.StatusTooManyRequests)
		return
	}
	err := r.ParseForm()
	if err != nil {
		render.Status(r, http.StatusBadRequest)
		return
	}
	username := r.Form.Get("username")
	password := r.Form.Get("password")

	user, err := c.service.Login(username, password)
	if err != nil {
		render.Status(r, http.StatusBadRequest)
		render.PlainText(w, r, err.Error())
		return
	}

	tokenString, err := session.CreateToken(user.ID, user.Username)
	if err != nil {
		render.Status(r, http.StatusInternalServerError)
		render.PlainText(w, r, err.Error())
		return
	}

	cookie := session.CreateLoginCookie(tokenString)

	http.SetCookie(w, &cookie)
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]bool{"mustChangePassword": user.MustChangePassword})
}

func (c *authController) HandleLogout(w http.ResponseWriter, r *http.Request) {
	cookie := session.CreateLogoutCookie()
	http.SetCookie(w, &cookie)
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

package main

import (
	"io"
	"net/http"
	"strconv"

	"github.com/brewbits-co/releasedesk/pkg/branding"
	"github.com/go-chi/chi/v5"
)

func serveBrandLogo(w http.ResponseWriter, r *http.Request) {
	serveBrand(w, r, "logo")
}

func serveBrandFavicon(w http.ResponseWriter, r *http.Request) {
	serveBrand(w, r, "favicon")
}

func serveBrand(w http.ResponseWriter, r *http.Request, kind string) {
	file, contentType, err := branding.OpenBrand(kind)
	if err != nil {
		http.Redirect(w, r, "/assets/logo.png", http.StatusTemporaryRedirect)
		return
	}
	defer file.Close()
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Cache-Control", "no-cache")
	_, _ = io.Copy(w, file)
}

func serveAvatar(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		http.NotFound(w, r)
		return
	}
	file, contentType, err := branding.OpenAvatar(id)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	defer file.Close()
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Cache-Control", "no-cache")
	_, _ = io.Copy(w, file)
}

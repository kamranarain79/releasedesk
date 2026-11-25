package views

import (
	"html/template"
	"log"
	"net/http"
)

// Execute writes a parsed console template to the response.
func Execute(w http.ResponseWriter, tmpl *template.Template, data any) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if tmpl == nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	if err := tmpl.ExecuteTemplate(w, "index.gohtml", data); err != nil {
		log.Println(err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}

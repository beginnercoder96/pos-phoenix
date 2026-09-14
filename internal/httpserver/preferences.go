package httpserver

import (
	"net/http"
	"strings"
)

type preferences struct {
	Language string
	Theme    string
}

func readPreferences(r *http.Request) preferences {
	p := preferences{Language: "en", Theme: "light"}
	if c, err := r.Cookie("language"); err == nil && (c.Value == "en" || c.Value == "id") {
		p.Language = c.Value
	}
	if c, err := r.Cookie("theme"); err == nil && (c.Value == "light" || c.Value == "dark") {
		p.Theme = c.Value
	}
	return p
}

func (s *Server) setPreferences(w http.ResponseWriter, r *http.Request) {
	if !s.validCSRF(r) {
		http.Error(w, "invalid request", http.StatusForbidden)
		return
	}
	language := r.FormValue("language")
	if language != "en" && language != "id" {
		language = "en"
	}
	theme := r.FormValue("theme")
	if theme != "light" && theme != "dark" {
		theme = "light"
	}
	for name, value := range map[string]string{"language": language, "theme": theme} {
		http.SetCookie(w, &http.Cookie{Name: name, Value: value, Path: "/", Secure: s.secure, SameSite: http.SameSiteLaxMode, MaxAge: 31536000})
	}
	redirect := r.FormValue("return_to")
	if !strings.HasPrefix(redirect, "/") || strings.HasPrefix(redirect, "//") {
		redirect = "/"
	}
	http.Redirect(w, r, redirect, http.StatusSeeOther)
}

func localizedData(r *http.Request, data pageData) pageData {
	p := readPreferences(r)
	data.Language, data.Theme = p.Language, p.Theme
	return data
}

func preferenceLabel(language, key string) string { return translate(language, key) }

package server

import (
	"embed"
	"html/template"
	"io/fs"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
)

// loginData holds template data for the login page.
type loginData struct {
	CSRFToken string
}

// NewRouter builds the chi router with all dashboard routes.
func NewRouter(content embed.FS) chi.Router {
	r := chi.NewRouter()
	r.Use(SecurityHeaders)

	// Sub-filesystems from the embed
	staticFS, err := fs.Sub(content, "web/static")
	if err != nil {
		slog.Error("failed to create static sub-FS", "error", err)
	}
	modulesFS, err := fs.Sub(content, "modules")
	if err != nil {
		slog.Error("failed to create modules sub-FS", "error", err)
	}
	i18nFS, err := fs.Sub(content, "i18n")
	if err != nil {
		slog.Error("failed to create i18n sub-FS", "error", err)
	}

	// Static file servers
	r.Handle("/static/*", http.StripPrefix("/static/", http.FileServerFS(staticFS)))
	r.Handle("/modules/*", http.StripPrefix("/modules/", http.FileServerFS(modulesFS)))
	r.Handle("/i18n/*", http.StripPrefix("/i18n/", http.FileServerFS(i18nFS)))

	// Read shell.html once at startup
	shellHTML, err := fs.ReadFile(content, "web/templates/shell.html")
	if err != nil {
		slog.Error("failed to read shell.html", "error", err)
	}

	// Parse login.html as a Go template
	loginTmplBytes, err := fs.ReadFile(content, "web/templates/login.html")
	if err != nil {
		slog.Error("failed to read login.html", "error", err)
	}
	loginTmpl, err := template.New("login").Parse(string(loginTmplBytes))
	if err != nil {
		slog.Error("failed to parse login.html template", "error", err)
	}

	// Dashboard shell — served as raw bytes (no template vars yet)
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write(shellHTML)
	})

	// Login page — rendered via html/template
	r.Get("/login", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if err := loginTmpl.Execute(w, loginData{CSRFToken: ""}); err != nil {
			slog.Error("failed to render login template", "error", err)
		}
	})

	return r
}

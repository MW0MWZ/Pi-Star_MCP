package server

import (
	"embed"
	"html/template"
	"io/fs"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/MW0MWZ/Pi-Star_MCP/internal/config"
	"github.com/MW0MWZ/Pi-Star_MCP/internal/hwdetect"
	"github.com/MW0MWZ/Pi-Star_MCP/internal/server/handlers"
)

// loginData holds template data for the login page.
type loginData struct {
	CSRFToken string
}

// NewRouter builds the chi router with public and admin route groups.
func NewRouter(content embed.FS, cfg *config.Config, configPath string, devices []hwdetect.DetectedDevice, i2cDevices []hwdetect.DetectedI2CDevice) chi.Router {
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

	// Static file servers (public — no auth)
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

	// Read admin.html once at startup
	adminHTML, err := fs.ReadFile(content, "web/templates/admin.html")
	if err != nil {
		slog.Error("failed to read admin.html", "error", err)
	}

	// ── Public routes (no auth) ──────────────────────────

	// Dashboard shell
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write(shellHTML)
	})

	// Login page
	r.Get("/login", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if err := loginTmpl.Execute(w, loginData{CSRFToken: ""}); err != nil {
			slog.Error("failed to render login template", "error", err)
		}
	})

	// Login POST
	r.Post("/login", handlers.LoginPost)

	// Logout (accept both POST and GET for robustness)
	r.Post("/logout", handlers.Logout)
	r.Get("/logout", handlers.Logout)

	// Hardware detection API (public — hardware info is not sensitive)
	hwHandler := &handlers.HardwareHandler{Devices: devices, I2CDevices: i2cDevices}
	r.Get("/api/hardware", hwHandler.ListHardware)

	// ── Admin routes (auth required) ─────────────────────

	svcHandlers := &handlers.ServiceHandlers{
		Cfg:        cfg,
		ConfigPath: configPath,
	}

	radioHandlers := &handlers.RadioHandlers{
		Cfg: cfg,
	}

	r.Route("/admin", func(admin chi.Router) {
		admin.Use(RequireAuth)

		// Admin SPA shell
		admin.Get("/", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.Write(adminHTML)
		})

		// Radio configuration API
		admin.Get("/api/radio/settings", radioHandlers.GetRadioSettings)
		admin.Put("/api/radio/settings", radioHandlers.PutRadioSettings)

		// Service API
		admin.Get("/api/services", svcHandlers.ListServices)
		admin.Put("/api/services/{svc}/enable", svcHandlers.EnableService)
		admin.Put("/api/services/{svc}/disable", svcHandlers.DisableService)
		admin.Get("/api/services/{svc}/settings", svcHandlers.GetServiceSettings)
		admin.Put("/api/services/{svc}/settings", svcHandlers.PutServiceSettings)

		// DStarRepeater hardware type
		admin.Put("/api/dstarrepeater/hwtype", svcHandlers.SetDStarHWType)

		// System API placeholders
		admin.Post("/api/system/reboot", handlers.Placeholder)
		admin.Post("/api/system/shutdown", handlers.Placeholder)
		admin.Post("/api/system/update", handlers.Placeholder)
	})

	return r
}

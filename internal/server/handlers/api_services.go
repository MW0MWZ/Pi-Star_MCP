package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"sort"

	"github.com/go-chi/chi/v5"

	"github.com/MW0MWZ/Pi-Star_MCP/internal/config"
	"github.com/MW0MWZ/Pi-Star_MCP/internal/svcconfig"
)

// ServiceHandlers holds dependencies for service API handlers.
type ServiceHandlers struct {
	Cfg        *config.Config
	ConfigPath string
}

// serviceListItem is the JSON shape returned by ListServices.
type serviceListItem struct {
	Name         string            `json:"name"`
	DisplayName  string            `json:"displayName"`
	Category     string            `json:"category"`
	Enabled      bool              `json:"enabled"`
	HasSettings  bool              `json:"hasSettings"`
	DependsOn    []string          `json:"dependsOn,omitempty"`
	HWType       string            `json:"hwType,omitempty"`       // DStarRepeater only
	HWVariants   []hwVariantItem   `json:"hwVariants,omitempty"`   // DStarRepeater only
}

// hwVariantItem describes a DStarRepeater hardware variant for the UI.
type hwVariantItem struct {
	Key         string `json:"key"`
	DisplayName string `json:"displayName"`
}

// ListServices returns all 22 services grouped by category.
func (h *ServiceHandlers) ListServices(w http.ResponseWriter, r *http.Request) {
	names := config.ServiceNames()
	items := make([]serviceListItem, 0, len(names))

	for _, name := range names {
		def, ok := config.LookupService(name)
		if !ok {
			continue
		}
		entry := h.Cfg.Services[name]
		_, hasSchema := svcconfig.LookupSchema(name)

		item := serviceListItem{
			Name:        name,
			DisplayName: def.DisplayName,
			Category:    categoryString(def.Category),
			Enabled:     entry != nil && entry.Enabled,
			HasSettings: hasSchema,
			DependsOn:   def.DependsOn,
		}

		// Include hardware variant info for DStarRepeater
		if name == "dstarrepeater" && entry != nil {
			item.HWType = entry.HWType
			for _, v := range config.DStarVariants {
				item.HWVariants = append(item.HWVariants, hwVariantItem{
					Key:         v.Key,
					DisplayName: v.DisplayName,
				})
			}
		}

		items = append(items, item)
	}

	// Sort by category order then by name
	sort.Slice(items, func(i, j int) bool {
		ci := categoryOrder(items[i].Category)
		cj := categoryOrder(items[j].Category)
		if ci != cj {
			return ci < cj
		}
		return items[i].Name < items[j].Name
	})

	writeJSON(w, http.StatusOK, items)
}

// EnableService enables a service, checking dependencies first.
func (h *ServiceHandlers) EnableService(w http.ResponseWriter, r *http.Request) {
	svc := chi.URLParam(r, "svc")
	if _, ok := config.LookupService(svc); !ok {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "unknown service"})
		return
	}

	entry := h.Cfg.Services[svc]
	if entry == nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "unknown service"})
		return
	}

	missing := config.MissingDeps(svc, h.Cfg.Services)
	if len(missing) > 0 {
		writeJSON(w, http.StatusConflict, map[string]interface{}{
			"error":      "missing dependencies",
			"missingDeps": missing,
		})
		return
	}

	entry.Enabled = true
	if err := config.SaveServices(h.ConfigPath, h.Cfg.Services); err != nil {
		slog.Error("failed to save services", "error", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to save configuration"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "enabled"})
}

// DisableService disables a service, checking dependents first.
func (h *ServiceHandlers) DisableService(w http.ResponseWriter, r *http.Request) {
	svc := chi.URLParam(r, "svc")
	if _, ok := config.LookupService(svc); !ok {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "unknown service"})
		return
	}

	entry := h.Cfg.Services[svc]
	if entry == nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "unknown service"})
		return
	}

	dependents := config.EnabledDependents(svc, h.Cfg.Services)
	if len(dependents) > 0 {
		writeJSON(w, http.StatusConflict, map[string]interface{}{
			"error":      "service has enabled dependents",
			"dependents": dependents,
		})
		return
	}

	entry.Enabled = false
	if err := config.SaveServices(h.ConfigPath, h.Cfg.Services); err != nil {
		slog.Error("failed to save services", "error", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to save configuration"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "disabled"})
}

// GetServiceSettings returns the schema and current values for a service.
func (h *ServiceHandlers) GetServiceSettings(w http.ResponseWriter, r *http.Request) {
	svc := chi.URLParam(r, "svc")
	schema, ok := svcconfig.LookupSchema(svc)
	if !ok {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "no settings schema for this service"})
		return
	}

	def, _ := config.LookupService(svc)
	entry := h.Cfg.Services[svc]
	iniPath := def.DefaultConfigPath
	if entry != nil && entry.ConfigPath != "" {
		iniPath = entry.ConfigPath
	}

	values, err := svcconfig.ReadSettings(schema, iniPath)
	if err != nil {
		slog.Error("failed to read service settings", "service", svc, "error", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to read settings"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"schema": schema,
		"values": values,
	})
}

// PutServiceSettings validates and writes curated settings for a service.
func (h *ServiceHandlers) PutServiceSettings(w http.ResponseWriter, r *http.Request) {
	svc := chi.URLParam(r, "svc")
	schema, ok := svcconfig.LookupSchema(svc)
	if !ok {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "no settings schema for this service"})
		return
	}

	var values map[string]string
	if err := json.NewDecoder(r.Body).Decode(&values); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON body"})
		return
	}

	// Server-side validation
	errs := svcconfig.ValidateSettings(schema, values)
	if len(errs) > 0 {
		writeJSON(w, http.StatusUnprocessableEntity, map[string]interface{}{
			"error":  "validation failed",
			"fields": errs,
		})
		return
	}

	def, _ := config.LookupService(svc)
	entry := h.Cfg.Services[svc]
	iniPath := def.DefaultConfigPath
	if entry != nil && entry.ConfigPath != "" {
		iniPath = entry.ConfigPath
	}

	if err := svcconfig.WriteSettings(schema, iniPath, values); err != nil {
		slog.Error("failed to write service settings", "service", svc, "error", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to save settings"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "saved"})
}

// SetDStarHWType sets the DStarRepeater hardware variant, updating the
// binary and config paths to match the selected hardware type.
func (h *ServiceHandlers) SetDStarHWType(w http.ResponseWriter, r *http.Request) {
	var body struct {
		HWType string `json:"hwType"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON body"})
		return
	}

	// Validate the key exists
	found := false
	for _, v := range config.DStarVariants {
		if v.Key == body.HWType {
			found = true
			break
		}
	}
	if !found {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "unknown hardware type"})
		return
	}

	entry := h.Cfg.Services["dstarrepeater"]
	if entry == nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "dstarrepeater service not found"})
		return
	}

	entry.HWType = body.HWType
	config.ResolveDStarPaths(entry)

	if err := config.SaveServices(h.ConfigPath, h.Cfg.Services); err != nil {
		slog.Error("failed to save dstarrepeater hwtype", "error", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to save configuration"})
		return
	}

	variant := config.LookupDStarVariant(body.HWType)
	slog.Info("dstarrepeater hardware type changed", "hwType", body.HWType, "binary", variant.BinaryName)
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"status":      "saved",
		"hwType":      body.HWType,
		"displayName": variant.DisplayName,
		"binary":      entry.BinaryPath,
		"config":      entry.ConfigPath,
	})
}

// Placeholder returns a 501 Not Implemented response for deferred features.
func Placeholder(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusNotImplemented, map[string]string{"error": "not implemented"})
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func categoryString(c config.ServiceCategory) string {
	switch c {
	case config.CategoryCore:
		return "core"
	case config.CategoryGateway:
		return "gateway"
	case config.CategoryBridge:
		return "bridge"
	case config.CategoryUtility:
		return "utility"
	default:
		return "unknown"
	}
}

func categoryOrder(s string) int {
	switch s {
	case "core":
		return 0
	case "gateway":
		return 1
	case "bridge":
		return 2
	case "utility":
		return 3
	default:
		return 99
	}
}

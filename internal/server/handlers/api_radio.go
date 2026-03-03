package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/MW0MWZ/Pi-Star_MCP/internal/config"
	"github.com/MW0MWZ/Pi-Star_MCP/internal/svcconfig"
)

// RadioHandlers holds dependencies for radio configuration API handlers.
type RadioHandlers struct {
	Cfg *config.Config
}

// GetRadioSettings returns the radio configuration schema and current
// values read from MMDVM.ini (the primary source of truth).
func (h *RadioHandlers) GetRadioSettings(w http.ResponseWriter, r *http.Request) {
	values, err := svcconfig.ReadRadioConfig(h.Cfg.Services)
	if err != nil {
		slog.Error("failed to read radio config", "error", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to read radio settings"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"schema": svcconfig.RadioSchema(),
		"values": values,
	})
}

// PutRadioSettings validates and writes radio configuration to all
// target INI files (fan-out write).
func (h *RadioHandlers) PutRadioSettings(w http.ResponseWriter, r *http.Request) {
	var values map[string]string
	if err := json.NewDecoder(r.Body).Decode(&values); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON body"})
		return
	}

	errs := svcconfig.ValidateRadioConfig(values)
	if len(errs) > 0 {
		writeJSON(w, http.StatusUnprocessableEntity, map[string]interface{}{
			"error":  "validation failed",
			"fields": errs,
		})
		return
	}

	written, err := svcconfig.WriteRadioConfig(h.Cfg.Services, values)
	if err != nil {
		slog.Error("failed to write radio config", "error", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to save radio settings"})
		return
	}

	slog.Info("radio config saved", "filesWritten", written)
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"status":       "saved",
		"filesWritten": written,
	})
}

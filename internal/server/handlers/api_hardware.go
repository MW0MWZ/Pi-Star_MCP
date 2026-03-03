package handlers

import (
	"net/http"

	"github.com/MW0MWZ/Pi-Star_MCP/internal/hwdetect"
)

// HardwareHandler serves detected hardware information.
type HardwareHandler struct {
	Devices    []hwdetect.DetectedDevice
	I2CDevices []hwdetect.DetectedI2CDevice
}

// ListHardware returns the list of detected devices as JSON.
func (h *HardwareHandler) ListHardware(w http.ResponseWriter, r *http.Request) {
	serial := h.Devices
	if serial == nil {
		serial = []hwdetect.DetectedDevice{}
	}
	i2c := h.I2CDevices
	if i2c == nil {
		i2c = []hwdetect.DetectedI2CDevice{}
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"serial": serial,
		"i2c":    i2c,
	})
}

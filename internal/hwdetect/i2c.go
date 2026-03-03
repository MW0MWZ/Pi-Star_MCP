package hwdetect

import (
	"fmt"
	"log/slog"
	"path/filepath"

	"golang.org/x/sys/unix"
)

// I2C ioctl constants (from linux/i2c-dev.h).
const (
	i2cSlave = 0x0703 // ioctl to set slave address
)

// Known I2C device addresses and their likely identities.
var knownI2CDevices = map[uint8]struct {
	name       string
	deviceType string
}{
	0x3C: {"OLED", "oled"},
	0x3D: {"OLED (alt addr)", "oled"},
	0x27: {"LCD", "lcd"},
	0x20: {"LCD", "lcd"},
}

// OLED controller chip type, mapped to Pi-Star's OLED type numbering.
const (
	OLEDTypeSSD1306 = 3 // Adafruit I2C 128x64 (SSD1306)
	OLEDTypeSH1106  = 6 // SH1106 I2C 128x64
)

// DetectedI2CDevice holds information about a device found on the I2C bus.
type DetectedI2CDevice struct {
	Bus           int    `json:"bus"`
	Address       string `json:"address"`         // hex, e.g. "0x3c"
	DeviceType    string `json:"deviceType"`       // "oled", "lcd", "unknown"
	Name          string `json:"name"`             // human-readable name
	OLEDChip      string `json:"oledChip,omitempty"`    // "SSD1306" or "SH1106"
	OLEDType      int    `json:"oledType,omitempty"`     // Pi-Star OLED type (3 or 6)
	LCDController string `json:"lcdController,omitempty"` // "MCP23017" or "PCF8574"
	LCDCols       int    `json:"lcdCols,omitempty"`       // e.g. 16
	LCDRows       int    `json:"lcdRows,omitempty"`       // e.g. 2
}

// DetectI2C scans I2C buses for known devices.
func DetectI2C() []DetectedI2CDevice {
	var devices []DetectedI2CDevice

	// Scan common I2C bus paths (bus 1 is standard on Pi)
	buses, _ := filepath.Glob("/dev/i2c-*")
	for _, busPath := range buses {
		busNum := 0
		fmt.Sscanf(filepath.Base(busPath), "i2c-%d", &busNum)

		found := scanI2CBus(busPath, busNum)
		devices = append(devices, found...)
	}

	return devices
}

// scanI2CBus probes all valid addresses on a single I2C bus.
func scanI2CBus(busPath string, busNum int) []DetectedI2CDevice {
	fd, err := unix.Open(busPath, unix.O_RDWR, 0)
	if err != nil {
		slog.Debug("cannot open i2c bus", "path", busPath, "error", err)
		return nil
	}
	defer unix.Close(fd)

	var devices []DetectedI2CDevice

	// Scan addresses 0x03-0x77 (standard 7-bit range, excluding reserved)
	for addr := uint8(0x03); addr <= 0x77; addr++ {
		if isReservedI2C(addr) {
			continue
		}

		if probeI2CAddress(fd, addr) {
			dev := DetectedI2CDevice{
				Bus:        busNum,
				Address:    fmt.Sprintf("0x%02x", addr),
				DeviceType: "unknown",
			}

			if known, ok := knownI2CDevices[addr]; ok {
				dev.Name = known.name
				dev.DeviceType = known.deviceType
			}

			// For LCD devices, distinguish MCP23017 (Adafruit plate) from PCF8574
			if dev.DeviceType == "lcd" {
				identifyLCDController(fd, &dev)
			}

			// For OLED devices, detect the controller chip
			if dev.DeviceType == "oled" {
				identifyOLEDChip(fd, &dev)
			}

			devices = append(devices, dev)
		}
	}

	return devices
}

// probeI2CAddress checks if a device responds at the given address.
func probeI2CAddress(fd int, addr uint8) bool {
	// Set slave address
	if err := unix.IoctlSetInt(fd, i2cSlave, int(addr)); err != nil {
		return false
	}

	// Try a zero-length write (SMBus quick write) to detect presence.
	// Some devices don't support this, so also try a single byte read.
	buf := make([]byte, 1)
	_, err := unix.Read(fd, buf)
	return err == nil
}

// identifyLCDController distinguishes MCP23017 from PCF8574 at LCD addresses.
// The slave address must already be set on fd.
func identifyLCDController(fd int, dev *DetectedI2CDevice) {
	if isMCP23017(fd) {
		dev.Name = "Adafruit LCD Plate"
		dev.LCDController = "MCP23017"
		dev.LCDCols = 16
		dev.LCDRows = 2
	} else {
		dev.Name = "PCF8574 LCD"
		dev.LCDController = "PCF8574"
		dev.LCDCols = 16
		dev.LCDRows = 2
	}
}

// identifyOLEDChip distinguishes SSD1306 from SH1106 by reading the I2C status byte.
//
// Detection relies on two observed differences:
//  1. SSD1306 does not support I2C reads — all reads return 0xFF (bus pull-ups).
//  2. SH1106 supports I2C status reads — bit 6 reflects display on/off state,
//     and bits 5-0 carry internal state (non-zero).
//
// The slave address must already be set on fd.
func identifyOLEDChip(fd int, dev *DetectedI2CDevice) {
	const rounds = 3

	// Toggle display off/on and read status each round.
	sh1106Votes := 0
	ssd1306Votes := 0
	for range rounds {
		writeI2CCmd(fd, 0xAE) // Display OFF
		bufOff := make([]byte, 1)
		if _, err := unix.Read(fd, bufOff); err != nil {
			return
		}

		writeI2CCmd(fd, 0xAF) // Display ON
		bufOn := make([]byte, 1)
		if _, err := unix.Read(fd, bufOn); err != nil {
			return
		}

		writeI2CCmd(fd, 0xAE) // Restore display OFF

		// SSD1306: no I2C read support → both bytes 0xFF, no toggle
		if bufOff[0] == 0xFF && bufOn[0] == 0xFF {
			ssd1306Votes++
			continue
		}

		// SH1106: bit 6 toggles with display on/off, lower bits non-zero
		if (bufOff[0]^bufOn[0])&0x40 != 0 && ((bufOff[0]&0x3F) != 0 || (bufOn[0]&0x3F) != 0) {
			sh1106Votes++
			continue
		}

		slog.Debug("oled status byte unexpected",
			"off", fmt.Sprintf("0x%02X", bufOff[0]),
			"on", fmt.Sprintf("0x%02X", bufOn[0]))
	}

	if ssd1306Votes == rounds {
		dev.OLEDChip = "SSD1306"
		dev.OLEDType = OLEDTypeSSD1306
		dev.Name = "SSD1306 OLED 128x64"
	} else if sh1106Votes == rounds {
		dev.OLEDChip = "SH1106"
		dev.OLEDType = OLEDTypeSH1106
		dev.Name = "SH1106 OLED 128x64"
	} else {
		slog.Debug("oled chip detection inconclusive",
			"ssd1306", ssd1306Votes, "sh1106", sh1106Votes, "rounds", rounds)
	}
}

// writeI2CCmd sends a single command byte to an I2C OLED display.
// The 0x00 prefix byte indicates a command (Co=0, D/C#=0).
func writeI2CCmd(fd int, cmd byte) {
	unix.Write(fd, []byte{0x00, cmd})
}

// isReservedI2C returns true for I2C addresses reserved by the spec.
func isReservedI2C(addr uint8) bool {
	// 0x00-0x02: general call, CBUS, reserved
	// 0x78-0x7F: 10-bit addressing, reserved
	return addr < 0x03 || addr > 0x77
}

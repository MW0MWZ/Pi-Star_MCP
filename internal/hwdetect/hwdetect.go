// Package hwdetect detects serial hardware connected to a Pi-Star hotspot.
//
// It enumerates /dev/ttyAMA*, /dev/ttyACM*, and /dev/ttyUSB* ports, reads
// USB metadata from sysfs, probes for MMDVM modems and Nextion displays,
// and classifies each port.
package hwdetect

import (
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// DeviceType identifies the class of a detected serial device.
type DeviceType string

const (
	DeviceUnknown DeviceType = "unknown"
	DeviceMMDVM   DeviceType = "mmdvm"
	DeviceNextion DeviceType = "nextion"
	DeviceDVAP    DeviceType = "dvap"
	DeviceGMSK    DeviceType = "gmsk"
)

// DetectedDevice holds all information gathered about a serial port.
type DetectedDevice struct {
	Port               string     `json:"port"`
	DeviceType         DeviceType `json:"deviceType"`
	USBVendor          string     `json:"usbVendor,omitempty"`
	USBProduct         string     `json:"usbProduct,omitempty"`
	USBName            string     `json:"usbName,omitempty"`
	Driver             string     `json:"driver,omitempty"`
	MMDVMDescription   string     `json:"mmdvmDescription,omitempty"`
	MMDVMHWType        string     `json:"mmdvmHwType,omitempty"`
	MMDVMProtocol      int        `json:"mmdvmProtocol,omitempty"`
	NextionModel       string     `json:"nextionModel,omitempty"`
	NextionSerial      string     `json:"nextionSerial,omitempty"`
	DVMegaFirmware     string     `json:"dvmegaFirmware,omitempty"`
	DVMegaHardware     string     `json:"dvmegaHardware,omitempty"`
}

// Known USB VID:PID mappings for quick identification.
var knownUSBDevices = map[string]struct {
	name       string
	likelyType DeviceType
}{
	"1eaf:0004": {"STM32/Maple (MMDVM)", DeviceMMDVM},
	"1a86:7523": {"CH340/CH341", DeviceUnknown}, // probe needed: could be Nextion or other
	"0403:6001": {"FTDI FT232R", DeviceUnknown},  // probe needed: could be DVAP, GMSK, or other
	"10c4:ea60": {"CP210x", DeviceUnknown},        // probe needed
	"1fc9:0083": {"NXP LPC (MMDVM)", DeviceMMDVM},
	"0483:5740": {"STM32 VCP (MMDVM)", DeviceMMDVM},
}

// DetectAll enumerates serial ports and probes each one.
func DetectAll() []DetectedDevice {
	// Reset GPIO-connected modems (MMDVM_HS hats, DV-Mega on GPIO)
	// before any probing, matching pistar-findmodem.
	resetGPIOModem()

	ports := enumeratePorts()
	devices := make([]DetectedDevice, 0, len(ports))

	// DTR-reset any USB serial ports to put Arduino-based devices (DV-Mega)
	// into a known state. The reset triggers the bootloader; by the time we
	// finish resetting all ports and start probing, it will have timed out.
	for _, port := range ports {
		if strings.HasPrefix(filepath.Base(port), "ttyUSB") {
			resetDTR(port)
		}
	}
	// Let Arduino bootloaders finish (Optiboot=250ms, Mega=1000ms)
	time.Sleep(1500 * time.Millisecond)

	for _, port := range ports {
		dev := DetectedDevice{Port: port}

		// Read USB metadata from sysfs (fast, no port open needed)
		ttyName := filepath.Base(port)
		readUSBInfo(ttyName, &dev)

		// Probe order matches pistar-findmodem: MMDVM first, then DV-Mega.
		// MMDVM tries 115200, 230400, 460800. DV-Mega uses 115200 only.

		// 1. MMDVM modem (most common, tries 3 baud rates)
		t0 := time.Now()
		mmdvm, err := ProbeMMDVM(port)
		slog.Debug("probe timing", "port", port, "probe", "mmdvm", "elapsed", time.Since(t0).Round(time.Millisecond))
		if err != nil {
			slog.Debug("mmdvm probe failed", "port", port, "error", err)
		}
		if mmdvm != nil {
			dev.DeviceType = DeviceMMDVM
			dev.MMDVMDescription = mmdvm.Description
			dev.MMDVMHWType = mmdvm.HWType
			dev.MMDVMProtocol = mmdvm.Protocol
			devices = append(devices, dev)
			continue
		}

		// 2. DV-Mega GMSK modem (DVRPTR protocol, 115200 only)
		t0 = time.Now()
		dvmega, err := ProbeDVMega(port)
		slog.Debug("probe timing", "port", port, "probe", "dvmega", "elapsed", time.Since(t0).Round(time.Millisecond))
		if err != nil {
			slog.Debug("dvmega probe failed", "port", port, "error", err)
		}
		if dvmega != nil {
			dev.DeviceType = DeviceGMSK
			dev.DVMegaFirmware = dvmega.FirmwareVersion
			dev.DVMegaHardware = dvmega.Hardware
			devices = append(devices, dev)
			continue
		}

		// 3. Nextion display (tries 9 baud rates, 9600 default first)
		t0 = time.Now()
		nextion, err := ProbeNextion(port)
		slog.Debug("probe timing", "port", port, "probe", "nextion", "elapsed", time.Since(t0).Round(time.Millisecond))
		if err != nil {
			slog.Debug("nextion probe failed", "port", port, "error", err)
		}
		if nextion != nil {
			dev.DeviceType = DeviceNextion
			dev.NextionModel = nextion.Model
			dev.NextionSerial = nextion.Serial
			devices = append(devices, dev)
			continue
		}

		// Classify based on USB identity alone
		dev.DeviceType = classifyByUSB(&dev)
		devices = append(devices, dev)
	}

	return devices
}

// enumeratePorts returns sorted serial port paths found on the system.
func enumeratePorts() []string {
	patterns := []string{
		"/dev/ttyAMA*",
		"/dev/ttyACM*",
		"/dev/ttyUSB*",
	}

	var ports []string
	for _, pattern := range patterns {
		matches, _ := filepath.Glob(pattern)
		ports = append(ports, matches...)
	}
	sort.Strings(ports)
	return ports
}

// readUSBInfo populates USB metadata from sysfs.
// For ttyACM devices: /sys/class/tty/<tty>/device/../{idVendor,idProduct,product}
// For ttyUSB devices: /sys/class/tty/<tty>/device/../../{idVendor,idProduct,product}
func readUSBInfo(ttyName string, dev *DetectedDevice) {
	deviceLink := "/sys/class/tty/" + ttyName + "/device"

	// Resolve the symlink to get an absolute path, then walk up from there.
	// filepath.Join would collapse ".." against the unresolved symlink path.
	resolved, err := filepath.EvalSymlinks(deviceLink)
	if err != nil {
		return
	}

	// Walk up 1-3 parent directories looking for USB attributes.
	// ttyACM: device -> USB interface (has idVendor)
	// ttyUSB: device -> ttyUSBN -> USB interface -> USB device (has idVendor)
	dir := resolved
	for range 3 {
		dir = filepath.Dir(dir)
		vid := readSysfs(filepath.Join(dir, "idVendor"))
		pid := readSysfs(filepath.Join(dir, "idProduct"))
		if vid != "" && pid != "" {
			dev.USBVendor = vid
			dev.USBProduct = pid
			dev.USBName = readSysfs(filepath.Join(dir, "product"))
			// Read driver name from the device node itself
			driverLink := filepath.Join(resolved, "driver")
			if target, err := os.Readlink(driverLink); err == nil {
				dev.Driver = filepath.Base(target)
			}
			return
		}
	}
}

// classifyByUSB uses USB VID:PID to make a best-guess classification
// when protocol probes didn't identify the device.
func classifyByUSB(dev *DetectedDevice) DeviceType {
	if dev.USBVendor == "" {
		return DeviceUnknown
	}

	vidpid := dev.USBVendor + ":" + dev.USBProduct
	if known, ok := knownUSBDevices[vidpid]; ok {
		if known.likelyType != DeviceUnknown {
			return known.likelyType
		}
	}

	return DeviceUnknown
}

// readSysfs reads a single-line sysfs attribute file, returning empty string on error.
func readSysfs(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}

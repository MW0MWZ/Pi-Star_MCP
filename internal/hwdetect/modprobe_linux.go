package hwdetect

import (
	"bufio"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// loadUSBDrivers scans connected USB devices, matches their modaliases against
// the kernel's modules.alias file, and loads any matching modules. This replaces
// the udev/kmod autoloading that isn't available on Alpine with busybox mdev.
func loadUSBDrivers() {
	// Read modaliases from all connected USB devices
	modaliases := readUSBModaliases()
	if len(modaliases) == 0 {
		return
	}

	// Read the kernel's alias-to-module mapping
	kernelRelease := readFileStr("/proc/sys/kernel/osrelease")
	if kernelRelease == "" {
		return
	}
	aliasFile := "/lib/modules/" + kernelRelease + "/modules.alias"
	aliasMap := parseModulesAlias(aliasFile)
	if len(aliasMap) == 0 {
		return
	}

	// Match each USB modalias against the alias patterns
	needed := map[string]bool{}
	for _, modalias := range modaliases {
		for pattern, module := range aliasMap {
			if matchModalias(modalias, pattern) {
				needed[module] = true
			}
		}
	}

	if len(needed) == 0 {
		return
	}

	// Load each needed module
	for module := range needed {
		out, err := exec.Command("/sbin/modprobe", "-q", module).CombinedOutput()
		if err != nil {
			slog.Debug("modprobe failed", "module", module, "error", err, "output", string(out))
		} else {
			slog.Info("loaded kernel module", "module", module)
		}
	}

	// Wait for device nodes to appear
	time.Sleep(500 * time.Millisecond)
}

// readUSBModaliases reads the modalias file from each USB device in sysfs.
func readUSBModaliases() []string {
	matches, _ := filepath.Glob("/sys/bus/usb/devices/*/modalias")
	var result []string
	for _, path := range matches {
		alias := readFileStr(path)
		if strings.HasPrefix(alias, "usb:") {
			result = append(result, alias)
		}
	}
	return result
}

// parseModulesAlias reads a modules.alias file and returns a map of
// pattern → module name, filtered to USB aliases only.
func parseModulesAlias(path string) map[string]string {
	f, err := os.Open(path)
	if err != nil {
		return nil
	}
	defer f.Close()

	result := make(map[string]string)
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "alias usb:") {
			continue
		}
		// Format: "alias <pattern> <module>"
		fields := strings.Fields(line)
		if len(fields) != 3 {
			continue
		}
		result[fields[1]] = fields[2]
	}
	return result
}

// matchModalias matches a USB modalias string against a modules.alias pattern.
// Patterns use simple glob wildcards where * matches any substring.
// Example pattern: "usb:v1A86p7523d*dc*dsc*dp*ic*isc*ip*in*"
// Example modalias: "usb:v1A86p7523d0254dcFFdsc00dp00icFFisc01ip02in00"
func matchModalias(modalias, pattern string) bool {
	matched, _ := filepath.Match(pattern, modalias)
	return matched
}

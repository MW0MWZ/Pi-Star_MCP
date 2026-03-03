package hwdetect

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// resetGPIOModem resets modems connected via GPIO hat (MMDVM_HS, DV-Mega etc.)
// by toggling GPIO pins 20 and 21 through the sysfs interface.
// Matches pistar-findmodem's resetGPIOModem function.
func resetGPIOModem() {
	base := gpioChipBase()

	pin20 := base + 20
	pin21 := base + 21

	// Export and configure pins — ignore errors (pins may already be exported
	// or GPIO may not be available on this platform)
	if !gpioExport(pin20) || !gpioExport(pin21) {
		return
	}
	defer gpioUnexport(pin20)
	defer gpioUnexport(pin21)

	gpioSetDirection(pin20, "out")
	gpioSetDirection(pin21, "out")
	time.Sleep(500 * time.Millisecond)

	// Reset sequence (matches pistar-findmodem)
	gpioWrite(pin20, 0)
	gpioWrite(pin21, 0)
	gpioWrite(pin21, 1)
	time.Sleep(1 * time.Second)

	gpioWrite(pin20, 0)
	gpioWrite(pin20, 1)
	gpioWrite(pin20, 0)
	time.Sleep(500 * time.Millisecond)

	slog.Debug("gpio modem reset complete", "pin20", pin20, "pin21", pin21)
}

// gpioChipBase returns the sysfs GPIO number base for the main BCM GPIO controller.
// On older kernels (Pi 3/4 with older DT) this is 0; on newer kernels the chip
// may be at 512 (Pi 4 with pinctrl-bcm2711) or 504 (Pi 5 with pinctrl-rp1).
// We read the "label" file of each gpiochip to find the main pinctrl-bcm* or
// pinctrl-rp1 controller, then read its "base" file.
func gpioChipBase() int {
	matches, _ := filepath.Glob("/sys/class/gpio/gpiochip*")
	for _, m := range matches {
		label := readFileStr(filepath.Join(m, "label"))
		if !strings.HasPrefix(label, "pinctrl-") {
			continue
		}

		baseStr := readFileStr(filepath.Join(m, "base"))
		base, err := strconv.Atoi(baseStr)
		if err != nil {
			continue
		}
		return base
	}
	return 0
}

// readFileStr reads a sysfs file and returns its trimmed content.
func readFileStr(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}

func gpioExport(pin int) bool {
	return os.WriteFile("/sys/class/gpio/export", []byte(strconv.Itoa(pin)), 0644) == nil
}

func gpioUnexport(pin int) {
	os.WriteFile("/sys/class/gpio/unexport", []byte(strconv.Itoa(pin)), 0644)
}

func gpioSetDirection(pin int, dir string) {
	os.WriteFile(fmt.Sprintf("/sys/class/gpio/gpio%d/direction", pin), []byte(dir), 0644)
}

func gpioWrite(pin int, val int) {
	os.WriteFile(fmt.Sprintf("/sys/class/gpio/gpio%d/value", pin), []byte(strconv.Itoa(val)), 0644)
}

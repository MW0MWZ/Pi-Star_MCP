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

// gpioChipBase finds the GPIO chip base number, handling Pi 5 (gpiochip4
// at base 504, mapped to 0) and older Pi boards.
func gpioChipBase() int {
	matches, _ := filepath.Glob("/sys/class/gpio/gpiochip*")
	for _, m := range matches {
		// Look for the main GPIO controller (contains "0000.gpio" in the link target)
		target, err := os.Readlink(m)
		if err != nil {
			continue
		}
		if !strings.Contains(target, "0000.gpio") {
			continue
		}

		name := filepath.Base(m)
		numStr := strings.TrimPrefix(name, "gpiochip")
		num, err := strconv.Atoi(numStr)
		if err != nil {
			continue
		}

		// Pi 5: gpiochip504 maps to base 0
		if num == 504 {
			return 0
		}
		return num
	}
	return 0
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

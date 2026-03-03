package hwdetect

import (
	"fmt"
	"log/slog"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/sys/unix"
)

const (
	nextionTimeout = 500 * time.Millisecond // short per-baud; tries 9 bauds
)

// NextionProbeResult holds parsed info from a Nextion "connect" response.
type NextionProbeResult struct {
	Model    string   // e.g. "NX4832K035_011R"
	Serial   string   // device serial number
	Raw      string   // full comok response line
	BaudRate baudRate // baud rate that worked during detection
}

// ProbeNextion sends the Nextion "connect" command at each baud rate and
// parses the response. Returns nil (no error) if the device doesn't respond.
func ProbeNextion(port string) (*NextionProbeResult, error) {
	for _, baud := range nextionBaudRates {
		result, err := probeNextionOnce(port, baud)
		if err != nil {
			return nil, err
		}
		if result != nil {
			result.BaudRate = baud
			return result, nil
		}
	}
	return nil, nil
}

func probeNextionOnce(port string, baud baudRate) (*NextionProbeResult, error) {
	fd, err := openSerialPort(port, baud)
	if err != nil {
		return nil, err
	}
	defer unix.Close(fd)

	flushSerial(fd)

	// Send terminators to flush any garbage from prior probes at different bauds
	unix.Write(fd, []byte{0xFF, 0xFF, 0xFF})
	time.Sleep(50 * time.Millisecond)

	// Drain any error responses
	drain := make([]byte, 256)
	readGeneric(fd, drain, 100*time.Millisecond)
	flushSerial(fd)

	// Send "connect" command terminated by 0xFF 0xFF 0xFF
	connectCmd := append([]byte("connect"), 0xFF, 0xFF, 0xFF)
	if _, err := unix.Write(fd, connectCmd); err != nil {
		return nil, fmt.Errorf("write %s: %w", port, err)
	}

	buf := make([]byte, 256)
	n := readGeneric(fd, buf, nextionTimeout)
	if n == 0 {
		return nil, nil
	}

	return parseNextionResponse(buf[:n])
}

// readGeneric reads from fd without MMDVM frame-awareness.
func readGeneric(fd int, buf []byte, timeout time.Duration) int {
	fds := []unix.PollFd{{Fd: int32(fd), Events: unix.POLLIN}}
	ms := int(timeout.Milliseconds())

	n, err := unix.Poll(fds, ms)
	if err != nil || n == 0 {
		return 0
	}

	total := 0
	for total < len(buf) {
		n, err := unix.Read(fd, buf[total:])
		if err != nil || n == 0 {
			break
		}
		total += n

		// Check for Nextion response terminator (0xFF 0xFF 0xFF)
		if total >= 3 && buf[total-1] == 0xFF && buf[total-2] == 0xFF && buf[total-3] == 0xFF {
			break
		}

		n2, _ := unix.Poll(fds, 200)
		if n2 == 0 {
			break
		}
	}
	return total
}

// parseNextionResponse parses a "comok" response from a Nextion display.
// Format: "comok 1,<touch>,<model>,<fw>,<serial>,<flash>\xff\xff\xff"
func parseNextionResponse(data []byte) (*NextionProbeResult, error) {
	// Strip trailing 0xFF terminators
	for len(data) > 0 && data[len(data)-1] == 0xFF {
		data = data[:len(data)-1]
	}

	resp := string(data)
	if !strings.Contains(resp, "comok") {
		return nil, nil // not a Nextion response
	}

	result := &NextionProbeResult{Raw: resp}

	// Parse "comok 1,<touchID>-0,<model>,<fw>,<mcuID>,<serial>,<flash>"
	idx := strings.Index(resp, "comok")
	fields := resp[idx:]

	// Split on comma after "comok X,"
	parts := strings.SplitN(fields, ",", 7)
	if len(parts) >= 3 {
		result.Model = parts[2]
	}
	if len(parts) >= 6 {
		result.Serial = parts[5]
	}

	return result, nil
}

// MMDVM serial passthrough command type.
const mmdvmSerial = 0x80

// nextionMakeCmd builds a Nextion command string with the 0xFF 0xFF 0xFF terminator.
func nextionMakeCmd(cmd string) []byte {
	buf := []byte(cmd)
	buf = append(buf, 0xFF, 0xFF, 0xFF)
	return buf
}

// nextionWrapMMDVM wraps a Nextion command in an MMDVM serial frame.
// Frame: [0xE0] [length] [0x80] [nextion data...]
func nextionWrapMMDVM(nextionCmd []byte) []byte {
	frameLen := len(nextionCmd) + 3
	buf := make([]byte, 0, frameLen)
	buf = append(buf, 0xE0, byte(frameLen), mmdvmSerial)
	buf = append(buf, nextionCmd...)
	return buf
}

// nextionSend sends a Nextion command, using the MMDVM serial wrapper for
// GPIO-attached displays (ttyAMA*) and raw commands for USB (ttyUSB*).
func nextionSend(fd int, cmd string, wrapMMDVM bool) {
	raw := nextionMakeCmd(cmd)
	if wrapMMDVM {
		raw = nextionWrapMMDVM(raw)
	}
	unix.Write(fd, raw)
}

// nextionSetText sends a text value to a Nextion field.
func nextionSetText(fd int, field, value string, wrapMMDVM bool) {
	nextionSend(fd, fmt.Sprintf("%s.txt=\"%s\"", field, value), wrapMMDVM)
}

// InitNextion finds any detected Nextion displays and shows a startup message.
func InitNextion(devices []DetectedDevice) {
	for _, dev := range devices {
		if dev.DeviceType != DeviceNextion {
			continue
		}

		baud := dev.NextionBaud
		if baud == 0 {
			baud = unix.B9600 // Nextion factory default
		}

		fd, err := openSerialPort(dev.Port, baud)
		if err != nil {
			continue
		}

		// GPIO-attached Nextions (ttyAMA*) go through the MMDVM firmware's
		// serial passthrough; USB-attached ones speak raw Nextion protocol.
		wrap := !strings.HasPrefix(filepath.Base(dev.Port), "ttyUSB")

		// Clear all standard fields
		for _, field := range []string{"t0", "t1", "t2", "t5", "t20", "t30", "t31", "t32"} {
			nextionSetText(fd, field, "", wrap)
		}

		// Display startup message
		nextionSetText(fd, "t0", "Nextion Detected", wrap)
		nextionSetText(fd, "t2", "Pi-Star MCP", wrap)
		nextionSend(fd, "ref 0", wrap)

		unix.Close(fd)

		slog.Info("nextion startup message displayed", "port", dev.Port, "model", dev.NextionModel)
	}
}

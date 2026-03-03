package hwdetect

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"golang.org/x/sys/unix"
)

// MMDVM protocol constants.
const (
	mmdvmFrameStart = 0xE0
	mmdvmGetVersion = 0x00
	mmdvmMaxFrame   = 256
	mmdvmTimeout    = 1500 * time.Millisecond
	mmdvmRetries    = 0 // real MMDVM responds in ~100ms; no need to retry
)

// MMDVMProbeResult holds the parsed response from a GET_VERSION probe.
type MMDVMProbeResult struct {
	Protocol    int    // 1 or 2
	Description string // firmware description string (e.g. "ZUMspot-v1.6.1")
	HWType      string // normalized hardware type (e.g. "zumspot", "mmdvm_hs")
}

// Known MMDVM hardware description prefixes, ordered longest-first
// so more specific matches win.
var mmdvmHWPrefixes = []struct {
	prefix string
	hwType string
}{
	{"MMDVM_HS_Dual_Hat", "mmdvm_hs_dual_hat"},
	{"MMDVM_HS_Hat", "mmdvm_hs_hat"},
	{"MMDVM_HS-", "mmdvm_hs"},
	{"MMDVM_HS", "mmdvm_hs"},
	{"MMDVM_RPT_Hat", "mmdvm_rpt_hat"},
	{"MMDVM ", "mmdvm"},
	{"DVMEGA", "dvmega"},
	{"ZUMspot", "zumspot"},
	{"NANO_hotSPOT", "nano_hotspot"},
	{"Nano hotSPOT", "nano_hotspot"},
	{"D2RG_MMDVM_HS", "d2rg_mmdvm_hs"},
	{"OpenGD77 Hotspot", "opengd77_hs"},
	{"SkyBridge", "skybridge"},
}

// ProbeMMDVM opens the given serial port, sends GET_VERSION, and parses
// the MMDVM response. Returns nil result (no error) if no response.
func ProbeMMDVM(port string) (*MMDVMProbeResult, error) {
	for attempt := 0; attempt <= mmdvmRetries; attempt++ {
		result, err := probeMMDVMOnce(port)
		if err != nil {
			return nil, err
		}
		if result != nil {
			return result, nil
		}
	}
	return nil, nil
}

func probeMMDVMOnce(port string) (*MMDVMProbeResult, error) {
	fd, err := openSerialPort(port, unix.B115200)
	if err != nil {
		return nil, err
	}
	defer unix.Close(fd)

	flushSerial(fd)
	time.Sleep(50 * time.Millisecond)

	// Send GET_VERSION: E0 03 00
	cmd := []byte{mmdvmFrameStart, 0x03, mmdvmGetVersion}
	if _, err := unix.Write(fd, cmd); err != nil {
		return nil, fmt.Errorf("write %s: %w", port, err)
	}

	buf := make([]byte, mmdvmMaxFrame)
	n := readWithTimeout(fd, buf, mmdvmTimeout)
	if n == 0 {
		return nil, nil // no response
	}

	return parseMMDVMResponse(buf[:n])
}

// openSerialPort opens a serial port in raw mode at the given baud rate.
func openSerialPort(port string, baud baudRate) (int, error) {
	fd, err := unix.Open(port, unix.O_RDWR|unix.O_NOCTTY|unix.O_NONBLOCK, 0)
	if err != nil {
		return -1, fmt.Errorf("open %s: %w", port, err)
	}

	if err := unix.SetNonblock(fd, false); err != nil {
		unix.Close(fd)
		return -1, fmt.Errorf("setnonblock %s: %w", port, err)
	}

	termios, err := unix.IoctlGetTermios(fd, ioctlGetTermios)
	if err != nil {
		unix.Close(fd)
		return -1, fmt.Errorf("get termios %s: %w", port, err)
	}

	// Raw mode
	termios.Iflag &^= unix.IGNBRK | unix.BRKINT | unix.PARMRK | unix.ISTRIP |
		unix.INLCR | unix.IGNCR | unix.ICRNL | unix.IXON
	termios.Oflag &^= unix.OPOST
	termios.Lflag &^= unix.ECHO | unix.ECHONL | unix.ICANON | unix.ISIG | unix.IEXTEN
	termios.Cflag &^= unix.CSIZE | unix.PARENB
	termios.Cflag |= unix.CS8 | unix.CLOCAL | unix.CREAD

	setBaud(termios, baud)

	// VMIN=0, VTIME=5 (500ms in deciseconds — Poll handles main timeout)
	termios.Cc[unix.VMIN] = 0
	termios.Cc[unix.VTIME] = 5

	if err := unix.IoctlSetTermios(fd, ioctlSetTermios, termios); err != nil {
		unix.Close(fd)
		return -1, fmt.Errorf("set termios %s: %w", port, err)
	}

	return fd, nil
}

// readWithTimeout reads from fd using poll + read, accumulating data
// until timeout or a complete frame is received.
func readWithTimeout(fd int, buf []byte, timeout time.Duration) int {
	fds := []unix.PollFd{{Fd: int32(fd), Events: unix.POLLIN}}
	ms := int(timeout.Milliseconds())

	n, err := unix.Poll(fds, ms)
	if err != nil || n == 0 {
		return 0
	}

	total := 0
	for total < len(buf) {
		n, err := unix.Read(fd, buf[total:])
		if err != nil {
			if errors.Is(err, unix.EAGAIN) || errors.Is(err, unix.EINTR) {
				break
			}
			return total
		}
		if n == 0 {
			break
		}
		total += n

		// Check for complete MMDVM frame
		if total >= 2 && buf[0] == mmdvmFrameStart {
			frameLen := int(buf[1])
			if total >= frameLen {
				break
			}
		}

		// Brief poll for more data
		n2, _ := unix.Poll(fds, 200)
		if n2 == 0 {
			break
		}
	}

	return total
}

// parseMMDVMResponse parses a raw MMDVM GET_VERSION response frame.
func parseMMDVMResponse(data []byte) (*MMDVMProbeResult, error) {
	if len(data) < 4 {
		return nil, fmt.Errorf("response too short: %d bytes", len(data))
	}

	// Find frame start
	start := -1
	for i, b := range data {
		if b == mmdvmFrameStart {
			start = i
			break
		}
	}
	if start < 0 {
		return nil, fmt.Errorf("no frame start marker found")
	}
	data = data[start:]

	if len(data) < 4 {
		return nil, fmt.Errorf("frame too short after start marker")
	}

	frameLen := int(data[1])
	if frameLen < 4 || frameLen > len(data) {
		return nil, fmt.Errorf("invalid frame length: %d", frameLen)
	}

	cmd := data[2]
	if cmd != mmdvmGetVersion {
		return nil, fmt.Errorf("unexpected command in response: 0x%02X", cmd)
	}

	protoVer := int(data[3])
	var desc string

	switch protoVer {
	case 1:
		if frameLen > 4 {
			desc = trimNul(data[4:frameLen])
		}
	case 2:
		// Proto v2: 4 bytes capabilities + 16 bytes UDID, description at offset 24
		const descOffset = 4 + 4 + 16
		if frameLen > descOffset {
			desc = trimNul(data[descOffset:frameLen])
		}
	default:
		if frameLen > 4 {
			desc = trimNul(data[4:frameLen])
		}
	}

	return &MMDVMProbeResult{
		Protocol:    protoVer,
		Description: desc,
		HWType:      parseMMDVMHWType(desc),
	}, nil
}

// parseMMDVMHWType matches a firmware description against known prefixes.
func parseMMDVMHWType(desc string) string {
	for _, p := range mmdvmHWPrefixes {
		if strings.HasPrefix(desc, p.prefix) {
			return p.hwType
		}
	}
	if desc != "" {
		return "unknown"
	}
	return ""
}

// trimNul extracts a string, stopping at the first NUL byte.
func trimNul(b []byte) string {
	for i, c := range b {
		if c == 0 {
			return string(b[:i])
		}
	}
	return string(b)
}

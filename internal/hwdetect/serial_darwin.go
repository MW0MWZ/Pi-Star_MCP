package hwdetect

import (
	"time"

	"golang.org/x/sys/unix"
)

type baudRate = uint64

const (
	ioctlGetTermios = unix.TIOCGETA
	ioctlSetTermios = unix.TIOCSETA
)

// setBaud sets the baud rate via Ispeed/Ospeed fields.
// On macOS, TIOCSETA reads baud from these fields directly (no CBAUD in cflag).
func setBaud(termios *unix.Termios, baud baudRate) {
	termios.Ispeed = baud
	termios.Ospeed = baud
}

// MMDVM baud rates to try (Darwin only has standard POSIX rates).
var mmdvmBaudRates = []baudRate{unix.B115200, unix.B230400}

// Nextion baud rates to try.
var nextionBaudRates = []baudRate{
	unix.B9600,
	unix.B2400,
	unix.B4800,
	unix.B19200,
	unix.B38400,
	unix.B57600,
	unix.B115200,
	unix.B230400,
}

func flushSerial(fd int) {
	// TIOCFLUSH with FREAD|FWRITE — on macOS these are 1|2
	unix.IoctlSetInt(fd, unix.TIOCFLUSH, 1|2)
}

// resetDTR toggles DTR to reset an Arduino-based device.
func resetDTR(port string) {
	fd, err := unix.Open(port, unix.O_RDWR|unix.O_NOCTTY|unix.O_NONBLOCK, 0)
	if err != nil {
		return
	}
	defer unix.Close(fd)

	bits, _ := unix.IoctlGetInt(fd, unix.TIOCMGET)
	bits &^= unix.TIOCM_DTR
	unix.IoctlSetInt(fd, unix.TIOCMSET, bits)

	time.Sleep(100 * time.Millisecond)

	bits |= unix.TIOCM_DTR | unix.TIOCM_RTS
	unix.IoctlSetInt(fd, unix.TIOCMSET, bits)
}

package hwdetect

import "golang.org/x/sys/unix"

type baudRate = uint32

const (
	ioctlGetTermios = unix.TCGETS
	ioctlSetTermios = unix.TCSETS
)

// setBaud sets the baud rate in both cflag (CBAUD mask) and Ispeed/Ospeed.
// On Linux, the kernel's TCSETS ioctl reads the baud from c_cflag & CBAUD,
// NOT from c_ispeed/c_ospeed (those are only used with TCSETS2/BOTHER).
// The C library cfsetispeed/cfsetospeed functions set both; we must do the same.
func setBaud(termios *unix.Termios, baud baudRate) {
	termios.Cflag &^= unix.CBAUD
	termios.Cflag |= uint32(baud)
	termios.Ispeed = baud
	termios.Ospeed = baud
}

// MMDVM baud rates to try, matching pistar-findmodem.
var mmdvmBaudRates = []baudRate{unix.B115200, unix.B230400, unix.B460800}

// Nextion baud rates to try, matching pistar-findmodem.
// Standard B* constants only; non-standard rates (31250, 250000, 256000, 512000)
// would require BOTHER and are extremely rare in practice.
var nextionBaudRates = []baudRate{
	unix.B9600, // Nextion factory default
	unix.B2400,
	unix.B4800,
	unix.B19200,
	unix.B38400,
	unix.B57600,
	unix.B115200,
	unix.B230400,
	unix.B921600,
}

func flushSerial(fd int) {
	unix.IoctlSetInt(fd, unix.TCFLSH, unix.TCIOFLUSH)
}

// resetDTR toggles DTR low then high on a serial port to reset an Arduino-based
// device (DV-Mega). The DTR line is coupled to the Arduino's RESET pin via a
// capacitor — the LOW→HIGH transition generates a reset pulse.
func resetDTR(port string) {
	fd, err := unix.Open(port, unix.O_RDWR|unix.O_NOCTTY|unix.O_NONBLOCK, 0)
	if err != nil {
		return
	}
	defer unix.Close(fd)

	bits, _ := unix.IoctlGetInt(fd, unix.TIOCMGET)

	// Drop DTR
	bits &^= unix.TIOCM_DTR
	unix.IoctlSetPointerInt(fd, unix.TIOCMSET, bits)

	// Brief pause then reassert — the rising edge triggers the reset
	unix.Nanosleep(&unix.Timespec{Nsec: 100_000_000}, nil) // 100ms

	bits |= unix.TIOCM_DTR | unix.TIOCM_RTS
	unix.IoctlSetPointerInt(fd, unix.TIOCMSET, bits)
}

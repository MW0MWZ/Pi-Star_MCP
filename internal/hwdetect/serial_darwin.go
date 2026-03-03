package hwdetect

import "golang.org/x/sys/unix"

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

func flushSerial(fd int) {
	// TIOCFLUSH with FREAD|FWRITE — on macOS these are 1|2
	unix.IoctlSetInt(fd, unix.TIOCFLUSH, 1|2)
}

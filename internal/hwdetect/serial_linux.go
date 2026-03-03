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

func flushSerial(fd int) {
	unix.IoctlSetInt(fd, unix.TCFLSH, unix.TCIOFLUSH)
}

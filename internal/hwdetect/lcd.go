package hwdetect

import (
	"fmt"
	"log/slog"
	"time"

	"golang.org/x/sys/unix"
)

// MCP23017 register addresses (BANK=0 mode).
const (
	mcp23017IODIRA = 0x00
	mcp23017IODIRB = 0x01
	mcp23017GPIOA  = 0x12
	mcp23017GPIOB  = 0x13
)

// Adafruit RGB Pi Plate MCP23017 Port B pin assignments.
// From the Adafruit_RGBLCDShield library:
//   RS=pin15(GPB7), RW=pin14(GPB6), EN=pin13(GPB5)
//   D4=pin12(GPB4), D5=pin11(GPB3), D6=pin10(GPB2), D7=pin9(GPB1)
//   Red=pin8(GPB0), Green=pin6(GPA6), Blue=pin7(GPA7)
// Backlight is active-low on this board.
const (
	lcdRS = 0x80 // GPB7
	lcdEN = 0x20 // GPB5
)

// lcdMapNibble maps a 4-bit value to the Adafruit plate's data pin positions.
// D4=GPB4(0x10), D5=GPB3(0x08), D6=GPB2(0x04), D7=GPB1(0x02).
func lcdMapNibble(n byte) byte {
	var d byte
	if n&1 != 0 {
		d |= 0x10
	}
	if n&2 != 0 {
		d |= 0x08
	}
	if n&4 != 0 {
		d |= 0x04
	}
	if n&8 != 0 {
		d |= 0x02
	}
	return d
}

// lcdWriteReg writes a single byte to an MCP23017 register.
func lcdWriteReg(fd int, reg, val byte) {
	unix.Write(fd, []byte{reg, val})
}

// lcdPulseEnable sends a value with EN pulsed high then low.
func lcdPulseEnable(fd int, val byte) {
	lcdWriteReg(fd, mcp23017GPIOB, val|lcdEN)
	lcdWriteReg(fd, mcp23017GPIOB, val&^lcdEN)
}

// lcdSendNibble sends a 4-bit nibble with the given mode (0=command, lcdRS=data).
func lcdSendNibble(fd int, n byte, mode byte) {
	lcdPulseEnable(fd, lcdMapNibble(n)|mode)
}

// lcdSendByte sends a full byte as two nibbles.
func lcdSendByte(fd int, b byte, mode byte) {
	lcdSendNibble(fd, (b>>4)&0x0F, mode)
	lcdSendNibble(fd, b&0x0F, mode)
}

// lcdCommand sends a command byte to the HD44780.
func lcdCommand(fd int, cmd byte) {
	lcdSendByte(fd, cmd, 0)
}

// lcdData sends a data byte (character) to the HD44780.
func lcdData(fd int, ch byte) {
	lcdSendByte(fd, ch, lcdRS)
}

// lcdInit initialises the MCP23017 and HD44780 in 4-bit mode with backlight on.
func lcdInit(fd int) {
	// Set both ports as outputs
	lcdWriteReg(fd, mcp23017IODIRA, 0x00)
	lcdWriteReg(fd, mcp23017IODIRB, 0x00)

	// Backlight on (active-low: all zeros)
	lcdWriteReg(fd, mcp23017GPIOA, 0x00)
	lcdWriteReg(fd, mcp23017GPIOB, 0x00)
	time.Sleep(100 * time.Millisecond)

	// HD44780 4-bit initialisation sequence
	lcdSendNibble(fd, 3, 0)
	time.Sleep(5 * time.Millisecond)
	lcdSendNibble(fd, 3, 0)
	time.Sleep(5 * time.Millisecond)
	lcdSendNibble(fd, 3, 0)
	time.Sleep(5 * time.Millisecond)
	lcdSendNibble(fd, 2, 0) // switch to 4-bit
	time.Sleep(5 * time.Millisecond)

	lcdCommand(fd, 0x28) // 4-bit, 2 lines, 5x8 font
	time.Sleep(2 * time.Millisecond)
	lcdCommand(fd, 0x08) // display off
	time.Sleep(2 * time.Millisecond)
	lcdCommand(fd, 0x01) // clear
	time.Sleep(5 * time.Millisecond)
	lcdCommand(fd, 0x06) // entry mode: increment, no shift
	time.Sleep(2 * time.Millisecond)
	lcdCommand(fd, 0x0C) // display on, cursor off, blink off
	time.Sleep(2 * time.Millisecond)
}

// lcdWriteString writes a string at the current cursor position.
func lcdWriteString(fd int, s string) {
	for i := 0; i < len(s); i++ {
		lcdData(fd, s[i])
	}
}

// lcdSetLine moves the cursor to the start of line 1 or 2.
func lcdSetLine(fd int, line int) {
	addr := byte(0x80) // line 1 (DDRAM address 0x00)
	if line == 2 {
		addr = 0xC0 // line 2 (DDRAM address 0x40)
	}
	lcdCommand(fd, addr)
	time.Sleep(1 * time.Millisecond)
}

// centerText centres a string within the given width, padding with spaces.
func centerText(s string, width int) string {
	if len(s) >= width {
		return s[:width]
	}
	pad := (width - len(s)) / 2
	result := make([]byte, width)
	for i := range result {
		result[i] = ' '
	}
	copy(result[pad:], s)
	return string(result)
}

// isMCP23017 checks whether the device at the current slave address is an
// MCP23017 by testing register-addressed I/O. We write a test pattern to
// IODIRB (reg 0x01), read it back, then restore the original value. PCF8574
// has no register addressing, so the round-trip will fail.
func isMCP23017(fd int) bool {
	// Read current value of register 0x01
	unix.Write(fd, []byte{0x01})
	orig := make([]byte, 1)
	if _, err := unix.Read(fd, orig); err != nil {
		return false
	}

	// Write a test pattern that differs from the current value
	test := byte(0xAA)
	if orig[0] == test {
		test = 0x55
	}
	unix.Write(fd, []byte{0x01, test})

	// Read it back
	unix.Write(fd, []byte{0x01})
	check := make([]byte, 1)
	if _, err := unix.Read(fd, check); err != nil {
		return false
	}

	// Restore original value
	unix.Write(fd, []byte{0x01, orig[0]})

	return check[0] == test
}

// InitLCD finds any detected MCP23017 LCD plate and displays a startup message.
// Call this after DetectI2C to show device info on the LCD.
func InitLCD(i2cDevices []DetectedI2CDevice) {
	for i := range i2cDevices {
		dev := &i2cDevices[i]
		if dev.DeviceType != "lcd" || dev.Name != "Adafruit LCD Plate" {
			continue
		}

		busPath := fmt.Sprintf("/dev/i2c-%d", dev.Bus)
		fd, err := unix.Open(busPath, unix.O_RDWR, 0)
		if err != nil {
			slog.Debug("lcd init: cannot open bus", "path", busPath, "error", err)
			return
		}
		defer unix.Close(fd)

		addr := uint8(0x20)
		fmt.Sscanf(dev.Address, "0x%x", &addr)
		if err := unix.IoctlSetInt(fd, i2cSlave, int(addr)); err != nil {
			slog.Debug("lcd init: cannot set slave", "addr", dev.Address, "error", err)
			return
		}

		lcdInit(fd)

		top := centerText("LCD Detected", 16)
		bottom := centerText("Pi-Star MCP", 16)

		lcdSetLine(fd, 1)
		lcdWriteString(fd, top)
		lcdSetLine(fd, 2)
		lcdWriteString(fd, bottom)

		slog.Info("lcd startup message displayed", "bus", dev.Bus, "address", dev.Address)
		return
	}
}

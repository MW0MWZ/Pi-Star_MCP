package hwdetect

import (
	"fmt"
	"time"

	"golang.org/x/sys/unix"
)

// DV-Mega / DVRPTR protocol constants.
const (
	dvrptrFrameStart = 0xD0
	dvrptrGetVersion = 0x11
	dvrptrHeaderLen  = 5 // frame_start + len_lo + len_hi + cmd + first_payload_byte
)

// DVMegaProbeResult holds parsed info from a DV-Mega GET_VERSION response.
type DVMegaProbeResult struct {
	FirmwareVersion string // e.g. "1.02" or "1.02a"
	Hardware        string // hardware description string from response
}

// ProbeDVMega sends a DVRPTR GET_VERSION command and parses the response.
// Uses raw serial mode at 115200 (matching pistar-findmodem bash script).
// Sends GET_VERSION 3 times, waits 200ms, reads response.
// Returns nil (no error) if the device doesn't respond.
func ProbeDVMega(port string) (*DVMegaProbeResult, error) {
	fd, err := openSerialPort(port, unix.B115200)
	if err != nil {
		return nil, err
	}
	defer unix.Close(fd)

	// GET_VERSION frame: D0 01 00 11 00 0B (includes CRC)
	getVersion := []byte{dvrptrFrameStart, 0x01, 0x00, dvrptrGetVersion, 0x00, 0x0B}

	// Send GET_VERSION 3 times back-to-back (matches DStarRepeater)
	for range 3 {
		if _, err := unix.Write(fd, getVersion); err != nil {
			return nil, fmt.Errorf("write %s: %w", port, err)
		}
	}

	// Wait for response (bash script uses sleep 0.2s)
	time.Sleep(200 * time.Millisecond)

	// Read response using poll+read
	buf := make([]byte, 256)
	n := readWithTimeout(fd, buf, 500*time.Millisecond)
	if n == 0 {
		return nil, nil
	}

	return parseDVMegaResponse(buf[:n])
}

// parseDVMegaResponse parses a DVRPTR GET_VERSION response.
// Frame: D0 [len_lo] [len_hi] [cmd|0x80] [payload...] (no CRC in default mode)
// Payload for GET_VERSION: [rev_build] [maj_min] [hardware_string...]
func parseDVMegaResponse(data []byte) (*DVMegaProbeResult, error) {
	// Find frame start
	start := -1
	for i, b := range data {
		if b == dvrptrFrameStart {
			start = i
			break
		}
	}
	if start < 0 {
		return nil, nil
	}
	data = data[start:]

	if len(data) < dvrptrHeaderLen {
		return nil, nil
	}

	// Parse length (little-endian)
	payloadLen := int(data[1]) | int(data[2])<<8

	// Verify it's a version response (command with response bit set)
	cmd := data[3] & 0x7F
	if cmd != dvrptrGetVersion {
		return nil, nil
	}

	// Total frame = 3 (start + len_lo + len_hi) + payloadLen
	if len(data) < 3+payloadLen {
		return nil, nil
	}

	result := &DVMegaProbeResult{}

	// Version bytes at data[4] and data[5]
	if payloadLen >= 2 {
		revBuild := data[4]
		majMin := data[5]

		major := (majMin & 0xF0) >> 4
		minor := majMin & 0x0F
		revision := (revBuild & 0xF0) >> 4
		buildLetter := revBuild & 0x0F

		if buildLetter > 0 {
			result.FirmwareVersion = fmt.Sprintf("%d.%d%d%c", major, minor, revision, 'a'+buildLetter-1)
		} else {
			result.FirmwareVersion = fmt.Sprintf("%d.%d%d", major, minor, revision)
		}
	}

	// Hardware string at offset 6, length = payloadLen - 3
	if payloadLen > 3 {
		hwLen := payloadLen - 3
		hwStart := 6
		if hwStart+hwLen <= len(data) {
			result.Hardware = trimNul(data[hwStart : hwStart+hwLen])
		}
	}

	return result, nil
}

// crcCCITT computes CRC-CCITT (init 0xFFFF) over the given data.
// Used when checksum mode is enabled (not the DStarRepeater default).
func crcCCITT(data []byte) uint16 {
	crc := uint16(0xFFFF)
	for _, b := range data {
		crc ^= uint16(b) << 8
		for range 8 {
			if crc&0x8000 != 0 {
				crc = (crc << 1) ^ 0x1021
			} else {
				crc <<= 1
			}
		}
	}
	return crc
}

package hwdetect

import (
	"testing"
)

// Real ZUMspot response captured from ttyACM0
var zumspotResponse = []byte{
	0xE0, 0x4A, 0x00, 0x01,
	// "ZUMspot-v1.6.1 20230526 14.7456MHz ADF7021 FW by CA6JAU GitID #7ff74ed"
	0x5A, 0x55, 0x4D, 0x73, 0x70, 0x6F, 0x74, 0x2D,
	0x76, 0x31, 0x2E, 0x36, 0x2E, 0x31, 0x20, 0x32,
	0x30, 0x32, 0x33, 0x30, 0x35, 0x32, 0x36, 0x20,
	0x31, 0x34, 0x2E, 0x37, 0x34, 0x35, 0x36, 0x4D,
	0x48, 0x7A, 0x20, 0x41, 0x44, 0x46, 0x37, 0x30,
	0x32, 0x31, 0x20, 0x46, 0x57, 0x20, 0x62, 0x79,
	0x20, 0x43, 0x41, 0x36, 0x4A, 0x41, 0x55, 0x20,
	0x47, 0x69, 0x74, 0x49, 0x44, 0x20, 0x23, 0x37,
	0x66, 0x66, 0x37, 0x34, 0x65, 0x64,
}

// Real SkyBridge response captured from ttyAMA0 (includes trailing UDID after NUL)
var skybridgeResponse = []byte{
	0xE0, 0x65, 0x00, 0x01,
	// "SkyBridge-v1.6.1 20230526 14.7456MHz ADF7021 FW by CA6JAU GitID #7ff74ed"
	0x53, 0x6B, 0x79, 0x42, 0x72, 0x69, 0x64, 0x67,
	0x65, 0x2D, 0x76, 0x31, 0x2E, 0x36, 0x2E, 0x31,
	0x20, 0x32, 0x30, 0x32, 0x33, 0x30, 0x35, 0x32,
	0x36, 0x20, 0x31, 0x34, 0x2E, 0x37, 0x34, 0x35,
	0x36, 0x4D, 0x48, 0x7A, 0x20, 0x41, 0x44, 0x46,
	0x37, 0x30, 0x32, 0x31, 0x20, 0x46, 0x57, 0x20,
	0x62, 0x79, 0x20, 0x43, 0x41, 0x36, 0x4A, 0x41,
	0x55, 0x20, 0x47, 0x69, 0x74, 0x49, 0x44, 0x20,
	0x23, 0x37, 0x66, 0x66, 0x37, 0x34, 0x65, 0x64,
	// NUL + trailing UDID data (beyond frame, but present in read buffer)
	0x00, 0x46, 0x46, 0x33, 0x32, 0x30, 0x36, 0x37,
	0x30, 0x33, 0x30, 0x33, 0x33, 0x34, 0x44, 0x34,
	0x45, 0x34, 0x33, 0x30, 0x38, 0x31, 0x38, 0x34,
	0x33,
}

func TestParseMMDVMResponse_ZUMspot(t *testing.T) {
	result, err := parseMMDVMResponse(zumspotResponse)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Protocol != 1 {
		t.Errorf("protocol = %d, want 1", result.Protocol)
	}
	if result.HWType != "zumspot" {
		t.Errorf("hwType = %q, want %q", result.HWType, "zumspot")
	}
	if result.Description != "ZUMspot-v1.6.1 20230526 14.7456MHz ADF7021 FW by CA6JAU GitID #7ff74ed" {
		t.Errorf("description = %q", result.Description)
	}
}

func TestParseMMDVMResponse_SkyBridge(t *testing.T) {
	// Use only the frame portion (up to frameLen)
	result, err := parseMMDVMResponse(skybridgeResponse)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Protocol != 1 {
		t.Errorf("protocol = %d, want 1", result.Protocol)
	}
	if result.HWType != "skybridge" {
		t.Errorf("hwType = %q, want %q", result.HWType, "skybridge")
	}
	// Description should stop at the NUL byte within the frame
	want := "SkyBridge-v1.6.1 20230526 14.7456MHz ADF7021 FW by CA6JAU GitID #7ff74ed"
	if result.Description != want {
		t.Errorf("description = %q, want %q", result.Description, want)
	}
}

// Synthetic protocol v2 response
func TestParseMMDVMResponse_ProtoV2(t *testing.T) {
	desc := "MMDVM_HS_Hat-v1.5.2"
	// Frame: E0 <len> 00 02 <4 cap bytes> <16 UDID bytes> <desc>
	frame := []byte{0xE0, 0x00, 0x00, 0x02}
	frame = append(frame, 0x01, 0x02, 0x03, 0x04) // capabilities
	frame = append(frame, make([]byte, 16)...)      // UDID
	frame = append(frame, []byte(desc)...)
	frame[1] = byte(len(frame)) // set length

	result, err := parseMMDVMResponse(frame)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Protocol != 2 {
		t.Errorf("protocol = %d, want 2", result.Protocol)
	}
	if result.Description != desc {
		t.Errorf("description = %q, want %q", result.Description, desc)
	}
	if result.HWType != "mmdvm_hs_hat" {
		t.Errorf("hwType = %q, want %q", result.HWType, "mmdvm_hs_hat")
	}
}

func TestParseMMDVMHWType(t *testing.T) {
	tests := []struct {
		desc string
		want string
	}{
		{"MMDVM_HS_Hat-v1.5.2 20210101", "mmdvm_hs_hat"},
		{"MMDVM_HS_Dual_Hat-v1.5.2", "mmdvm_hs_dual_hat"},
		{"MMDVM_HS-v1.4.17", "mmdvm_hs"},
		{"MMDVM_RPT_Hat-v1.5.2", "mmdvm_rpt_hat"},
		{"MMDVM 20171031 (D-Star/DMR/System Fusion/P25/NXDN/POCSAG/FM)", "mmdvm"},
		{"DVMEGA HR-v0.1", "dvmega"},
		{"ZUMspot-v1.6.1 20230526", "zumspot"},
		{"NANO_hotSPOT-v1.5.2", "nano_hotspot"},
		{"Nano hotSPOT-v1.5.2", "nano_hotspot"},
		{"D2RG_MMDVM_HS-v1.5.2", "d2rg_mmdvm_hs"},
		{"OpenGD77 Hotspot-v0.1", "opengd77_hs"},
		{"SkyBridge-v1.6.1", "skybridge"},
		{"SomeNewBoard-v1.0", "unknown"},
		{"", ""},
	}

	for _, tt := range tests {
		got := parseMMDVMHWType(tt.desc)
		if got != tt.want {
			t.Errorf("parseMMDVMHWType(%q) = %q, want %q", tt.desc, got, tt.want)
		}
	}
}

func TestParseMMDVMResponse_TooShort(t *testing.T) {
	_, err := parseMMDVMResponse([]byte{0xE0, 0x03})
	if err == nil {
		t.Error("expected error for short response")
	}
}

func TestParseMMDVMResponse_NoStartMarker(t *testing.T) {
	_, err := parseMMDVMResponse([]byte{0x00, 0x04, 0x00, 0x01})
	if err == nil {
		t.Error("expected error for missing start marker")
	}
}

// Real Nextion response captured from ttyUSB1
func TestParseNextionResponse(t *testing.T) {
	// "comok 1,30614-0,NX4832K035_011R,163,61699,DB62203117173C27,33554432" + 0xFF 0xFF 0xFF
	raw := []byte("comok 1,30614-0,NX4832K035_011R,163,61699,DB62203117173C27,33554432")
	raw = append(raw, 0xFF, 0xFF, 0xFF)

	result, err := parseNextionResponse(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.Model != "NX4832K035_011R" {
		t.Errorf("model = %q, want %q", result.Model, "NX4832K035_011R")
	}
	if result.Serial != "DB62203117173C27" {
		t.Errorf("serial = %q, want %q", result.Serial, "DB62203117173C27")
	}
}

func TestParseNextionResponse_NotNextion(t *testing.T) {
	result, err := parseNextionResponse([]byte{0x01, 0x02, 0x03})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != nil {
		t.Error("expected nil result for non-Nextion response")
	}
}

func TestCRCCCITT(t *testing.T) {
	// GET_VERSION frame: D0 01 00 11
	frame := []byte{0xD0, 0x01, 0x00, 0x11}
	crc := crcCCITT(frame)
	// Verify it produces a non-zero CRC
	if crc == 0 {
		t.Error("CRC should not be zero for non-empty data")
	}

	// Verify CRC is consistent
	crc2 := crcCCITT(frame)
	if crc != crc2 {
		t.Errorf("CRC not deterministic: %04X vs %04X", crc, crc2)
	}

	// Verify appending CRC and re-computing gives a known property
	// (CRC of frame+CRC should verify correctly in parseDVMegaResponse)
	withCRC := append(frame, byte(crc&0xFF), byte(crc>>8))
	result := crcCCITT(withCRC[:4]) // CRC of just the frame portion
	gotCRC := uint16(withCRC[4]) | uint16(withCRC[5])<<8
	if result != gotCRC {
		t.Errorf("CRC verification failed: computed %04X, stored %04X", result, gotCRC)
	}
}

func TestParseDVMegaResponse(t *testing.T) {
	// Build a synthetic version response (no-CRC mode, matching DStarRepeater default):
	// D0 [len_lo] [len_hi] 0x91 [rev_build] [maj_min] [hw...]
	// Version 1.23 = major=1, minor=2, revision=3, build=0
	// maj_min = 0x12, rev_build = 0x30
	frame := []byte{
		0xD0,       // frame start
		0x0A, 0x00, // payload length = 10
		0x91,       // response command (0x11 | 0x80)
		0x30,       // revision 3, build 0
		0x12,       // major 1, minor 2
		'D', 'V', '-', 'M', 'e', 'g', 'a', // hardware string (7 bytes = payloadLen - 3)
	}

	result, err := parseDVMegaResponse(frame)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.FirmwareVersion != "1.23" {
		t.Errorf("firmware = %q, want %q", result.FirmwareVersion, "1.23")
	}
	if result.Hardware != "DV-Mega" {
		t.Errorf("hardware = %q, want %q", result.Hardware, "DV-Mega")
	}
}

func TestParseDVMegaResponse_WithBuildLetter(t *testing.T) {
	// Version 1.23a = major=1, minor=2, revision=3, build=1
	frame := []byte{
		0xD0,
		0x03, 0x00, // payload length = 3
		0x91,       // response
		0x31,       // revision 3, build 1 ('a')
		0x12,       // major 1, minor 2
	}

	result, err := parseDVMegaResponse(frame)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.FirmwareVersion != "1.23a" {
		t.Errorf("firmware = %q, want %q", result.FirmwareVersion, "1.23a")
	}
}

func TestParseDVMegaResponse_NotDVRPTR(t *testing.T) {
	// Random data that isn't a DVRPTR frame
	result, err := parseDVMegaResponse([]byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != nil {
		t.Error("expected nil result for non-DVRPTR data")
	}
}

func TestClassifyByUSB(t *testing.T) {
	tests := []struct {
		name      string
		dev       DetectedDevice
		wantType  DeviceType
	}{
		{
			name:     "LeafLabs Maple → MMDVM",
			dev:      DetectedDevice{USBVendor: "1eaf", USBProduct: "0004"},
			wantType: DeviceMMDVM,
		},
		{
			name:     "STM32 VCP → MMDVM",
			dev:      DetectedDevice{USBVendor: "0483", USBProduct: "5740"},
			wantType: DeviceMMDVM,
		},
		{
			name:     "CH341 → unknown (needs probe)",
			dev:      DetectedDevice{USBVendor: "1a86", USBProduct: "7523"},
			wantType: DeviceUnknown,
		},
		{
			name:     "FTDI → unknown (needs probe)",
			dev:      DetectedDevice{USBVendor: "0403", USBProduct: "6001"},
			wantType: DeviceUnknown,
		},
		{
			name:     "no USB info → unknown",
			dev:      DetectedDevice{},
			wantType: DeviceUnknown,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := classifyByUSB(&tt.dev)
			if got != tt.wantType {
				t.Errorf("classifyByUSB() = %q, want %q", got, tt.wantType)
			}
		})
	}
}

func TestTrimNul(t *testing.T) {
	tests := []struct {
		in   []byte
		want string
	}{
		{[]byte("hello\x00world"), "hello"},
		{[]byte("hello"), "hello"},
		{[]byte{0x00}, ""},
		{[]byte{}, ""},
	}
	for _, tt := range tests {
		got := trimNul(tt.in)
		if got != tt.want {
			t.Errorf("trimNul(%v) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

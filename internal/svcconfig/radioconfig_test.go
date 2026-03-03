package svcconfig

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/go-ini/ini"

	"github.com/MW0MWZ/Pi-Star_MCP/internal/config"
)

func TestRadioSchemaValidity(t *testing.T) {
	schema := RadioSchema()
	if len(schema) == 0 {
		t.Fatal("radio schema has no groups")
	}

	keys := make(map[string]bool)
	for _, g := range schema {
		if g.Name == "" {
			t.Error("group has empty name")
		}
		if g.I18nKey == "" {
			t.Error("group has empty i18nKey")
		}
		for _, f := range g.Fields {
			if f.Key == "" {
				t.Error("field has empty key")
			}
			if keys[f.Key] {
				t.Errorf("duplicate field key: %s", f.Key)
			}
			keys[f.Key] = true

			if len(f.Targets) == 0 {
				t.Errorf("field %s has no targets", f.Key)
			}
			if f.Label == "" {
				t.Errorf("field %s has empty label", f.Key)
			}
			validTypes := map[string]bool{"text": true, "number": true, "boolean": true, "select": true, "duplex": true}
			if !validTypes[f.FieldType] {
				t.Errorf("field %s has invalid type: %s", f.Key, f.FieldType)
			}
			for _, target := range f.Targets {
				if _, ok := config.LookupService(target.Service); !ok {
					t.Errorf("field %s target references unknown service: %s", f.Key, target.Service)
				}
			}
		}
	}
}

func TestReadRadioConfig(t *testing.T) {
	dir := t.TempDir()
	mmdvmPath := filepath.Join(dir, "MMDVM.ini")

	writeINI(t, mmdvmPath, map[string]map[string]string{
		"General": {
			"Callsign":    "MW0MWZ",
			"Id":          "2342009",
			"Duplex":      "1",
			"RXFrequency": "439500000",
			"TXFrequency": "430500000",
			"Latitude":    "51.5",
			"Longitude":   "-0.1",
			"Location":    "London",
		},
		"DMR":           {"Enable": "1"},
		"D-Star":        {"Enable": "0"},
		"System Fusion": {"Enable": "1"},
		"P25":           {"Enable": "0"},
		"NXDN":          {"Enable": "0"},
		"M17":           {"Enable": "0"},
		"POCSAG":        {"Enable": "0"},
		"FM":            {"Enable": "0"},
	})

	services := map[string]*config.ServiceEntry{
		"mmdvmhost": {ConfigPath: mmdvmPath},
	}

	values, err := ReadRadioConfig(services)
	if err != nil {
		t.Fatalf("ReadRadioConfig error: %v", err)
	}

	checks := map[string]string{
		"callsign":    "MW0MWZ",
		"dmrId":       "2342009",
		"duplex":      "1",
		"rxFrequency": "439500000",
		"txFrequency": "430500000",
		"latitude":    "51.5",
		"longitude":   "-0.1",
		"location":    "London",
		"dmrEnable":   "1",
		"dstarEnable": "0",
		"ysfEnable":   "1",
	}

	for key, want := range checks {
		if got := values[key]; got != want {
			t.Errorf("values[%s] = %q, want %q", key, got, want)
		}
	}
}

func TestReadRadioConfigMissingFile(t *testing.T) {
	services := map[string]*config.ServiceEntry{
		"mmdvmhost": {ConfigPath: "/nonexistent/MMDVM.ini"},
	}

	values, err := ReadRadioConfig(services)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should return defaults
	if values["callsign"] != "M0ABC" {
		t.Errorf("expected default callsign M0ABC, got %q", values["callsign"])
	}
}

func TestWriteRadioConfigFanOut(t *testing.T) {
	dir := t.TempDir()

	// Create 3 INI files
	mmdvmPath := filepath.Join(dir, "MMDVM.ini")
	ysfPath := filepath.Join(dir, "YSFGateway.ini")
	dmrgwPath := filepath.Join(dir, "DMRGateway.ini")

	writeINI(t, mmdvmPath, map[string]map[string]string{
		"General": {"Callsign": "OLD", "Id": "0"},
		"Info":    {"Height": "0"},
	})
	writeINI(t, ysfPath, map[string]map[string]string{
		"General": {"Callsign": "OLD"},
		"Info":    {"Latitude": "0"},
	})
	writeINI(t, dmrgwPath, map[string]map[string]string{
		"Info": {"Latitude": "0"},
	})

	services := map[string]*config.ServiceEntry{
		"mmdvmhost":  {ConfigPath: mmdvmPath},
		"ysfgateway": {ConfigPath: ysfPath},
		"dmrgateway": {ConfigPath: dmrgwPath},
	}

	values := map[string]string{
		"callsign": "MW0MWZ",
		"latitude": "52.5",
	}

	written, err := WriteRadioConfig(services, values)
	if err != nil {
		t.Fatalf("WriteRadioConfig error: %v", err)
	}
	if written != 3 {
		t.Errorf("expected 3 files written, got %d", written)
	}

	// Verify MMDVM.ini
	checkINIValue(t, mmdvmPath, "General", "Callsign", "MW0MWZ")
	checkINIValue(t, mmdvmPath, "General", "Latitude", "52.5")
	checkINIValue(t, mmdvmPath, "Info", "Latitude", "52.5")

	// Verify YSFGateway.ini
	checkINIValue(t, ysfPath, "General", "Callsign", "MW0MWZ")
	checkINIValue(t, ysfPath, "Info", "Latitude", "52.5")

	// Verify DMRGateway.ini
	checkINIValue(t, dmrgwPath, "Info", "Latitude", "52.5")
}

func TestWriteRadioConfigSkipsMissingFiles(t *testing.T) {
	dir := t.TempDir()
	mmdvmPath := filepath.Join(dir, "MMDVM.ini")

	writeINI(t, mmdvmPath, map[string]map[string]string{
		"General": {"Callsign": "OLD"},
	})

	services := map[string]*config.ServiceEntry{
		"mmdvmhost":  {ConfigPath: mmdvmPath},
		"ysfgateway": {ConfigPath: filepath.Join(dir, "nonexistent.ini")},
	}

	values := map[string]string{
		"callsign": "MW0MWZ",
	}

	written, err := WriteRadioConfig(services, values)
	if err != nil {
		t.Fatalf("WriteRadioConfig error: %v", err)
	}
	if written != 1 {
		t.Errorf("expected 1 file written (skipping missing), got %d", written)
	}

	checkINIValue(t, mmdvmPath, "General", "Callsign", "MW0MWZ")
}

func TestValidateRadioConfigSimplex(t *testing.T) {
	// Simplex mode: TX must equal RX
	values := map[string]string{
		"callsign":    "MW0MWZ",
		"dmrId":       "2342009",
		"rxFrequency": "430100000",
		"txFrequency": "439500000",
		"duplex":      "0",
	}

	errs := ValidateRadioConfig(values)
	found := false
	for _, e := range errs {
		if e.Key == "txFrequency" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected validation error for txFrequency in simplex mode")
	}

	// Should pass when frequencies match
	values["txFrequency"] = "430100000"
	errs = ValidateRadioConfig(values)
	for _, e := range errs {
		if e.Key == "txFrequency" {
			t.Errorf("unexpected txFrequency error when frequencies match: %s", e.Message)
		}
	}
}

func TestValidateRadioConfigDuplex(t *testing.T) {
	// Duplex mode: different frequencies are fine
	values := map[string]string{
		"callsign":    "MW0MWZ",
		"dmrId":       "2342009",
		"rxFrequency": "430100000",
		"txFrequency": "439500000",
		"duplex":      "1",
	}

	errs := ValidateRadioConfig(values)
	for _, e := range errs {
		if e.Key == "txFrequency" {
			t.Errorf("unexpected txFrequency error in duplex mode: %s", e.Message)
		}
	}
}

// --- helpers ---

func writeINI(t *testing.T, path string, sections map[string]map[string]string) {
	t.Helper()
	f := ini.Empty()
	for secName, keys := range sections {
		sec, _ := f.NewSection(secName)
		for k, v := range keys {
			sec.Key(k).SetValue(v)
		}
	}
	if err := f.SaveTo(path); err != nil {
		t.Fatalf("failed to write test INI %s: %v", path, err)
	}
}

func checkINIValue(t *testing.T, path, section, key, want string) {
	t.Helper()
	f, err := ini.Load(path)
	if err != nil {
		t.Fatalf("failed to load %s: %v", path, err)
	}
	got := f.Section(section).Key(key).String()
	if got != want {
		t.Errorf("%s [%s]%s = %q, want %q", filepath.Base(path), section, key, got, want)
	}
}

// Ensure the test binary doesn't complain about unused os import.
var _ = os.TempDir

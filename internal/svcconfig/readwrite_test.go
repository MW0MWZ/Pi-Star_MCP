package svcconfig

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestReadSettingsFromMissingFile(t *testing.T) {
	schema := &SettingsSchema{
		ServiceName: "test",
		Groups: []SettingsGroup{
			{
				Name:    "General",
				I18nKey: "test.general",
				Fields: []SettingsField{
					{Key: "callsign", INISection: "General", INIKey: "Callsign", FieldType: "text", Default: "M0TEST"},
					{Key: "port", INISection: "General", INIKey: "Port", FieldType: "number", Default: "8080"},
				},
			},
		},
	}

	values, err := ReadSettings(schema, "/tmp/nonexistent_test_ini_file.ini")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if values["callsign"] != "M0TEST" {
		t.Errorf("expected default callsign M0TEST, got %q", values["callsign"])
	}
	if values["port"] != "8080" {
		t.Errorf("expected default port 8080, got %q", values["port"])
	}
}

func TestReadSettingsFromFile(t *testing.T) {
	dir := t.TempDir()
	iniPath := filepath.Join(dir, "test.ini")
	iniContent := `[General]
Callsign=MW0MWZ
Port=9090

[Other]
Something=preserved
`
	if err := os.WriteFile(iniPath, []byte(iniContent), 0644); err != nil {
		t.Fatal(err)
	}

	schema := &SettingsSchema{
		ServiceName: "test",
		Groups: []SettingsGroup{
			{
				Name:    "General",
				I18nKey: "test.general",
				Fields: []SettingsField{
					{Key: "callsign", INISection: "General", INIKey: "Callsign", FieldType: "text", Default: "M0TEST"},
					{Key: "port", INISection: "General", INIKey: "Port", FieldType: "number", Default: "8080"},
				},
			},
		},
	}

	values, err := ReadSettings(schema, iniPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if values["callsign"] != "MW0MWZ" {
		t.Errorf("expected callsign MW0MWZ, got %q", values["callsign"])
	}
	if values["port"] != "9090" {
		t.Errorf("expected port 9090, got %q", values["port"])
	}
}

func TestWriteSettingsPreservesOtherKeys(t *testing.T) {
	dir := t.TempDir()
	iniPath := filepath.Join(dir, "test.ini")

	// Write initial file with extra content
	iniContent := `[General]
Callsign = MW0MWZ
ExtraKey = should_be_preserved
Port = 9090

[OtherSection]
Foo = bar
`
	if err := os.WriteFile(iniPath, []byte(iniContent), 0644); err != nil {
		t.Fatal(err)
	}

	schema := &SettingsSchema{
		ServiceName: "test",
		Groups: []SettingsGroup{
			{
				Name:    "General",
				I18nKey: "test.general",
				Fields: []SettingsField{
					{Key: "callsign", INISection: "General", INIKey: "Callsign", FieldType: "text"},
					{Key: "port", INISection: "General", INIKey: "Port", FieldType: "number"},
				},
			},
		},
	}

	// Write new values for only callsign
	err := WriteSettings(schema, iniPath, map[string]string{
		"callsign": "G0ABC",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Read back and verify
	data, err := os.ReadFile(iniPath)
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)

	if !strings.Contains(content, "G0ABC") {
		t.Error("expected callsign G0ABC in written file")
	}
	if !strings.Contains(content, "ExtraKey") {
		t.Error("expected ExtraKey to be preserved")
	}
	if !strings.Contains(content, "OtherSection") {
		t.Error("expected OtherSection to be preserved")
	}
	if !strings.Contains(content, "Foo") {
		t.Error("expected Foo key in OtherSection to be preserved")
	}
}

func TestWriteSettingsCreatesNewFile(t *testing.T) {
	dir := t.TempDir()
	iniPath := filepath.Join(dir, "new.ini")

	schema := &SettingsSchema{
		ServiceName: "test",
		Groups: []SettingsGroup{
			{
				Name:    "General",
				I18nKey: "test.general",
				Fields: []SettingsField{
					{Key: "callsign", INISection: "General", INIKey: "Callsign", FieldType: "text"},
				},
			},
		},
	}

	err := WriteSettings(schema, iniPath, map[string]string{
		"callsign": "MW0MWZ",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile(iniPath)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "MW0MWZ") {
		t.Error("expected MW0MWZ in new file")
	}
}

func TestWriteSettingsIgnoresUnknownKeys(t *testing.T) {
	dir := t.TempDir()
	iniPath := filepath.Join(dir, "test.ini")

	schema := &SettingsSchema{
		ServiceName: "test",
		Groups: []SettingsGroup{
			{
				Name:    "General",
				I18nKey: "test.general",
				Fields: []SettingsField{
					{Key: "callsign", INISection: "General", INIKey: "Callsign", FieldType: "text"},
				},
			},
		},
	}

	err := WriteSettings(schema, iniPath, map[string]string{
		"callsign":    "MW0MWZ",
		"unknown_key": "should_be_ignored",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile(iniPath)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(data), "unknown_key") {
		t.Error("unknown_key should not appear in output")
	}
}

func TestReadWriteRoundTrip(t *testing.T) {
	dir := t.TempDir()
	iniPath := filepath.Join(dir, "roundtrip.ini")

	schema, ok := LookupSchema("mmdvmhost")
	if !ok {
		t.Fatal("mmdvmhost schema not found")
	}

	// Write settings (only service-specific fields remain in mmdvmhost schema)
	values := map[string]string{
		"colorCode":  "4",
		"dstarModule": "B",
	}
	if err := WriteSettings(schema, iniPath, values); err != nil {
		t.Fatalf("write error: %v", err)
	}

	// Read back
	got, err := ReadSettings(schema, iniPath)
	if err != nil {
		t.Fatalf("read error: %v", err)
	}

	for key, want := range values {
		if got[key] != want {
			t.Errorf("round-trip mismatch for %q: want %q, got %q", key, want, got[key])
		}
	}
}

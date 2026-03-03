// Package svcconfig provides schema-driven reading and writing of
// curated service settings. Each service registers a SettingsSchema
// that describes which INI keys are exposed via the admin API. The
// frontend renders forms dynamically from these schemas.
package svcconfig

import (
	"fmt"
	"os"

	"github.com/go-ini/ini"
)

// SettingsSchema describes the curated settings for one service.
type SettingsSchema struct {
	ServiceName string          `json:"serviceName"`
	Groups      []SettingsGroup `json:"groups"`
}

// SettingsGroup is a labelled cluster of related fields.
type SettingsGroup struct {
	Name    string          `json:"name"`
	I18nKey string          `json:"i18nKey"`
	Fields  []SettingsField `json:"fields"`
}

// SettingsField describes a single editable setting.
type SettingsField struct {
	Key        string   `json:"key"`                  // API key: "callsign"
	I18nLabel  string   `json:"i18nLabel"`             // i18n key for display
	INISection string   `json:"iniSection"`            // section in service INI file
	INIKey     string   `json:"iniKey"`                // key in service INI file
	FieldType  string   `json:"fieldType"`             // "text", "number", "boolean", "select"
	Validate   string   `json:"validate"`              // data-validate rules
	Options    []Option `json:"options,omitempty"`      // for select fields
	Default    string   `json:"default"`               // default value
	HelpI18n   string   `json:"helpI18n,omitempty"`    // i18n key for help text
}

// Option is a choice for select-type fields.
type Option struct {
	Value   string `json:"value"`
	I18nKey string `json:"i18nKey"`
}

// SchemaRegistry maps service name to its settings schema.
var SchemaRegistry = map[string]*SettingsSchema{}

// LookupSchema returns the settings schema for a named service.
func LookupSchema(name string) (*SettingsSchema, bool) {
	s, ok := SchemaRegistry[name]
	return s, ok
}

// ReadSettings reads curated field values from a service INI file.
// Returns a map of field Key → current value. Missing keys use the
// field's default value.
func ReadSettings(schema *SettingsSchema, iniPath string) (map[string]string, error) {
	values := make(map[string]string)

	// If the file does not exist, return all defaults.
	if _, err := os.Stat(iniPath); os.IsNotExist(err) {
		for _, g := range schema.Groups {
			for _, f := range g.Fields {
				values[f.Key] = f.Default
			}
		}
		return values, nil
	}

	f, err := ini.LoadSources(ini.LoadOptions{
		SkipUnrecognizableLines: true,
	}, iniPath)
	if err != nil {
		return nil, fmt.Errorf("read settings from %s: %w", iniPath, err)
	}

	for _, g := range schema.Groups {
		for _, field := range g.Fields {
			sec := f.Section(field.INISection)
			if sec.HasKey(field.INIKey) {
				values[field.Key] = sec.Key(field.INIKey).String()
			} else {
				values[field.Key] = field.Default
			}
		}
	}

	return values, nil
}

// WriteSettings writes curated field values to a service INI file.
// Only keys defined in the schema are touched — all other sections,
// keys, and comments are preserved.
func WriteSettings(schema *SettingsSchema, iniPath string, values map[string]string) error {
	var f *ini.File
	var err error

	if _, statErr := os.Stat(iniPath); os.IsNotExist(statErr) {
		f = ini.Empty()
	} else {
		f, err = ini.LoadSources(ini.LoadOptions{
			SkipUnrecognizableLines: true,
		}, iniPath)
		if err != nil {
			return fmt.Errorf("load INI for write %s: %w", iniPath, err)
		}
	}

	// Build a lookup of field Key → SettingsField for quick access.
	fieldMap := make(map[string]SettingsField)
	for _, g := range schema.Groups {
		for _, field := range g.Fields {
			fieldMap[field.Key] = field
		}
	}

	// Write only the values that were provided.
	for key, value := range values {
		field, ok := fieldMap[key]
		if !ok {
			continue // ignore unknown keys
		}
		sec := f.Section(field.INISection)
		sec.Key(field.INIKey).SetValue(value)
	}

	return f.SaveTo(iniPath)
}

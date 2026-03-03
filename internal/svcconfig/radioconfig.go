package svcconfig

import (
	"fmt"
	"os"

	"github.com/go-ini/ini"

	"github.com/MW0MWZ/Pi-Star_MCP/internal/config"
)

// RadioTarget describes a single INI location that a radio field writes to.
type RadioTarget struct {
	Service    string // config.Registry name (e.g. "mmdvmhost")
	INISection string
	INIKey     string
}

// BridgeService describes a cross-mode bridge that belongs under a
// parent mode enable (e.g. DMR2YSF under DMR).
type BridgeService struct {
	Service string `json:"service"` // service registry name
	Label   string `json:"label"`   // display name, e.g. "DMR → YSF"
}

// RadioField describes a single radio configuration field that fans out
// to one or more service INI files.
type RadioField struct {
	Key       string          `json:"key"`
	Label     string          `json:"label"`               // English display name (i18n fallback)
	I18nLabel string          `json:"i18nLabel"`
	FieldType string          `json:"fieldType"`
	Validate  string          `json:"validate"`
	Default   string          `json:"default"`
	HelpI18n  string          `json:"helpI18n,omitempty"`
	Options   []Option        `json:"options,omitempty"`
	Bridges   []BridgeService `json:"bridges,omitempty"`   // cross-mode bridges for mode enables
	Targets   []RadioTarget   `json:"-"`
}

// RadioGroup is a labelled cluster of related radio fields.
type RadioGroup struct {
	Name    string       `json:"name"`
	I18nKey string       `json:"i18nKey"`
	Fields  []RadioField `json:"fields"`
}

// radioSchema is the singleton schema built once at init time.
var radioSchema []RadioGroup

func init() {
	radioSchema = buildRadioSchema()
}

// RadioSchema returns the radio configuration schema for the frontend.
func RadioSchema() []RadioGroup {
	return radioSchema
}

func buildRadioSchema() []RadioGroup {
	return []RadioGroup{
		{
			Name:    "Station Identity",
			I18nKey: "radio.identity",
			Fields: []RadioField{
				{
					Key:       "callsign",
					Label:     "Callsign",
					I18nLabel: "radio.callsign",
					FieldType: "text",
					Validate:  "required,callsign",
					Default:   "M0ABC",
					HelpI18n:  "radio.callsign.help",
					Targets: []RadioTarget{
						{"mmdvmhost", "General", "Callsign"},
						{"ysfgateway", "General", "Callsign"},
						{"dgidgateway", "General", "Callsign"},
						{"p25gateway", "General", "Callsign"},
						{"nxdngateway", "General", "Callsign"},
						{"aprsgateway", "General", "Callsign"},
						{"dapnetgateway", "General", "Callsign"},
						{"fmgateway", "General", "Callsign"},
						{"dmr2ysf", "YSF Network", "Callsign"},
						{"ysf2dmr", "YSF Network", "Callsign"},
						{"ysf2nxdn", "YSF Network", "Callsign"},
						{"ysf2p25", "YSF Network", "Callsign"},
						{"nxdn2dmr", "NXDN Network", "Callsign"},
					},
				},
				{
					Key:       "dmrId",
					Label:     "Radio ID",
					I18nLabel: "radio.dmrId",
					FieldType: "number",
					Validate:  "required,numeric,range:1:9999999",
					Default:   "1234567",
					HelpI18n:  "radio.dmrId.help",
					Targets: []RadioTarget{
						{"mmdvmhost", "General", "Id"},
						{"ysfgateway", "General", "Id"},
						{"dgidgateway", "General", "Id"},
						{"dmr2ysf", "DMR Network", "Id"},
						{"dmr2nxdn", "DMR Network", "Id"},
						{"ysf2dmr", "DMR Network", "Id"},
						{"ysf2p25", "P25 Network", "Id"},
						{"nxdn2dmr", "DMR Network", "Id"},
					},
				},
				{
					Key:       "nxdnId",
					Label:     "NXDN ID",
					I18nLabel: "radio.nxdnId",
					FieldType: "number",
					Validate:  "numeric,range:1:65535",
					Default:   "",
					HelpI18n:  "radio.nxdnId.help",
					Targets: []RadioTarget{
						{"ysf2nxdn", "NXDN Network", "Id"},
					},
				},
			},
		},
		{
			Name:    "Frequencies",
			I18nKey: "radio.frequencies",
			Fields: []RadioField{
				{
					Key:       "duplex",
					Label:     "Duplex Mode",
					I18nLabel: "radio.duplex",
					FieldType: "duplex",
					Default:   "0",
					Targets: []RadioTarget{
						{"mmdvmhost", "General", "Duplex"},
					},
				},
				{
					Key:       "rxFrequency",
					Label:     "RX Frequency",
					I18nLabel: "radio.rxFrequency",
					FieldType: "text",
					Validate:  "required,numeric",
					Default:   "430100000",
					HelpI18n:  "radio.rxFrequency.help",
					Targets: []RadioTarget{
						{"mmdvmhost", "General", "RXFrequency"},
						{"mmdvmhost", "Info", "RXFrequency"},
						{"ysfgateway", "Info", "RXFrequency"},
						{"nxdngateway", "Info", "RXFrequency"},
						{"dgidgateway", "Info", "RXFrequency"},
						{"ysf2dmr", "Info", "RXFrequency"},
						{"ysf2nxdn", "Info", "RXFrequency"},
						{"ysf2p25", "Info", "RXFrequency"},
						{"nxdn2dmr", "Info", "RXFrequency"},
					},
				},
				{
					Key:       "txFrequency",
					Label:     "TX Frequency",
					I18nLabel: "radio.txFrequency",
					FieldType: "text",
					Validate:  "required,numeric",
					Default:   "430100000",
					HelpI18n:  "radio.txFrequency.help",
					Targets: []RadioTarget{
						{"mmdvmhost", "General", "TXFrequency"},
						{"mmdvmhost", "Info", "TXFrequency"},
						{"ysfgateway", "Info", "TXFrequency"},
						{"nxdngateway", "Info", "TXFrequency"},
						{"dgidgateway", "Info", "TXFrequency"},
						{"ysf2dmr", "Info", "TXFrequency"},
						{"ysf2nxdn", "Info", "TXFrequency"},
						{"ysf2p25", "Info", "TXFrequency"},
						{"nxdn2dmr", "Info", "TXFrequency"},
					},
				},
			},
		},
		{
			Name:    "Location",
			I18nKey: "radio.location",
			Fields: []RadioField{
				{
					Key:       "latitude",
					Label:     "Latitude",
					I18nLabel: "radio.latitude",
					FieldType: "text",
					Validate:  "decimal,range:-90:90",
					Default:   "0.0",
					Targets: []RadioTarget{
						{"mmdvmhost", "General", "Latitude"},
						{"mmdvmhost", "Info", "Latitude"},
						{"dmrgateway", "Info", "Latitude"},
						{"ysfgateway", "Info", "Latitude"},
						{"nxdngateway", "Info", "Latitude"},
						{"dgidgateway", "Info", "Latitude"},
						{"ysf2dmr", "Info", "Latitude"},
						{"ysf2nxdn", "Info", "Latitude"},
						{"nxdn2dmr", "Info", "Latitude"},
					},
				},
				{
					Key:       "longitude",
					Label:     "Longitude",
					I18nLabel: "radio.longitude",
					FieldType: "text",
					Validate:  "decimal,range:-180:180",
					Default:   "0.0",
					Targets: []RadioTarget{
						{"mmdvmhost", "General", "Longitude"},
						{"mmdvmhost", "Info", "Longitude"},
						{"dmrgateway", "Info", "Longitude"},
						{"ysfgateway", "Info", "Longitude"},
						{"nxdngateway", "Info", "Longitude"},
						{"dgidgateway", "Info", "Longitude"},
						{"ysf2dmr", "Info", "Longitude"},
						{"ysf2nxdn", "Info", "Longitude"},
						{"nxdn2dmr", "Info", "Longitude"},
					},
				},
				{
					Key:       "height",
					Label:     "Height (metres)",
					I18nLabel: "radio.height",
					FieldType: "number",
					Validate:  "numeric",
					Default:   "0",
					Targets: []RadioTarget{
						{"mmdvmhost", "Info", "Height"},
						{"dmrgateway", "Info", "Height"},
						{"ysfgateway", "Info", "Height"},
						{"nxdngateway", "Info", "Height"},
						{"dgidgateway", "Info", "Height"},
						{"ysf2dmr", "Info", "Height"},
						{"ysf2nxdn", "Info", "Height"},
						{"nxdn2dmr", "Info", "Height"},
					},
				},
				{
					Key:       "location",
					Label:     "Location Name",
					I18nLabel: "radio.locationName",
					FieldType: "text",
					Validate:  "maxlen:20",
					Default:   "",
					Targets: []RadioTarget{
						{"mmdvmhost", "General", "Location"},
						{"mmdvmhost", "Info", "Location"},
						{"dmrgateway", "Info", "Location"},
						{"ysfgateway", "Info", "Name"},
						{"nxdngateway", "Info", "Name"},
						{"ysf2dmr", "Info", "Location"},
						{"nxdn2dmr", "Info", "Location"},
					},
				},
				{
					Key:       "description",
					Label:     "Description",
					I18nLabel: "radio.description",
					FieldType: "text",
					Validate:  "maxlen:40",
					Default:   "",
					Targets: []RadioTarget{
						{"mmdvmhost", "General", "Description"},
						{"mmdvmhost", "Info", "Description"},
						{"dmrgateway", "Info", "Description"},
						{"ysfgateway", "Info", "Description"},
						{"nxdngateway", "Info", "Description"},
						{"dgidgateway", "Info", "Description"},
						{"ysf2dmr", "Info", "Description"},
						{"ysf2nxdn", "Info", "Description"},
						{"nxdn2dmr", "Info", "Description"},
					},
				},
				{
					Key:       "url",
					Label:     "URL",
					I18nLabel: "radio.url",
					FieldType: "text",
					Default:   "",
					Targets: []RadioTarget{
						{"mmdvmhost", "General", "URL"},
						{"mmdvmhost", "Info", "URL"},
						{"dmrgateway", "Info", "URL"},
						{"ysf2dmr", "Info", "URL"},
						{"nxdn2dmr", "Info", "URL"},
					},
				},
				{
					Key:       "power",
					Label:     "Power (watts)",
					I18nLabel: "radio.power",
					FieldType: "number",
					Validate:  "numeric",
					Default:   "1",
					Targets: []RadioTarget{
						{"mmdvmhost", "Info", "Power"},
						{"ysfgateway", "Info", "Power"},
						{"nxdngateway", "Info", "Power"},
						{"ysf2dmr", "Info", "Power"},
						{"nxdn2dmr", "Info", "Power"},
					},
				},
			},
		},
		{
			Name:    "Mode Enables",
			I18nKey: "radio.modes",
			Fields: []RadioField{
				{
					Key:       "dstarEnable",
					Label:     "D-Star",
					I18nLabel: "radio.dstarEnable",
					FieldType: "boolean",
					Default:   "0",
					Targets:   []RadioTarget{{"mmdvmhost", "D-Star", "Enable"}},
				},
				{
					Key:       "dmrEnable",
					Label:     "DMR",
					I18nLabel: "radio.dmrEnable",
					FieldType: "boolean",
					Default:   "1",
					Targets:   []RadioTarget{{"mmdvmhost", "DMR", "Enable"}},
					Bridges: []BridgeService{
						{Service: "dmr2ysf", Label: "DMR \u2192 YSF"},
						{Service: "dmr2nxdn", Label: "DMR \u2192 NXDN"},
					},
				},
				{
					Key:       "ysfEnable",
					Label:     "System Fusion",
					I18nLabel: "radio.ysfEnable",
					FieldType: "boolean",
					Default:   "0",
					Targets:   []RadioTarget{{"mmdvmhost", "System Fusion", "Enable"}},
					Bridges: []BridgeService{
						{Service: "ysf2dmr", Label: "YSF \u2192 DMR"},
						{Service: "ysf2nxdn", Label: "YSF \u2192 NXDN"},
						{Service: "ysf2p25", Label: "YSF \u2192 P25"},
					},
				},
				{
					Key:       "p25Enable",
					Label:     "P25",
					I18nLabel: "radio.p25Enable",
					FieldType: "boolean",
					Default:   "0",
					Targets:   []RadioTarget{{"mmdvmhost", "P25", "Enable"}},
				},
				{
					Key:       "nxdnEnable",
					Label:     "NXDN",
					I18nLabel: "radio.nxdnEnable",
					FieldType: "boolean",
					Default:   "0",
					Targets:   []RadioTarget{{"mmdvmhost", "NXDN", "Enable"}},
					Bridges: []BridgeService{
						{Service: "nxdn2dmr", Label: "NXDN \u2192 DMR"},
					},
				},
				{
					Key:       "pocsagEnable",
					Label:     "POCSAG",
					I18nLabel: "radio.pocsagEnable",
					FieldType: "boolean",
					Default:   "0",
					Targets:   []RadioTarget{{"mmdvmhost", "POCSAG", "Enable"}},
				},
				{
					Key:       "fmEnable",
					Label:     "FM",
					I18nLabel: "radio.fmEnable",
					FieldType: "boolean",
					Default:   "0",
					Targets:   []RadioTarget{{"mmdvmhost", "FM", "Enable"}},
				},
			},
		},
	}
}

// primaryReadTargets maps each radio field key to the INI section+key
// used when reading from MMDVM.ini (the primary source of truth).
var primaryReadTargets = map[string][2]string{
	"callsign":    {"General", "Callsign"},
	"dmrId":       {"General", "Id"},
	"duplex":      {"General", "Duplex"},
	"rxFrequency": {"General", "RXFrequency"},
	"txFrequency": {"General", "TXFrequency"},
	"latitude":    {"General", "Latitude"},
	"longitude":   {"General", "Longitude"},
	"location":    {"General", "Location"},
	"description": {"General", "Description"},
	"url":         {"General", "URL"},
	"height":      {"Info", "Height"},
	"power":       {"Info", "Power"},
	"dstarEnable": {"D-Star", "Enable"},
	"dmrEnable":   {"DMR", "Enable"},
	"ysfEnable":   {"System Fusion", "Enable"},
	"p25Enable":   {"P25", "Enable"},
	"nxdnEnable":  {"NXDN", "Enable"},
	"pocsagEnable": {"POCSAG", "Enable"},
	"fmEnable":    {"FM", "Enable"},
}

// ReadRadioConfig reads the current radio configuration values from
// MMDVM.ini (the primary source of truth). Fields not present in
// MMDVM.ini (like nxdnId) return their defaults.
func ReadRadioConfig(services map[string]*config.ServiceEntry) (map[string]string, error) {
	values := make(map[string]string)

	// Collect all field defaults
	for _, g := range radioSchema {
		for _, f := range g.Fields {
			values[f.Key] = f.Default
		}
	}

	// Read from MMDVM.ini
	mmdvmPath := configPath(services, "mmdvmhost")
	if mmdvmPath == "" {
		return values, nil
	}
	if _, err := os.Stat(mmdvmPath); os.IsNotExist(err) {
		return values, nil
	}

	f, err := ini.LoadSources(ini.LoadOptions{
		SkipUnrecognizableLines: true,
	}, mmdvmPath)
	if err != nil {
		return nil, fmt.Errorf("read radio config from %s: %w", mmdvmPath, err)
	}

	for key, target := range primaryReadTargets {
		sec := f.Section(target[0])
		if sec.HasKey(target[1]) {
			values[key] = sec.Key(target[1]).String()
		}
	}

	// nxdnId comes from YSF2NXDN.ini, not MMDVM.ini
	nxdnPath := configPath(services, "ysf2nxdn")
	if nxdnPath != "" {
		if _, err := os.Stat(nxdnPath); err == nil {
			nf, err := ini.LoadSources(ini.LoadOptions{
				SkipUnrecognizableLines: true,
			}, nxdnPath)
			if err == nil {
				sec := nf.Section("NXDN Network")
				if sec.HasKey("Id") {
					values["nxdnId"] = sec.Key("Id").String()
				}
			}
		}
	}

	return values, nil
}

// WriteRadioConfig writes radio configuration values to all target INI
// files. Only files that exist on disk are updated (missing = service
// not installed). Returns the number of files written.
func WriteRadioConfig(services map[string]*config.ServiceEntry, values map[string]string) (int, error) {
	// Build a map of field key → RadioField for lookup.
	fieldMap := make(map[string]RadioField)
	for _, g := range radioSchema {
		for _, f := range g.Fields {
			fieldMap[f.Key] = f
		}
	}

	// Group writes by config file path: path → list of (section, key, value).
	type iniWrite struct {
		Section string
		Key     string
		Value   string
	}
	fileWrites := make(map[string][]iniWrite)

	for key, value := range values {
		field, ok := fieldMap[key]
		if !ok {
			continue
		}
		for _, target := range field.Targets {
			path := configPath(services, target.Service)
			if path == "" {
				continue
			}
			if _, err := os.Stat(path); os.IsNotExist(err) {
				continue
			}
			fileWrites[path] = append(fileWrites[path], iniWrite{
				Section: target.INISection,
				Key:     target.INIKey,
				Value:   value,
			})
		}
	}

	// Write each file once.
	written := 0
	for path, writes := range fileWrites {
		f, err := ini.LoadSources(ini.LoadOptions{
			SkipUnrecognizableLines: true,
		}, path)
		if err != nil {
			return written, fmt.Errorf("load %s for radio write: %w", path, err)
		}
		for _, w := range writes {
			f.Section(w.Section).Key(w.Key).SetValue(w.Value)
		}
		if err := f.SaveTo(path); err != nil {
			return written, fmt.Errorf("save %s: %w", path, err)
		}
		written++
	}

	return written, nil
}

// ValidateRadioConfig checks radio values against the schema's
// validation rules plus cross-field rules (simplex frequency sync).
func ValidateRadioConfig(values map[string]string) []FieldError {
	var errs []FieldError

	for _, g := range radioSchema {
		for _, field := range g.Fields {
			if field.Validate == "" {
				continue
			}
			value := values[field.Key]
			if err := validateValue(field.Key, value, field.Validate); err != nil {
				errs = append(errs, *err)
			}
		}
	}

	// Simplex rule: when duplex=0, txFrequency must equal rxFrequency.
	if values["duplex"] == "0" && values["txFrequency"] != values["rxFrequency"] {
		errs = append(errs, FieldError{
			Key:     "txFrequency",
			Message: "must equal RX frequency in simplex mode",
		})
	}

	return errs
}

// configPath returns the INI file path for a service, preferring the
// entry's ConfigPath over the registry default.
func configPath(services map[string]*config.ServiceEntry, name string) string {
	entry := services[name]
	if entry != nil && entry.ConfigPath != "" {
		return entry.ConfigPath
	}
	def, ok := config.LookupService(name)
	if !ok {
		return ""
	}
	return def.DefaultConfigPath
}

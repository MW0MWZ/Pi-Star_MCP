// Package config handles loading, validation, and defaults for the
// Pi-Star dashboard configuration file (dashboard.ini).
package config

import (
	"fmt"
	"os"

	"github.com/go-ini/ini"
)

// Config holds all configuration sections for the dashboard.
type Config struct {
	Dashboard DashboardConfig
	Security  SecurityConfig
	TLS       TLSConfig
	Paths     PathsConfig
	MQTT      MQTTConfig
	Services  map[string]*ServiceEntry
}

// DashboardConfig holds the [dashboard] section.
type DashboardConfig struct {
	ListenHTTP  string
	ListenHTTPS string
	ModulesDir  string
}

// SecurityConfig holds the [security] section.
type SecurityConfig struct {
	AuthUser       string
	SessionTimeout int
	SessionSecret  string
}

// TLSConfig holds the [tls] section.
type TLSConfig struct {
	CertFile    string
	KeyFile     string
	AutoGenerate bool
	MinVersion  string
}

// PathsConfig holds the [paths] section.
type PathsConfig struct {
	CertsDir   string
	DBDir      string
	BackupDir  string
	AuditLog   string
	RuntimeDir string
}

// MQTTConfig holds the [mqtt] section.
type MQTTConfig struct {
	Port          int
	FallbackPort  int
	MosquittoPath string
}

// ServiceEntry describes a single managed service.
type ServiceEntry struct {
	Enabled    bool
	BinaryPath string
	ConfigPath string
	HWType     string // DStarRepeater only: hardware variant key (e.g. "dvap", "gmsk")
}

// Load reads an INI config file and returns a Config with defaults
// applied for any missing values. If the file does not exist, a
// Config with all defaults is returned (first-boot behaviour).
func Load(path string) (*Config, error) {
	cfg := defaults()

	if _, err := os.Stat(path); os.IsNotExist(err) {
		// No config file — use defaults (first boot)
		return cfg, nil
	}

	f, err := ini.Load(path)
	if err != nil {
		return nil, fmt.Errorf("parse config %s: %w", path, err)
	}

	loadDashboard(f, cfg)
	loadSecurity(f, cfg)
	loadTLS(f, cfg)
	loadPaths(f, cfg)
	loadMQTT(f, cfg)
	loadServices(f, cfg)

	if err := validate(cfg); err != nil {
		return nil, fmt.Errorf("config validation: %w", err)
	}

	return cfg, nil
}

// defaults returns a Config populated with sensible default values
// so the dashboard can start with a minimal or absent config file.
func defaults() *Config {
	return &Config{
		Dashboard: DashboardConfig{
			ListenHTTP:  ":80",
			ListenHTTPS: ":443",
			ModulesDir:  "/opt/pistar/modules",
		},
		Security: SecurityConfig{
			AuthUser:       "pi-star",
			SessionTimeout: 1800,
		},
		TLS: TLSConfig{
			CertFile:     "/etc/pistar-dashboard/certs/server.crt",
			KeyFile:      "/etc/pistar-dashboard/certs/server.key",
			AutoGenerate: true,
			MinVersion:   "1.2",
		},
		Paths: PathsConfig{
			CertsDir:   "/etc/pistar-dashboard/certs",
			DBDir:      "/var/lib/pistar-dashboard",
			BackupDir:  "/var/lib/pistar-dashboard/backups",
			AuditLog:   "/var/log/pistar-dashboard/audit.log",
			RuntimeDir: "/run/pistar",
		},
		MQTT: MQTTConfig{
			Port:          1883,
			FallbackPort:  1884,
			MosquittoPath: "/usr/sbin/mosquitto",
		},
		Services: defaultServices(),
	}
}

func loadDashboard(f *ini.File, cfg *Config) {
	s := f.Section("dashboard")
	if v := s.Key("listen_http").String(); v != "" {
		cfg.Dashboard.ListenHTTP = v
	}
	if v := s.Key("listen_https").String(); v != "" {
		cfg.Dashboard.ListenHTTPS = v
	}
	if v := s.Key("modules_dir").String(); v != "" {
		cfg.Dashboard.ModulesDir = v
	}
}

func loadSecurity(f *ini.File, cfg *Config) {
	s := f.Section("security")
	if v := s.Key("auth_user").String(); v != "" {
		cfg.Security.AuthUser = v
	}
	if v, err := s.Key("session_timeout").Int(); err == nil {
		cfg.Security.SessionTimeout = v
	}
	if v := s.Key("session_secret").String(); v != "" {
		cfg.Security.SessionSecret = v
	}
}

func loadTLS(f *ini.File, cfg *Config) {
	s := f.Section("tls")
	if v := s.Key("cert_file").String(); v != "" {
		cfg.TLS.CertFile = v
	}
	if v := s.Key("key_file").String(); v != "" {
		cfg.TLS.KeyFile = v
	}
	if v, err := s.Key("auto_generate").Int(); err == nil {
		cfg.TLS.AutoGenerate = v != 0
	}
	if v := s.Key("min_version").String(); v != "" {
		cfg.TLS.MinVersion = v
	}
}

func loadPaths(f *ini.File, cfg *Config) {
	s := f.Section("paths")
	if v := s.Key("certs_dir").String(); v != "" {
		cfg.Paths.CertsDir = v
	}
	if v := s.Key("db_dir").String(); v != "" {
		cfg.Paths.DBDir = v
	}
	if v := s.Key("backup_dir").String(); v != "" {
		cfg.Paths.BackupDir = v
	}
	if v := s.Key("audit_log").String(); v != "" {
		cfg.Paths.AuditLog = v
	}
	if v := s.Key("runtime_dir").String(); v != "" {
		cfg.Paths.RuntimeDir = v
	}
}

func loadMQTT(f *ini.File, cfg *Config) {
	s := f.Section("mqtt")
	if v, err := s.Key("port").Int(); err == nil {
		cfg.MQTT.Port = v
	}
	if v, err := s.Key("fallback_port").Int(); err == nil {
		cfg.MQTT.FallbackPort = v
	}
	if v := s.Key("mosquitto_path").String(); v != "" {
		cfg.MQTT.MosquittoPath = v
	}
}

func loadServices(f *ini.File, cfg *Config) {
	s := f.Section("services")
	for _, name := range ServiceNames() {
		entry := cfg.Services[name]
		loadServiceEntry(s, name, entry)
	}

	// Resolve DStarRepeater variant paths from hwtype.
	if entry := cfg.Services["dstarrepeater"]; entry != nil {
		if v := s.Key("dstarrepeater_hwtype").String(); v != "" {
			entry.HWType = v
		}
		if entry.HWType == "" {
			entry.HWType = DStarVariants[0].Key
		}
		ResolveDStarPaths(entry)
	}
}

func loadServiceEntry(s *ini.Section, prefix string, entry *ServiceEntry) {
	if v, err := s.Key(prefix + "_enabled").Int(); err == nil {
		entry.Enabled = v != 0
	}
	if v := s.Key(prefix + "_path").String(); v != "" {
		entry.BinaryPath = v
	}
	if v := s.Key(prefix + "_config").String(); v != "" {
		entry.ConfigPath = v
	}
}

// defaultServices builds the Services map from the registry defaults.
func defaultServices() map[string]*ServiceEntry {
	services := make(map[string]*ServiceEntry, len(Registry))
	for name, def := range Registry {
		services[name] = &ServiceEntry{
			Enabled:    def.DefaultEnabled,
			BinaryPath: def.DefaultBinaryPath,
			ConfigPath: def.DefaultConfigPath,
		}
	}
	return services
}

// SaveServices writes the [services] section of the config to the INI
// file at path, preserving all other sections. Only keys that differ
// from registry defaults are written.
func SaveServices(path string, services map[string]*ServiceEntry) error {
	f, err := ini.Load(path)
	if err != nil {
		if os.IsNotExist(err) {
			f = ini.Empty()
		} else {
			return fmt.Errorf("load config for save: %w", err)
		}
	}

	s := f.Section("services")

	// Clear existing service keys so removed services don't linger.
	for _, key := range s.Keys() {
		s.DeleteKey(key.Name())
	}

	for _, name := range ServiceNames() {
		entry, ok := services[name]
		if !ok {
			continue
		}
		def, _ := LookupService(name)

		if entry.Enabled != def.DefaultEnabled {
			enabledVal := "0"
			if entry.Enabled {
				enabledVal = "1"
			}
			s.Key(name + "_enabled").SetValue(enabledVal)
		} else if entry.Enabled {
			// Always write enabled=1 for enabled services for clarity.
			s.Key(name + "_enabled").SetValue("1")
		}

		// DStarRepeater paths are resolved from hwtype, not stored directly.
		if name == "dstarrepeater" {
			continue
		}

		if entry.BinaryPath != def.DefaultBinaryPath {
			s.Key(name + "_path").SetValue(entry.BinaryPath)
		}
		if entry.ConfigPath != def.DefaultConfigPath {
			s.Key(name + "_config").SetValue(entry.ConfigPath)
		}
	}

	// Write DStarRepeater hardware type.
	if entry, ok := services["dstarrepeater"]; ok && entry.HWType != "" {
		s.Key("dstarrepeater_hwtype").SetValue(entry.HWType)
	}

	return f.SaveTo(path)
}

// ResolveDStarPaths updates a DStarRepeater ServiceEntry's BinaryPath
// and ConfigPath based on the selected hardware variant (HWType).
func ResolveDStarPaths(entry *ServiceEntry) {
	variant := LookupDStarVariant(entry.HWType)
	entry.BinaryPath = "/usr/bin/" + variant.BinaryName
	entry.ConfigPath = "/etc/dstarrepeater/" + variant.ConfigFile
}

// validate checks that configuration values are within acceptable ranges.
func validate(cfg *Config) error {
	if cfg.MQTT.Port < 1 || cfg.MQTT.Port > 65535 {
		return fmt.Errorf("mqtt.port %d out of range (1-65535)", cfg.MQTT.Port)
	}
	if cfg.MQTT.FallbackPort < 1 || cfg.MQTT.FallbackPort > 65535 {
		return fmt.Errorf("mqtt.fallback_port %d out of range (1-65535)", cfg.MQTT.FallbackPort)
	}
	if cfg.Security.SessionTimeout < 60 {
		return fmt.Errorf("security.session_timeout %d too low (minimum 60)", cfg.Security.SessionTimeout)
	}
	if cfg.TLS.MinVersion != "1.2" && cfg.TLS.MinVersion != "1.3" {
		return fmt.Errorf("tls.min_version must be 1.2 or 1.3, got %q", cfg.TLS.MinVersion)
	}
	return nil
}

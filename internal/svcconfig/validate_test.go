package svcconfig

import "testing"

func TestValidateRequired(t *testing.T) {
	err := validateValue("test", "", "required")
	if err == nil {
		t.Error("expected error for empty required field")
	}
	err = validateValue("test", "  ", "required")
	if err == nil {
		t.Error("expected error for whitespace-only required field")
	}
	err = validateValue("test", "hello", "required")
	if err != nil {
		t.Errorf("unexpected error for valid required field: %v", err)
	}
}

func TestValidateNumeric(t *testing.T) {
	tests := []struct {
		value string
		ok    bool
	}{
		{"123", true},
		{"0", true},
		{"", true}, // empty skipped
		{"12.3", false},
		{"abc", false},
		{"-1", false},
	}
	for _, tt := range tests {
		err := validateValue("test", tt.value, "numeric")
		if tt.ok && err != nil {
			t.Errorf("numeric(%q): unexpected error %v", tt.value, err)
		}
		if !tt.ok && err == nil {
			t.Errorf("numeric(%q): expected error", tt.value)
		}
	}
}

func TestValidateDecimal(t *testing.T) {
	tests := []struct {
		value string
		ok    bool
	}{
		{"123", true},
		{"12.34", true},
		{"0.5", true},
		{"", true},
		{"abc", false},
		{".5", false},
	}
	for _, tt := range tests {
		err := validateValue("test", tt.value, "decimal")
		if tt.ok && err != nil {
			t.Errorf("decimal(%q): unexpected error %v", tt.value, err)
		}
		if !tt.ok && err == nil {
			t.Errorf("decimal(%q): expected error", tt.value)
		}
	}
}

func TestValidateRange(t *testing.T) {
	tests := []struct {
		value string
		rules string
		ok    bool
	}{
		{"5", "range:1:10", true},
		{"1", "range:1:10", true},
		{"10", "range:1:10", true},
		{"0", "range:1:10", false},
		{"11", "range:1:10", false},
		{"abc", "range:1:10", false},
		{"", "range:1:10", true}, // empty skipped
	}
	for _, tt := range tests {
		err := validateValue("test", tt.value, tt.rules)
		if tt.ok && err != nil {
			t.Errorf("range(%q, %q): unexpected error %v", tt.value, tt.rules, err)
		}
		if !tt.ok && err == nil {
			t.Errorf("range(%q, %q): expected error", tt.value, tt.rules)
		}
	}
}

func TestValidateCallsign(t *testing.T) {
	tests := []struct {
		value string
		ok    bool
	}{
		{"M0ABC", true},
		{"W1AW", true},
		{"VK2RG", true},
		{"2E0ABC", true},
		{"", true},
		{"toolong123", false},
	}
	for _, tt := range tests {
		err := validateValue("test", tt.value, "callsign")
		if tt.ok && err != nil {
			t.Errorf("callsign(%q): unexpected error %v", tt.value, err)
		}
		if !tt.ok && err == nil {
			t.Errorf("callsign(%q): expected error", tt.value)
		}
	}
}

func TestValidateIP(t *testing.T) {
	tests := []struct {
		value string
		ok    bool
	}{
		{"127.0.0.1", true},
		{"192.168.1.100", true},
		{"255.255.255.255", true},
		{"", true},
		{"256.0.0.1", false},
		{"1.2.3", false},
		{"abc", false},
	}
	for _, tt := range tests {
		err := validateValue("test", tt.value, "ip")
		if tt.ok && err != nil {
			t.Errorf("ip(%q): unexpected error %v", tt.value, err)
		}
		if !tt.ok && err == nil {
			t.Errorf("ip(%q): expected error", tt.value)
		}
	}
}

func TestValidatePort(t *testing.T) {
	tests := []struct {
		value string
		ok    bool
	}{
		{"1", true},
		{"80", true},
		{"65535", true},
		{"", true},
		{"0", false},
		{"65536", false},
		{"abc", false},
	}
	for _, tt := range tests {
		err := validateValue("test", tt.value, "port")
		if tt.ok && err != nil {
			t.Errorf("port(%q): unexpected error %v", tt.value, err)
		}
		if !tt.ok && err == nil {
			t.Errorf("port(%q): expected error", tt.value)
		}
	}
}

func TestValidateHostname(t *testing.T) {
	tests := []struct {
		value string
		ok    bool
	}{
		{"localhost", true},
		{"example.com", true},
		{"sub.domain.co.uk", true},
		{"my-host", true},
		{"", true},
		{"-invalid", false},
		{".invalid", false},
	}
	for _, tt := range tests {
		err := validateValue("test", tt.value, "hostname")
		if tt.ok && err != nil {
			t.Errorf("hostname(%q): unexpected error %v", tt.value, err)
		}
		if !tt.ok && err == nil {
			t.Errorf("hostname(%q): expected error", tt.value)
		}
	}
}

func TestValidateMinLen(t *testing.T) {
	err := validateValue("test", "ab", "minlen:3")
	if err == nil {
		t.Error("expected error for string shorter than minlen")
	}
	err = validateValue("test", "abc", "minlen:3")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestValidateMaxLen(t *testing.T) {
	err := validateValue("test", "abcdef", "maxlen:5")
	if err == nil {
		t.Error("expected error for string longer than maxlen")
	}
	err = validateValue("test", "abcde", "maxlen:5")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestValidateMultipleRules(t *testing.T) {
	err := validateValue("test", "", "required,callsign")
	if err == nil {
		t.Error("expected error for empty required+callsign")
	}
	if err.Message != "required" {
		t.Errorf("expected 'required' error, got %q", err.Message)
	}

	err = validateValue("test", "INVALID!!!", "required,callsign")
	if err == nil {
		t.Error("expected error for invalid callsign")
	}
	if err.Message != "must be a valid callsign" {
		t.Errorf("expected callsign error, got %q", err.Message)
	}
}

func TestValidateSettingsIntegration(t *testing.T) {
	schema, ok := LookupSchema("mmdvmhost")
	if !ok {
		t.Fatal("mmdvmhost schema not found")
	}

	// Valid values (only service-specific fields remain in mmdvmhost schema)
	values := map[string]string{
		"colorCode": "1",
	}
	errs := ValidateSettings(schema, values)
	if len(errs) != 0 {
		t.Errorf("expected no errors for valid values, got %v", errs)
	}

	// Invalid colorCode (out of range)
	values["colorCode"] = "20"
	errs = ValidateSettings(schema, values)
	if len(errs) == 0 {
		t.Error("expected error for colorCode out of range")
	}
}

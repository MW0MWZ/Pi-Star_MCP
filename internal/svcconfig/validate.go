package svcconfig

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// FieldError describes a validation failure for a single field.
type FieldError struct {
	Key     string `json:"key"`
	Message string `json:"message"`
}

// ValidateSettings checks values against the schema's validation rules.
// Returns nil if all values are valid.
func ValidateSettings(schema *SettingsSchema, values map[string]string) []FieldError {
	var errs []FieldError

	for _, g := range schema.Groups {
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

	return errs
}

func validateValue(key, value, rules string) *FieldError {
	for _, part := range strings.Split(rules, ",") {
		rule, args := parseRule(part)

		switch rule {
		case "required":
			if strings.TrimSpace(value) == "" {
				return &FieldError{Key: key, Message: "required"}
			}
		case "numeric":
			if value == "" {
				continue
			}
			if !isNumeric(value) {
				return &FieldError{Key: key, Message: "must be a whole number"}
			}
		case "decimal":
			if value == "" {
				continue
			}
			if !isDecimal(value) {
				return &FieldError{Key: key, Message: "must be a number"}
			}
		case "range":
			if value == "" {
				continue
			}
			if len(args) != 2 {
				continue
			}
			n, err := strconv.ParseFloat(value, 64)
			if err != nil {
				return &FieldError{Key: key, Message: "must be a number"}
			}
			lo, _ := strconv.ParseFloat(args[0], 64)
			hi, _ := strconv.ParseFloat(args[1], 64)
			if n < lo || n > hi {
				return &FieldError{Key: key, Message: fmt.Sprintf("must be between %s and %s", args[0], args[1])}
			}
		case "callsign":
			if value == "" {
				continue
			}
			if !isCallsign(value) {
				return &FieldError{Key: key, Message: "must be a valid callsign"}
			}
		case "ip":
			if value == "" {
				continue
			}
			if !isIP(value) {
				return &FieldError{Key: key, Message: "must be a valid IP address"}
			}
		case "port":
			if value == "" {
				continue
			}
			if !isPort(value) {
				return &FieldError{Key: key, Message: "must be a valid port (1-65535)"}
			}
		case "hostname":
			if value == "" {
				continue
			}
			if !isHostname(value) {
				return &FieldError{Key: key, Message: "must be a valid hostname"}
			}
		case "minlen":
			if value == "" {
				continue
			}
			if len(args) == 1 {
				n, _ := strconv.Atoi(args[0])
				if len(value) < n {
					return &FieldError{Key: key, Message: fmt.Sprintf("must be at least %d characters", n)}
				}
			}
		case "maxlen":
			if value == "" {
				continue
			}
			if len(args) == 1 {
				n, _ := strconv.Atoi(args[0])
				if len(value) > n {
					return &FieldError{Key: key, Message: fmt.Sprintf("must be at most %d characters", n)}
				}
			}
		}
	}

	return nil
}

func parseRule(s string) (string, []string) {
	parts := strings.Split(strings.TrimSpace(s), ":")
	if len(parts) == 1 {
		return parts[0], nil
	}
	return parts[0], parts[1:]
}

var (
	numericRe  = regexp.MustCompile(`^\d+$`)
	decimalRe  = regexp.MustCompile(`^-?\d+(\.\d+)?$`)
	callsignRe = regexp.MustCompile(`(?i)^[A-Z0-9]{1,3}[0-9][A-Z0-9]{0,4}$`)
	ipRe       = regexp.MustCompile(`^(\d{1,3})\.(\d{1,3})\.(\d{1,3})\.(\d{1,3})$`)
	hostnameRe = regexp.MustCompile(`^[a-zA-Z0-9]([a-zA-Z0-9-]*[a-zA-Z0-9])?(\.[a-zA-Z0-9]([a-zA-Z0-9-]*[a-zA-Z0-9])?)*$`)
)

func isNumeric(s string) bool  { return numericRe.MatchString(s) }
func isDecimal(s string) bool  { return decimalRe.MatchString(s) }
func isCallsign(s string) bool { return callsignRe.MatchString(s) }

func isIP(s string) bool {
	m := ipRe.FindStringSubmatch(s)
	if m == nil {
		return false
	}
	for i := 1; i <= 4; i++ {
		n, _ := strconv.Atoi(m[i])
		if n > 255 {
			return false
		}
	}
	return true
}

func isPort(s string) bool {
	n, err := strconv.Atoi(s)
	return err == nil && n >= 1 && n <= 65535
}

func isHostname(s string) bool { return hostnameRe.MatchString(s) }

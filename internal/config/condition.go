package config

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/rztaylor/GoDotFiles/internal/platform"
)

// ConditionEvaluator is an interface for mockability, though simple function might suffice.
// Let's stick to function for now.

// EvaluateCondition checks if a condition string matches the current platform.
// Format: field op value (e.g. "os == 'macos'")
// Supported fields: os, distro, hostname, arch
// Supported ops: ==, !=, =~ (regex)
func EvaluateCondition(condition string, p *platform.Platform) (bool, error) {
	// Simple parsing: split by spaces
	parts := strings.Fields(condition)
	if len(parts) < 3 {
		return false, fmt.Errorf("invalid condition format: %s", condition)
	}

	field := parts[0]
	op := parts[1]
	// Join rest as value (in case of spaces in value, though quotes handle it)
	valueRaw := strings.Join(parts[2:], " ")
	// Handle quoting
	value := strings.Trim(valueRaw, "'\"")

	var actual string
	switch field {
	case "os":
		actual = p.OS
	case "distro":
		actual = p.Distro
	case "hostname":
		actual = p.Hostname
	case "arch":
		actual = p.Arch
	default:
		// Unknown field
		return false, fmt.Errorf("unknown field: %s", field)
	}

	switch op {
	case "==":
		return actual == value, nil
	case "!=":
		return actual != value, nil
	case "=~":
		matched, err := regexp.MatchString(value, actual)
		return matched, err
	default:
		return false, fmt.Errorf("unknown operator: %s", op)
	}
}

// CheckConditions evaluates a list of ProfileConditions against the platform.
func CheckConditions(conditions []ProfileCondition, p *platform.Platform) (includes, includeApps, excludeApps []string, err error) {
	for _, c := range conditions {
		match, evalErr := EvaluateCondition(c.If, p)
		if evalErr != nil {
			return nil, nil, nil, evalErr
		}
		if match {
			includes = append(includes, c.Includes...)
			includeApps = append(includeApps, c.IncludeApps...)
			excludeApps = append(excludeApps, c.ExcludeApps...)
		}
	}
	return
}

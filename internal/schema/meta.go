package schema

import (
	"fmt"
	"strings"
)

// TypeMeta handles the versioning header for all GDF YAML files.
// It is embedded in other structs to enforce the presence of the "kind" field.
type TypeMeta struct {
	// Kind matches the format "<Type>/<Version>" (e.g., "App/v1").
	// This field is required for all GDF files.
	Kind string `yaml:"kind"`
}

// ParseKind parses the Kind field into its component type and version.
// It returns an error if the format is invalid.
//
// Example:
//
//	"App/v1" -> "App", "v1", nil
//	"Profile/v2" -> "Profile", "v2", nil
//	"Invalid" -> "", "", error
func (t TypeMeta) ParseKind() (typeName, version string, err error) {
	parts := strings.Split(t.Kind, "/")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid kind format: expected <Type>/<Version>, got %q", t.Kind)
	}

	if parts[0] == "" {
		return "", "", fmt.Errorf("invalid kind format: type cannot be empty")
	}

	if !strings.HasPrefix(parts[1], "v") {
		return "", "", fmt.Errorf("invalid version format: version must start with 'v', got %q", parts[1])
	}

	return parts[0], parts[1], nil
}

// ValidateKind checks if the Kind matches the expected type and a supported version.
// Currently only "v1" is supported for all types.
func (t TypeMeta) ValidateKind(expectedType string) error {
	typeName, version, err := t.ParseKind()
	if err != nil {
		return err
	}

	if typeName != expectedType {
		return fmt.Errorf("kind mismatch: expected type %q, got %q", expectedType, typeName)
	}

	// For now, we only support v1. In the future, this could be a switch or map.
	if version != "v1" {
		return fmt.Errorf("unsupported version %q for type %q (supported: v1)", version, typeName)
	}

	return nil
}

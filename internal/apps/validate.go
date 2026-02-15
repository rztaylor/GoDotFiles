package apps

import (
	"errors"
	"fmt"
	"regexp"
)

// ValidationError represents a validation failure.
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// Validate checks that a bundle is valid.
// Returns an error if validation fails.
func (b *Bundle) Validate() error {
	var errs []error

	// Name is required
	if b.Name == "" {
		errs = append(errs, &ValidationError{
			Field:   "name",
			Message: "is required",
		})
	} else if !isValidName(b.Name) {
		errs = append(errs, &ValidationError{
			Field:   "name",
			Message: "must be lowercase alphanumeric with hyphens",
		})
	}

	// Validate dotfiles
	for i, df := range b.Dotfiles {
		if df.Source == "" {
			errs = append(errs, &ValidationError{
				Field:   fmt.Sprintf("dotfiles[%d].source", i),
				Message: "is required",
			})
		}
		if df.Target == "" && df.TargetMap == nil {
			errs = append(errs, &ValidationError{
				Field:   fmt.Sprintf("dotfiles[%d].target", i),
				Message: "is required",
			})
		}
	}

	// Validate plugins
	for i, p := range b.Plugins {
		if p.Name == "" {
			errs = append(errs, &ValidationError{
				Field:   fmt.Sprintf("plugins[%d].name", i),
				Message: "is required",
			})
		}
		if p.Install == "" {
			errs = append(errs, &ValidationError{
				Field:   fmt.Sprintf("plugins[%d].install", i),
				Message: "is required",
			})
		}
	}

	// Validate custom install
	if b.Package != nil && b.Package.Custom != nil {
		if b.Package.Custom.Script == "" {
			errs = append(errs, &ValidationError{
				Field:   "package.custom.script",
				Message: "is required when using custom install",
			})
		}
	}

	// Validate shell init snippets
	if b.Shell != nil && len(b.Shell.Init) > 0 {
		seenInitNames := make(map[string]struct{}, len(b.Shell.Init))
		for i, snippet := range b.Shell.Init {
			if snippet.Name == "" {
				errs = append(errs, &ValidationError{
					Field:   fmt.Sprintf("shell.init[%d].name", i),
					Message: "is required",
				})
			} else {
				if _, exists := seenInitNames[snippet.Name]; exists {
					errs = append(errs, &ValidationError{
						Field:   fmt.Sprintf("shell.init[%d].name", i),
						Message: "must be unique within shell.init",
					})
				}
				seenInitNames[snippet.Name] = struct{}{}
			}

			if snippet.Common == "" && snippet.Bash == "" && snippet.Zsh == "" {
				errs = append(errs, &ValidationError{
					Field:   fmt.Sprintf("shell.init[%d]", i),
					Message: "must define at least one of common, bash, or zsh",
				})
			}
		}
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
	}
	return nil
}

// isValidName checks if a bundle name is valid.
// Names must be lowercase alphanumeric with hyphens.
var nameRegex = regexp.MustCompile(`^[a-z][a-z0-9-]*$`)

func isValidName(name string) bool {
	return nameRegex.MatchString(name)
}

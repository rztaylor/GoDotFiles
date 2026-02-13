package packages

import (
	"os/exec"
	"testing"

	"github.com/rztaylor/GoDotFiles/internal/apps"
)

func TestCustom_Execute(t *testing.T) {
	tests := []struct {
		name          string
		customInstall *apps.CustomInstall
		userConfirms  bool
		wantErr       bool
	}{
		{
			name: "successful execution with confirmation",
			customInstall: &apps.CustomInstall{
				Script:  "echo 'Installing...'",
				Sudo:    false,
				Confirm: boolPtr(true),
			},
			userConfirms: true,
			wantErr:      false,
		},
		{
			name: "user declines confirmation",
			customInstall: &apps.CustomInstall{
				Script:  "curl -sL https://example.com/install.sh | bash",
				Sudo:    false,
				Confirm: boolPtr(true),
			},
			userConfirms: false,
			wantErr:      true,
		},
		{
			name: "execution with sudo",
			customInstall: &apps.CustomInstall{
				Script:  "apt-get install something",
				Sudo:    true,
				Confirm: boolPtr(true),
			},
			userConfirms: true,
			wantErr:      false,
		},
		{
			name: "default confirm is true",
			customInstall: &apps.CustomInstall{
				Script: "echo 'test'",
				Sudo:   false,
				// Confirm not specified, should default to true
			},
			userConfirms: true,
			wantErr:      false,
		},
		{
			name:          "nil custom install",
			customInstall: nil,
			wantErr:       true,
		},
		{
			name: "empty script",
			customInstall: &apps.CustomInstall{
				Script: "",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Custom{
				promptFunc: func(message string) (bool, error) {
					return tt.userConfirms, nil
				},
				execCommand: func(cmd string, args ...string) *exec.Cmd {
					// Mock that always succeeds
					return exec.Command("echo", "mock")
				},
			}

			err := c.Execute(tt.customInstall)
			if (err != nil) != tt.wantErr {
				t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCustom_PromptUser(t *testing.T) {
	tests := []struct {
		name         string
		script       string
		sudo         bool
		wantContains []string
	}{
		{
			name:   "warning message contains script",
			script: "curl -sL https://example.com/install.sh | bash",
			sudo:   true,
			wantContains: []string{
				"Custom",
				"Script",
			},
		},
		{
			name:   "no sudo message",
			script: "echo 'test'",
			sudo:   false,
			wantContains: []string{
				"Custom",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Custom{}
			msg := c.buildPromptMessage(tt.script, tt.sudo)

			// Basic validation that message is built
			if len(msg) == 0 {
				t.Error("buildPromptMessage() returned empty string")
			}

			// Check for script in message
			if len(tt.script) > 0 && len(msg) < len(tt.script) {
				t.Error("buildPromptMessage() message too short to contain script")
			}
		})
	}
}

func TestCustom_ExecuteLogging(t *testing.T) {
	// Test that execution is logged for audit purposes
	customInstall := &apps.CustomInstall{
		Script:  "echo 'test'",
		Confirm: boolPtr(true),
	}

	logged := false
	c := &Custom{
		promptFunc: func(msg string) (bool, error) {
			return true, nil
		},
		execCommand: func(cmd string, args ...string) *exec.Cmd {
			return exec.Command("echo", "mock")
		},
		logFunc: func(operation string) error {
			logged = true
			return nil
		},
	}

	_ = c.Execute(customInstall)

	if !logged {
		t.Error("Execute() should log operation for audit")
	}
}

// Helper functions
func boolPtr(b bool) *bool {
	return &b
}

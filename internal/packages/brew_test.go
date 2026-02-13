package packages

import (
	"testing"
)

func TestBrew_Name(t *testing.T) {
	b := NewBrew()
	if got := b.Name(); got != "brew" {
		t.Errorf("Name() = %q, want %q", got, "brew")
	}
}

func TestBrew_Install_Validation(t *testing.T) {
	tests := []struct {
		name    string
		pkg     string
		wantErr bool
	}{
		{
			name:    "empty package name",
			pkg:     "",
			wantErr: true,
		},
		{
			name:    "valid package name",
			pkg:     "tree",
			wantErr: false, // Would succeed with mock
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := NewBrew()

			// For valid packages, we can't test without brew installed
			// Just test validation logic
			if tt.pkg == "" {
				err := b.Install(tt.pkg)
				if (err != nil) != tt.wantErr {
					t.Errorf("Install() error = %v, wantErr %v", err, tt.wantErr)
				}
			}
		})
	}
}

func TestBrew_IsInstalled_Validation(t *testing.T) {
	b := NewBrew()

	// Test empty package name validation
	_, err := b.IsInstalled("")
	if err == nil {
		t.Error("IsInstalled(\"\") should return error for empty package name")
	}
}

func TestBrew_IsAvailable(t *testing.T) {
	b := NewBrew()
	// Just test that it doesn't crash
	_ = b.IsAvailable()
}

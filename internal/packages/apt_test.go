package packages

import (
	"testing"

	"github.com/rztaylor/GoDotFiles/internal/apps"
)

func TestApt_Name(t *testing.T) {
	a := NewApt()
	if got := a.Name(); got != "apt" {
		t.Errorf("Name() = %q, want %q", got, "apt")
	}
}

func TestApt_Install_Validation(t *testing.T) {
	a := NewApt()

	// Test empty package name validation
	err := a.Install("")
	if err == nil {
		t.Error("Install(\"\") should return error for empty package name")
	}
}

func TestApt_InstallWithRepo_Validation(t *testing.T) {
	tests := []struct {
		name    string
		aptPkg  *apps.AptPackage
		wantErr bool
	}{
		{
			name:    "nil package",
			aptPkg:  nil,
			wantErr: true,
		},
		{
			name: "empty package name",
			aptPkg: &apps.AptPackage{
				Name: "",
			},
			wantErr: true,
		},
		{
			name: "valid simple package",
			aptPkg: &apps.AptPackage{
				Name: "tree",
			},
			wantErr: false, // Would succeed with proper mock
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := NewApt()

			// Only test validation errors, not actual installation
			if tt.wantErr {
				err := a.InstallWithRepo(tt.aptPkg)
				if err == nil {
					t.Error("InstallWithRepo() expected error, got nil")
				}
			}
		})
	}
}

func TestApt_Uninstall_Validation(t *testing.T) {
	a := NewApt()

	err := a.Uninstall("")
	if err == nil {
		t.Error("Uninstall(\"\") should return error for empty package name")
	}
}

func TestApt_IsInstalled_Validation(t *testing.T) {
	a := NewApt()

	// Test empty package name validation
	_, err := a.IsInstalled("")
	if err == nil {
		t.Error("IsInstalled(\"\") should return error for empty package name")
	}
}

func TestApt_IsAvailable(t *testing.T) {
	a := NewApt()
	// Just test that it doesn't crash
	_ = a.IsAvailable()
}

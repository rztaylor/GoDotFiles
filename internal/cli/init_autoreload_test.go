package cli

import (
	"path/filepath"
	"testing"

	"github.com/rztaylor/GoDotFiles/internal/config"
)

func TestSetAutoReloadEnabled(t *testing.T) {
	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "config.yaml")

	cfg := &config.Config{}
	if err := cfg.Save(cfgPath); err != nil {
		t.Fatalf("saving config: %v", err)
	}

	if err := setAutoReloadEnabled(cfgPath, true); err != nil {
		t.Fatalf("setAutoReloadEnabled(true): %v", err)
	}

	loaded, err := config.LoadConfig(cfgPath)
	if err != nil {
		t.Fatalf("loading config: %v", err)
	}
	if loaded.ShellIntegration == nil || !loaded.ShellIntegration.AutoReloadEnabledDefault() {
		t.Fatalf("expected shell_integration.auto_reload.enabled=true, got %#v", loaded.ShellIntegration)
	}
}

package cli

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestShouldSkipInitCheck(t *testing.T) {
	tests := []struct {
		name string
		cmd  *cobra.Command
		want bool
	}{
		{name: "init", cmd: initCmd, want: true},
		{name: "version", cmd: versionCmd, want: true},
		{name: "update", cmd: updateCmd, want: true},
		{name: "shell", cmd: shellCmd, want: true},
		{name: "shell completion subcommand", cmd: shellCompletionCmd, want: true},
		{name: "profile list", cmd: profileListCmd, want: false},
		{name: "status", cmd: statusCmd, want: false},
		{name: "health doctor", cmd: healthDoctorCmd, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := shouldSkipInitCheck(tt.cmd)
			if got != tt.want {
				t.Fatalf("shouldSkipInitCheck(%s) = %v, want %v", tt.cmd.Name(), got, tt.want)
			}
		})
	}
}

func TestPersistentPreRunE_InitRequirement(t *testing.T) {
	tests := []struct {
		name      string
		cmd       *cobra.Command
		initRepo  bool
		wantError bool
	}{
		{name: "profile list requires init", cmd: profileListCmd, initRepo: false, wantError: true},
		{name: "status requires init", cmd: statusCmd, initRepo: false, wantError: true},
		{name: "health doctor requires init", cmd: healthDoctorCmd, initRepo: false, wantError: true},
		{name: "init exempted", cmd: initCmd, initRepo: false, wantError: false},
		{name: "version exempted", cmd: versionCmd, initRepo: false, wantError: false},
		{name: "shell completion exempted", cmd: shellCompletionCmd, initRepo: false, wantError: false},
		{name: "profile list succeeds when initialized", cmd: profileListCmd, initRepo: true, wantError: false},
		{name: "status succeeds when initialized", cmd: statusCmd, initRepo: true, wantError: false},
		{name: "health doctor succeeds when initialized", cmd: healthDoctorCmd, initRepo: true, wantError: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			home := t.TempDir()
			t.Setenv("HOME", home)

			if tt.initRepo {
				gdfDir := filepath.Join(home, ".gdf")
				configureGitUserGlobal(t, home)
				if err := createNewRepo(gdfDir); err != nil {
					t.Fatalf("createNewRepo() error = %v", err)
				}
			}

			err := rootCmd.PersistentPreRunE(tt.cmd, []string{})
			if tt.wantError {
				if err == nil {
					t.Fatalf("PersistentPreRunE(%s) expected error, got nil", tt.cmd.Name())
				}
				return
			}
			if err != nil {
				t.Fatalf("PersistentPreRunE(%s) unexpected error = %v", tt.cmd.Name(), err)
			}
		})
	}
}

func TestPersistentPreRunE_UninitializedMessage(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	err := rootCmd.PersistentPreRunE(statusCmd, []string{})
	if err == nil {
		t.Fatal("PersistentPreRunE(status) expected error, got nil")
	}
	if got := err.Error(); got == "" || !strings.Contains(got, "Please run 'gdf init' first.") {
		t.Fatalf("unexpected init error message: %q", got)
	}
}

func TestCommandHierarchy_GroupsAppAndRecover(t *testing.T) {
	for _, name := range []string{"init", "save", "push", "pull", "sync"} {
		if findSubcommand(rootCmd, name) == nil {
			t.Fatalf("expected top-level '%s' command", name)
		}
	}

	app := findSubcommand(rootCmd, "app")
	if app == nil {
		t.Fatal("expected top-level 'app' command")
	}

	recoverCmd := findSubcommand(rootCmd, "recover")
	if recoverCmd == nil {
		t.Fatal("expected top-level 'recover' command")
	}

	for _, name := range []string{"add", "remove", "list", "install", "track", "move", "library"} {
		if findSubcommand(app, name) == nil {
			t.Fatalf("expected 'gdf app %s' command", name)
		}
	}

	for _, name := range []string{"rollback", "restore"} {
		if findSubcommand(recoverCmd, name) == nil {
			t.Fatalf("expected 'gdf recover %s' command", name)
		}
	}
}

func TestCommandHierarchy_RemovesOldTopLevelCommands(t *testing.T) {
	for _, name := range []string{"add", "remove", "list", "install", "track", "move", "library", "rollback", "restore"} {
		if findSubcommand(rootCmd, name) != nil {
			t.Fatalf("did not expect top-level '%s' command after regrouping", name)
		}
	}
}

func findSubcommand(cmd *cobra.Command, name string) *cobra.Command {
	for _, c := range cmd.Commands() {
		if c.Name() == name {
			return c
		}
	}
	return nil
}

package cli

import (
	"fmt"

	"github.com/rztaylor/GoDotFiles/internal/platform"
	"github.com/rztaylor/GoDotFiles/internal/shell"
	"github.com/spf13/cobra"
)

var shellCmd = &cobra.Command{
	Use:   "shell",
	Short: "Shell integration commands",
}

var shellReloadCmd = &cobra.Command{
	Use:   "reload",
	Short: "Reload shell integration in current session",
	Long: `Provides instructions for reloading GDF's shell integration.

Since a subprocess cannot modify the parent shell environment, this command
prints the source command for you to copy and paste or eval.`,
	RunE: runShellReload,
}

func init() {
	rootCmd.AddCommand(shellCmd)
	shellCmd.AddCommand(shellReloadCmd)
}

func runShellReload(cmd *cobra.Command, args []string) error {
	detectedShell := platform.DetectShell()
	shellType := shell.ParseShellType(detectedShell)

	initPath := "~/.gdf/generated/init.sh"

	fmt.Println("To reload shell integration, run:")
	fmt.Println()

	if shellType == shell.Bash || shellType == shell.Zsh {
		fmt.Printf("  source %s\n", initPath)
	} else {
		fmt.Printf("  source %s  # (for bash/zsh)\n", initPath)
	}

	fmt.Println()
	fmt.Println("Or restart your shell.")

	return nil
}

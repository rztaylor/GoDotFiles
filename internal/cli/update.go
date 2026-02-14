package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/rztaylor/GoDotFiles/internal/config"
	"github.com/rztaylor/GoDotFiles/internal/platform"
	"github.com/rztaylor/GoDotFiles/internal/state"
	"github.com/rztaylor/GoDotFiles/internal/updater"
)

var neverCheck bool

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Check for updates and self-update",
	Long:  `Check for updates to GDF and install them if available.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfgPath := platform.ConfigFile()

		cfg, err := config.LoadConfig(cfgPath)
		if err != nil {
			return err
		}

		// Handle --never flag
		if neverCheck {
			if cfg.Updates == nil {
				cfg.Updates = &config.UpdatesConfig{}
			}
			cfg.Updates.Disabled = true
			if err := os.MkdirAll(platform.ConfigDir(), 0755); err != nil {
				return fmt.Errorf("creating config directory: %w", err)
			}
			if err := cfg.Save(cfgPath); err != nil {
				return fmt.Errorf("saving config: %w", err)
			}
			fmt.Println("Auto-update checks disabled.")
			return nil
		}

		// Check for updates (ignoring disabled/snoozed status)
		fmt.Println("Checking for updates...")

		info, err := updater.CheckForUpdate(cfg, &state.State{UpdateCheck: state.UpdateCheck{}}, true)
		if err != nil {
			return err
		}

		if info == nil {
			fmt.Println("No update available.")
			return nil
		}

		fmt.Printf("New version available: %s\n", info.Version)
		fmt.Println("Updating...")
		return updater.Update()
	},
}

func init() {
	rootCmd.AddCommand(updateCmd)
	updateCmd.Flags().BoolVar(&neverCheck, "never", false, "Disable auto-update checks permanently")
}

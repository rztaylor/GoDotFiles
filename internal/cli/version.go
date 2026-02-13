package cli

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
)

var (
	// Version is the current version of the application
	Version = "0.0.0-dev"
	// Commit is the git commit hash
	Commit = "none"
	// Date is the build date
	Date = "unknown"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of gdf",
	Long:  `All software has versions. This is gdf's.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("gdf version %s\n", Version)
		fmt.Printf("commit: %s\n", Commit)
		fmt.Printf("built at: %s\n", Date)
		fmt.Printf("go version: %s\n", runtime.Version())
		fmt.Printf("os/arch: %s/%s\n", runtime.GOOS, runtime.GOARCH)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}

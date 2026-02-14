package cli

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
)

var (
	// Version is the current version of the application
	Version = "0.6.0-dev"
	// Commit is the git commit hash
	Commit = "none"
	// Date is the build date
	Date = "unknown"
	// verbose controls whether to show detailed version info
	verbose bool
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of gdf",
	Long:  `All software has versions. This is gdf's.`,
	Run: func(cmd *cobra.Command, args []string) {
		if verbose {
			fmt.Printf("gdf version: %s\n", Version)
			fmt.Printf("commit: %s\n", Commit)
			fmt.Printf("built at: %s\n", Date)
			fmt.Printf("go version: %s\n", runtime.Version())
			fmt.Printf("os/arch: %s/%s\n", runtime.GOOS, runtime.GOARCH)
			return
		}

		v := Version
		// If it's a dev version and we have a date, show it to help differentiate builds
		if (Version == "0.6.0-dev" || Commit == "none") && Date != "unknown" {
			v = fmt.Sprintf("%s (%s)", v, Date)
		}
		fmt.Printf("gdf version %s\n", v)
	},
}

func init() {
	versionCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "show detailed version information")
	rootCmd.AddCommand(versionCmd)
}

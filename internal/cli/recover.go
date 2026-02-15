package cli

import "github.com/spf13/cobra"

var recoverCmd = &cobra.Command{
	Use:   "recover",
	Short: "Restore or rollback managed state",
	Long:  `Recovery tools for undoing apply operations or restoring managed files.`,
}

func init() {
	rootCmd.AddCommand(recoverCmd)
}

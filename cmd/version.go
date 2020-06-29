package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "putt version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("putt v1.1, support for dpl-v2.2")
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}

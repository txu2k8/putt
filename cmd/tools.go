package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// toolsCmd represents the tools command
var toolsCmd = &cobra.Command{
	Use:   "tools",
	Short: "DevOps tools",
	Long:  `DevOps tools include deploy/maint/check ...`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("tools called")
	},
}

func init() {
	rootCmd.AddCommand(toolsCmd)
}

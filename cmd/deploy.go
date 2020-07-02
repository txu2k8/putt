package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// deployCmd represents the deploy command
var deployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Deploy test env",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("deploy called, use -h or --help for help")
	},
}

func init() {
	rootCmd.AddCommand(deployCmd)
}

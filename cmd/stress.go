package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// stressCmd represents the stress command
var stressCmd = &cobra.Command{
	Use:   "stress",
	Short: "Stress test",
	Long:  `Stress test cases S3/ES/HA ...`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("stress called, use -h or --help for help")
	},
}

func init() {
	rootCmd.AddCommand(stressCmd)
}

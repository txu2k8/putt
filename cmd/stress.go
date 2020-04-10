package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// stressCmd represents the stress command
var stressCmd = &cobra.Command{
	Use:   "stress",
	Short: "Vizion Stress test",
	Long:  `Vizion Stress test include S3/ES/HA ...`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("stress called")
	},
}

func init() {
	rootCmd.AddCommand(stressCmd)
}

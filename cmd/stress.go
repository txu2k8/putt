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

// s3Cmd represents the s3 command
var s3Cmd = &cobra.Command{
	Use:   "s3",
	Short: "S3 IO Stress",
	Long:  `S3 upload/download files. --help for detail args.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("s3 called")
	},
}

// esCmd represents the es command
var esCmd = &cobra.Command{
	Use:   "es",
	Short: "ES Index/Search",
	Long:  `ES Index/Search(multi progress). --help for detail args.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("es called")
	},
}

func init() {
	rootCmd.AddCommand(stressCmd)
	stressCmd.AddCommand(s3Cmd)
	stressCmd.AddCommand(esCmd)

}

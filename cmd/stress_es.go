package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// esCmd represents the es command
var esCmd = &cobra.Command{
	Use:   "es",
	Short: "Vizion ES Index/Search",
	Long:  `Vizion ES Index/Search(multi progress). --help for detail args.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("es called")
	},
}

func init() {
	stressCmd.AddCommand(esCmd)

	esCmd.PersistentFlags().String("foo", "", "A help for foo")
	esCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

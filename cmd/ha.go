package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// haCmd represents the ha command
var haCmd = &cobra.Command{
	Use:   "ha",
	Short: "Vizion HA",
	Long:  `Vizion high availability test. --help for detail args.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("es called")
	},
}

// restartNodeCmd represents the restart_node command
var restartNodeCmd = &cobra.Command{
	Use:   "restart_node",
	Short: "Vizion HA: Restart Nodes",
	Long:  `Vizion high availability test. --help for detail args.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("ha restart_node called")
	},
}

func init() {
	haCmd.AddCommand(haCmd)
}

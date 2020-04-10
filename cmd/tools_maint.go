package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// maintCmd represents the maint command
var maintCmd = &cobra.Command{
	Use:   "maint",
	Short: "Maintaince mode tools",
	Long: `Maintaince mode operations, subCommand:
	stop/start/restart: stop/start/restart specified services
	cleanup: cleanup specified options: log/journal/etcd ...
	upgrade: upgrade env to specified image
	rolling_update: rolling update env services`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("maint called")
	},
}

// stopCmd represents the stop command
var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Maintaince mode tools: stop",
	Long:  `stop specified services`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("maint stop called")
	},
}

// startCmd represents the start command
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Maintaince mode tools: start",
	Long:  `start specified services`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("maint start called")
	},
}

// restartCmd represents the restart command
var restartCmd = &cobra.Command{
	Use:   "restart",
	Short: "Maintaince mode tools: restart",
	Long:  `restart specified services`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("maint restart called")
	},
}

// cleanupCmd represents the cleanup command
var cleanupCmd = &cobra.Command{
	Use:   "cleanup",
	Short: "Maintaince mode tools: cleanup",
	Long:  `cleanup specified items`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("maint cleanup called")
	},
}

// cleanupCmd represents the make_binary command
var makeBinaryCmd = &cobra.Command{
	Use:   "make_binary",
	Short: "Maintaince mode tools: make_binary",
	Long:  `make binary from git server`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("maint make_binary called")
	},
}

func init() {
	toolsCmd.AddCommand(maintCmd)
	maintCmd.AddCommand(stopCmd)
	maintCmd.AddCommand(startCmd)
	maintCmd.AddCommand(restartCmd)
	maintCmd.AddCommand(cleanupCmd)
	maintCmd.AddCommand(makeBinaryCmd)
}

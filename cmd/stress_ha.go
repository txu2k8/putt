package cmd

import (
	"fmt"
	"putt/libs/runner/stress"
	"putt/vizion/testcase"

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

// RestartNodeTestConf ...
var RestartNodeTestConf = testcase.RestartNodeTestInput{}

// restartNodeCmd represents the restart_node command
var restartNodeCmd = &cobra.Command{
	Use:   "restart_node",
	Short: "Restart Env Nodes",
	Long:  `Vizion high availability test. --help for detail args.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(caseList) == 0 {
			caseList = []string{"restart_node"}
		}
		logger.Infof("Case List(ha): %s", caseList)
		testJobs := []stress.Job{}
		var haTester testcase.HATester
		for _, tc := range caseList {
			jobs := []stress.Job{}
			switch tc {
			case "restart_node":
				haTester = testcase.NewHANode(RestartNodeTestConf)
				jobs = []stress.Job{
					{
						Fn:       haTester.Run,
						Name:     "RestartNode",
						RunTimes: runTimes,
					},
				}
			}
			testJobs = append(testJobs, jobs...)
		}
		stress.Run(testJobs)
	},
}

// restartServiceCmd represents the restart_service command
var restartServiceCmd = &cobra.Command{
	Use:   "restart_service",
	Short: "VRestart Env Services",
	Long:  `Vizion high availability test. --help for detail args.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(caseList) == 0 {
			caseList = []string{"restart_service"}
		}
		logger.Infof("Case List(ha): %s", caseList)
		testJobs := []stress.Job{}
		for _, tc := range caseList {
			jobs := []stress.Job{}
			switch tc {
			case "restart_service":
				debug := func() error {
					return testcase.Debug(baseConf)
				}
				jobs = []stress.Job{
					{
						Fn:       debug,
						Name:     "RestartService",
						RunTimes: runTimes,
					},
				}
			}
			testJobs = append(testJobs, jobs...)
		}
		stress.Run(testJobs)
	},
}

func init() {
	stressCmd.AddCommand(haCmd)
	haCmd.AddCommand(restartNodeCmd)
	restartNodeCmd.PersistentFlags().StringArrayVar(&RestartNodeTestConf.NodeIPs, "node_ip", []string{}, "To restart node IP address Array")
	restartNodeCmd.PersistentFlags().StringArrayVar(&RestartNodeTestConf.VMNames, "vm_name", []string{}, "To restart node VM name Array")
	restartNodeCmd.PersistentFlags().StringVar(&RestartNodeTestConf.Platform, "platform", "vsphere", "Test VM platfor: vsphere | aws")
	restartNodeCmd.PersistentFlags().StringArrayVar(&RestartNodeTestConf.PowerOpts, "power_opt", []string{}, "Power opts: shoutdwon|poweroff|reset|reboot")
	restartNodeCmd.PersistentFlags().IntVar(&RestartNodeTestConf.RestartNumMin, "restart_num_min", 0, "Min Restart VM number")
	restartNodeCmd.PersistentFlags().IntVar(&RestartNodeTestConf.RestartNumMax, "restart_num_max", 0, "Max Restart VM number")
	AddFlagsVsphere(restartNodeCmd)

	haCmd.AddCommand(restartServiceCmd)
}

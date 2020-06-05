package cmd

import (
	"fmt"
	"pzatest/libs/runner/stress"
	"pzatest/vizion/maintenance"

	"github.com/spf13/cobra"
)

var maintConf maintenance.MaintTestInput

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
	Use:     "stop",
	Short:   "Maintaince mode tools: stop service",
	Long:    "Stop specified services(default:All DPL+APP)",
	Example: "pzatest maint stop --master_ips 10.25.119.71 --vset_ids 1 --services es --clean all",
	Run: func(cmd *cobra.Command, args []string) {
		logger.Infof("maint stop services ...")
		var maintainer maintenance.Maintainer
		maintainer = maintenance.NewMaint(vizionBaseConf, maintConf)
		stop := func() error {
			err := maintainer.Stop()
			return err
		}
		jobs := []stress.Job{
			{
				Fn:       stop,
				Name:     "Stop-Service",
				RunTimes: 1,
			},
		}
		stress.Run(jobs)
	},
}

// startCmd represents the start command
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Maintaince mode tools: start",
	Long:  `start specified services`,
	Run: func(cmd *cobra.Command, args []string) {
		logger.Infof("maint start services ...")
		var maintainer maintenance.Maintainer
		maintainer = maintenance.NewMaint(vizionBaseConf, maintConf)
		start := func() error {
			err := maintainer.Start()
			return err
		}
		jobs := []stress.Job{
			{
				Fn:       start,
				Name:     "Start-Service",
				RunTimes: 1,
			},
		}
		stress.Run(jobs)
	},
}

// restartCmd represents the restart command
var restartCmd = &cobra.Command{
	Use:   "restart",
	Short: "Maintaince mode tools: restart",
	Long:  `restart specified services`,
	Run: func(cmd *cobra.Command, args []string) {
		logger.Infof("maint restart services ...")
		var maintainer maintenance.Maintainer
		maintainer = maintenance.NewMaint(vizionBaseConf, maintConf)
		restart := func() error {
			err := maintainer.Restart()
			return err
		}
		jobs := []stress.Job{
			{
				Fn:       restart,
				Name:     "Restart-Service",
				RunTimes: 1,
			},
		}
		stress.Run(jobs)
	},
}

// cleanupCmd represents the cleanup command
var cleanupCmd = &cobra.Command{
	Use:   "cleanup",
	Short: "Maintaince mode tools: cleanup",
	Long:  `cleanup specified items`,
	Run: func(cmd *cobra.Command, args []string) {
		logger.Infof("maint clean up ...")
		var maintainer maintenance.Maintainer
		maintainer = maintenance.NewMaint(vizionBaseConf, maintConf)
		cleanup := func() error {
			err := maintainer.Cleanup()
			return err
		}
		jobs := []stress.Job{
			{
				Fn:       cleanup,
				Name:     "Clean Up",
				RunTimes: 1,
			},
		}
		stress.Run(jobs)
	},
}

// cleanupCmd represents the make_binary command
var makeBinaryCmd = &cobra.Command{
	Use:   "make_binary",
	Short: "Maintaince mode tools: make_binary",
	Long:  `make binary from git server`,
	Run: func(cmd *cobra.Command, args []string) {
		logger.Infof("maint clean up ...")
		var maintainer maintenance.Maintainer
		maintainer = maintenance.NewMaint(vizionBaseConf, maintConf)
		cleanup := func() error {
			err := maintainer.Cleanup()
			return err
		}
		jobs := []stress.Job{
			{
				Fn:       cleanup,
				Name:     "Clean Up",
				RunTimes: 1,
			},
		}
		stress.Run(jobs)
	},
}

// applyImageCmd represents the make_binary command
var applyImageCmd = &cobra.Command{
	Use:   "apply_image",
	Short: "Maintaince mode tools: apply_image",
	Long:  `apply service container image`,
	Run: func(cmd *cobra.Command, args []string) {
		logger.Infof("maint apply service image ...")
		var maintainer maintenance.Maintainer
		maintainer = maintenance.NewMaint(vizionBaseConf, maintConf)
		applyImg := func() error {
			err := maintainer.ApplyImage()
			return err
		}
		jobs := []stress.Job{
			{
				Fn:       applyImg,
				Name:     "Apply Service Image",
				RunTimes: 1,
			},
		}
		stress.Run(jobs)
	},
}

// AddFlagsMaintService ...
func AddFlagsMaintService(cmd *cobra.Command) {
	cmd.PersistentFlags().StringArrayVar(&maintConf.SvNameArr, "services", []string{}, "Service Name List")
	cmd.PersistentFlags().StringArrayVar(&maintConf.ExculdeSvNameArr, "services_exclude", []string{}, "Service Name List which excluded")
}

// AddFlagsMaintClean ...
func AddFlagsMaintClean(cmd *cobra.Command) {
	cmd.PersistentFlags().StringArrayVar(&maintConf.CleanNameArr, "clean", []string{}, "Clean item Name List")
}

// AddFlagsMaintImage ...
func AddFlagsMaintImage(cmd *cobra.Command) {
	cmd.PersistentFlags().StringArrayVar(&maintConf.Image, "image", "", "core image")
}

func init() {
	rootCmd.AddCommand(maintCmd)
	maintCmd.AddCommand(stopCmd)
	maintCmd.AddCommand(startCmd)
	maintCmd.AddCommand(restartCmd)
	maintCmd.AddCommand(cleanupCmd)
	maintCmd.AddCommand(makeBinaryCmd)

	// stop
	AddFlagsMaintService(stopCmd)
	AddFlagsMaintClean(stopCmd)
	// start
	AddFlagsMaintService(startCmd)
	AddFlagsMaintClean(stopCmd)
	// restart
	AddFlagsMaintService(restartCmd)
	AddFlagsMaintClean(restartCmd)
	// apply image
	AddFlagsMaintImage(applyImageCmd)
	AddFlagsMaintService(applyImageCmd)
	AddFlagsMaintClean(applyImageCmd)

}

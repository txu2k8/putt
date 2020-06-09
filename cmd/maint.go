package cmd

import (
	"fmt"
	"pzatest/config"
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

// makeImageCmd represents the make_binary command
var makeImageCmd = &cobra.Command{
	Use:   "make_image",
	Short: "Maintaince mode tools: make_image",
	Long:  `make image by push tag to gitlab server`,
	Run: func(cmd *cobra.Command, args []string) {
		logger.Infof("maint make image ...")
		var maintainer maintenance.Maintainer
		maintainer = maintenance.NewMaint(vizionBaseConf, maintConf)
		makeImg := func() error {
			err := maintainer.MakeImage()
			return err
		}
		jobs := []stress.Job{
			{
				Fn:       makeImg,
				Name:     "Apply Service Image",
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

// AddFlagsMaintService Service
func AddFlagsMaintService(cmd *cobra.Command) {
	cmd.PersistentFlags().StringArrayVar(&maintConf.SvNameArr, "services", []string{}, "Service Name List (default [])")
	cmd.PersistentFlags().StringArrayVar(&maintConf.ExculdeSvNameArr, "services_exclude", []string{}, "Service Name List which excluded (default [])")
}

// AddFlagsMaintClean Clean
func AddFlagsMaintClean(cmd *cobra.Command) {
	cmd.PersistentFlags().StringArrayVar(&maintConf.CleanNameArr, "clean", []string{}, "Clean item Name List (default [])")
}

// AddFlagsMaintImage Image
func AddFlagsMaintImage(cmd *cobra.Command) {
	cmd.PersistentFlags().StringVar(&maintConf.Image, "image", "", "core image (default \"\")")
}

// AddFlagsMaintGit Git
func AddFlagsMaintGit(cmd *cobra.Command) {
	cmd.PersistentFlags().BoolVar(&maintConf.GitCfg.Pull, "pull", false, "git pull if true (default false)")
	cmd.PersistentFlags().BoolVar(&maintConf.GitCfg.Tag, "tag", false, "git tag if true (default false)")
	cmd.PersistentFlags().BoolVar(&maintConf.GitCfg.Make, "make", false, "make file if true (default false)")

	cmd.PersistentFlags().StringVar(&maintConf.GitCfg.BuildIP, "build_ip", config.DplBuildIP, "build ip)")
	cmd.PersistentFlags().StringVar(&maintConf.GitCfg.BuildPath, "build_path", config.DplBuildPath, "build path")
	cmd.PersistentFlags().StringVar(&maintConf.GitCfg.BuildNum, "build_num", "", "build number,eg: 2.1.0.100 (default \"\")")
	cmd.PersistentFlags().StringVar(&maintConf.GitCfg.BuildServerUser, "build_server_user", "", "build server user (default root)")
	cmd.PersistentFlags().StringVar(&maintConf.GitCfg.BuildServerPwd, "build_server_pwd", "", "build server pwd (default password)")
	cmd.PersistentFlags().StringVar(&maintConf.GitCfg.BuildServerKey, "build_server_key", "", "build server key, (default \"\")")
}

func init() {
	rootCmd.AddCommand(maintCmd)
	maintCmd.AddCommand(stopCmd)
	maintCmd.AddCommand(startCmd)
	maintCmd.AddCommand(restartCmd)
	maintCmd.AddCommand(cleanupCmd)
	maintCmd.AddCommand(makeBinaryCmd)
	maintCmd.AddCommand(makeImageCmd)
	maintCmd.AddCommand(applyImageCmd)

	// clean
	AddFlagsMaintClean(cleanupCmd)
	// stop
	AddFlagsMaintService(stopCmd)
	AddFlagsMaintClean(stopCmd)
	// start
	AddFlagsMaintService(startCmd)
	AddFlagsMaintClean(startCmd)
	// restart
	AddFlagsMaintService(restartCmd)
	AddFlagsMaintClean(restartCmd)
	// make image
	AddFlagsMaintGit(makeImageCmd)
	// apply image
	AddFlagsMaintImage(applyImageCmd)
	AddFlagsMaintService(applyImageCmd)
	AddFlagsMaintClean(applyImageCmd)

}

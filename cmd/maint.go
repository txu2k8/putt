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
		jobs := []stress.Job{
			{
				Fn:       maintainer.StopC,
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
		jobs := []stress.Job{
			{
				Fn:       maintainer.Start,
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
		jobs := []stress.Job{
			{
				Fn:       maintainer.Restart,
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
		jobs := []stress.Job{
			{
				Fn:       maintainer.Cleanup,
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
	Short: "Maintaince mode tools: make_binary --TODO",
	Long:  `make binary from git server`,
	Run: func(cmd *cobra.Command, args []string) {
		logger.Infof("maint Make Binary ...")
		var maintainer maintenance.Maintainer
		maintainer = maintenance.NewMaint(vizionBaseConf, maintConf)
		jobs := []stress.Job{
			{
				Fn:       maintainer.MakeBinary,
				Name:     "Make Binary",
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
		jobs := []stress.Job{
			{
				Fn:       maintainer.MakeImage,
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
		jobs := []stress.Job{
			{
				Fn:       maintainer.ApplyImage,
				Name:     "Apply Service Image",
				RunTimes: 1,
			},
		}
		stress.Run(jobs)
	},
}

// AddFlagsMaintService Service
func AddFlagsMaintService(cmd *cobra.Command) {
	var DefaultCoreServiceNameArray = []string{}
	for _, sv := range config.DefaultCoreServiceArray {
		DefaultCoreServiceNameArray = append(DefaultCoreServiceNameArray, sv.Name)
	}
	cmd.PersistentFlags().StringArrayVar(&maintConf.SvNameArr, "services", []string{}, fmt.Sprintf("Service Name List (default [])\nchoice:%s", DefaultCoreServiceNameArray))
	cmd.PersistentFlags().StringArrayVar(&maintConf.ExculdeSvNameArr, "services_exclude", []string{}, "Service Name List which excluded (default [])")
}

// AddFlagsMaintBinary Service
func AddFlagsMaintBinary(cmd *cobra.Command) {
	var DefaultCoreBinaryNameArray = []string{}
	for _, sv := range config.DefaultDplServiceArray {
		DefaultCoreBinaryNameArray = append(DefaultCoreBinaryNameArray, sv.Name)
	}
	cmd.PersistentFlags().StringArrayVar(&maintConf.BinNameArr, "binarys", []string{}, fmt.Sprintf("Binarys Name List (default [])\nchoice:%s", DefaultCoreBinaryNameArray))
	cmd.PersistentFlags().StringArrayVar(&maintConf.ExculdeBinNameArr, "binarys_exclude", []string{}, "Binarys Name List which excluded (default [])")
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
	cmd.PersistentFlags().StringVar(&maintConf.GitCfg.BuildServerIP, "build_server_ip", config.DplBuildIP, "git build server ip")
	cmd.PersistentFlags().StringVar(&maintConf.GitCfg.BuildServerUser, "build_server_user", "root", "git build server user")
	cmd.PersistentFlags().StringVar(&maintConf.GitCfg.BuildServerPwd, "build_server_pwd", "password", "git build server pwd")
	cmd.PersistentFlags().StringVar(&maintConf.GitCfg.BuildServerKey, "build_server_key", "", "git build server key (default \"\")")
	cmd.PersistentFlags().StringVar(&maintConf.GitCfg.BuildPath, "build_path", config.DplBuildPath, "git build project path")
	cmd.PersistentFlags().StringVar(&maintConf.GitCfg.BuildNum, "build_num", "", "build number,eg: 2.1.0.100 (default \"\")")
	cmd.PersistentFlags().StringVar(&maintConf.GitCfg.LocalBinPath, "local_bin_path", config.DplBuildLocalPath, "local path for store dpl binarys")
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
	// make binary
	AddFlagsMaintGit(makeBinaryCmd)
	AddFlagsMaintBinary(makeBinaryCmd)
	// make image
	AddFlagsMaintGit(makeImageCmd)
	// apply image
	AddFlagsMaintImage(applyImageCmd)
	AddFlagsMaintService(applyImageCmd)
	AddFlagsMaintClean(applyImageCmd)
}

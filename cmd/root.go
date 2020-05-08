package cmd

import (
	"fmt"
	"gtest/models"
	"os"

	"github.com/op/go-logging"
	"github.com/spf13/cobra"
)

var (
	logger   = logging.MustGetLogger("test")
	runTimes int  // runTimes
	debug    bool // debug modle
	// var cfgFile string
	caseList   []string // Case List
	vizionBase models.VizionBaseInput
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "gtest",
	Short: "The rootCmd of this test project",
	Long:  `example: vztest stress --ssh_pwd password`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Vizion Base Info: %v", vizionBase)
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	// cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().IntVar(&runTimes, "run_times", 10, "Run test case with iteration loop")
	rootCmd.PersistentFlags().BoolVar(&s3TestConf.RenameFile, "rename", true, "Rename files name each time if true")

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.vztest.yaml)")
	rootCmd.PersistentFlags().StringArrayVar(&vizionBase.MasterIPs, "master_ips", []string{}, "Master nodes IP address Array")
	rootCmd.PersistentFlags().IntSliceVar(&vizionBase.VsetIDs, "vset_ids", []int{}, "vset IDs array")
	rootCmd.PersistentFlags().IntSliceVar(&vizionBase.DPLGroupIDs, "dpl_group_ids", []int{1}, "dpl group ids array")
	rootCmd.PersistentFlags().IntSliceVar(&vizionBase.JDGroupIDs, "jd_group_ids", []int{1}, "jd group ids array")
	// rootCmd.MarkPersistentFlagRequired("master_ips")
	// rootCmd.MarkPersistentFlagRequired("vset_ids")

	rootCmd.PersistentFlags().StringVar(&vizionBase.SSHKey.UserName, "ssh_user", "root", "ssh login user")
	rootCmd.PersistentFlags().StringVar(&vizionBase.SSHKey.Password, "ssh_pwd", "password", "ssh login password")
	rootCmd.PersistentFlags().IntVar(&vizionBase.SSHKey.Port, "ssh_port", 22, "ssh login port")
	rootCmd.PersistentFlags().StringVar(&vizionBase.SSHKey.KeyFile, "ssh_key", "", "ssh login PrivateKey file full path")
}

// initConfig reads in config file and ENV variables if set.
/*
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".vztest" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".vztest")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}
*/

// CaseMapToString ...
func CaseMapToString(caseMap map[string]string) string {
	caseString := fmt.Sprintf("\n  %-3s %-20s  CaseDescription\n", "NO.", "CaseName")
	idx := 1
	for k, v := range caseMap {
		caseString += fmt.Sprintf("  %-3d %-20s  %s\n", idx, k, v)
		idx++
	}

	return caseString
}

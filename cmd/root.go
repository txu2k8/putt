package cmd

import (
	"fmt"
	"os"
	"path"
	"pzatest/libs/tlog"
	"strings"
	"time"

	"github.com/op/go-logging"
	"github.com/spf13/cobra"
)

// SSHKey ...
type SSHKey struct {
	UserName string // ssh login username
	Password string // ssh loging password
	Port     int    // ssh login port, default: 22
	KeyFile  string // ssh login PrivateKey file full path
}

// VizionBaseInput ...
type VizionBaseInput struct {
	MasterIPs    []string // Master nodes ips array
	VsetIDs      []int    // vset ids array
	DPLGroupIDs  []int    // dpl group ids array
	JDGroupIDs   []int    // jd group ids array
	K8sNameSpace string   // k8s namespace
	SSHKey                // ssh keys for connect to nodes
}

var (
	logger     = logging.MustGetLogger("test")
	runTimes   int      // runTimes
	debug      bool     // debug modle
	caseList   []string // Case List
	vizionBase VizionBaseInput
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "pzatest",
	Short: "The rootCmd of this test project",
	Long:  `pzatest project for "Stress | DevOps | Maintenance | ..."`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Vizion Base Info: %v", vizionBase)
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	initLogging()
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

	rootCmd.PersistentFlags().IntVar(&runTimes, "run_times", 1, "Run test case with iteration loop")
	rootCmd.PersistentFlags().StringArrayVar(&caseList, "case", []string{}, "Test Case Array")

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.vztest.yaml)")
	rootCmd.PersistentFlags().StringArrayVar(&vizionBase.MasterIPs, "master_ips", []string{}, "Master nodes IP address Array")
	rootCmd.PersistentFlags().IntSliceVar(&vizionBase.VsetIDs, "vset_ids", []int{}, "vset IDs array")
	rootCmd.PersistentFlags().IntSliceVar(&vizionBase.DPLGroupIDs, "dpl_group_ids", []int{1}, "dpl group ids array")
	rootCmd.PersistentFlags().IntSliceVar(&vizionBase.JDGroupIDs, "jd_group_ids", []int{1}, "jd group ids array")
	rootCmd.PersistentFlags().StringVar(&vizionBase.K8sNameSpace, "k8s_np", "vizion", "k8s namespace")
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

// initLogging initialize the logging configs
func initLogging() {
	dir, _ := os.Getwd()
	fileLogName := "vizion"
	fileLogPath := path.Join(dir, "log")
	timeStr := time.Now().Format("20060102150405")
	for _, v := range stripArgs() {
		fileLogName = fmt.Sprintf("%s-%s", fileLogName, v)
		fileLogPath = path.Join(fileLogPath, v)
	}
	fileLogName = fmt.Sprintf("%s-%s.log", fileLogName, timeStr)
	fileLogPath = path.Join(fileLogPath, fileLogName)

	conf := tlog.NewOptions(
		tlog.OptionSetFileLogPath(fileLogPath),
	)
	conf.InitLogging()
	logger.Infof("Args: pzatest %s", strings.Join(os.Args[1:], " "))
}

// ========== Common functions ==========

// CaseMapToString ...
func caseMapToString(caseMap map[string]string) string {
	caseString := fmt.Sprintf("\n  %-3s %-20s  CaseDescription\n", "NO.", "CaseName")
	idx := 1
	for k, v := range caseMap {
		caseString += fmt.Sprintf("  %-3d %-20s  %s\n", idx, k, v)
		idx++
	}

	return caseString
}

func stripArgs() []string {
	commands := []string{}
	args := os.Args[1:]
	ps := ""
	for len(args) > 0 {
		s := args[0]
		args = args[1:]
		switch {
		case s == "--":
			// "--" terminates the flags
			break
		case strings.HasPrefix(s, "--") && !strings.Contains(s, "="):
			// If '--flag arg' then
			// delete arg from args.
			fallthrough // (do the same as below)
		case strings.HasPrefix(s, "-") && !strings.Contains(s, "=") && len(s) == 2:
			// If '-f arg' then
			// delete 'arg' from args or break the loop if len(args) <= 1.
			if len(args) <= 1 {
				break
			} else {
				args = args[1:]
				continue
			}
		case s != "" && !strings.HasPrefix(s, "-") && !strings.HasPrefix(ps, "-"):
			commands = append(commands, s)
		}
		ps = s
	}

	return commands
}

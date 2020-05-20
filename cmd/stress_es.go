package cmd

import (
	"fmt"
	"pzatest/libs/runner/stress"
	"pzatest/vizion/testcase"

	"github.com/spf13/cobra"
)

var esTestConf = testcase.ESTestInput{}
var esTestCaseArray = map[string]string{
	"index":   "es index test (default)",
	"search":  "es search test",
	"cleanup": "cleanup exist es index",
	"stress":  "es index stress test: index && search",
}

// esCmd represents the es command
var esCmd = &cobra.Command{
	Use:   "es",
	Short: "Vizion ES Index/Search",
	Long:  fmt.Sprintf(`Vizion ES Index/Search(multi progress).%s`, caseMapToString(esTestCaseArray)),
	Run: func(cmd *cobra.Command, args []string) {
		if len(caseList) == 0 {
			caseList = []string{"index"}
		}
		logger.Infof("Case List(es): %s", caseList)
		testJobs := []stress.Job{}
		var esTester testcase.ESTester
		esTester = &esTestConf
		for _, tc := range caseList {
			logger.Warning(tc)
			jobs := []stress.Job{}
			switch tc {
			case "index":
				jobs = []stress.Job{
					{
						Fn:       esTester.ESIndex,
						Name:     "ES Index",
						RunTimes: runTimes,
					},
				}
			case "search":
				jobs = []stress.Job{
					{
						Fn:       esTester.ESSearch,
						Name:     "ES Search",
						RunTimes: runTimes,
					},
				}
			case "stress":
				jobs = []stress.Job{
					{
						Fn:       esTester.ESIndex,
						Name:     "ES Index",
						RunTimes: runTimes,
					},
					{
						Fn:       esTester.ESSearch,
						Name:     "ES Search",
						RunTimes: runTimes,
					},
				}
			case "cleanup":
				jobs = []stress.Job{
					{
						Fn:       esTester.ESCleanup,
						Name:     "Cleanup ES Index",
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
	stressCmd.AddCommand(esCmd)

	esCmd.PersistentFlags().StringVar(&esTestConf.IP, "es_ip", "", "ES server IP address")
	esCmd.PersistentFlags().StringVar(&esTestConf.UserName, "es_user", "root", "ES login username")
	esCmd.PersistentFlags().StringVar(&esTestConf.Password, "es_pwd", "password", "ES login password")
	esCmd.PersistentFlags().IntVar(&esTestConf.Port, "es_port", 9211, "ES server access port")
	esCmd.PersistentFlags().StringVar(&esTestConf.IndexNamePrefix, "index_name", "pzatest", "index name prefix")
	esCmd.PersistentFlags().IntVar(&esTestConf.Indices, "indice", 50, "Number of indices to write")
	esCmd.PersistentFlags().IntVar(&esTestConf.Documents, "document", 100000, "Number of template documents that hold the same mapping")
	esCmd.PersistentFlags().IntVar(&esTestConf.BulkSize, "bulk_size", 1000, "How many documents each bulk request should contain")

	esCmd.MarkPersistentFlagRequired("es_ip")
	esCmd.MarkPersistentFlagRequired("es_port")
}

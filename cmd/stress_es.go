package cmd

import (
	"fmt"
	"gtest/libs/runner/stress"
	"gtest/models"
	"gtest/vizion/resources"

	"github.com/spf13/cobra"
)

var esTestConf = models.S3TestInput{}
var esTestCaseArray = map[string]string{
	"index":   "es index test (default)",
	"search":  "es search test",
	"stress":  "es index stress test: index && search",
	"cleanup": "cleanup exist es index",
}

// esCmd represents the es command
var esCmd = &cobra.Command{
	Use:   "es",
	Short: "Vizion ES Index/Search",
	Long:  fmt.Sprintf(`Vizion ES Index/Search(multi progress).%s`, CaseMapToString(esTestCaseArray)),
	Run: func(cmd *cobra.Command, args []string) {
		// fmt.Println("es called")
		// fmt.Println(caseList)
		logger.Infof("es: %s", caseList)
		for _, tc := range caseList {
			switch tc {
			case "index":
				jobs := []stress.Job{
					{
						Fn:       resources.ESIndex,
						Name:     "ES Index",
						RunTimes: runTimes,
					},
				}
				stress.Run(jobs)
			}
		}
	},
}

func init() {
	stressCmd.AddCommand(esCmd)

	esCmd.PersistentFlags().StringArrayVar(&caseList, "case", []string{}, "ES test case array")
	// esCmd.PersistentFlags().String("foo", "", "A help for foo")
	// esCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

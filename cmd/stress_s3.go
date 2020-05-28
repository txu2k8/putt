package cmd

import (
	"fmt"

	"pzatest/libs/runner/stress"
	"pzatest/vizion/testcase"

	"github.com/spf13/cobra"
)

var s3TestConf = testcase.S3TestInput{}

var s3TestCaseArray = map[string]string{
	"upload":          "s3 upload test",
	"download":        "s3 download test:TODO",
	"upload_download": "s3 upload/download test (default)",
	"multi_users":     "multi users s3 upload/download test(TODO)",
}

// s3Cmd represents the s3 command
var s3Cmd = &cobra.Command{
	Use:   "s3",
	Short: "Vizion S3 IO Stress",
	Long:  fmt.Sprintf(`Vizion S3 upload/download files.%s`, caseMapToString(s3TestCaseArray)),
	Run: func(cmd *cobra.Command, args []string) {
		if len(caseList) == 0 {
			caseList = []string{"upload_download"}
		}
		logger.Infof("Case List(s3): %s", caseList)
		testJobs := []stress.Job{}
		var s3Tester testcase.S3Tester
		s3Tester = &s3TestConf
		for _, tc := range caseList {
			jobs := []stress.Job{}
			switch tc {
			case "upload":
				upload := func() error {
					_, err := s3Tester.MultiS3UploadFiles()
					return err
				}
				jobs = []stress.Job{
					{
						Fn:       upload,
						Name:     "Multi S3 Upload",
						RunTimes: runTimes,
					},
				}
			case "upload_download":
				jobs = []stress.Job{
					{
						Fn:       s3Tester.MultiS3UploadDownloadListDeleteFiles,
						Name:     "Multi S3 Upload/List/Download/Delete",
						RunTimes: runTimes,
					},
				}
			case "multi_users":
				jobs = []stress.Job{
					{
						Fn:       s3Tester.MultiUserS3UploadDownloadListDeleteFiles,
						Name:     "Multi-Users S3 Upload/List/Download/Delete",
						RunTimes: runTimes,
					},
				}

			}
			testJobs = append(testJobs, jobs...)
		}
		stress.Run(testJobs)
	},
}

// AddFlagsS3 ...
func AddFlagsS3(cmd *cobra.Command) {
	cmd.PersistentFlags().StringVar(&s3TestConf.S3Ip, "s3_ip", "", "S3 server IP address")
	cmd.PersistentFlags().StringVar(&s3TestConf.S3AccessID, "s3_access_id", "", "S3 access ID")
	cmd.PersistentFlags().StringVar(&s3TestConf.S3SecretKey, "s3_secret_key", "", "S3 access secret key")
	cmd.PersistentFlags().IntVar(&s3TestConf.S3Port, "s3_port", 443, "S3 server access port")
	cmd.PersistentFlags().StringVar(&s3TestConf.S3Bucket, "s3_bucket", "", "S3 bucket name")
	cmd.PersistentFlags().StringVar(&s3TestConf.LocalDataDir, "s3_local_dir", "/tmp/", "S3 test Local dir for save files")
	cmd.PersistentFlags().IntVar(&s3TestConf.RandomPercent, "s3_random_percent", 100, "S3 test Percent of files with random data")
	cmd.PersistentFlags().IntVar(&s3TestConf.EmptyPercent, "s3_empty_percent", 0, "S3 test Percent of files with empty data(0~100)% (default 0)")
	cmd.PersistentFlags().BoolVar(&s3TestConf.RenameFile, "s3_rename", false, "S3 test Rename files name each time if true (default false)")
	cmd.PersistentFlags().BoolVar(&s3TestConf.DeleteFile, "s3_delete", false, "S3 test Delete files from bucket after test if true (default false)")
	cmd.PersistentFlags().IntVar(&s3TestConf.Clients, "s3_client", 1, "S3 test Client number exec at the same time")
	cmd.PersistentFlags().StringArrayVar(&s3TestConf.FileInputs, "s3_files", []string{"txt:20:1k-10k", "dd:1:100mb"}, "S3 test files config array")

	// cmd.MarkPersistentFlagRequired("s3_ip")
	// cmd.MarkPersistentFlagRequired("s3_bucket")
}

func init() {
	stressCmd.AddCommand(s3Cmd)
	AddFlagsS3(s3Cmd)
	s3Cmd.MarkPersistentFlagRequired("s3_ip")
	s3Cmd.MarkPersistentFlagRequired("s3_bucket")
}

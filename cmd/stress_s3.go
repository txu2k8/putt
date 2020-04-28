package cmd

import (
	"fmt"

	"gtest/models"
	"gtest/vizion/resources"

	"github.com/spf13/cobra"
)

var (
	caseList   []string
	s3TestConf = models.S3TestInput{}
)

// s3Cmd represents the s3 command
var s3Cmd = &cobra.Command{
	Use:   "s3",
	Short: "Vizion S3 IO Stress",
	Long:  `Vizion S3 upload/download files. --help for detail args.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("s3 called", s3TestConf)
		for _, tc := range caseList {
			switch tc {
			case "upload":
				resources.S3UploadFiles(s3TestConf)
			}
		}
	},
}

func init() {
	stressCmd.AddCommand(s3Cmd)

	s3Cmd.PersistentFlags().StringArrayVar(&caseList, "case", []string{"upload"}, "S3 test case array")
	s3Cmd.PersistentFlags().StringVar(&s3TestConf.S3Ip, "s3_ip", "", "S3 server IP address")
	s3Cmd.PersistentFlags().StringVar(&s3TestConf.S3AccessID, "s3_access_id", "", "S3 access ID")
	s3Cmd.PersistentFlags().StringVar(&s3TestConf.S3SecretKey, "s3_secret_key", "", "S3 access secret key")
	s3Cmd.PersistentFlags().IntVar(&s3TestConf.S3Port, "s3_port", 443, "S3 server access port")
	s3Cmd.PersistentFlags().StringVar(&s3TestConf.S3Bucket, "s3_bucket", "", "S3 bucket name")
	s3Cmd.PersistentFlags().StringVar(&s3TestConf.LocalDataDir, "local_dir", "/tmp/", "Local dir for store S3 test files")
	s3Cmd.PersistentFlags().IntVar(&s3TestConf.RandomPercent, "random_percent", 100, "Percent of S3 test files with random data")
	s3Cmd.PersistentFlags().IntVar(&s3TestConf.EmptyPercent, "empty_percent", 0, "Percent of S3 test files with empty data")
	s3Cmd.PersistentFlags().BoolVar(&s3TestConf.RenameFile, "rename", true, "Rename files name each time if true")
	s3Cmd.PersistentFlags().BoolVar(&s3TestConf.DeleteFile, "delete", true, "Delete files from s3 bucket after test if true")
	s3Cmd.PersistentFlags().IntVar(&s3TestConf.Clients, "client", 1, "S3 Client number for test at the same time")
	s3Cmd.PersistentFlags().StringArrayVar(&s3TestConf.FileInputs, "files", []string{"txt:20:1k-10k", "dd:1:100mb"}, "S3 files config array")

	s3Cmd.MarkPersistentFlagRequired("s3_ip")
	s3Cmd.MarkPersistentFlagRequired("s3_bucket")
}

package testcases

import (
	"flag"
	"fmt"
	"gtest/vizion/resources"

	"gopkg.in/alecthomas/kingpin.v2"
)

// S3TestArgs ...
type S3TestArgs resources.S3TestConfig

func (arg *S3TestArgs) run(c *kingpin.ParseContext) error {
	fmt.Printf("Would xxxxxxxxx")
	return nil
}

// S3Command ...
func S3Command(parentCmd *kingpin.CmdClause, s3args *S3TestArgs) *kingpin.CmdClause {
	s3Cmd := parentCmd.Command("s3", "Vizion S3 Stress Test")
	s3Cmd.Arg("S3Ip", "host login username").Default("root").String()
	s3Cmd.Arg("S3AccessID", "host login password").Default("password").String()
	s3Cmd.Arg("S3SecretKey", "host login key files").Default("").String()
	return s3Cmd
}

func addSubCommand(app *kingpin.Application, name string, description string) {
	c := app.Command(name, description).Action(func(c *kingpin.ParseContext) error {
		fmt.Printf("Would have run command %s.\n", name)
		return nil
	})
	c.Flag("nop-flag", "Example of a flag with no options").Bool()
}

func initArgs(s3c *resources.S3TestConfig) {
	flag.StringVar(&s3c.S3Ip, "s3_ip", "", "s3 server ip address")
	flag.StringVar(&s3c.S3AccessID, "s3_access_id", "", "s3 server ip address")
	flag.StringVar(&s3c.S3SecretKey, "s3ip", "", "s3 server ip address")
	flag.IntVar(&s3c.S3Port, "s3_port", 443, "s3 server port")
	flag.StringVar(&s3c.S3Bucket, "s3_bucket", "", "s3 bucket name")
	flag.StringVar(&s3c.LocalDataDir, "local_path", "/tmp/", "local path for s3 test")
	flag.IntVar(&s3c.RandomPercent, "rand_percent", 100, "s3 server ip address")
	flag.IntVar(&s3c.EmptyPercent, "empty_percent", 0, "s3 server ip address")
	flag.BoolVar(&s3c.RenameFile, "rename", false, "s3 server ip address")
	flag.BoolVar(&s3c.DeleteFile, "delete", false, "s3 server ip address")
	flag.IntVar(&s3c.Clients, "clients", 1, "s3 server ip address")
}

package testcases

import (
	"flag"
	"testing"

	"gtest/vizion/resources"

	. "github.com/smartystreets/goconvey/convey"
)

// TestS3UploadFiles ...
func TestS3UploadFiles(t *testing.T) {
	var s3TestConf resources.S3TestConfig
	initArgs(&s3TestConf)
	flag.Parse()

	Convey("Upload files to vizionS3", t, func() {
		resources.S3UploadFiles(s3TestConf)
	})
}

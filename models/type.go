package models

import (
	"fmt"
	"path"
	"pzatest/libs/utils"
	"strconv"
	"strings"

	"github.com/gocql/gocql"
	"github.com/op/go-logging"
)

var logger = logging.MustGetLogger("test")

// SSHKey ...
type SSHKey struct {
	UserName string // ssh login username
	Password string // ssh loging password
	Port     int    // ssh login port, default: 22
	KeyFile  string // ssh login PrivateKey file full path
}

// VizionBaseInput ...
type VizionBaseInput struct {
	MasterIPs   []string       // Master nodes ips array
	VsetIDs     []int          // vset ids array
	DPLGroupIDs []int          // dpl group ids array
	JDGroupIDs  []int          // jd group ids array
	SSHKey                     // ssh keys for connect to nodes
	MasterCass  *gocql.Session // master cassandra session
}

// S3TestFileInput define s3 test file config
type S3TestFileInput struct {
	FileType       string // txt or dd
	FileNum        int    // file number
	FileSizeMin    int64  // the min size of file
	FileSizeMax    int64  // the max size of the file
	FileNamePrefix string // the file name prefix
	FileDir        string // the file dir path
}

// S3TestInput define S3 test config
type S3TestInput struct {
	S3Ip             string // endpoint: https://<S3Ip>:<S3Port>, eg: https://10.25.119.86:443
	S3AccessID       string
	S3SecretKey      string
	S3Port           int               // port (default: 443)
	S3Bucket         string            // s3 bucket for test
	LocalDataDir     string            // The local data Dir
	FileInputs       []string          // S3 files config array,eg: {"txt:20:1k-10k", "dd:1:100mb"}
	RandomPercent    int               // percent of files with random data
	EmptyPercent     int               // percent of files with empty data
	RenameFile       bool              // rename files name each time if true
	DeleteFile       bool              // delete files from s3 bucket after test if true
	Clients          int               // S3 Client number for test at the same time
	S3TestFileInputs []S3TestFileInput // Parse(FileInputs) --> S3TestFileInputs
	Endpoint         string            // Parse(S3Ip,S3Port) --> Endpoint
}

// ParseS3Input ...
func (conf *S3TestInput) ParseS3Input() {
	// Parse S3Ip S3Port to conf.endpoint
	conf.Endpoint = fmt.Sprintf("https://%s:%d", conf.S3Ip, conf.S3Port)
	// Parse FileInputs to conf.S3TestFileInputs
	strS3Ip := strings.Replace(conf.S3Ip, ".", "", -1)
	conf.S3TestFileInputs = make([]S3TestFileInput, len(conf.FileInputs))
	for i, v := range conf.FileInputs {
		fArr := strings.Split(v, ":")
		// fmt.Println(fArr)
		conf.S3TestFileInputs[i].FileType = fArr[0]
		conf.S3TestFileInputs[i].FileNum, _ = strconv.Atoi(fArr[1])
		nArr := strings.Split(fArr[2], "-")
		conf.S3TestFileInputs[i].FileSizeMin = utils.SizeCountToByte(nArr[0])
		if len(nArr) > 1 {
			conf.S3TestFileInputs[i].FileSizeMax = utils.SizeCountToByte(nArr[1])
		} else {
			conf.S3TestFileInputs[i].FileSizeMax = conf.S3TestFileInputs[i].FileSizeMin
		}
		conf.S3TestFileInputs[i].FileNamePrefix = fmt.Sprintf("s3stress_%s", strS3Ip)
		conf.S3TestFileInputs[i].FileDir = path.Join(conf.LocalDataDir, strS3Ip)
	}
	logger.Debugf("S3TestInput:%v", utils.Prettify(conf))
}

// ESTestInput ...
type ESTestInput struct {
	IP              string
	UserName        string
	Password        string
	Port            int
	URL             string // Parse(IP,Port) --> URL
	IndexNamePrefix string // index name prefix
	Indices         int    // Number of indices to write
	Documents       int    // Number of template documents that hold the same mapping
	BulkSize        int    // How many documents each bulk request should contain
}

// ParseESInput ...
func (conf *ESTestInput) ParseESInput() {
	// Parse ES Ip Port to conf.URL
	conf.URL = fmt.Sprintf("http://%s:%d", conf.IP, conf.Port)
	logger.Debugf("ESTestInput:%v", utils.Prettify(conf))
}

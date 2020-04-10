package models

// SSHKey ...
type SSHKey struct {
	UserName string // login username
	Password string // loging password
	KeyFile  string // The login key file full path
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
	S3TestFileInputs []S3TestFileInput // S3 files config list
	RandomPercent    int               // percent of files with random data
	EmptyPercent     int               // percent of files with empty data
	RenameFile       bool              // rename files name each time if true
	DeleteFile       bool              // delete files from s3 bucket after test if true
	Clients          int               // S3 Client number for test at the same time
}

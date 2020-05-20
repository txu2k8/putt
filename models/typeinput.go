package models

import (
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

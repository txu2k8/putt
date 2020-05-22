package types

import (
	"pzatest/libs/sshmgr"
)

// VizionBaseInput ...
type VizionBaseInput struct {
	MasterIPs     []string // Master nodes ips array
	VsetIDs       []int    // vset ids array
	DPLGroupIDs   []int    // dpl group ids array
	JDGroupIDs    []int    // jd group ids array
	K8sNameSpace  string   // k8s namespace
	KubeConfig    string   // kubeconfig file path
	sshmgr.SSHKey          // ssh keys for connect to nodes
}

package etcd

import (
	"testing"
)

func TestEtcd(t *testing.T) {
	cfg := Config{
		Endpoints: []string{
			"10.25.119.71:2379",
			"10.25.119.72:2379",
			"10.25.119.73:2379",
		},
		CertFile:      "/tmp/certs/peer.crt",
		KeyFile:       "/tmp/certs/peer.key",
		TrustedCAFile: "/tmp/certs/ca.crt",
	}

	cli, _ := NewClientWithRetry(cfg)
	resp, _ := cli.GetPrefix("/vizion/dpl/jnl")
	logger.Info(resp.Header)
}

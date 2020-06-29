package etcd

import (
	"context"
	"fmt"
	"time"

	"github.com/op/go-logging"
	"go.etcd.io/etcd/clientv3"
	"go.etcd.io/etcd/pkg/transport"
)

var logger = logging.MustGetLogger("test")

var (
	// ConnectTimeout .
	ConnectTimeout int = 600 // Second
	client         *Client
)

// Client .
type Client struct {
	*clientv3.Client
}

// Config config for etcd
type Config struct {
	Endpoints     []string // eg: localhost:2379
	CertFile      string   // TLSConfig: certs/test-name-1.pem
	KeyFile       string   // TLSConfig: certs/test-name-1-key.pem
	TrustedCAFile string   // TLSConfig: certs/trusted-ca.pem
	Username      string   // login etcd user
	Password      string   // login etcd pwd
}

// NewClient .
func NewClient(cfg Config) (*clientv3.Client, error) {
	tlsInfo := transport.TLSInfo{
		CertFile:      cfg.CertFile,
		KeyFile:       cfg.KeyFile,
		TrustedCAFile: cfg.TrustedCAFile,
	}
	tlsConfig, err := tlsInfo.ClientConfig()
	if err != nil {
		return nil, err
	}

	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   cfg.Endpoints,
		DialTimeout: time.Duration(ConnectTimeout) * time.Second,
		TLS:         tlsConfig,
	})
	if err != nil {
		return nil, err
	}
	return cli, nil
}

// NewClientWithRetry return the Client
func NewClientWithRetry(cfg Config) (*Client, error) {
	if client != nil {
		return client, nil
	}
	interval := time.Duration(15)
	timeout := time.NewTimer(30 * time.Minute)
	var err error
loop:
	for {
		cli, err := NewClient(cfg)
		client = &Client{cli}
		if err == nil && client != nil {
			break loop
		}
		logger.Warningf("new etcd clientv3 failed, %v", err)

		// retry or timeout
		select {
		case <-time.After(interval * time.Second):
			logger.Infof("retry new etcd clientv3 after %d second", interval)
		case <-timeout.C:
			err = fmt.Errorf("new etcd clientv3 failed after retry many times, cause by %v", err)
			break loop
		}
	}
	return client, err
}

// GetPrefix .
func (cli *Client) GetPrefix(prefix string) (resp *clientv3.GetResponse, err error) {
	resp, err = cli.Get(context.TODO(), prefix, clientv3.WithPrefix())
	logger.Info(resp.Count)
	return
}

// DeletePrefix .
func (cli *Client) DeletePrefix(prefix string) (resp *clientv3.DeleteResponse, err error) {
	resp, err = cli.Delete(context.TODO(), prefix, clientv3.WithPrefix())
	logger.Info(resp.Deleted)
	return
}

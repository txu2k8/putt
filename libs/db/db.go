package db

import (
	"fmt"
	"time"

	"github.com/gocql/gocql"
	"github.com/op/go-logging"
)

var logger = logging.MustGetLogger("test")

var (
	cassTimeout        int = 600 // Second
	cassConnectTimeout int = 600 // Second
	session            *gocql.Session
)

// CassConfig config for cassandra
type CassConfig struct {
	host     string
	user     string
	pwd      string
	keyspace string
	port     int
}

func connectCluster(cf *CassConfig) *gocql.ClusterConfig {
	// connect to the cluster
	cluster := gocql.NewCluster(cf.host)
	cluster.Port = cf.port
	cluster.Keyspace = cf.keyspace
	cluster.Timeout = time.Duration(cassTimeout) * time.Second
	cluster.ConnectTimeout = time.Duration(cassConnectTimeout) * time.Second
	cluster.Authenticator = gocql.PasswordAuthenticator{
		Username: cf.user,
		Password: cf.pwd,
	}
	cluster.Consistency = gocql.LocalQuorum
	cluster.NumConns = 3 // set connection pool num
	return cluster
}

// NewSession return the cassandra session
func NewSession(cf *CassConfig) (*gocql.Session, error) {
	cassCluster := connectCluster(cf)
	return cassCluster.CreateSession()
}

// NewSessionWithRetry return the cassandra session
func NewSessionWithRetry(cf *CassConfig) (*gocql.Session, error) {
	if session != nil {
		return session, nil
	}
	interval := time.Duration(15)
	timeout := time.NewTimer(30 * time.Minute)
	var err error

loop:
	for {
		session, err = NewSession(cf)
		if err == nil && session != nil {
			break loop
		}
		logger.Warningf("new cassandra session failed, %v", err)

		// retry or timeout
		select {
		case <-time.After(interval * time.Second):
			logger.Infof("retry new cassandra session after %d second", interval)
		case <-timeout.C:
			err = fmt.Errorf("new cassandra session failed after retry many times, cause by %v", err)
			break loop
		}
	}
	return session, err
}

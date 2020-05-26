package sshmgr

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"time"

	"github.com/op/go-logging"
	"golang.org/x/crypto/ssh"
)

var logger = logging.MustGetLogger("test")

// SSHManager ...
type SSHManager interface {
	RunCmd(cmdSpec string) (int, string) // session.RunCmd
	SCPGet(localPath, remotePath string) error
}

// SSHMgr .
type SSHMgr struct {
	*ssh.Session
	*SSHConfig
}

// SSHKey ssh login keys
type SSHKey struct {
	UserName string // ssh login username
	Password string // ssh loging password
	Port     int    // ssh login port, default: 22
	KeyFile  string // ssh login PrivateKey file full path
}

// SSHConfig ssh login config
type SSHConfig struct {
	Host           string        // ssh target host ip address
	SSHKey                       // ssh login keys
	Timeout        time.Duration // connection timeout (default: 600ms)
	ConnectTimeout time.Duration // initial connection timeout, used during initial dial to server (default: 600ms)
}

// NewSSHConfig generates a new config for the default ssh implementation.
func NewSSHConfig(host string, sshKey SSHKey) *SSHConfig {
	cfg := &SSHConfig{
		Host:           host,
		SSHKey:         sshKey,
		Timeout:        600 * time.Second,
		ConnectTimeout: 600 * time.Second,
	}
	return cfg
}

// CreateSession initializes the cluster based on this config and returns a
// session object that can be used to interact with the database.
func (cfg *SSHConfig) CreateSession() (*ssh.Session, error) {
	return cfg.NewSessionWithRetry()
}

// NewClient return the ssh client
func (cfg *SSHConfig) NewClient() (*ssh.Client, error) {
	// get auth method
	auth := make([]ssh.AuthMethod, 0)
	if cfg.KeyFile != "" {
		// Use the PublicKeys method for remote authentication.
		key, err := ioutil.ReadFile(cfg.KeyFile) // privateKey file path,eg:/home/user/.ssh/id_rsa
		if err != nil {
			log.Fatalf("unable to read private key: %v", err)
		}
		// Create the Signer for this private key.
		signer, err := ssh.ParsePrivateKey(key)
		if err != nil {
			log.Fatalf("unable to parse private key: %v", err)
		}
		auth = append(auth, ssh.PublicKeys(signer))
	} else {
		// Use the password for remote authentication.
		auth = append(auth, ssh.Password(cfg.Password))
	}

	hostKeyCallbk := func(hostname string, remote net.Addr, key ssh.PublicKey) error {
		return nil
	}

	config := &ssh.ClientConfig{
		User:            cfg.UserName,
		Auth:            auth,
		Timeout:         cfg.Timeout,
		HostKeyCallback: hostKeyCallbk,
	}

	// connet to ssh
	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	logger.Infof("SSH Connect to %s@%s(pwd:%s, privateKey:%s)",
		cfg.UserName, addr, cfg.Password, cfg.KeyFile)
	client, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		return nil, err
	}
	return client, nil
}

// NewSession return the cassandra session
func (cfg *SSHConfig) NewSession() (*ssh.Session, error) {
	// connet to ssh
	client, err := cfg.NewClient()
	if err != nil {
		return nil, err
	}

	// create session
	session, err := client.NewSession()
	if err != nil {
		return nil, err
	}

	return session, nil
}

// NewSessionWithRetry return the cassandra session
func (cfg *SSHConfig) NewSessionWithRetry() (*ssh.Session, error) {
	interval := time.Duration(15)
	timeout := time.NewTimer(30 * time.Minute)
	var session *ssh.Session
	var err error

loop:
	for {
		session, err = cfg.NewSession()
		if err == nil && session != nil {
			break loop
		}
		logger.Warningf("New ssh session failed, %v", err)

		// retry or timeout
		select {
		case <-time.After(interval * time.Second):
			logger.Infof("retry new ssh session after %d second", interval)
		case <-timeout.C:
			err = fmt.Errorf("new ssh session failed after retry many times, cause by %v", err)
			break loop
		}
	}
	return session, err
}

// RunCmd ...
func (session *SSHMgr) RunCmd(cmdSpec string) (int, string) {
	var rc int
	var stdOut, stdErr bytes.Buffer

	session.Stdout = &stdOut
	session.Stderr = &stdErr

	logger.Infof("SSH Execute: ssh %s@%s# %s", session.UserName, session.Host, cmdSpec)
	if err := session.Run(cmdSpec); err != nil {
		logger.Debugf("Failed to run: %s", err.Error())
		// Process exited with status 1
		rc = 1
	}
	if stdErr.Len() == 0 {
		rc = 0
	} else {
		rc = -1
	}
	output := stdOut.String() + stdErr.String()
	// logger.Infof("%d, %s\n", rc, stdOut.String())
	return rc, output
}

// // RunCmd ...
// func (s *sshMgr) RunCmd(cmdSpec string) (int, string) {
// 	logger.Infof("Execute: ssh %s@%s# %s", cfg.UserName, cfg.Host, cmdSpec)
// 	return RunCmdWithOutput(session, cmdSpec)
// }

// SCPGet ...
func (session *SSHMgr) SCPGet(localPath, remotePath string) error {
	logger.Info("SCPGet ...")
	return nil
}

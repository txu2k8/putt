package sshmgr

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"time"

	"github.com/op/go-logging"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

var logger = logging.MustGetLogger("test")

// SSHManager ...
type SSHManager interface {
	RunCmd(cmdSpec string) (int, string) // session.RunCmd
	ScpGet(localPath, remotePath string) error
	ScpPut(localPath, remotePath string) error
}

// SSHMgr .
type SSHMgr struct {
	Client     *ssh.Client  //ssh client
	Cfg        *SSHConfig   // ssh config
	SftpClient *sftp.Client // sftp client for scp copy files
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

// NewSSHMgr generates a new SSHManager for the default ssh implementation.
func NewSSHMgr(host string, sshKey SSHKey) *SSHMgr {
	cfg := &SSHConfig{
		Host:           host,
		SSHKey:         sshKey,
		Timeout:        600 * time.Second,
		ConnectTimeout: 600 * time.Second,
	}

	client, err := cfg.NewClientWithRetry()
	if err != nil {
		panic(err)
	}

	return &SSHMgr{Client: client, Cfg: cfg}
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

// NewSftpClient .
func (cfg *SSHConfig) NewSftpClient() (*sftp.Client, error) {
	// connet to ssh
	client, err := cfg.NewClient()
	if err != nil {
		return nil, err
	}

	// create sftp client
	sftpClient, err := sftp.NewClient(client)
	if err != nil {
		return nil, err
	}

	return sftpClient, nil
}

// NewClientWithRetry return the ssh client
func (cfg *SSHConfig) NewClientWithRetry() (*ssh.Client, error) {
	interval := time.Duration(15)
	timeout := time.NewTimer(30 * time.Minute)
	var client *ssh.Client
	var err error

loop:
	for {
		client, err = cfg.NewClient()
		if err == nil && client != nil {
			break loop
		}
		logger.Warningf("New ssh client failed, %v", err)

		// retry or timeout
		select {
		case <-time.After(interval * time.Second):
			logger.Infof("retry new ssh client after %d second", interval)
		case <-timeout.C:
			err = fmt.Errorf("new ssh client failed after retry many times, cause by %v", err)
			break loop
		}
	}
	return client, err
}

// NewSftpClientWithRetry return the sftp.Client
func (cfg *SSHConfig) NewSftpClientWithRetry() (*sftp.Client, error) {
	interval := time.Duration(15)
	timeout := time.NewTimer(30 * time.Minute)
	var sftpClient *sftp.Client
	var err error

loop:
	for {
		sftpClient, err = cfg.NewSftpClient()
		if err == nil && sftpClient != nil {
			break loop
		}
		logger.Warningf("New ssh sftpClient failed, %v", err)

		// retry or timeout
		select {
		case <-time.After(interval * time.Second):
			logger.Infof("retry new ssh session after %d second", interval)
		case <-timeout.C:
			err = fmt.Errorf("new ssh session failed after retry many times, cause by %v", err)
			break loop
		}
	}
	return sftpClient, err
}

// RunCmd ...
func (sshMgr *SSHMgr) RunCmd(cmdSpec string) (int, string) {
	var rc int
	var stdOut, stdErr bytes.Buffer
	// create session
	session, err := sshMgr.Client.NewSession()
	if err != nil {
		return -1, fmt.Sprintf("%s", err)
	}
	session.Stdout = &stdOut
	session.Stderr = &stdErr

	logger.Infof("SSH Execute: ssh %s@%s# %s", sshMgr.Cfg.UserName, sshMgr.Cfg.Host, cmdSpec)
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
	// logger.Debugf("%d, %s\n", rc, output)
	return rc, output
}

// ConnectSftpClient ...
func (sshMgr *SSHMgr) ConnectSftpClient() {
	sftpClient, err := sshMgr.Cfg.NewSftpClientWithRetry()
	if err != nil {
		panic(err)
	}
	sshMgr.SftpClient = sftpClient
}

// ScpGet ...
func (sshMgr *SSHMgr) ScpGet(localPath, remotePath string) error {
	logger.Infof("scp %s@%s:%s %s ...", sshMgr.Cfg.UserName, sshMgr.Cfg.Host, remotePath, localPath)
	dstFile, err := os.Create(localPath)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	srcFile, err := sshMgr.SftpClient.Open(remotePath)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	buf := make([]byte, 1024)
	for {
		n, _ := srcFile.Read(buf)
		if n == 0 {
			break
		}
		dstFile.Write(buf[0:n])
	}

	return nil
}

// ScpPut ...
func (sshMgr *SSHMgr) ScpPut(localPath, remotePath string) error {
	logger.Infof("scpPut %s -> %s@%s:%s ...", localPath, sshMgr.Cfg.UserName, sshMgr.Cfg.Host, remotePath)
	srcFile, err := os.Open(localPath)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := sshMgr.SftpClient.Create(remotePath)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	buf := make([]byte, 1024)
	for {
		n, _ := srcFile.Read(buf)
		if n == 0 {
			break
		}
		dstFile.Write(buf[0:n])
	}

	return nil
}

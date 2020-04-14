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
	NewSession() (*ssh.Session, error)          // SSH connect session
	NewSessionWithRetry() (*ssh.Session, error) // SSH connect session, retry when failed
}

// SSHInput ssh login input keys
type SSHInput struct {
	Host     string // ssh node host ip address
	UserName string // ssh login username
	Password string // ssh loging password
	Port     int    // ssh login port, default: 22
	KeyFile  string // ssh login PrivateKey file full path
}

var (
	connectTimeout int = 600 // Second
	session        *ssh.Session
)

// NewSession return the cassandra session
func (conf *SSHInput) NewSession() (*ssh.Session, error) {
	// get auth method
	auth := make([]ssh.AuthMethod, 0)
	if conf.KeyFile != "" {
		// Use the PublicKeys method for remote authentication.
		key, err := ioutil.ReadFile(conf.KeyFile) // privateKey file path,eg:/home/user/.ssh/id_rsa
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
		auth = append(auth, ssh.Password(conf.Password))
	}

	hostKeyCallbk := func(hostname string, remote net.Addr, key ssh.PublicKey) error {
		return nil
	}

	config := &ssh.ClientConfig{
		User:            conf.UserName,
		Auth:            auth,
		Timeout:         time.Duration(connectTimeout) * time.Second,
		HostKeyCallback: hostKeyCallbk,
	}

	// connet to ssh
	addr := fmt.Sprintf("%s:%d", conf.Host, conf.Port)
	logger.Infof("SSH Connect to %s@%s(pwd:%s, privateKey:%s)",
		conf.UserName, addr, conf.Password, conf.KeyFile)
	client, err := ssh.Dial("tcp", addr, config)
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
func (conf *SSHInput) NewSessionWithRetry() (*ssh.Session, error) {
	if session != nil {
		return session, nil
	}
	interval := time.Duration(15)
	timeout := time.NewTimer(30 * time.Minute)
	var err error

loop:
	for {
		session, err = conf.NewSession()
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

// SessionRun ...
func SessionRun(session *ssh.Session, cmdSpec string) (int, string) {
	var rc int
	var stdOut, stdErr bytes.Buffer

	session.Stdout = &stdOut
	session.Stderr = &stdErr

	if err := session.Run(cmdSpec); err != nil {
		logger.Fatal("Failed to run: " + err.Error())
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

// RunCmd ...
func (conf *SSHInput) RunCmd(cmdSpec string) (int, string) {
	session, err := conf.NewSessionWithRetry()
	if err != nil {
		logger.Fatal(err)
	}
	defer session.Close()
	logger.Infof("Execute: ssh %s@%s# %s", conf.UserName, conf.Host, cmdSpec)
	return SessionRun(session, cmdSpec)
}

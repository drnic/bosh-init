package sshtunnel

import (
	"code.google.com/p/go.crypto/ssh"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"time"

	bosherr "github.com/cloudfoundry/bosh-agent/errors"
	boshlog "github.com/cloudfoundry/bosh-agent/logger"
)

type SSHTunnel interface {
	Start(chan<- error, chan<- error)
	Stop() error
}

type sshTunnel struct {
	startDialMaxTries int
	startDialDelay    time.Duration
	options           Options
	remoteListener    net.Listener
	logger            boshlog.Logger
	logTag            string
}

func (s *sshTunnel) Start(readyErrCh chan<- error, errCh chan<- error) {
	authMethods := []ssh.AuthMethod{}

	if s.options.PrivateKey != "" {
		s.logger.Debug(s.logTag, "Reading private key file")
		keyContents, err := ioutil.ReadFile(s.options.PrivateKey)
		if err != nil {
			readyErrCh <- bosherr.WrapError(err, "Reading private key file")
			return
		}

		s.logger.Debug(s.logTag, "Parsing private key file")
		signer, err := ssh.ParsePrivateKey(keyContents)
		if err != nil {
			readyErrCh <- bosherr.WrapError(err, "Parsing private key file")
			return
		}

		authMethods = append(authMethods, ssh.PublicKeys(signer))
	}

	if s.options.Password != "" {
		s.logger.Debug(s.logTag, "Adding password auth method to ssh tunnel config")

		keyboardInteractiveChallenge := func(
			user,
			instruction string,
			questions []string,
			echos []bool,
		) (answers []string, err error) {
			if len(questions) == 0 {
				return []string{}, nil
			}
			return []string{s.options.Password}, nil
		}
		authMethods = append(authMethods, ssh.KeyboardInteractive(keyboardInteractiveChallenge))
		authMethods = append(authMethods, ssh.Password(s.options.Password))
	}

	sshConfig := &ssh.ClientConfig{
		User: s.options.User,
		Auth: authMethods,
	}

	s.logger.Debug(s.logTag, "Dialing remote server at %s:%d", s.options.Host, s.options.Port)
	remoteAddr := fmt.Sprintf("%s:%d", s.options.Host, s.options.Port)

	var conn *ssh.Client
	var err error
	for i := 0; i < s.startDialMaxTries; i++ {
		conn, err = ssh.Dial(
			"tcp",
			remoteAddr,
			sshConfig,
		)

		if err != nil && i == s.startDialMaxTries-1 {
			readyErrCh <- bosherr.WrapError(err, "Timed out dialing remote server")
			return
		}

		if err == nil {
			break
		}

		time.Sleep(s.startDialDelay)
	}

	remoteListenAddr := fmt.Sprintf("127.0.0.1:%d", s.options.RemoteForwardPort)
	s.logger.Debug(s.logTag, "Listening on remote server %s", remoteListenAddr)
	s.remoteListener, err = conn.Listen("tcp", remoteListenAddr)
	if err != nil {
		readyErrCh <- bosherr.WrapError(err, "Listening on remote server")
		return
	}

	readyErrCh <- nil
	for {
		remoteConn, err := s.remoteListener.Accept()
		s.logger.Debug(s.logTag, "Received connection")
		if err != nil {
			errCh <- bosherr.WrapError(err, "Accepting connection on remote server")
		}
		defer remoteConn.Close()

		s.logger.Debug(s.logTag, "Dialing local server")
		localDialAddr := fmt.Sprintf("127.0.0.1:%d", s.options.LocalForwardPort)
		localConn, err := net.Dial("tcp", localDialAddr)
		if err != nil {
			errCh <- bosherr.WrapError(err, "Dialing local server")
			return
		}

		go func() {
			bytesNum, err := io.Copy(remoteConn, localConn)
			defer localConn.Close()
			s.logger.Debug(s.logTag, "Copying bytes from local to remote %d", bytesNum)
			if err != nil {
				errCh <- bosherr.WrapError(err, "Copying bytes from local to remote")
			}
		}()

		go func() {
			bytesNum, err := io.Copy(localConn, remoteConn)
			defer localConn.Close()
			s.logger.Debug(s.logTag, "Copying bytes from remote to local %d", bytesNum)
			if err != nil {
				errCh <- bosherr.WrapError(err, "Copying bytes from remote to local")
			}
		}()
	}
}

func (s *sshTunnel) Stop() error {
	if s.remoteListener == nil {
		return nil
	}

	return s.remoteListener.Close()
}

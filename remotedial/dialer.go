package remotedial

import (
	"io"
	"io/ioutil"
	"time"

	"golang.org/x/crypto/ssh"
)

// NewClient ...
func NewClient(user, remoteAddress string) (*ssh.Client, error) {
	key, err := ioutil.ReadFile("/etc/spotcluster/id_rsa")
	if err != nil {
		return nil, err
	}
	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		return nil, err
	}
	config := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         10 * time.Minute,
	}
	return ssh.Dial("tcp", remoteAddress, config)
}

// NewSession ...
func NewSession(in io.Reader, out, err io.Writer) *ssh.Session {
	return &ssh.Session{
		Stdin:  in,
		Stdout: out,
		Stderr: err,
	}
}

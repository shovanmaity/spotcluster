package remotedial

import (
	"io/ioutil"
	"time"

	"golang.org/x/crypto/ssh"
)

// NewSSHClient returns a ssh client for given user and remote address
func NewSSHClient(user, remoteAddress string) (*ssh.Client, error) {
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

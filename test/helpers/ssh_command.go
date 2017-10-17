// Copyright 2017 Authors of Cilium
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package helpers

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"os"
	"strconv"

	"github.com/kevinburke/ssh_config"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"

	log "github.com/sirupsen/logrus"
)

//SSHConfigPath ssh-config temp path for the different scopes
var SSHConfigPath = "ssh-config"

//SSHCommand struct to send commands over SSHClient
type SSHCommand struct {
	Path   string
	Env    []string
	Stdin  io.Reader
	Stdout io.Writer
	Stderr io.Writer
}

//SSHClient configuration
type SSHClient struct {
	Config *ssh.ClientConfig
	Host   string
	Port   int
	client *ssh.Client
}

//SSHConfig contains metadata for running an SSH session .
type SSHConfig struct {
	target       string
	host         string
	user         string
	port         int
	identityFile string
}

//SSHConfigs map with all sshconfig
type SSHConfigs map[string]*SSHConfig

//GetSSHClient initializes an SSHClient based on the provided SSHConfig
func (cfg *SSHConfig) GetSSHClient() *SSHClient {

	sshConfig := &ssh.ClientConfig{
		User: cfg.user,
		Auth: []ssh.AuthMethod{
			cfg.GetSSHAgent(),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	return &SSHClient{
		Config: sshConfig,
		Host:   cfg.host,
		Port:   cfg.port,
	}
}

//GetSSHAgent returns the ssh.AuthMethod corresponding to SSHConfig cfg
func (cfg *SSHConfig) GetSSHAgent() ssh.AuthMethod {
	key, err := ioutil.ReadFile(cfg.identityFile)
	if err != nil {
		log.Fatalf("unable to retrieve ssh-key on target '%s': %s", cfg.target, err)
	}

	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		log.Fatalf("unable to parse private key on target '%s': %s", cfg.target, err)
	}
	return ssh.PublicKeys(signer)
}

//ImportSSHconfig imports the SSH configuration stored at the provided path.
//Returns an error if the SSH configuration could not be instantiated.
func ImportSSHconfig(path string) (SSHConfigs, error) {
	result := make(SSHConfigs)
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	cfg, err := ssh_config.Decode(f)
	if err != nil {
		return nil, err
	}

	for _, host := range cfg.Hosts {
		key := host.Patterns[0].String()
		if key == "*" {
			continue
		}
		port, _ := cfg.Get(key, "Port")
		hostConfig := SSHConfig{target: key}
		hostConfig.host, _ = cfg.Get(key, "Hostname")
		hostConfig.identityFile, _ = cfg.Get(key, "identityFile")
		hostConfig.user, _ = cfg.Get(key, "User")
		hostConfig.port, _ = strconv.Atoi(port)
		result[key] = &hostConfig
	}
	return result, nil
}

//RunCommand runs a SSHCommand using SSHClient client. It will return the
//stdout and a error.
//Error will happen when a session can't be initialized correctly
func (client *SSHClient) RunCommand(cmd *SSHCommand) ([]byte, error) {
	var (
		session *ssh.Session
		err     error
	)

	if session, err = client.newSession(); err != nil {
		return nil, err
	}
	defer session.Close()

	stderr, err := session.StderrPipe()
	if err != nil {
		return nil, fmt.Errorf("Unable to setup stderr for session: %v", err)
	}
	go io.Copy(cmd.Stderr, stderr)
	return session.Output(cmd.Path)
}

//RunCommandContext run a ssh command but with a context, so can be cancel when is needed
func (client *SSHClient) RunCommandContext(ctx context.Context, cmd *SSHCommand) error {
	if ctx == nil {
		panic("nil Context")
	}

	var (
		session *ssh.Session
		err     error
	)

	if session, err = client.newSession(); err != nil {
		return err
	}

	modes := ssh.TerminalModes{
		ssh.ECHO:          1,     // enable echoing
		ssh.TTY_OP_ISPEED: 14400, // input speed = 14.4kbaud
		ssh.TTY_OP_OSPEED: 14400, // output speed = 14.4kbaud
	}
	session.RequestPty("xterm-256color", 80, 80, modes)
	defer session.Close()

	stderr, err := session.StderrPipe()
	if err != nil {
		return fmt.Errorf("Unable to setup stderr for session: %v", err)
	}
	go io.Copy(cmd.Stderr, stderr)

	stdout, err := session.StdoutPipe()
	if err != nil {
		return fmt.Errorf("Unable to setup stdout for session: %v", err)
	}
	go io.Copy(cmd.Stdout, stdout)

	go func() {
		select {
		case <-ctx.Done():
			if err := session.Signal(ssh.SIGHUP); err != nil {
				log.Errorf("failed to kill command: %s", err)
			}
			session.Close()
		}
	}()
	err = session.Run(cmd.Path)
	return nil
}

func (client *SSHClient) newSession() (*ssh.Session, error) {
	var connection *ssh.Client
	var err error

	if client.client != nil {
		connection = client.client
	} else {
		connection, err = ssh.Dial(
			"tcp",
			fmt.Sprintf("%s:%d", client.Host, client.Port),
			client.Config)

		if err != nil {
			return nil, fmt.Errorf("failed to dial: %s", err)
		}
		client.client = connection
	}

	session, err := connection.NewSession()
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %s", err)
	}

	return session, nil
}

//SSHAgent return the ssh.Authmethod using the Public keys. If can connect to
//SSH_AUTH_SHOCK it will return nil
func SSHAgent() ssh.AuthMethod {
	if sshAgent, err := net.Dial("unix", os.Getenv("SSH_AUTH_SOCK")); err == nil {
		return ssh.PublicKeysCallback(agent.NewClient(sshAgent).Signers)
	}
	return nil
}

//GetSSHclient initializes an SSHClient for the specified host/port/user
//combination.
func GetSSHclient(host string, port int, user string) *SSHClient {

	sshConfig := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			SSHAgent(),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	return &SSHClient{
		Config: sshConfig,
		Host:   host,
		Port:   port,
	}

}

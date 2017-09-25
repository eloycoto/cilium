package helpers

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"os"
	"strconv"

	"github.com/kevinburke/ssh_config"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
)

//SSHConfigPath is where the vagrant ssh-config is located.
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

//SSHConfig parser
type SSHConfig struct {
	target       string
	host         string
	user         string
	port         int
	identityFile string
}

//SSHConfigs map with all sshconfig
type SSHConfigs map[string]*SSHConfig

//GetSSHClient return the SSHClient for a SSHConfig
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

//GetSSHAgent return the sshAuthmethod for a SSHConfig
func (cfg *SSHConfig) GetSSHAgent() ssh.AuthMethod {
	key, err := ioutil.ReadFile(cfg.identityFile)
	if err != nil {
		log.Fatalf("Unable to retrieve ssh-key: %s", err)
	}

	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		log.Fatalf("unable to parse private key: %s", err)
	}
	return ssh.PublicKeys(signer)
}

//ImportSSHconfig import path and create all SSHConfigs needed
func ImportSSHconfig(path string) (SSHConfigs, error) {
	// var result SSHConfigs
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

//RunCommand run SSHCommand over ssh
func (client *SSHClient) RunCommand(cmd *SSHCommand) ([]byte, error) {
	var (
		session *ssh.Session
		err     error
	)

	if session, err = client.newSession(); err != nil {
		return nil, err
	}
	defer session.Close()

	return session.Output(cmd.Path)
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
			return nil, fmt.Errorf("Failed to dial: %s", err)
		}
		client.client = connection
	}

	session, err := connection.NewSession()
	if err != nil {
		return nil, fmt.Errorf("Failed to create session: %s", err)
	}

	return session, nil
}

//SSHAgent return the ssh.Authmethod
func SSHAgent() ssh.AuthMethod {
	if sshAgent, err := net.Dial("unix", os.Getenv("SSH_AUTH_SOCK")); err == nil {
		return ssh.PublicKeysCallback(agent.NewClient(sshAgent).Signers)
	}
	return nil
}

//GetSSHclient return a SSHClient for a specific host/port
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

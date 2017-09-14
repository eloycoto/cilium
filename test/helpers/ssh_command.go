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

var SSHConfigPath = "ssh-config"

type SSHCommand struct {
	Path   string
	Env    []string
	Stdin  io.Reader
	Stdout io.Writer
	Stderr io.Writer
}

type SSHClient struct {
	Config *ssh.ClientConfig
	Host   string
	Port   int
}

type SSHConfig struct {
	target       string
	host         string
	user         string
	port         int
	identityFile string
}

type SSHConfigs map[string]*SSHConfig

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

// func (client *SSHClient) prepareCommand(session *ssh.Session, cmd *SSHCommand) error {
// 	for _, env := range cmd.Env {
// 		variable := strings.Split(env, "=")
// 		if len(variable) != 2 {
// 			continue
// 		}

// 		if err := session.Setenv(variable[0], variable[1]); err != nil {
// 			return err
// 		}
// 	}

// 	if cmd.Stdin != nil {
// 		stdin, err := session.StdinPipe()
// 		if err != nil {
// 			return fmt.Errorf("Unable to setup stdin for session: %v", err)
// 		}
// 		go io.Copy(stdin, cmd.Stdin)
// 	}

// 	if cmd.Stdout != nil {
// 		stdout, err := session.StdoutPipe()
// 		if err != nil {
// 			return fmt.Errorf("Unable to setup stdout for session: %v", err)
// 		}
// 		go io.Copy(cmd.Stdout, stdout)
// 	}

// 	if cmd.Stderr != nil {
// 		stderr, err := session.StderrPipe()
// 		if err != nil {
// 			return fmt.Errorf("Unable to setup stderr for session: %v", err)
// 		}
// 		go io.Copy(cmd.Stderr, stderr)
// 	}

// 	return nil
// }

func (client *SSHClient) newSession() (*ssh.Session, error) {
	connection, err := ssh.Dial(
		"tcp",
		fmt.Sprintf("%s:%d", client.Host, client.Port),
		client.Config)
	if err != nil {
		return nil, fmt.Errorf("Failed to dial: %s", err)
	}

	session, err := connection.NewSession()
	if err != nil {
		return nil, fmt.Errorf("Failed to create session: %s", err)
	}

	// modes := ssh.TerminalModes{
	// 	// ssh.ECHO:          0,     // disable echoing
	// 	ssh.TTY_OP_ISPEED: 14400, // input speed = 14.4kbaud
	// 	ssh.TTY_OP_OSPEED: 14400, // output speed = 14.4kbaud
	// }

	// if err := session.RequestPty("xterm", 80, 40, modes); err != nil {
	// 	session.Close()
	// 	return nil, fmt.Errorf("request for pseudo terminal failed: %s", err)
	// }
	// err = session.Shell()
	// if err != nil {
	// 	return nil, fmt.Errorf("request for pseudo terminal failed: %s", err)
	// }
	return session, nil
}

// func PublicKeyFile(file string) ssh.AuthMethod {
// 	buffer, err := ioutil.ReadFile(file)
// 	if err != nil {
// 		return nil
// 	}

// 	key, err := ssh.ParsePrivateKey(buffer)
// 	if err != nil {
// 		return nil
// 	}
// 	return ssh.PublicKeys(key)
// }

func SSHAgent() ssh.AuthMethod {
	if sshAgent, err := net.Dial("unix", os.Getenv("SSH_AUTH_SOCK")); err == nil {
		return ssh.PublicKeysCallback(agent.NewClient(sshAgent).Signers)
	}
	return nil
}

func GetSSHClient(host string, port int, user string) *SSHClient {

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

// func Connect(ip string, port int, user string, command string) {
// 	sshConfig := &ssh.ClientConfig{
// 		User: user,
// 		Auth: []ssh.AuthMethod{
// 			SSHAgent(),
// 		},
// 		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
// 	}

// 	client := &SSHClient{
// 		Config: sshConfig,
// 		Host:   ip,
// 		Port:   port,
// 	}

// 	cmd := &SSHCommand{
// 		Path: command,
// 		// Env:    []string{"LC_DIR=/"},
// 		Stdin:  os.Stdin,
// 		Stdout: os.Stdout,
// 		Stderr: os.Stderr,
// 	}

// 	fmt.Printf("Running command: %s\n", cmd.Path)
// 	if err := client.RunCommand(cmd); err != nil {
// 		fmt.Fprintf(os.Stderr, "command run error: %s\n", err)
// 		os.Exit(1)
// 	}
// }

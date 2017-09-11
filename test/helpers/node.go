package helpers

import (
	"fmt"
	"io"
	"os"
)

type Node struct {
	sshClient *SSHClient
	host      string
	port      int
	env       []string
}

func CreateNode(host string, port int) *Node {
	return &Node{
		host:      host,
		port:      port,
		sshClient: GetSSHClient(host, port, "ubuntu"),
	}
}

func (node *Node) Execute(cmd string, stdout io.Writer, stderr io.Writer) bool {
	if stdout == nil {
		stdout = os.Stdout
	}

	if stderr == nil {
		stderr = os.Stderr
	}

	command := &SSHCommand{
		Path:   cmd,
		Env:    node.env,
		Stdin:  os.Stdin,
		Stdout: stdout,
		Stderr: stderr,
	}
	if err := node.sshClient.RunCommand(command); err != nil {
		fmt.Fprintf(os.Stderr, "command run error '%s': %s\n", command.Path, err)
		return false
	}
	return true
}

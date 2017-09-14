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

func CreateNode(host string, port int, user string) *Node {
	return &Node{
		host:      host,
		port:      port,
		sshClient: GetSSHClient(host, port, user),
	}
}

func CreateNodeFromTarget(target string) *Node {
	nodes, err := ImportSSHconfig(SSHConfigPath)
	if err != nil {
		return nil
	}

	node := nodes[target]
	if node == nil {
		return nil
	}

	return &Node{
		host:      node.host,
		port:      node.port,
		sshClient: node.GetSSHClient(),
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
		Path: cmd,
		// Env:    node.env,
		Stdin:  os.Stdin,
		Stdout: stdout,
		Stderr: stderr,
	}
	result, err := node.sshClient.RunCommand(command)
	if err != nil {
		fmt.Fprintf(os.Stderr, "command run error '%s': %s\n", command.Path, err)
		return false
	}
	stdout.Write(result)
	return true
}

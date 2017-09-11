package helpers

import (
	"bytes"
	"fmt"
	"io"
	"os"
)

//Node struct to have the info for each vagrant box
type Node struct {
	sshClient *SSHClient
	host      string
	port      int
	env       []string
}

//CreateNode return a Node
func CreateNode(host string, port int, user string) *Node {
	return &Node{
		host:      host,
		port:      port,
		sshClient: GetSSHclient(host, port, user),
	}
}

//CreateNodeFromTarget create node from a ssh-config target
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

//Execute execute a command in the node
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
	stdout.Write(result)
	if err != nil {
		fmt.Fprintf(os.Stderr, "command run error '%s': %s\n", command.Path, err)
		return false
	}
	return true
}

//ExecWithSudo execute a new command  using sudo
func (node *Node) ExecWithSudo(cmd string, stdout io.Writer, stderr io.Writer) bool {
	command := fmt.Sprintf("sudo %s", cmd)
	return node.Execute(command, stdout, stderr)
}

//Exec a function and return a cmdRes command
func (node *Node) Exec(cmd string) *CmdRes {
	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)

	exit := node.Execute(cmd, stdout, stderr)

	return &CmdRes{
		cmd:    cmd,
		stdout: stdout,
		stderr: stderr,
		exit:   exit,
	}
}

//ExecWithTimeout a function and return data to the out chan CmdRes
func (node *Node) ExecWithTimeout(cmd string, out chan *CmdRes) {
	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)

	command := &SSHCommand{
		Path:   cmd,
		Stdin:  os.Stdin,
		Stdout: stdout,
		Stderr: stderr,
	}
	go func(sshCmd *SSHCommand) {
		err := node.sshClient.RunCommandWithTimeout(sshCmd, 5)
		if err != nil {
			fmt.Fprintf(os.Stderr, "command run error '%s': %s\n", command.Path, err)
			return
		}
		out <- &CmdRes{
			cmd:    sshCmd.Path,
			stdout: stdout,
			stderr: stderr,
			exit:   true,
		}
	}(command)
}

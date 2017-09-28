package helpers

import (
	"bytes"
	"fmt"

	log "github.com/sirupsen/logrus"
)

//Docker kubectl command helper
type Docker struct {
	Node *Node

	logCxt *log.Entry
}

//CreateDocker return a new Docker instance based on target
func CreateDocker(target string, log *log.Entry) *Docker {
	log.Infof("Docker: set target to '%s'", target)
	node := CreateNodeFromTarget(target)
	if node == nil {
		return nil
	}

	return &Docker{
		Node:   node,
		logCxt: log,
	}
}

//ContainerExec: execute a command in a container
func (do *Docker) ContainerExec(name string, cmd string) *cmdRes {
	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)

	command := fmt.Sprintf("docker exec -i %s %s", name, cmd)
	exit := do.Node.ExecWithSudo(command, stdout, stderr)
	return &cmdRes{
		cmd:    command,
		stdout: stdout,
		stderr: stderr,
		exit:   exit,
	}
}

func (do *Docker) ContainerCreate(name, image, net, options string) *cmdRes {

	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)

	cmd := fmt.Sprintf(
		"docker run -d --name %s --net %s %s %s", name, net, options, image)
	exit := do.Node.ExecWithSudo(cmd, stdout, stderr)

	return &cmdRes{
		cmd:    cmd,
		stdout: stdout,
		stderr: stderr,
		exit:   exit,
	}
}

func (do *Docker) ContainerRm(name string) *cmdRes {

	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)

	cmd := fmt.Sprintf("docker rm -f %s", name)
	exit := do.Node.ExecWithSudo(cmd, stdout, stderr)

	return &cmdRes{
		cmd:    cmd,
		stdout: stdout,
		stderr: stderr,
		exit:   exit,
	}
}

//ContainerInspect: Inspect a container
func (do *Docker) ContainerInspect(name string) *cmdRes {
	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)

	cmd := fmt.Sprintf("docker inspect %s", name)
	exit := do.Node.ExecWithSudo(cmd, stdout, stderr)

	return &cmdRes{
		cmd:    cmd,
		stdout: stdout,
		stderr: stderr,
		exit:   exit,
	}
}

//NetworkCreate create a new docker network
func (do *Docker) NetworkCreate(name string, subnet string) *cmdRes {
	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)
	if subnet == "" {
		subnet = "::1/112"
	}
	cmd := fmt.Sprintf(
		"docker network create --ipv6 --subnet %s --driver cilium --ipam-driver cilium %s",
		subnet, name)
	exit := do.Node.ExecWithSudo(cmd, stdout, stderr)

	return &cmdRes{
		cmd:    cmd,
		stdout: stdout,
		stderr: stderr,
		exit:   exit,
	}
}

//NetworkCreate create a new docker network
func (do *Docker) NetworkDelete(name string) *cmdRes {
	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)
	cmd := fmt.Sprintf("docker network rm  %s", name)
	exit := do.Node.ExecWithSudo(cmd, stdout, stderr)
	return &cmdRes{
		cmd:    cmd,
		stdout: stdout,
		stderr: stderr,
		exit:   exit,
	}
}

//NetworkGet: return all the docker network information
func (do *Docker) NetworkGet(name string) *cmdRes {
	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)
	cmd := fmt.Sprintf("docker network inspect %s", name)
	exit := do.Node.ExecWithSudo(cmd, stdout, stderr)

	return &cmdRes{
		cmd:    cmd,
		stdout: stdout,
		stderr: stderr,
		exit:   exit,
	}
}

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

//ContainerExec execute a command in a container
func (do *Docker) ContainerExec(name string, cmd string) *CmdRes {
	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)

	command := fmt.Sprintf("docker exec -i %s %s", name, cmd)
	exit := do.Node.ExecWithSudo(command, stdout, stderr)
	return &CmdRes{
		cmd:    command,
		stdout: stdout,
		stderr: stderr,
		exit:   exit,
	}
}

//ContainerCreate create a container on docker
func (do *Docker) ContainerCreate(name, image, net, options string) *CmdRes {

	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)

	cmd := fmt.Sprintf(
		"docker run -d --name %s --net %s %s %s", name, net, options, image)
	exit := do.Node.ExecWithSudo(cmd, stdout, stderr)

	return &CmdRes{
		cmd:    cmd,
		stdout: stdout,
		stderr: stderr,
		exit:   exit,
	}
}

//ContainerRm force a deletigion of a container based on a name
func (do *Docker) ContainerRm(name string) *CmdRes {

	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)

	cmd := fmt.Sprintf("docker rm -f %s", name)
	exit := do.Node.ExecWithSudo(cmd, stdout, stderr)

	return &CmdRes{
		cmd:    cmd,
		stdout: stdout,
		stderr: stderr,
		exit:   exit,
	}
}

//ContainerInspect Inspect a container
func (do *Docker) ContainerInspect(name string) *CmdRes {
	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)

	cmd := fmt.Sprintf("docker inspect %s", name)
	exit := do.Node.ExecWithSudo(cmd, stdout, stderr)

	return &CmdRes{
		cmd:    cmd,
		stdout: stdout,
		stderr: stderr,
		exit:   exit,
	}
}

//ContainerInspectNet return the Net details for a container
func (do *Docker) ContainerInspectNet(name string) (map[string]string, error) {
	res := do.ContainerInspect(name)
	properties := map[string]string{
		"EndpointID":        "EndpointID",
		"GlobalIPv6Address": "IPv6",
		"IPAddress":         "IPv4",
		"NetworkID":         "NetworkID",
	}

	if !res.Correct() {
		return nil, fmt.Errorf("Can't get the container")
	}
	filter := `{ [0].NetworkSettings.Networks.cilium-net }`
	result := make(map[string]string)
	data, err := res.FindResults(filter)
	if err != nil {
		return nil, err
	}
	for _, val := range data {
		iface := val.Interface()
		for k, v := range iface.(map[string]interface{}) {
			if key, ok := properties[k]; ok {
				result[key] = fmt.Sprintf("%s", v)
			}
		}
	}
	return result, nil
}

//NetworkCreate create a new docker network
func (do *Docker) NetworkCreate(name string, subnet string) *CmdRes {
	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)
	if subnet == "" {
		subnet = "::1/112"
	}
	cmd := fmt.Sprintf(
		"docker network create --ipv6 --subnet %s --driver cilium --ipam-driver cilium %s",
		subnet, name)
	exit := do.Node.ExecWithSudo(cmd, stdout, stderr)

	return &CmdRes{
		cmd:    cmd,
		stdout: stdout,
		stderr: stderr,
		exit:   exit,
	}
}

//NetworkDelete create a new docker network
func (do *Docker) NetworkDelete(name string) *CmdRes {
	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)
	cmd := fmt.Sprintf("docker network rm  %s", name)
	exit := do.Node.ExecWithSudo(cmd, stdout, stderr)
	return &CmdRes{
		cmd:    cmd,
		stdout: stdout,
		stderr: stderr,
		exit:   exit,
	}
}

//NetworkGet return all the docker network information
func (do *Docker) NetworkGet(name string) *CmdRes {
	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)
	cmd := fmt.Sprintf("docker network inspect %s", name)
	exit := do.Node.ExecWithSudo(cmd, stdout, stderr)

	return &CmdRes{
		cmd:    cmd,
		stdout: stdout,
		stderr: stderr,
		exit:   exit,
	}
}

//SampleContainersActions This create/delete a bunch of containers to test
func (do *Docker) SampleContainersActions(mode string, networkName string) {
	images := map[string]string{
		"httpd1": "cilium/demo-httpd",
		"httpd2": "cilium/demo-httpd",
		"httpd3": "cilium/demo-httpd",
		"app1":   "tgraf/netperf",
		"app2":   "tgraf/netperf",
		"app3":   "tgraf/netperf",
	}

	switch mode {
	case "create":
		for k, v := range images {
			do.ContainerCreate(k, v, networkName, fmt.Sprintf("-l id.%s", k))
		}
	case "delete":
		for k := range images {
			do.ContainerRm(k)
		}
	}
}

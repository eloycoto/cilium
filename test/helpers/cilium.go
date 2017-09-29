package helpers

import (
	"bytes"
	"fmt"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

//Docker kubectl command helper
type Cilium struct {
	Node *Node

	logCxt *log.Entry
}

func CreateCilium(target string, log *log.Entry) *Cilium {
	log.Infof("Cilium: set target to '%s'", target)
	node := CreateNodeFromTarget(target)
	if node == nil {
		return nil
	}

	return &Cilium{
		Node:   node,
		logCxt: log,
	}
}

func (c *Cilium) Exec(cmd string) *cmdRes {
	// c.Node.Execute(fmt.Sprintf("cilium %s", name))

	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)
	command := fmt.Sprintf("cilium %s", cmd)
	exit := c.Node.ExecWithSudo(command, stdout, stderr)
	return &cmdRes{
		cmd:    command,
		stdout: stdout,
		stderr: stderr,
		exit:   exit,
	}
}

//EndPointSetConfig: set to a container endpoint a new config
func (c *Cilium) EndpointSetConfig(container, option, value string) bool {
	data := c.Exec(fmt.Sprintf("endpoint config %s %s=%s", container, option, value))
	fmt.Printf("%v", data)
	return data.Correct()
}

func (c *Cilium) GetEndpoints() *cmdRes {
	return c.Exec("endpoint list -o json")
}

func (c *Cilium) GetEndpointsIds() (map[string]string, error) {
	// cilium endpoint list -o jsonpath='{range [*]}{@.container-name}{"="}{@.id}{"\n"}{end}'
	filter := `{range [*]}{@.container-name}{"="}{@.id}{"\n"}{end}`
	endpoints := c.Exec(fmt.Sprintf("endpoint list -o jsonpath='%s'", filter))
	if !endpoints.Correct() {
		return nil, fmt.Errorf("Can't get endpoint list")
	}
	return endpoints.KVOutput(), nil
}

func (c *Cilium) GetEndpointsNames() ([]string, error) {
	data := c.GetEndpoints()
	if data.Correct() == false {
		return nil, fmt.Errorf("Could't get endpoints")
	}
	result, err := data.Filter("{ [*].container-name }")
	if err != nil {
		return nil, err
	}

	return strings.Split(result.String(), " "), nil
}

func (c *Cilium) GetPolicyRevision() (int, error) {
	rev := c.Exec("policy get | grep Revision| awk '{print $2}'")
	return rev.IntOutput()
}

func (c *Cilium) ImportPolicy(path string, timeout int) (int, error) {
	var wait int
	revision, err := c.GetPolicyRevision()
	if err != nil {
		return -1, fmt.Errorf("Can't get policy revision: %s", err)
	}

	fmt.Printf("\n---->policy import %s \n", path)

	res := c.Exec(fmt.Sprintf("policy import %s", path))
	if res.Correct() == false {
		return -1, fmt.Errorf("Can't import policy %s", path)
	}

	for wait < timeout {
		current_rev, _ := c.GetPolicyRevision()
		if current_rev > revision {
			return current_rev, nil
		}
		time.Sleep(1 * time.Second)
		wait++
	}
	return -1, fmt.Errorf("Can't import Policy revision %s", path)
}

func (c *Cilium) ManifestsPath() string {
	return fmt.Sprintf("%s/runtime/manifests/", basePath)
}

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

//Exec: run a cilium command and return a cmdRes
func (c *Cilium) Exec(cmd string) *cmdRes {
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

//EndpointWaitUntilReady: This function wait until all the endpoints are in the ready status
func (c *Cilium) EndpointWaitUntilReady() bool {
	for {
		status, err := c.GetEndpoints().Filter("{[*].state}")
		if err != nil {
			Sleep(1)
			continue
		}
		var valid, invalid int
		for _, endpoint := range strings.Split(status.String(), " ") {
			if endpoint != "ready" {
				invalid++
			} else {
				valid++
			}
		}
		if invalid == 0 {
			return true
		}
		c.logCxt.Infof("Endpoints are not ready valid='%d' invalid='%d'", valid, invalid)
		Sleep(1)
	}
	return false
}

//GetEndpoints: Return the endpoints in jsonFormat
func (c *Cilium) GetEndpoints() *cmdRes {
	return c.Exec("endpoint list -o json")
}

//GetEndpointsIds: return a map with with docker container name and the endpoint id
func (c *Cilium) GetEndpointsIds() (map[string]string, error) {
	// cilium endpoint list -o jsonpath='{range [*]}{@.container-name}{"="}{@.id}{"\n"}{end}'
	filter := `{range [*]}{@.container-name}{"="}{@.id}{"\n"}{end}`
	endpoints := c.Exec(fmt.Sprintf("endpoint list -o jsonpath='%s'", filter))
	if !endpoints.Correct() {
		return nil, fmt.Errorf("Can't get endpoint list")
	}
	return endpoints.KVOutput(), nil
}

//GetEndpointsNames: Return the list of containers from cilium endpoint
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

func (c *Cilium) ManifestsPath() string {
	return fmt.Sprintf("%s/runtime/manifests/", basePath)
}

func (c *Cilium) GetFullPath(name string) string {
	return fmt.Sprintf("%s%s", c.ManifestsPath(), name)
}

//PolicyEndpointsSummary: Return a map of the status of the policies
func (c *Cilium) PolicyEndpointsSummary() (map[string]int, error) {
	var result map[string]int = map[string]int{
		"enabled":  0,
		"disabled": 0,
		"total":    0,
	}
	endpoints, err := c.GetEndpoints().Filter("{ [*].policy-enabled }")
	if err != nil {
		return result, fmt.Errorf("Can't get the endpoints")
	}
	status := strings.Split(endpoints.String(), " ")
	result["enabled"], result["total"] = CountValues("true", status)
	result["disabled"], result["total"] = CountValues("false", status)
	return result, nil
}

func (c *Cilium) PolicyEnforcementSet(status string, waitReady ...bool) *cmdRes {
	res := c.Exec(fmt.Sprintf("config PolicyEnforcement=%s", status))
	if len(waitReady) > 0 && waitReady[0] {
		c.EndpointWaitUntilReady()
	}
	return res
}

//PolicyGetRevision: Get the current Policy revision
func (c *Cilium) PolicyGetRevision() (int, error) {
	rev := c.Exec("policy get | grep Revision| awk '{print $2}'")
	return rev.IntOutput()
}

//PolicyImport: Import a new policy in cilium and wait until all endpoints
// get the policy apply.
func (c *Cilium) PolicyImport(path string, timeout int) (int, error) {
	var wait int
	revision, err := c.PolicyGetRevision()
	if err != nil {
		return -1, fmt.Errorf("Can't get policy revision: %s", err)
	}

	res := c.Exec(fmt.Sprintf("policy import %s", path))
	if res.Correct() == false {
		return -1, fmt.Errorf("Can't import policy %s", path)
	}

	for wait < timeout {
		current_rev, _ := c.PolicyGetRevision()
		if current_rev > revision {
			c.PolicyWait(current_rev)
			return current_rev, nil
		}
		time.Sleep(1 * time.Second)
		wait++
	}
	return -1, fmt.Errorf("Can't import Policy revision %s", path)
}

func (c *Cilium) PolicyWait(id int) *cmdRes {
	return c.Exec(fmt.Sprintf("policy wait %d", id))
}

//ServiceAdd: Create a new cilium service
func (c *Cilium) ServiceAdd(id int, frontend string, backends []string, rev int) *cmdRes {
	cmd := fmt.Sprintf(
		"service update --frontend '%s' --backends '%s' --id '%d' --rev '%d'",
		frontend, strings.Join(backends, ","), id, rev)
	return c.Exec(cmd)
}

//ServiceGet: Get a service from cilium
func (c *Cilium) ServiceGet(id int) *cmdRes {
	return c.Exec(fmt.Sprintf("service get '%d'", id))
}

//ServiceDel: delete the service ID
func (c *Cilium) ServiceDel(id int) *cmdRes {
	return c.Exec(fmt.Sprintf("service delete '%d'", id))
}

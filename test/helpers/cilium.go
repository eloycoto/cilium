package helpers

import (
	"bytes"
	"fmt"
	"strings"
	"time"

	"github.com/cilium/cilium/api/v1/models"
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

func (c *Cilium) EndpointsGet(id string) *models.Endpoint {

	var data []models.Endpoint
	err := c.Exec(fmt.Sprintf("endpoint get %s", id)).UnMarshal(&data)
	if err != nil {
		c.logCxt.Infof("EndpointsGet fail %d: %s", id, err)
		return nil
	}
	if len(data) > 0 {
		return &data[0]
	}
	return nil
}

//EndPointSetConfig: set to a container endpoint a new config
func (c *Cilium) EndpointSetConfig(container, option, value string) bool {
	// on grep we use an space, so we are sure that only match the key that we want.
	res := c.Exec(fmt.Sprintf(
		"endpoint config %s | grep '%s ' | awk '{print $2}'", container, option))
	if res.SingleOut() == value {
		return res.Correct()
	}

	before := c.EndpointsGet(container)
	if before == nil {
		return false
	}
	data := c.Exec(fmt.Sprintf("endpoint config %s %s=%s", container, option, value))
	if !data.Correct() {
		c.logCxt.Infof("Can't set endoint '%d' config %s=%s", container, option, value)
		return false
	}
	err := WithTimeout(func() bool {
		status := c.EndpointsGet(container)
		if len(status.Status) > len(before.Status) {
			return true
		}
		c.logCxt.Infof("Endpoint '%s' is not regenerated", container)
		return false
	}, "Endpoint is not regenerated", &TimeoutConfig{Timeout: 100})
	if err != nil {
		c.logCxt.Infof("Endpoint set failed:%s", err)
		return false
	}
	return true
}

//EndpointWaitUntilReady: This function wait until all the endpoints are in the ready status
func (c *Cilium) EndpointWaitUntilReady(validation ...bool) bool {

	logger := c.logCxt.WithFields(log.Fields{"EndpointWaitReady": ""})

	getEpsStatus := func(data []models.Endpoint) map[int64]int {
		result := make(map[int64]int)
		for _, v := range data {
			result[v.ID] = len(v.Status)
		}
		return result
	}

	var data []models.Endpoint

	if err := c.GetEndpoints().UnMarshal(&data); err != nil {
		logger.Info("Can't get original endpoints: %s", err)
		Sleep(5)
		return c.EndpointWaitUntilReady(validation...)
	}
	epsStatus := getEpsStatus(data)

	body := func() bool {
		var data []models.Endpoint

		if err := c.GetEndpoints().UnMarshal(&data); err != nil {
			logger.Info("Can't get endpoints: %s", err)
			return false
		}
		var valid, invalid int
		for _, eps := range data {
			if eps.State != "ready" {
				invalid++
			} else {
				valid++
			}
			if len(validation) > 0 && validation[0] {
				if originalVal, _ := epsStatus[eps.ID]; len(eps.Status) <= originalVal {
					logger.Infof("Endpoint '%d' is not regenerated", eps.ID)
					return false
				}
			}
		}

		if invalid == 0 {
			return true
		}
		logger.Infof("Endpoints are not ready valid='%d' invalid='%d'", valid, invalid)
		return false
	}
	err := WithTimeout(body, "Endpoints are not ready", &TimeoutConfig{Timeout: 300})
	if err != nil {
		return false
	}
	return true
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
	//We check before, if not EndpointWait will be fail due no changes on eps.Status
	res := c.Exec("config | grep PolicyEnforcement | awk '{print $2}'")
	if res.SingleOut() == status {
		return res
	}
	res = c.Exec(fmt.Sprintf("config PolicyEnforcement=%s", status))
	if len(waitReady) > 0 && waitReady[0] {
		c.EndpointWaitUntilReady(true)
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
func (c *Cilium) PolicyImport(path string, timeout time.Duration) (int, error) {
	revision, err := c.PolicyGetRevision()
	if err != nil {
		return -1, fmt.Errorf("Can't get policy revision: %s", err)
	}

	res := c.Exec(fmt.Sprintf("policy import %s", path))
	if res.Correct() == false {
		c.logCxt.Errorf("Couldn't import policy %s", res.Output())
		return -1, fmt.Errorf("Couldn't import policy %s", path)
	}
	body := func() bool {
		current_rev, _ := c.PolicyGetRevision()
		if current_rev > revision {
			c.PolicyWait(current_rev)
			return true
		}
		c.logCxt.Infof("PolicyImport: current revision %d same as %d", current_rev, revision)
		return false
	}
	err = WithTimeout(body, "Couldn't import policy revision", &TimeoutConfig{Timeout: timeout})
	if err != nil {
		return -1, err
	}
	revision, err = c.PolicyGetRevision()
	return revision, err
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

func (c *Cilium) WaitUntilReady(timeout time.Duration) error {

	body := func() bool {
		res := c.Node.Exec("sudo cilium status")
		c.logCxt.Infof("Cilium status is %t", res.Correct())
		return res.Correct()
	}
	err := WithTimeout(body, "Cilium is not ready", &TimeoutConfig{Timeout: timeout})
	return err
}

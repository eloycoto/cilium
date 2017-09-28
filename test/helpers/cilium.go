package helpers

import (
	"bytes"
	"fmt"
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

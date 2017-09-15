package helpers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/asaskevich/govalidator"
	"k8s.io/client-go/util/jsonpath"
)

type Kubectl struct {
	Node *Node

	logCxt *log.Entry
}

type KubectlRes struct {
	cmd    string
	params []string
	stdout *bytes.Buffer
	stderr *bytes.Buffer
	exit   bool
}

func (res *KubectlRes) Filter(filter string) (*bytes.Buffer, error) {
	var data interface{}
	result := new(bytes.Buffer)
	err := json.Unmarshal(res.stdout.Bytes(), &data)
	if err != nil {
		return nil, fmt.Errorf("Coundn't parse json")
	}
	parser := jsonpath.New("").AllowMissingKeys(true)
	parser.Parse(filter)
	err = parser.Execute(result, data)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (res *KubectlRes) Output() *bytes.Buffer {
	return res.stdout
}

var timeout = 300 * time.Second

func CreateKubectl(target string, log *log.Entry) *Kubectl {
	node := CreateNodeFromTarget(target)
	if node == nil {
		return nil
	}

	return &Kubectl{
		Node:   node,
		logCxt: log,
	}
}

// func (kubectl *Kubectl) Execute(cmd string, stdout io.Writer, stderr io.Writer) bool {
// 	return kubectl.Node.Execute(cmd, stdout, stderr)
// }

func (kubectl *Kubectl) GetPods(namespace string, filter string) *KubectlRes {
	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)

	exit := kubectl.Node.Execute(
		fmt.Sprintf("kubectl -n %s get pods %s -o json", namespace, filter),
		stdout, stderr)
	return &KubectlRes{
		cmd:    "",
		stdout: stdout,
		stderr: stderr,
		exit:   exit,
	}
}

func (kubectl *Kubectl) GetCiliumPodOnNode(namespace string, node string) (string, error) {

	stdout := new(bytes.Buffer)
	filter := fmt.Sprintf(
		"-o jsonpath='{.items[?(@.spec.nodeName == \"%s\")].metadata.name}'", node)
	exit := kubectl.Node.Execute(
		fmt.Sprintf("kubectl -n %s get pods -l k8s-app=cilium %s", namespace, filter),
		stdout, nil)
	if exit == false {
		return "", fmt.Errorf("Cilium pod not found on node '%s'", node)
	}
	return stdout.String(), nil
}

func (kubectl *Kubectl) GetCiliumPods(namespace string) ([]string, error) {

	stdout := new(bytes.Buffer)
	filter := "-o jsonpath='{.items[*].metadata.name}'"
	exit := kubectl.Node.Execute(
		fmt.Sprintf("kubectl -n %s get pods -l k8s-app=cilium %s", namespace, filter),
		stdout, nil)
	if exit == false {
		return nil, fmt.Errorf("Cilium pods not found on namespace '%s'", namespace)
	}
	result := strings.Split(strings.Trim(stdout.String(), "\n"), " ")
	return result, nil
}

func (kubectl *Kubectl) WaitforPods(namespace string, filter string, timeout int) (bool, error) {
	wait := 0
	var jsonPath = "{.items[*].status.containerStatuses[*].ready}"
	for wait < timeout {
		data, err := kubectl.GetPods(namespace, filter).Filter(jsonPath)
		if err == nil {
			valid := true
			result := strings.Split(data.String(), " ")
			for _, v := range result {
				if val, _ := govalidator.ToBoolean(v); val == false {
					valid = false
					break
				}
			}
			if valid == true {
				return true, nil
			}
		}
		time.Sleep(1)
		wait++
	}

	return false, nil
}

func (kubectl *Kubectl) CiliumExec(pod string, cmd string) (string, error) {
	command := fmt.Sprintf("kubectl exec -n kube-system %s -- %s", pod, cmd)
	stdout := new(bytes.Buffer)

	exit := kubectl.Node.Execute(command, stdout, nil)
	if exit == false {
		// FIXME: Output here is important.
		// Return the string is not fired on the assertion :\ Need to check
		return "", fmt.Errorf("CiliumExec: command '%s' failed", command)
	}
	return stdout.String(), nil
}

func (kubectl *Kubectl) CiliumImportPolicy(namespace string, filepath string, timeout int) (string, error) {
	var revision int
	var wait int = 0
	pods, err := kubectl.GetCiliumPods(namespace)
	if err != nil {
		return "", err
	}

	//FIXME use channels here?
	for _, v := range pods {
		rev, err := kubectl.CiliumExec(v, "cilium policy get | grep Revision | awk '{print $2}'")
		if err != nil {
			return "", err
		}
		//FIXME: Log here with the pod rev value
		revi, err := strconv.Atoi(strings.Trim(rev, "\n"))
		if err != nil {
			return "", err
		}

		if revi > revision {
			revision = revi
		}
	}

	if kubectl.Apply(filepath) == false {
		return "", fmt.Errorf("Can't apply the policy '%s'", filepath)
	}

	for wait < timeout {
		valid := true
		for _, v := range pods {
			rev, err := kubectl.CiliumExec(v, "cilium policy get | grep Revision | awk '{print $2}'")
			if err != nil {
				return "", err
			}
			//FIXME: Log here with the pod rev value
			revi, err := strconv.Atoi(strings.Trim(rev, "\n"))
			if err != nil {
				return "", err
			}
			if revi <= revision {
				valid = false
			}
		}
		if valid == true {
			//FIXME: Check if something need to be return here
			return "", nil
		}
		time.Sleep(1 * time.Second)
		wait++
	}

	return "", fmt.Errorf("ImportPolicy error due timeout '%d'", timeout)
}

func (kubectl *Kubectl) Apply(filepath string) bool {
	return kubectl.Node.Execute(
		fmt.Sprintf("kubectl apply -f  %s", filepath), nil, nil)
}

func (kubectl *Kubectl) Delete(filepath string) bool {
	return kubectl.Node.Execute(
		fmt.Sprintf("kubectl delete -f  %s", filepath), nil, nil)
}

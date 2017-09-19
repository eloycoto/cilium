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

var timeout = 300 * time.Second
var basePath = "/vagrant/"

//Kubectl kubectl command helper
type Kubectl struct {
	Node *Node

	logCxt *log.Entry
}

//KubectlRes Kubectl command response
type KubectlRes struct {
	cmd    string
	params []string
	stdout *bytes.Buffer
	stderr *bytes.Buffer
	exit   bool
}

//Filter Filter json using jsonpath
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

//Output Return the command Output
func (res *KubectlRes) Output() *bytes.Buffer {
	return res.stdout
}

//CreateKubectl  Create a new Kubectl helper with a proper log
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

func (kubectl *Kubectl) Exec(namespace string, pod string, cmd string) (string, error) {
	command := fmt.Sprintf("kubectl exec -n %s %s -- %s", namespace, pod, cmd)
	stdout := new(bytes.Buffer)

	exit := kubectl.Node.Execute(command, stdout, nil)
	if exit == false {
		// FIXME: Output here is important.
		// Return the string is not fired on the assertion :\ Need to check
		kubectl.logCxt.Infof(
			"Exec command failed '%s' pod='%s' erro='%s'",
			cmd, pod, stdout.String())
		return "", fmt.Errorf("Exec: command '%s' failed '%s'", command, stdout.String())
	}
	return stdout.String(), nil
}

func (kubectl *Kubectl) Get(namespace string, command string) *KubectlRes {
	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)

	exit := kubectl.Node.Execute(
		fmt.Sprintf("kubectl -n %s get %s -o json", namespace, command),
		stdout, stderr)

	return &KubectlRes{
		cmd:    "",
		stdout: stdout,
		stderr: stderr,
		exit:   exit,
	}
}

//GetPods return all the pods for a namespace. Kubectl filter can be passed
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

//GetPodsNames return a name of all the pods for a filter
func (kubectl *Kubectl) GetPodsNames(namespace string, label string) ([]string, error) {
	stdout := new(bytes.Buffer)
	filter := "-o jsonpath='{.items[*].metadata.name}'"
	exit := kubectl.Node.Execute(
		fmt.Sprintf("kubectl -n %s get pods -l %s %s", namespace, label, filter),
		stdout, nil)

	if exit == false {
		return nil, fmt.Errorf(
			"Pods can't be found on namespace '%s' with label %s", namespace, label)
	}

	out := strings.Trim(stdout.String(), "\n")
	if len(out) == 0 {
		//Small hack. String split always return array with an empty string
		return []string{}, nil
	}
	return strings.Split(out, " "), nil
}

func (kubectl *Kubectl) ManifestsPath() string {
	return fmt.Sprintf("%s/k8sT/manifests/%s", basePath, "1.7")
}

//WaitforPods wait during timeout to get a pod ready
func (kubectl *Kubectl) WaitforPods(namespace string, filter string, timeout int) (bool, error) {
	wait := 0
	var jsonPath = "{.items[*].status.containerStatuses[*].ready}"
	for wait < timeout {
		data, err := kubectl.GetPods(namespace, filter).Filter(jsonPath)
		if err != nil {
			kubectl.logCxt.Warnf("WaitforPods: GetPods failed err='%s'", err)
		} else {
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
		kubectl.logCxt.Infof(
			"WaitForPods on namespace '%s' with filter '%s' is not ready timeout='%d' data='%s'",
			namespace, filter, wait, data)
		time.Sleep(1)
		wait++
	}

	return false, nil
}

//Apply a new manifest using kubectl
func (kubectl *Kubectl) Apply(filepath string) bool {
	return kubectl.Node.Execute(
		fmt.Sprintf("kubectl apply -f  %s", filepath), nil, nil)
}

//Delete a manifest using kubectl
func (kubectl *Kubectl) Delete(filepath string) bool {
	return kubectl.Node.Execute(
		fmt.Sprintf("kubectl delete -f  %s", filepath), nil, nil)
}

//GetCiliumPods return all cilium pods
func (kubectl *Kubectl) GetCiliumPods(namespace string) ([]string, error) {
	return kubectl.GetPodsNames(namespace, "k8s-app=cilium")
}

//CiliumExec run command into a cilium pod
func (kubectl *Kubectl) CiliumExec(pod string, cmd string) (string, error) {
	command := fmt.Sprintf("kubectl exec -n kube-system %s -- %s", pod, cmd)
	stdout := new(bytes.Buffer)

	exit := kubectl.Node.Execute(command, stdout, nil)
	if exit == false {
		// FIXME: Output here is important.
		// Return the string is not fired on the assertion :\ Need to check
		kubectl.logCxt.Infof(
			"CiliumExec command failed '%s' pod='%s' erro='%s'",
			cmd, pod, stdout.String())
		return "", fmt.Errorf("CiliumExec: command '%s' failed '%s'", command, stdout.String())
	}
	return stdout.String(), nil
}

//CiliumImportPolicy import a new policy to cilium
func (kubectl *Kubectl) CiliumImportPolicy(namespace string, filepath string, timeout int) (string, error) {
	var revision int
	var wait int
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

//GetCiliumPodOnNode Returns cilium pod name that is running on specific node
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

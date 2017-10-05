package helpers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/asaskevich/govalidator"
	"github.com/cilium/cilium/api/v1/models"
	"k8s.io/client-go/util/jsonpath"
)

//GetCurentK8SEnv Return the value of the K8S_VERSION
func GetCurrentK8SEnv() string { return os.Getenv("K8S_VERSION") }

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
	versionTarget := fmt.Sprintf("%s-%s", target, GetCurrentK8SEnv())
	log.Infof("Kubectl: Set target to '%s'", versionTarget)
	node := CreateNodeFromTarget(versionTarget)
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

func (kub *Kubectl) Get(namespace string, command string) *cmdRes {
	return kub.Node.Exec(fmt.Sprintf(
		"kubectl -n %s get %s -o json", namespace, command))
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
func (kub *Kubectl) GetPodsNames(namespace string, label string) ([]string, error) {
	stdout := new(bytes.Buffer)
	filter := "-o jsonpath='{.items[*].metadata.name}'"
	exit := kub.Node.Execute(
		fmt.Sprintf("kubectl -n %s get pods -l %s %s", namespace, label, filter),
		stdout, nil)

	if exit == false {
		return nil, fmt.Errorf(
			"Pods couldn't be found on namespace '%s' with label %s", namespace, label)
	}

	out := strings.Trim(stdout.String(), "\n")
	if len(out) == 0 {
		//Small hack. String split always return array with an empty string
		return []string{}, nil
	}
	return strings.Split(out, " "), nil
}

func (kubectl *Kubectl) ManifestsPath() string {
	return fmt.Sprintf("%s/k8sT/manifests/%s", basePath, GetCurrentK8SEnv())
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
func (kubectl *Kubectl) Apply(filepath string) *cmdRes {
	return kubectl.Node.Exec(
		fmt.Sprintf("kubectl apply -f  %s", filepath))
}

//Delete a manifest using kubectl
func (kubectl *Kubectl) Delete(filepath string) *cmdRes {
	return kubectl.Node.Exec(
		fmt.Sprintf("kubectl delete -f  %s", filepath))
}

//GetCiliumPods return all cilium pods
func (kubectl *Kubectl) GetCiliumPods(namespace string) ([]string, error) {
	return kubectl.GetPodsNames(namespace, "k8s-app=cilium")
}

func (kub *Kubectl) CiliumEndpointsGet(pod string) *cmdRes {
	return kub.CiliumExec(pod, "cilium endpoint list -o json")
}

func (kub *Kubectl) CiliumEndpointsGetByTag(pod, tag string) (EndPointMap, error) {
	result := make(EndPointMap)
	var data []models.Endpoint
	eps := kub.CiliumEndpointsGet(pod)

	err := eps.UnMarshal(&data)
	if err != nil {
		return nil, err
	}

	for _, ep := range data {
		for _, label := range ep.Labels.OrchestrationIdentity {
			if tag == label {
				result[ep.ContainerName] = ep
				break
			}
		}

	}
	return result, nil
}

func (kub *Kubectl) CiliumEndpointWait(pod string) bool {

	body := func() bool {
		status, err := kub.CiliumEndpointsGet(pod).Filter("{[*].state}")
		if err != nil {
			return false
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

		kub.logCxt.Infof(
			"waiting for cilium endpoints pod=%s valid='%d' invalid='%s'",
			pod, valid, invalid)
		return false
	}

	err := WithTimeout(body, "Can't retrieve endpoints", &TimeoutConfig{Timeout: 100})
	if err != nil {
		return false
	}
	return true
}

//CiliumExec run command into a cilium pod
func (kub *Kubectl) CiliumExec(pod string, cmd string) *cmdRes {
	cmd = fmt.Sprintf("kubectl exec -n kube-system %s -- %s", pod, cmd)
	return kub.Node.Exec(cmd)
}

//CiliumPolicyRevision: Get the policy revision for a pod
func (kub *Kubectl) CiliumPolicyRevision(pod string) (int, error) {

	res := kub.CiliumExec(pod, "cilium policy get | grep Revision | awk '{print $2}'")

	if !res.Correct() {
		return -1, fmt.Errorf("Can't get the revision %s", res.Output())
	}

	revi, err := res.IntOutput()
	if err != nil {
		return revi, err
	}
	return revi, nil
}

//CiliumImportPolicy import a new policy to cilium
func (kub *Kubectl) CiliumImportPolicy(namespace string, filepath string, timeout int) (string, error) {
	var revision, revi int
	var wait int
	pods, err := kub.GetCiliumPods(namespace)
	if err != nil {
		return "", err
	}

	for _, v := range pods {
		revi, err := kub.CiliumPolicyRevision(v)
		if err != nil {
			return "", err
		}

		if revi > revision {
			revision = revi
		}
	}

	if status := kub.Apply(filepath); !status.Correct() {
		return "", fmt.Errorf("Can't apply the policy '%s'", filepath)
	}

	for wait < timeout {
		valid := true
		for _, v := range pods {

			revi, err := kub.CiliumPolicyRevision(v)
			if err != nil {
				return "", err
			}

			if revi <= revision {
				valid = false
			}
		}
		if valid == true {
			//Wait until all the pods are synced
			for _, v := range pods {
				kub.Exec(namespace, v, fmt.Sprintf("cilium policy wait %d", revi))
			}
			//FIXME: Check if something need to be return here
			return "", nil
		}
		time.Sleep(1 * time.Second)
		wait++
	}
	return "", fmt.Errorf("ImportPolicy error due timeout '%d'", timeout)
}

//GetCiliumPodOnNode Returns cilium pod name that is running on specific node
func (kub *Kubectl) GetCiliumPodOnNode(namespace string, node string) (string, error) {
	filter := fmt.Sprintf(
		"-o jsonpath='{.items[?(@.spec.nodeName == \"%s\")].metadata.name}'", node)

	res := kub.Node.Exec(fmt.Sprintf(
		"kubectl -n %s get pods -l k8s-app=cilium %s", namespace, filter))
	if !res.Correct() {
		return "", fmt.Errorf("Cilium pod not found on node '%s'", node)
	}

	return res.Output().String(), nil
}

//EndPointMap Map with all the endpoints in cilium
type EndPointMap map[string]models.Endpoint

func (epMap *EndPointMap) GetPolicyStatus() map[string]int {
	var result map[string]int = map[string]int{
		"enabled":  0,
		"disabled": 0,
	}

	for _, ep := range *epMap {
		if *ep.PolicyEnabled == true {
			result["enabled"]++
		} else {
			result["disabled"]++
		}
	}
	return result
}

func (epMap *EndPointMap) AreReady() bool {
	for _, ep := range *epMap {
		if ep.State != "ready" {
			return false
		}
	}
	return true
}

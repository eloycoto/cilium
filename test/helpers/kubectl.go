package helpers

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/asaskevich/govalidator"
	log "github.com/sirupsen/logrus"

	"github.com/cilium/cilium/api/v1/models"
)

//GetCurrentK8SEnv returns the value of K8S_VERSION from the OS environment
func GetCurrentK8SEnv() string { return os.Getenv("K8S_VERSION") }

//Kubectl kubectl command helper to connect to Kubectl instance
type Kubectl struct {
	Node *Node //helpers.Node struct to connect to ssh

	logCxt *log.Entry //log context with test fields
}

// CreateKubectl initializes a Kubectl helper with the provided target and log
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

//Exec runs the provided command in a pod running in a specific namespace
func (kub *Kubectl) Exec(namespace string, pod string, cmd string) (string, error) {
	command := fmt.Sprintf("kubectl exec -n %s %s -- %s", namespace, pod, cmd)
	stdout := new(bytes.Buffer)

	exit := kub.Node.Execute(command, stdout, nil)
	if exit == false {
		// TODO: Return CmdRes here
		// Return the string is not fired on the assertion :\ Need to check
		kub.logCxt.Errorf(
			"Exec command failed '%s' pod='%s' error='%s'",
			cmd, pod, stdout.String())
		return "", fmt.Errorf("Exec: command '%s' failed '%s'", command, stdout.String())
	}
	return stdout.String(), nil
}

// Get retrieves the provided Kubernetes objects from the specified namespace
func (kub *Kubectl) Get(namespace string, command string) *CmdRes {
	return kub.Node.Exec(fmt.Sprintf(
		"kubectl -n %s get %s -o json", namespace, command))
}

//GetPods return all the pods for a namespace. Kubectl options/filter can be passed
func (kub *Kubectl) GetPods(namespace string, filter string) *CmdRes {
	return kub.Node.Exec(fmt.Sprintf("kubectl -n %s get pods %s -o json", namespace, filter))
}

//GetPodNames returns the names of all of the pods that are labelled with label
//in the specified namespace, along with an error if the pod names cannot be
//retrieved.
func (kub *Kubectl) GetPodNames(namespace string, label string) ([]string, error) {
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
		//Small hack. String split always return an array with an empty string
		return []string{}, nil
	}
	return strings.Split(out, " "), nil
}

//Logs returns CmdRes with the output of kubectl logs <pod> command
func (kub *Kubectl) Logs(namespace string, pod string) *CmdRes {
	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)

	exit := kub.Node.Execute(
		fmt.Sprintf("kubectl -n %s logs %s", namespace, pod),
		stdout, stderr)
	return &CmdRes{
		cmd:    "",
		stdout: stdout,
		stderr: stderr,
		exit:   exit,
	}
}

//ManifestsPath returns the manifests path for the k8s version
func (kub *Kubectl) ManifestsPath() string {
	return fmt.Sprintf("%s/k8sT/manifests/%s", basePath, GetCurrentK8SEnv())
}

//WaitforPods waits until pods are ready in the specified namespace using the provided JSONPath filter. It waits up to timeout seconds for the command to complete. Returns true if the command succeeded within the specified timeout, and an error if the command failed or did not succeed during the specified timeout.
func (kub *Kubectl) WaitforPods(namespace string, filter string, timeout time.Duration) (bool, error) {
	body := func() bool {
		var jsonPath = "{.items[*].status.containerStatuses[*].ready}"
		data, err := kub.GetPods(namespace, filter).Filter(jsonPath)
		if err != nil {
			kub.logCxt.Errorf("Could not get pods: %s", err)
			return false
		}

		valid := true
		result := strings.Split(data.String(), " ")
		for _, v := range result {
			if val, _ := govalidator.ToBoolean(v); val == false {
				valid = false
				break
			}
		}
		if valid == true {
			return true
		}

		kub.logCxt.Infof(
			"WaitForPods on namespace '%s' with filter '%s' is not ready data='%s'",
			namespace, filter, data)
		return false
	}
	err := WithTimeout(body, "Could not get Pods", &TimeoutConfig{Timeout: timeout})
	if err != nil {
		return false, err
	}
	return true, nil
}

//Apply applies the Kubernetes manifest located at path filepath
func (kub *Kubectl) Apply(filepath string) *CmdRes {
	return kub.Node.Exec(
		fmt.Sprintf("kubectl apply -f  %s", filepath))
}

//Delete Deletes the Kubernetes manifest at path filepath.
func (kub *Kubectl) Delete(filepath string) *CmdRes {
	return kub.Node.Exec(
		fmt.Sprintf("kubectl delete -f  %s", filepath))
}

//GetCiliumPods returns a list of all Cilium pods in the specified namespace,
//and an error if the Cilium pods were not able to be retrieved.
func (kub *Kubectl) GetCiliumPods(namespace string) ([]string, error) {
	return kub.GetPodNames(namespace, "k8s-app=cilium")
}

//CiliumEndpointsList return the list of cilium endpoint list
func (kub *Kubectl) CiliumEndpointsList(pod string) *CmdRes {
	return kub.CiliumExec(pod, "cilium endpoint list -o json")
}

//CiliumEndpointsListByTag return the list of endpoints filter by a tag
func (kub *Kubectl) CiliumEndpointsListByTag(pod, tag string) (EndPointMap, error) {
	result := make(EndPointMap)
	var data []models.Endpoint
	eps := kub.CiliumEndpointsList(pod)

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

//CiliumEndpointWait waits until all endpoints managed by the specified Cilium
//pod are ready. Returns false if the command to retrieve the state of the
//endpoints times out.
func (kub *Kubectl) CiliumEndpointWait(pod string) bool {

	body := func() bool {
		status, err := kub.CiliumEndpointsList(pod).Filter("{[*].state}")
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
			"waiting for cilium endpoints pod='%s' valid='%d' invalid='%d'",
			pod, valid, invalid)
		return false
	}

	err := WithTimeout(body, "Can't retrieve endpoints", &TimeoutConfig{Timeout: timeout})
	if err != nil {
		return false
	}
	return true
}

//CiliumExec runs cmd in the specified Cilium pod
func (kub *Kubectl) CiliumExec(pod string, cmd string) *CmdRes {
	cmd = fmt.Sprintf("kubectl exec -n kube-system %s -- %s", pod, cmd)
	return kub.Node.Exec(cmd)
}

//CiliumPolicyRevision returns the policy revision in the specified Cilium pod.
//If the policy revision cannot be retrieved, returns an error.
func (kub *Kubectl) CiliumPolicyRevision(pod string) (int, error) {
	// FIXME GH-1725
	res := kub.CiliumExec(pod, "cilium policy get | grep Revision | awk '{print $2}'")

	if !res.WasSuccessful() {
		return -1, fmt.Errorf("Cannot get the revision %s", res.Output())
	}

	revi, err := res.IntOutput()
	if err != nil {
		return revi, err
	}
	return revi, nil
}

//CiliumImportPolicy imports the policy stored in path filepath and waits up
//until timeout seconds for the policy to be applied in all Cilium endpoints.
//Returns an error if the command fails or times out.
func (kub *Kubectl) CiliumImportPolicy(namespace string, filepath string, timeout time.Duration) (string, error) {
	var revision, revi int

	kub.logCxt.Infof("Importing policy '%s'", filepath)
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
	kub.logCxt.Infof("CiliumImportPolicy: path='%s' with revision '%d'", filepath, revision)
	if status := kub.Apply(filepath); !status.WasSuccessful() {
		return "", fmt.Errorf("Can't apply the policy '%s'", filepath)
	}

	body := func() bool {
		valid := true
		for _, v := range pods {

			revi, err := kub.CiliumPolicyRevision(v)
			if err != nil {
				kub.logCxt.Errorf("CiliumImportPolicy: error on get revision %s", err)
				return false
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
			kub.logCxt.Infof("CiliumImportPolicy: reivision %d is ready", revi)
			return true
		}
		return false
	}
	err = WithTimeout(
		body,
		"cannot import policy correctly; command timed out",
		&TimeoutConfig{Timeout: timeout})
	if err != nil {
		return "", err
	}
	return "", nil
}

//CiliumReport report the cilium pod to the log and apppend the logs for the
//given commands. Return err in case of any problem
func (kub *Kubectl) CiliumReport(namespace string, pod string, commands []string) error {
	wr := kub.logCxt.Logger.Out
	fmt.Fprint(wr, "StackTrace Begin\n")
	data := kub.Logs(namespace, pod)
	fmt.Fprintln(wr, data.Output())

	for _, cmd := range commands {
		out, err := kub.Exec(namespace, pod, cmd)
		if err != nil {
			kub.logCxt.Errorf("Error executing command '%s' on pod %s: %s", cmd, pod, err)
		}
		fmt.Fprintln(wr, out)
	}
	fmt.Fprint(wr, "StackTrace Ends\n")
	return nil
}

//GetCiliumPodOnNode returns the name of the Cilium pod that is running on / in
//the specified node / namespace.
func (kub *Kubectl) GetCiliumPodOnNode(namespace string, node string) (string, error) {
	filter := fmt.Sprintf(
		"-o jsonpath='{.items[?(@.spec.nodeName == \"%s\")].metadata.name}'", node)

	res := kub.Node.Exec(fmt.Sprintf(
		"kubectl -n %s get pods -l k8s-app=cilium %s", namespace, filter))
	if !res.WasSuccessful() {
		return "", fmt.Errorf("Cilium pod not found on node '%s'", node)
	}

	return res.Output().String(), nil
}

//EndPointMap Map with all the endpoints in cilium
type EndPointMap map[string]models.Endpoint

//GetPolicyStatus returns a mapping of how many endpoints have policy
//enforcement enabled and disabled.
func (epMap *EndPointMap) GetPolicyStatus() map[string]int {
	result := map[string]int{
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

//AreReady return true if all cilium endpoints are in 'ready' state
func (epMap *EndPointMap) AreReady() bool {
	for _, ep := range *epMap {
		if ep.State != "ready" {
			return false
		}
	}
	return true
}

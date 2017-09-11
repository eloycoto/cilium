package helpers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"time"

	"k8s.io/client-go/util/jsonpath"
)

type Kubectl struct {
	Node *Node
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

var timeout = 300 * time.Second

func CreateKubectl(host string, port int) *Kubectl {
	return &Kubectl{
		Node: CreateNode(host, port),
	}
}

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

func (kubectl *Kubectl) WaitforPods(namespace string, filter string, timeout int) (*KubectlRes, error) {
	wait := 0
	var jsonPath = "{.items[*].status.containerStatuses[*].ready}"
	for wait < timeout {
		data, err := kubectl.GetPods(namespace, filter).Filter(jsonPath)
		if err != nil {
			fmt.Printf("Result --%v \n", data)
			fmt.Printf("#############################\n")
		}
		time.Sleep(1)
		wait++

		return nil, nil
	}

	return &KubectlRes{}, nil

}

func (kubectl *Kubectl) Apply(filepath string) bool {
	return kubectl.Node.Execute(
		fmt.Sprintf("kubectl apply -f  %s", filepath), nil, nil)
}

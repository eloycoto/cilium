package helpers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"k8s.io/client-go/util/jsonpath"
)

type cmdRes struct {
	cmd    string
	params []string
	stdout *bytes.Buffer
	stderr *bytes.Buffer
	exit   bool
}

func (res *cmdRes) Output() *bytes.Buffer {
	return res.stdout
}

func (res *cmdRes) IntOutput() (int, error) {
	return strconv.Atoi(strings.Trim(res.stdout.String(), "\n"))
}

//Correct: return true if the command was sucessfull
func (res *cmdRes) Correct() bool {
	return res.exit
}

func (res *cmdRes) Filter(filter string) (*bytes.Buffer, error) {
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

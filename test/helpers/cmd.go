package helpers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
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

//Correct: return true if the command was sucessfull
func (res *cmdRes) Correct() bool {
	return res.exit
}

func (res *cmdRes) IntOutput() (int, error) {
	return strconv.Atoi(strings.Trim(res.stdout.String(), "\n"))
}

func (res *cmdRes) SingleOut() string {
	return strings.Trim(res.stdout.String(), "\n")
}

func (res *cmdRes) FindResults(filter string) ([]reflect.Value, error) {

	var data interface{}
	var result []reflect.Value

	err := json.Unmarshal(res.stdout.Bytes(), &data)
	if err != nil {
		return nil, err
	}
	parser := jsonpath.New("").AllowMissingKeys(true)
	parser.Parse(filter)
	fullResults, _ := parser.FindResults(data)
	for _, res := range fullResults {
		for _, val := range res {
			result = append(result, val)
		}
	}
	return result, nil
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

//KVOutput: This is a helper functon that return a map with the key=val output.
// This is going to be used when the output will be like this:
// 		a=1
// 		b=2
// 		c=3
// This funtion will return a map with the values in the stdout output
func (res *cmdRes) KVOutput() map[string]string {
	result := make(map[string]string)
	for _, line := range strings.Split(res.stdout.String(), "\n") {
		vals := strings.Split(line, "=")
		if len(vals) == 2 {
			result[vals[0]] = vals[1]
		}
	}
	return result
}

func (res *cmdRes) Output() *bytes.Buffer {
	return res.stdout
}

func (res *cmdRes) UnMarshal(data interface{}) error {
	err := json.Unmarshal(res.stdout.Bytes(), &data)
	return err
}

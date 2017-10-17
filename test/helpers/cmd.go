// Copyright 2017 Authors of Cilium
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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

//CmdRes Command response with all the data
type CmdRes struct {
	cmd    string
	params []string
	stdout *bytes.Buffer
	stderr *bytes.Buffer
	exit   bool
}

//WasSuccessful returns true if the command was sucessfull
func (res *CmdRes) WasSuccessful() bool {
	return res.exit
}

//CountLines return the number of stdout lines
func (res *CmdRes) CountLines() int {
	return len(strings.Split(res.stdout.String(), "\n"))
}

//IntOutput returns the stdout of res as integer
func (res *CmdRes) IntOutput() (int, error) {
	return strconv.Atoi(strings.Trim(res.stdout.String(), "\n"))
}

//SingleOut returns the stdout of res without any newline characters
func (res *CmdRes) SingleOut() string {
	return strings.Trim(res.stdout.String(), "\n")
}

//FindResults filter CmdRes using jsonpath and returns an interface with the values
func (res *CmdRes) FindResults(filter string) ([]reflect.Value, error) {

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

//Filter filters cmdRes using the provided JSONPath filter.
func (res *CmdRes) Filter(filter string) (*bytes.Buffer, error) {
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

// KVOutput returns a map of the stdout of the provided CmdRes split based on
// the separator '='.
// This is going to be used when the output will be like this:
// 		a=1
// 		b=2
// 		c=3
// This funtion will return a map with the values in the stdout output
func (res *CmdRes) KVOutput() map[string]string {
	result := make(map[string]string)
	for _, line := range strings.Split(res.stdout.String(), "\n") {
		vals := strings.Split(line, "=")
		if len(vals) == 2 {
			result[vals[0]] = vals[1]
		}
	}
	return result
}

//Output returns the contents of stdout
func (res *CmdRes) Output() *bytes.Buffer {
	return res.stdout
}

//CombineOutput returns the combined output of stdout and stderr
func (res *CmdRes) CombineOutput() *bytes.Buffer {
	result := res.stdout
	result.WriteString(res.stderr.String())
	return result
}

//UnMarshal unmarshals res's stdout into data
func (res *CmdRes) UnMarshal(data interface{}) error {
	err := json.Unmarshal(res.stdout.Bytes(), &data)
	return err
}

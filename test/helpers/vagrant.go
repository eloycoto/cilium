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
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"

	log "github.com/sirupsen/logrus"
)

//Vagrant helper struct
type Vagrant struct{}

//Create a new vagrant server. Receives and scope that it's the target server that need to be created.
// If ssh is true, the ssh-config data will be append to the ssh-config.
// In case of any error on vagrant [provision|up|ssh-config] error will be returned.
func (vagrant *Vagrant) Create(scope string, ssh ...bool) error {
	createCMD := "vagrant up %s --provision"
	for _, v := range vagrant.Status(scope) {
		switch v {
		case "running":
			createCMD = "vagrant provision %s"
		case "not_created":
			createCMD = "vagrant up %s --provision"
		default:
			//Sometimes server are stoped and not destroyed. Destroy just in case
			vagrant.Destroy(scope)
		}
	}
	createCMD = fmt.Sprintf(createCMD, scope)
	log.Infof("Vagrant:Create: running %s", createCMD)
	cmd := vagrant.getCmd(createCMD)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}

	if err := cmd.Start(); err != nil {
		log.Errorf("Create error on start command='%s' error=%s", createCMD, err)
		return err
	}

	in := bufio.NewScanner(stdout)
	for in.Scan() {
		log.Infof(in.Text()) // write each line to your log
	}

	if err := cmd.Wait(); err != nil {
		return err
	}

	if len(ssh) > 0 && ssh[0] == true {
		err = vagrant.createConfig(scope)
		if err != nil {
			return err
		}
	}
	return nil
}

func (vagrant *Vagrant) createConfig(scope string) error {
	cmd := vagrant.getCmd(fmt.Sprintf("vagrant ssh-config %s >> ssh-config", scope))
	_, err := cmd.CombinedOutput()
	if err != nil {
		return err
	}
	return nil
}

func (vagrant *Vagrant) deleteConfig() error {

	cmd := vagrant.getCmd("rm ssh-config")
	_, err := cmd.CombinedOutput()
	if err != nil {
		return err
	}
	return nil
}

//Destroy destroys all running Vagrant VMs in the provided scope. It returns an
//error if deletion of either the VMs fails
func (vagrant *Vagrant) Destroy(scope string) error {
	command := fmt.Sprintf("vagrant destroy -f %s ", scope)
	log.Infof("Vagrant:Destroy: running '%s'", command)
	cmd := vagrant.getCmd(command)
	_, err := cmd.CombinedOutput()
	if err != nil {
		return err
	}
	return nil
}

func (vagrant *Vagrant) getCmd(op string) *exec.Cmd {
	cmd := exec.Command(vagrant.getPath("bash"), "-c", op)
	cmd.Dir = vagrant.getDir()
	return cmd
}

func (vagrant *Vagrant) getDir() string {
	dir, err := os.Getwd()
	if err != nil {
		return "/tmp/"
	}
	return fmt.Sprintf("%s/", dir)
}

func (vagrant *Vagrant) getPath(prog string) string {
	path, err := exec.LookPath(prog)
	if err != nil {
		return ""
	}
	return path
}

//Status returns a mapping of Vagrant VM name to its status
func (vagrant *Vagrant) Status(key string) map[string]string {
	result := map[string]string{}

	cmd := vagrant.getCmd(fmt.Sprintf("vagrant status %s --machine-readable", key))
	data, err := cmd.CombinedOutput()
	if err != nil {
		return result
	}
	for _, line := range strings.Split(string(data), "\n") {
		val := strings.Split(line, ",")
		if len(val) > 2 && val[2] == "state" {
			result[val[1]] = val[3]
		}
	}
	return result
}

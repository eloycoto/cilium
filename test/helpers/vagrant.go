package helpers

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

//Vagrant helper struct
type Vagrant struct{}

//Create a new vagrant server
func (vagrant *Vagrant) Create() error {
	createCMD := "vagrant up"

	for _, v := range vagrant.Status() {
		if v == "running" {
			createCMD = "vagrant provision"
			break
		}
	}
	cmd := vagrant.getCMD(createCMD)
	// FIXME: output to log with proper header
	out, err := cmd.CombinedOutput()
	fmt.Printf("%s", out)
	if err != nil {
		return err
	}

	err = vagrant.createConfig()
	if err != nil {
		return err
	}
	return nil
}

func (vagrant *Vagrant) createConfig() error {
	cmd := vagrant.getCMD("vagrant ssh-config > ssh-config")
	_, err := cmd.CombinedOutput()
	if err != nil {
		return err
	}
	return nil
}

func (vagrant *Vagrant) deleteConfig() error {

	cmd := vagrant.getCMD("rm ssh-config")
	_, err := cmd.CombinedOutput()
	if err != nil {
		return err
	}
	return nil
}

//Destroy all the vms
func (vagrant *Vagrant) Destroy() error {

	cmd := vagrant.getCMD("vagrant destroy -f")
	_, err := cmd.CombinedOutput()
	if err != nil {
		return err
	}
	err = vagrant.deleteConfig()
	if err != nil {
		return err
	}
	return nil
}

func (vagrant *Vagrant) getCMD(op string) *exec.Cmd {
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

//Status return a map with the server name (key) and the status as value
func (vagrant *Vagrant) Status(key ...string) map[string]string {
	var result map[string]string = map[string]string{}

	cmd := vagrant.getCMD("vagrant status --machine-readable")
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

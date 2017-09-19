package helpers

import (
	"fmt"
	"os"
	"os/exec"
)

//Vagrant helper struct
type Vagrant struct{}

func (vagrant *Vagrant) getPath(prog string) string {
	path, err := exec.LookPath(prog)
	if err != nil {
		return ""
	}
	return path
}

func (vagrant *Vagrant) getDir() string {
	dir, err := os.Getwd()
	if err != nil {
		return "/tmp/"
	}
	return fmt.Sprintf("%s/", dir)
}

func (vagrant *Vagrant) getCMD(op string) *exec.Cmd {
	cmd := exec.Command(vagrant.getPath("bash"), "-c", op)
	cmd.Dir = vagrant.getDir()
	return cmd
}

//Create a new vagrant server
func (vagrant *Vagrant) Create() error {
	cmd := vagrant.getCMD("vagrant up")
	// FIXME: output to log with proper header
	// FIXME: Check if the VM is up, if up set reload
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

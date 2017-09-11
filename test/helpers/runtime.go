package helpers

import (
	log "github.com/sirupsen/logrus"
)

//CreateNewRuntimeHelper return a docker, cilium helper
func CreateNewRuntimeHelper(target string, log *log.Entry) (*Docker, *Cilium) {
	return CreateDocker(target, log), CreateCilium(target, log)
}

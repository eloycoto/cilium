package RunT

import (
	"fmt"
	"time"

	"github.com/cilium/cilium/test/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	log "github.com/sirupsen/logrus"
)

var _ = Describe("RunKVStoreTest", func() {

	var initilized bool
	var networkName string = "cilium-net"

	var netperfImage string = "tgraf/netperf"
	var logger *log.Entry
	var docker *helpers.Docker
	var cilium *helpers.Cilium

	initilize := func() {
		if initilized == true {
			return
		}
		logger = log.WithFields(log.Fields{"test": "RunKVStoreTest"})
		logger.Info("Starting")
		docker, cilium = helpers.CreateNewRuntimeHelper("runtime", logger)
		initilized = true
	}
	containers := func(option string) {
		switch option {
		case "create":
			docker.NetworkCreate(networkName, "")
			docker.ContainerCreate("client", netperfImage, networkName, "-l id.client")
		case "delete":
			docker.ContainerRm("client")

		}
	}

	done := make(chan string, 1)
	agent := func(option string) {
		cmd := fmt.Sprintf("sudo /usr/bin/cilium-agent %s --debug", option)
		killCmd := "sudo kill -9 $(pgrep cilium-agent)"
		go docker.Node.Exec(cmd)
		timeout := time.After(300 * time.Second)
		for {
			select {
			case <-done:
				logger.Info("killing cilium-agent daemon")
				docker.Node.Exec(killCmd)
				return
			case <-timeout:
				logger.Info("killing cilium-agent due timeout")
				fmt.Printf("Timeout")
				docker.Node.Exec(killCmd)
				return
			}
		}
	}

	BeforeEach(func() {
		initilize()
		docker.Node.Exec("sudo systemctl stop cilium")
	}, 150)

	AfterEach(func() {
		containers("delete")
		done <- "Fail"
		docker.Node.Exec("sudo systemctl start cilium")
	})

	It("Consul KVStore", func() {
		go agent("--kvstore consul --kvstore-opt consul.address=127.0.0.1:8500")
		cilium.WaitUntilReady(150)
		docker.Node.Exec("sudo systemctl restart cilium-docker")
		helpers.Sleep(2)
		containers("create")
		cilium.EndpointWaitUntilReady()
		eps, err := cilium.GetEndpointsNames()
		Expect(err).Should(BeNil())
		Expect(len(eps)).To(Equal(1))

	})

	It("Etcd KVStore", func() {
		go agent("--kvstore etcd --kvstore-opt etcd.address=127.0.0.1:4001")
		cilium.WaitUntilReady(150)
		docker.Node.Exec("sudo systemctl restart cilium-docker")
		helpers.Sleep(2)
		containers("create")
		cilium.EndpointWaitUntilReady()

		eps, err := cilium.GetEndpointsNames()
		Expect(err).Should(BeNil())
		Expect(len(eps)).To(Equal(1))
	})
})

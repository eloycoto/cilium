package RunT

import (
	"github.com/cilium/cilium/test/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	log "github.com/sirupsen/logrus"
)

var _ = Describe("RunChaosMonkey", func() {

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
		logger = log.WithFields(log.Fields{"test": "RunChaosMonkey"})
		logger.Info("Starting")
		docker, cilium = helpers.CreateNewRuntimeHelper("runtime", logger)
		docker.NetworkCreate(networkName, "")
		initilized = true
	}

	wait_for_cilium := func() {
		var wait int
		var timeout int = 100

		for wait < timeout {
			res := cilium.Node.ExecWithSudo("cilium status", nil, nil)
			if res {
				cilium.EndpointWaitUntilReady()
				// Sometimes system fail on jenkins. Looks like a race condition.
				// so we sleep for 5 seconds.
				helpers.Sleep(5)
				break
			}
			logger.Infof("Cilium is not ready yet wait='%d'", wait)
			helpers.Sleep(1)
			wait++
		}
	}

	BeforeEach(func() {
		initilize()
		docker.ContainerCreate("client", netperfImage, networkName, "-l id.client")
		docker.ContainerCreate("server", netperfImage, networkName, "-l id.server")

	})

	AfterEach(func() {
		docker.ContainerRm("client")
		docker.ContainerRm("server")
	})

	It("Endpoint recovery on restart", func() {
		endpoints, err := cilium.GetEndpointsNames()
		Expect(err).Should(BeNil())
		originalEndpoins := len(endpoints)
		cilium.Node.ExecWithSudo("systemctl restart cilium", nil, nil)

		wait_for_cilium()

		endpoints, err = cilium.GetEndpointsNames()
		Expect(err).Should(BeNil())
		Expect(len(endpoints)).To(Equal(originalEndpoins))
		for _, container := range endpoints {
			docker.ContainerRm(container)
		}
	}, 300)

	It("Interfaces chaos", func() {
		originalLinks, err := docker.Node.Exec("sudo ip link show | wc -l").IntOutput()
		Expect(err).Should(BeNil())

		_ = docker.Node.Exec("sudo ip link add lxc12345 type veth peer name tmp54321")

		status := docker.Node.Exec("sudo systemctl restart cilium")
		Expect(status.Correct()).Should(BeTrue())

		wait_for_cilium()

		status = docker.Node.Exec("sudo ip link show lxc12345")
		Expect(status.Correct()).Should(BeFalse(),
			"leftover interface were not properly cleaned up")

		links, err := docker.Node.Exec("sudo ip link show | wc -l").IntOutput()
		Expect(links).Should(Equal(originalLinks),
			"Some network interfaces were accidentally removed!")
	}, 300)
})

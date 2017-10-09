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
		err := cilium.WaitUntilReady(100)
		Expect(err).Should(BeNil())

		status := cilium.EndpointWaitUntilReady()
		Expect(status).Should(BeTrue())

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
		original_ips := cilium.Node.Exec(`
		curl -s --unix-socket /var/run/cilium/cilium.sock \
		http://localhost/v1beta/healthz/ | jq ".ipam.ipv4|length"`)

		res := cilium.Node.Exec("sudo systemctl restart cilium")
		Expect(res.Correct()).Should(BeTrue())

		wait_for_cilium()

		ips := cilium.Node.Exec(`
		curl -s --unix-socket /var/run/cilium/cilium.sock \
		http://localhost/v1beta/healthz/ | jq ".ipam.ipv4|length"`)
		Expect(original_ips.Output()).To(Equal(ips.Output()))

	}, 300)

	It("Interfaces chaos", func() {
		originalLinks, err := docker.Node.Exec("sudo ip link show | wc -l").IntOutput()
		Expect(err).Should(BeNil())

		_ = docker.Node.Exec("sudo ip link add lxc12345 type veth peer name tmp54321")

		res := cilium.Node.Exec("sudo systemctl restart cilium")
		Expect(res.Correct()).Should(BeTrue())

		wait_for_cilium()

		status := docker.Node.Exec("sudo ip link show lxc12345")
		Expect(status.Correct()).Should(BeFalse(),
			"leftover interface were not properly cleaned up")

		links, err := docker.Node.Exec("sudo ip link show | wc -l").IntOutput()
		Expect(links).Should(Equal(originalLinks),
			"Some network interfaces were accidentally removed!")
	}, 300)
})

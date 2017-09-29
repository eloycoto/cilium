package RunT

import (
	"time"

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

	BeforeEach(func() {
		initilize()
	})

	It("Endpoint recovery on restart", func() {
		docker.ContainerCreate("client", netperfImage, networkName, "-l id.client")
		docker.ContainerCreate("server", netperfImage, networkName, "-l id.server")

		endpoints, err := cilium.GetEndpointsNames()
		Expect(err).Should(BeNil())
		original_endpoins := len(endpoints)
		cilium.Node.ExecWithSudo("systemctl restart cilium", nil, nil)

		var wait int
		var timeout int = 100

		for wait < timeout {
			res := cilium.Node.ExecWithSudo("cilium status", nil, nil)
			if res {
				break
			}
			logger.Infof("Cilium is not ready yet wait='%d'", wait)
			wait++
			time.Sleep(1 * time.Second)
		}

		endpoints, err = cilium.GetEndpointsNames()
		Expect(err).Should(BeNil())
		Expect(len(endpoints)).To(Equal(original_endpoins))
		for _, container := range endpoints {
			docker.ContainerRm(container)
		}
	})
})

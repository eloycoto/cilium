package RunT

import (
	"fmt"
	"time"

	ginkgoext "github.com/cilium/cilium/test/ginkgo-ext"
	"github.com/cilium/cilium/test/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	log "github.com/sirupsen/logrus"
)

var _ = FDescribe("RunConnectivyTest", func() {

	var initilized bool
	var networkName string = "cilium-net"

	var netperfImage string = "tgraf/netperf"
	var logger *log.Entry
	var docker *helpers.Docker
	var cilium *helpers.Cilium

	afterAll := &ginkgoext.AfterAll{
		Body: func() {
			//AfterEach will be executed after that. So wait until delete the containers
			go func() {
				var wait int = 0
				var timeout int = 150

				for wait < timeout {
					res, _ := docker.NetworkGet(networkName).Filter("{ [0].Containers}")
					if res.String() == "map[]" {
						//FIXME: This should be .[0].Containers|len
						break
					}
					wait++
					time.Sleep(1 * time.Second)
					log.Infof("Waiting for containers to delete wait '%d'", wait)
				}
				docker.NetworkDelete(networkName)
			}()
		},
	}

	initilize := func() {
		if initilized == true {
			return
		}
		logger = log.WithFields(log.Fields{"test": "RunConnectivyTest"})
		logger.Info("Starting")
		docker, cilium = helpers.CreateNewRuntimeHelper("runtime", logger)
		docker.NetworkCreate(networkName, "")
		initilized = true
	}

	BeforeEach(func() {
		initilize()
		docker.ContainerCreate("client", netperfImage, networkName, "-l id.client")
		docker.ContainerCreate("server", netperfImage, networkName, "-l id.server")
		cilium.Exec("policy delete --all")
	})

	AfterEach(func() {
		docker.ContainerRm("client")
		docker.ContainerRm("server")
		return
	})

	ginkgoext.It("Test containers connectivity without policy", func() {
		serverData := docker.ContainerInspect("server")
		serverIP, err := serverData.Filter("{[0].NetworkSettings.Networks.cilium-net.IPAddress}")
		Expect(err).Should(BeNil())
		serverIPv6, err := serverData.Filter("{[0].NetworkSettings.Networks.cilium-net.GlobalIPv6Address}")
		Expect(err).Should(BeNil())

		By("Client can ping to server IPV6")
		res := docker.ContainerExec("client", fmt.Sprintf("ping6 -c 4 %s", serverIPv6))
		Expect(res.Correct()).Should(BeTrue())

		By("Client can ping to server Ipv4")
		res = docker.ContainerExec("client", fmt.Sprintf("ping -c 5 %s", serverIP))
		Expect(res.Correct()).Should(BeTrue())

		By("Netperf to server from client IPv6")
		cmd := fmt.Sprintf(
			"netperf -c -C -t TCP_SENDFILE -H %s", serverIPv6)
		res = docker.ContainerExec("client", cmd)
		Expect(res.Correct()).Should(BeTrue())
	}, afterAll)

	ginkgoext.It("Test containers connectivity WITH policy", func() {
		policyID, _ := cilium.ImportPolicy(
			fmt.Sprintf("%s/test.policy", cilium.ManifestsPath()), 150)
		logger.Debug("New policy created with id '%d'", policyID)

		serverData := docker.ContainerInspect("server")
		serverIP, err := serverData.Filter("{[0].NetworkSettings.Networks.cilium-net.IPAddress}")
		Expect(err).Should(BeNil())
		serverIPv6, err := serverData.Filter("{[0].NetworkSettings.Networks.cilium-net.GlobalIPv6Address}")
		Expect(err).Should(BeNil())

		By("Client can ping to server IPV6")
		res := docker.ContainerExec("client", fmt.Sprintf("ping6 -c 4 %s", serverIPv6))
		Expect(res.Correct()).Should(BeTrue())

		By("Client can ping to server Ipv4")
		res = docker.ContainerExec("client", fmt.Sprintf("ping -c 5 %s", serverIP))
		Expect(res.Correct()).Should(BeTrue())

		By("Netperf to server from client IPv6")
		cmd := fmt.Sprintf(
			"netperf -c -C -t TCP_SENDFILE -H %s", serverIPv6)
		res = docker.ContainerExec("client", cmd)
		Expect(res.Correct()).Should(BeTrue())

		By("Ping from host to server")
		ping := docker.Node.Execute(fmt.Sprintf("ping -c 4 %s", serverIP), nil, nil)
		Expect(ping).Should(BeTrue())
	}, afterAll)
})

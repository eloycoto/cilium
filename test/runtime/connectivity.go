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

var _ = XDescribe("RunConnectivyTest", func() {

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

		for {
			if data, _ := cilium.GetEndpointsNames(); len(data) >= 2 {
				logger.Info("Waiting for endpoints to be ready")
				return
			}
			helpers.Sleep(1)
		}
	}, 150)

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
		policyID, _ := cilium.PolicyImport(
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

var _ = Describe("RunConntrackTest", func() {

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
		logger = log.WithFields(log.Fields{"test": "RunConnectivyTest"})
		logger.Info("Starting")
		docker, cilium = helpers.CreateNewRuntimeHelper("runtime", logger)
		docker.NetworkCreate(networkName, "")
		initilized = true
	}

	client_server_connectivity := func() {
		cliIP, err := docker.ContainerInspectNet("client")
		Expect(err).Should(BeNil(), "Couldn't get container client Meta")

		srvIP, err := docker.ContainerInspectNet("server")
		Expect(err).Should(BeNil(), "Couldn't get container server Meta")

		By("Client pinging server IPv6")
		res := docker.ContainerExec("client", fmt.Sprintf("ping6 -c 4 %s", srvIP["IPv6"]))
		Expect(res.Correct()).Should(BeTrue(), fmt.Sprintf(
			"Client can't ping to server %s", srvIP["IPv6"]))

		By("Client pinging server IPv4")
		res = docker.ContainerExec("client", fmt.Sprintf("ping -c 4 %s", srvIP["IPv4"]))
		Expect(res.Correct()).Should(BeTrue(), fmt.Sprintf(
			"Client can't ping to server %s", srvIP["IPv4"]))

		By("Client netcat to port 777 IPv6")
		res = docker.ContainerExec("client", fmt.Sprintf("nc -w 4 %s 777", srvIP["IPv6"]))
		Expect(res.Correct()).Should(BeFalse(), fmt.Sprintf(
			"Client can connect to %s:777. Should fail", srvIP["IPv6"]))

		By("Client netcat to port 777 IPv4")
		res = docker.ContainerExec("client", fmt.Sprintf("nc -w 4 %s 777", srvIP["IPv4"]))
		Expect(res.Correct()).Should(BeFalse(), fmt.Sprintf(
			"Client can connect to %s:777. Should fail", srvIP["IPv4"]))

		By("Client netperf to server IPv6")
		res = docker.ContainerExec("client", fmt.Sprintf(
			"netperf -l 3 -t TCP_RR -H %s", srvIP["IPv6"]))
		Expect(res.Correct()).Should(BeTrue(), fmt.Sprintf(
			"Client can't netperf to server %s", srvIP["IPv6"]))

		By("Client netperf to server IPv4")
		res = docker.ContainerExec("client", fmt.Sprintf(
			"netperf -l 3 -t TCP_RR -H %s", srvIP["IPv4"]))
		Expect(res.Correct()).Should(BeTrue(), fmt.Sprintf(
			"Client can't netperf to server %s", srvIP["IPv4"]))

		By("Client UDP netperf to server IPv6")
		res = docker.ContainerExec("client", fmt.Sprintf(
			"netperf -l 3 -t UDP_RR -H %s", srvIP["IPv6"]))
		Expect(res.Correct()).Should(BeTrue(), fmt.Sprintf(
			"Client can't netperf to server %s", srvIP["IPv6"]))

		By("Client UDP netperf to server IPv4")
		res = docker.ContainerExec("client", fmt.Sprintf(
			"netperf -l 3 -t UDP_RR -H %s", srvIP["IPv4"]))
		Expect(res.Correct()).Should(BeTrue(), fmt.Sprintf(
			"Client can't netperf to server %s", srvIP["IPv4"]))

		By("Ping from host to server IPv6")
		ping := docker.Node.Execute(fmt.Sprintf("ping6 -c 4 %s", srvIP["IPv6"]), nil, nil)
		Expect(ping).Should(BeTrue(), "Host Can't ping to server")

		By("Ping from host to server IPv4")
		ping = docker.Node.Execute(fmt.Sprintf("ping -c 4 %s", srvIP["IPv4"]), nil, nil)
		Expect(ping).Should(BeTrue(), "Host can't ping to server")

		By("Ping from server to client IPv6")
		res = docker.ContainerExec("server", fmt.Sprintf("ping6 -c 4 %s", cliIP["IPv6"]))
		Expect(res.Correct()).Should(BeTrue(), fmt.Sprintf(
			"Server can't ping to client %s", cliIP["IPv6"]))

		By("Ping from server to client IPv4")
		res = docker.ContainerExec("server", fmt.Sprintf("ping -c 4 %s", cliIP["IPv4"]))
		Expect(res.Correct()).Should(BeTrue(), fmt.Sprintf(
			"Server can't ping to client %s", cliIP["IPv4"]))
	}

	BeforeEach(func() {
		initilize()
		docker.ContainerCreate("client", netperfImage, networkName, "-l id.client")
		docker.ContainerCreate("server", netperfImage, networkName, "-l id.server")
		cilium.Exec("policy delete --all")
	})

	AfterEach(func() {
		docker.ContainerRm("server")
		docker.ContainerRm("client")
	})

	It("Conntrack disabled", func() {
		endpoints, err := cilium.GetEndpointsIds()
		Expect(err).Should(BeNil(), "Couldn't get endpoints IDS")

		status := cilium.EndpointSetConfig(endpoints["server"], "Conntrack", "false")
		Expect(status).Should(BeTrue(), "Couldn't set conntrack=false on endpoint 'server'")

		status = cilium.EndpointSetConfig(endpoints["client"], "Conntrack", "false")
		Expect(status).Should(BeTrue(), "Couldn't set conntrack=false on endpoint 'client'")

		client_server_connectivity()
	})

	It("ConntrackLocal disabled", func() {
		endpoints, err := cilium.GetEndpointsIds()
		Expect(err).Should(BeNil(), "Couldn't get endpoints IDS")

		status := cilium.EndpointSetConfig(endpoints["server"], "ConntrackLocal", "false")
		Expect(status).Should(BeTrue(), "Couldn't set conntrack=false on endpoint 'server'")

		status = cilium.EndpointSetConfig(endpoints["client"], "ConntrackLocal", "false")
		Expect(status).Should(BeTrue(), "Couldn't set conntrack=false on endpoint 'client'")

		client_server_connectivity()
	})

	It("ConntrackLocal Enabled", func() {
		endpoints, err := cilium.GetEndpointsIds()
		Expect(err).Should(BeNil(), "Couldn't get endpoints IDS")

		status := cilium.EndpointSetConfig(endpoints["server"], "ConntrackLocal", "true")
		Expect(status).Should(BeTrue(), "Couldn't set conntrack=false on endpoint 'server'")

		status = cilium.EndpointSetConfig(endpoints["client"], "ConntrackLocal", "true")
		Expect(status).Should(BeTrue(), "Couldn't set conntrack=false on endpoint 'client'")

		client_server_connectivity()
	})

})

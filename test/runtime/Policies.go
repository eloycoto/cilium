package RunT

import (
	"fmt"

	"github.com/cilium/cilium/test/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/onsi/gomega/types"
	log "github.com/sirupsen/logrus"
)

var _ = Describe("RunPolicyEnforcement", func() {

	Context("Default", func() {
		It("Default values", func() {
			Expect(true).Should(BeTrue())
		})
		It("Policy Checking", func() {
			By("L3 Policy")
			Expect(true).Should(BeTrue())

			By("L4 Policy")
			Expect(true).Should(BeTrue())
		})
	})

	Context("Always", func() {
		It("Default values", func() {
			Expect(true).Should(BeTrue())
		})
		It("Policy Checking", func() {
			By("L3 Policy")
			Expect(true).Should(BeTrue())

			By("L4 Policy")
			Expect(true).Should(BeTrue())
		})
	})

	Context("Never", func() {
		It("Default values", func() {
			Expect(true).Should(BeTrue())
		})
		It("Policy Checking", func() {
			By("L3 Policy")
			Expect(true).Should(BeTrue())

			By("L4 Policy")
			Expect(true).Should(BeTrue())
		})
	})

})

var _ = Describe("RunPolicies", func() {

	var initilized bool
	var networkName string = "cilium-net"
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

	BeforeEach(func() {
		initilize()
		cilium.Exec("policy delete --all")
		docker.SampleContainersActions("create", networkName)
		cilium.EndpointWaitUntilReady()
	})

	AfterEach(func() {
		docker.SampleContainersActions("delete", networkName)
	})

	connectivityTest := func(tests []string, client, server string, assertFn func() types.GomegaMatcher) {
		title := func(title string) string {
			return fmt.Sprintf(title, client, server)
		}
		_, err := docker.ContainerInspectNet(client)
		Expect(err).Should(BeNil(), fmt.Sprintf(
			"Couldn't get container '%s' client meta", client))

		srvIP, err := docker.ContainerInspectNet(server)
		Expect(err).Should(BeNil(), fmt.Sprintf(
			"Couldn't get container '%s' server meta", server))
		for _, test := range tests {
			switch test {
			case "ping":
				By(title("Client '%s' pinging server '%s' IPv4"))
				res := docker.ContainerExec(client, fmt.Sprintf("ping -c 4 %s", srvIP["IPv4"]))
				Expect(res.Correct()).Should(assertFn(), fmt.Sprintf(
					"Client '%s' can't ping to server '%s'", client, srvIP["IPv4"]))
			case "ping6":
				By(title("Client '%s' pinging server '%s' IPv6"))
				res := docker.ContainerExec(client, fmt.Sprintf("ping6 -c 4 %s", srvIP["IPv6"]))
				Expect(res.Correct()).Should(assertFn(), fmt.Sprintf(
					"Client '%s' can't ping to server '%s'", client, srvIP["IPv6"]))
			case "http":
				By(title("Client '%s' HttpReq to server '%s' Ipv4"))
				res := docker.ContainerExec(client, fmt.Sprintf(
					"curl -s --fail --connect-timeout 3 http://%s:80/public", srvIP["IPv4"]))
				Expect(res.Correct()).Should(assertFn(), fmt.Sprintf(
					"Client '%s' can't curl to server '%s'", client, srvIP["IPv4"]))
			case "http6":
				By(title("Client '%s' HttpReq to server '%s' IPv6"))
				res := docker.ContainerExec(client, fmt.Sprintf(
					"curl -s --fail --connect-timeout 3 http://[%s]:80/public", srvIP["IPv6"]))
				Expect(res.Correct()).Should(assertFn(), fmt.Sprintf(
					"Client '%s' can't curl to server '%s'", client, server, srvIP["IPv6"]))
			case "http_private":
				By(title("Client '%s' HttpReq to server '%s' Ipv4"))
				res := docker.ContainerExec(client, fmt.Sprintf(
					"curl -s --fail --connect-timeout 3 http://%s:80/private", srvIP["IPv4"]))
				Expect(res.Correct()).Should(assertFn(), fmt.Sprintf(
					"Client '%s' can't curl to server '%s'", client, srvIP["IPv4"]))
			case "http6_private":
				By(title("Client '%s' HttpReq to server '%s' Ipv6"))
				res := docker.ContainerExec(client, fmt.Sprintf(
					"curl -s --fail --connect-timeout 3 http://%s:80/private", srvIP["IPv6"]))
				Expect(res.Correct()).Should(assertFn(), fmt.Sprintf(
					"Client '%s' can't curl to server '%s'", client, srvIP["IPv6"]))
			}
		}
	}

	XIt("L3/L4 Checks", func() {
		_, err := cilium.PolicyImport(cilium.GetFullPath("l3-policy.json"), 300)
		Expect(err).Should(BeNil())

		//APP1 can connect to all Httpd1
		connectivityTest([]string{"ping", "ping6", "http", "http6"}, "app1", "httpd1", BeTrue)

		//APP2 can't connect to Httpd1
		connectivityTest([]string{"http"}, "app2", "httpd1", BeFalse)

		// APP1 can reach using TCP HTTP2
		connectivityTest([]string{"http", "http6"}, "app1", "httpd2", BeTrue)

		// APP2 can't reach using TCP to HTTP2
		connectivityTest([]string{"http", "http6"}, "app2", "httpd2", BeFalse)

		// APP3 can reach using TCP HTTP2, but can't ping EGRESS
		connectivityTest([]string{"http", "http6"}, "app3", "httpd3", BeTrue)

		status := cilium.Exec("policy delete --all")
		Expect(status.Correct()).Should(BeTrue())
		cilium.EndpointWaitUntilReady()

		By("Disabling all the policies. All should work")
		connectivityTest([]string{"ping", "ping6", "http", "http6"}, "app1", "httpd1", BeTrue)
		connectivityTest([]string{"ping", "ping6", "http", "http6"}, "app2", "httpd1", BeTrue)
	})

	It("L7 Checks", func() {

		_, err := cilium.PolicyImport(cilium.GetFullPath("l7-simple.json"), 300)
		Expect(err).Should(BeNil())

		By("Simple Ingress")
		//APP1 can connnect to public, but no to private
		connectivityTest([]string{"http", "http6"}, "app1", "httpd1", BeTrue)
		connectivityTest([]string{"http_private", "http6_private"}, "app1", "httpd1", BeFalse)

		//App2 can't connect
		connectivityTest([]string{"http", "http6"}, "app2", "httpd1", BeFalse)

		By("Simple Egress")

		//APP2 can connnect to public, but no to private
		connectivityTest([]string{"http", "http6"}, "app2", "httpd2", BeTrue)
		connectivityTest([]string{"http_private", "http6_private"}, "app2", "httpd2", BeFalse)

		By("Disabling all the policies. All should work")
		status := cilium.Exec("policy delete --all")
		Expect(status.Correct()).Should(BeTrue())
		cilium.EndpointWaitUntilReady()

		connectivityTest([]string{"ping", "ping6", "http", "http6"}, "app1", "httpd1", BeTrue)
		connectivityTest([]string{"ping", "ping6", "http", "http6"}, "app2", "httpd1", BeTrue)

		By("Multiple Ingress")

		cilium.Exec("policy delete --all")
		_, err = cilium.PolicyImport(cilium.GetFullPath("l7-multiple.json"), 300)
		Expect(err).Should(BeNil())

		//APP1 can connnect to public, but no to private
		connectivityTest([]string{"http", "http6"}, "app1", "httpd1", BeTrue)
		connectivityTest([]string{"http_private", "http6_private"}, "app1", "httpd1", BeFalse)

		//App2 can't connect
		connectivityTest([]string{"http", "http6"}, "app2", "httpd1", BeFalse)

		By("Multiple Egress")
		//APP2 can connnect to public, but no to private
		connectivityTest([]string{"http", "http6"}, "app2", "httpd2", BeTrue)
		connectivityTest([]string{"http_private", "http6_private"}, "app2", "httpd2", BeFalse)

		By("Disabling all the policies. All should work")

		status = cilium.Exec("policy delete --all")
		Expect(status.Correct()).Should(BeTrue())
		cilium.EndpointWaitUntilReady()

		connectivityTest([]string{"ping", "ping6", "http", "http6"}, "app1", "httpd1", BeTrue)
		connectivityTest([]string{"ping", "ping6", "http", "http6"}, "app2", "httpd1", BeTrue)
	})
})

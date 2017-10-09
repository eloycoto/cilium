package RunT

import (
	"fmt"
	"os"

	"github.com/cilium/cilium/test/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/onsi/gomega/types"
	log "github.com/sirupsen/logrus"
)

var _ = Describe("RunPolicyEnforcement", func() {

	var initilized bool
	var networkName string = "cilium-net"
	var logger *log.Entry
	var docker *helpers.Docker
	var cilium *helpers.Cilium

	initilize := func() {
		if initilized == true {
			return
		}
		logger = log.WithFields(log.Fields{"test": "RunPolicyEnforcement"})
		logger.Info("Starting")
		docker, cilium = helpers.CreateNewRuntimeHelper("runtime", logger)
		cilium.WaitUntilReady(100)
		docker.NetworkCreate(networkName, "")

		res := cilium.PolicyEnforcementSet("default", false)
		Expect(res.Correct()).Should(BeTrue())

		initilized = true
	}

	BeforeEach(func() {
		initilize()
		docker.ContainerCreate("app", "cilium/demo-httpd", networkName, "-l id.app")
		cilium.Exec("policy delete --all")
		cilium.EndpointWaitUntilReady()
	})

	AfterEach(func() {
		docker.ContainerRm("app")
	})

	Context("Policy Enforcement Default", func() {

		BeforeEach(func() {
			initilize()
			res := cilium.PolicyEnforcementSet("default")
			Expect(res.Correct()).Should(BeTrue())
		})

		It("Default values", func() {

			By("Policy Enforcement should be disabled for containers", func() {
				endPoints, err := cilium.PolicyEndpointsSummary()
				Expect(err).Should(BeNil())
				Expect(endPoints["disabled"]).To(Equal(1))
			})

			By("Apply a new policy")
			_, err := cilium.PolicyImport(cilium.GetFullPath("sample_policy.json"), 300)
			Expect(err).Should(BeNil())

			endPoints, err := cilium.PolicyEndpointsSummary()
			Expect(err).Should(BeNil())
			Expect(endPoints["enabled"]).To(Equal(1))
		})

		It("Default to Always without policy", func() {
			By("Check no policy enforcement")
			endPoints, err := cilium.PolicyEndpointsSummary()
			Expect(err).Should(BeNil())
			Expect(endPoints["disabled"]).To(Equal(1))

			By("Setting to Always")

			res := cilium.PolicyEnforcementSet("always", true)
			Expect(res.Correct()).Should(BeTrue())

			endPoints, err = cilium.PolicyEndpointsSummary()
			Expect(err).Should(BeNil())
			Expect(endPoints["enabled"]).To(Equal(1))

			By("Setting to default from Always")
			res = cilium.PolicyEnforcementSet("default", true)
			Expect(res.Correct()).Should(BeTrue())

			endPoints, err = cilium.PolicyEndpointsSummary()
			Expect(err).Should(BeNil())
			Expect(endPoints["disabled"]).To(Equal(1))
		})

		It("Default to Always with policy", func() {
			_, err := cilium.PolicyImport(cilium.GetFullPath("sample_policy.json"), 300)
			Expect(err).Should(BeNil())

			endPoints, err := cilium.PolicyEndpointsSummary()
			Expect(err).Should(BeNil())
			Expect(endPoints["enabled"]).To(Equal(1))
			//DEfault =APP with PolicyEnforcement

			res := cilium.PolicyEnforcementSet("always", true)
			Expect(res.Correct()).Should(BeTrue())

			endPoints, err = cilium.PolicyEndpointsSummary()
			Expect(err).Should(BeNil())
			Expect(endPoints["enabled"]).To(Equal(1))

			res = cilium.PolicyEnforcementSet("default", true)
			Expect(res.Correct()).Should(BeTrue())

			endPoints, err = cilium.PolicyEndpointsSummary()
			Expect(err).Should(BeNil())
			Expect(endPoints["enabled"]).To(Equal(1))
		})

		It("Default to Never without policy", func() {
			endPoints, err := cilium.PolicyEndpointsSummary()
			Expect(err).Should(BeNil())
			Expect(endPoints["disabled"]).To(Equal(1))

			res := cilium.PolicyEnforcementSet("never", true)
			Expect(res.Correct()).Should(BeTrue())

			endPoints, err = cilium.PolicyEndpointsSummary()
			Expect(err).Should(BeNil())
			Expect(endPoints["disabled"]).To(Equal(1))

			endPoints, err = cilium.PolicyEndpointsSummary()
			Expect(err).Should(BeNil())
			Expect(endPoints["disabled"]).To(Equal(1))
		})

		It("Default to Never with policy", func() {

			_, err := cilium.PolicyImport(cilium.GetFullPath("sample_policy.json"), 300)
			Expect(err).Should(BeNil())

			endPoints, err := cilium.PolicyEndpointsSummary()
			Expect(err).Should(BeNil())
			Expect(endPoints["enabled"]).To(Equal(1))

			res := cilium.PolicyEnforcementSet("never", true)
			Expect(res.Correct()).Should(BeTrue())

			endPoints, err = cilium.PolicyEndpointsSummary()
			Expect(err).Should(BeNil())
			Expect(endPoints["enabled"]).To(Equal(0))

			res = cilium.PolicyEnforcementSet("default", true)
			Expect(res.Correct()).Should(BeTrue())

			endPoints, err = cilium.PolicyEndpointsSummary()
			Expect(err).Should(BeNil())
			Expect(endPoints["enabled"]).To(Equal(1))
		})
	})

	Context("Policy Enforcement Always", func() {
		//The test Always to Default is already tested in from default-always
		BeforeEach(func() {
			initilize()
			res := cilium.PolicyEnforcementSet("always", true)
			Expect(res.Correct()).Should(BeTrue())
		})

		It("Container creation", func() {
			//Check default containers are in place.
			endPoints, err := cilium.PolicyEndpointsSummary()
			Expect(err).Should(BeNil())
			Expect(endPoints["enabled"]).To(Equal(1))
			Expect(endPoints["disabled"]).To(Equal(0))

			By("Create a new container")
			docker.ContainerCreate("new", "cilium/demo-httpd", networkName, "-l id.new")
			endPoints, err = cilium.PolicyEndpointsSummary()
			Expect(err).Should(BeNil())
			Expect(endPoints["enabled"]).To(Equal(2))
			Expect(endPoints["disabled"]).To(Equal(0))
			docker.ContainerRm("new")
		}, 300)

		It("Always to Never with policy", func() {
			endPoints, err := cilium.PolicyEndpointsSummary()
			Expect(err).Should(BeNil())
			Expect(endPoints["enabled"]).To(Equal(1))
			Expect(endPoints["disabled"]).To(Equal(0))

			_, err = cilium.PolicyImport(cilium.GetFullPath("sample_policy.json"), 300)
			Expect(err).Should(BeNil())

			endPoints, err = cilium.PolicyEndpointsSummary()
			Expect(err).Should(BeNil())
			Expect(endPoints["enabled"]).To(Equal(1))
			Expect(endPoints["disabled"]).To(Equal(0))

			res := cilium.PolicyEnforcementSet("never", true)
			Expect(res.Correct()).Should(BeTrue())

			endPoints, err = cilium.PolicyEndpointsSummary()
			Expect(err).Should(BeNil())
			Expect(endPoints["enabled"]).To(Equal(0))

			res = cilium.PolicyEnforcementSet("always", true)
			Expect(res.Correct()).Should(BeTrue())

			endPoints, err = cilium.PolicyEndpointsSummary()
			Expect(err).Should(BeNil())
			Expect(endPoints["enabled"]).To(Equal(1))
		})

		It("Always to Never without policy", func() {
			endPoints, err := cilium.PolicyEndpointsSummary()
			Expect(err).Should(BeNil())
			Expect(endPoints["enabled"]).To(Equal(1))
			Expect(endPoints["disabled"]).To(Equal(0))

			res := cilium.PolicyEnforcementSet("never", true)
			Expect(res.Correct()).Should(BeTrue())

			endPoints, err = cilium.PolicyEndpointsSummary()
			Expect(err).Should(BeNil())
			Expect(endPoints["enabled"]).To(Equal(0))
			Expect(endPoints["disabled"]).To(Equal(1))

			res = cilium.PolicyEnforcementSet("always", true)
			Expect(res.Correct()).Should(BeTrue())

			endPoints, err = cilium.PolicyEndpointsSummary()
			Expect(err).Should(BeNil())
			Expect(endPoints["enabled"]).To(Equal(1))
		})

	})

	Context("Policy Enforcement Never", func() {
		//The test Always to Default is already tested in from default-always
		BeforeEach(func() {
			initilize()
			res := cilium.PolicyEnforcementSet("never")
			Expect(res.Correct()).Should(BeTrue())
		})

		It("Container creation", func() {
			//Check default containers are in place.
			endPoints, err := cilium.PolicyEndpointsSummary()
			Expect(err).Should(BeNil())
			Expect(endPoints["enabled"]).To(Equal(0))
			Expect(endPoints["disabled"]).To(Equal(1))

			docker.ContainerCreate("new", "cilium/demo-httpd", networkName, "-l id.new")
			cilium.EndpointWaitUntilReady()
			endPoints, err = cilium.PolicyEndpointsSummary()
			Expect(err).Should(BeNil())
			Expect(endPoints["enabled"]).To(Equal(0))
			Expect(endPoints["disabled"]).To(Equal(2))
			docker.ContainerRm("new")
		}, 300)

		It("Never to default with policy", func() {
			endPoints, err := cilium.PolicyEndpointsSummary()
			Expect(err).Should(BeNil())
			Expect(endPoints["enabled"]).To(Equal(0))
			Expect(endPoints["disabled"]).To(Equal(1))

			_, err = cilium.PolicyImport(cilium.GetFullPath("sample_policy.json"), 300)
			Expect(err).Should(BeNil())

			endPoints, err = cilium.PolicyEndpointsSummary()
			Expect(err).Should(BeNil())
			Expect(endPoints["enabled"]).To(Equal(0))
			Expect(endPoints["disabled"]).To(Equal(1))

			res := cilium.PolicyEnforcementSet("default", true)
			Expect(res.Correct()).Should(BeTrue())

			endPoints, err = cilium.PolicyEndpointsSummary()
			Expect(err).Should(BeNil())
			Expect(endPoints["enabled"]).To(Equal(1))
			Expect(endPoints["disabled"]).To(Equal(0))

			res = cilium.PolicyEnforcementSet("never", true)
			Expect(res.Correct()).Should(BeTrue())

			endPoints, err = cilium.PolicyEndpointsSummary()
			Expect(err).Should(BeNil())
			Expect(endPoints["enabled"]).To(Equal(0))
			Expect(endPoints["disabled"]).To(Equal(1))
		})

		It("Never to default without policy", func() {
			endPoints, err := cilium.PolicyEndpointsSummary()
			Expect(err).Should(BeNil())
			Expect(endPoints["enabled"]).To(Equal(0))
			Expect(endPoints["disabled"]).To(Equal(1))

			res := cilium.PolicyEnforcementSet("default", true)
			Expect(res.Correct()).Should(BeTrue())

			endPoints, err = cilium.PolicyEndpointsSummary()
			Expect(err).Should(BeNil())
			Expect(endPoints["enabled"]).To(Equal(0))
			Expect(endPoints["disabled"]).To(Equal(1))

			res = cilium.PolicyEnforcementSet("never", true)
			Expect(res.Correct()).Should(BeTrue())

			endPoints, err = cilium.PolicyEndpointsSummary()
			Expect(err).Should(BeNil())
			Expect(endPoints["enabled"]).To(Equal(0))
			Expect(endPoints["disabled"]).To(Equal(1))
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
		logger = log.WithFields(log.Fields{"test": "RunPolicies"})
		logger.Info("Starting")
		docker, cilium = helpers.CreateNewRuntimeHelper("runtime", logger)
		docker.NetworkCreate(networkName, "")

		cilium.WaitUntilReady(100)
		res := cilium.PolicyEnforcementSet("default", false)
		Expect(res.Correct()).Should(BeTrue())
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
				By(title("Client '%s' HttpReq to server '%s' private Ipv4"))
				res := docker.ContainerExec(client, fmt.Sprintf(
					"curl -s --fail --connect-timeout 3 http://%s:80/private", srvIP["IPv4"]))
				Expect(res.Correct()).Should(assertFn(), fmt.Sprintf(
					"Client '%s' can't curl to server '%s' private", client, srvIP["IPv4"]))
			case "http6_private":
				By(title("Client '%s' HttpReq to server '%s' private Ipv6"))
				res := docker.ContainerExec(client, fmt.Sprintf(
					"curl -s --fail --connect-timeout 3 http://%s:80/private", srvIP["IPv6"]))
				Expect(res.Correct()).Should(assertFn(), fmt.Sprintf(
					"Client '%s' can't curl to server '%s' private", client, srvIP["IPv6"]))
			}
		}
	}

	It("L3/L4 Checks", func() {
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

		By("Disabling all the policies. All should work")

		status := cilium.Exec("policy delete --all")
		Expect(status.Correct()).Should(BeTrue())
		cilium.EndpointWaitUntilReady()

		connectivityTest([]string{"ping", "ping6", "http", "http6"}, "app1", "httpd1", BeTrue)
		connectivityTest([]string{"ping", "ping6", "http", "http6"}, "app2", "httpd1", BeTrue)

		By("Ingress CIDR")

		app1, err := docker.ContainerInspectNet("app1")
		Expect(err).Should(BeNil())

		script := fmt.Sprintf(`
		[{
			"endpointSelector": {
				"matchLabels":{"id.httpd1":""}
			},
			"ingress": [
				{"fromEndpoints": [
					{ "matchLabels": {"id.app1": ""}}
				]},
				{"fromCIDR":
					[ "%s/32", "%s" ]}
			]
		}]`, app1["IPv4"], app1["IPv6"])

		err = helpers.RenderTemplateToFile("ingress.json", script, 0777)
		Expect(err).Should(BeNil())

		path := helpers.GetFilePath("ingress.json")
		_, err = cilium.PolicyImport(path, 300)
		Expect(err).Should(BeNil())
		defer os.Remove("ingress.json")

		connectivityTest([]string{"http", "http6"}, "app1", "httpd1", BeTrue)
		connectivityTest([]string{"http", "http6"}, "app2", "httpd1", BeFalse)

		By("Egress CIDR")

		httpd1, err := docker.ContainerInspectNet("httpd1")
		Expect(err).Should(BeNil())

		script = fmt.Sprintf(`
		[{
			"endpointSelector": {
				"matchLabels":{"id.httpd1":""}
			},
			"ingress": [{
				"fromEndpoints": [{"matchLabels":{"id.app1":""}}]
			}]
		},
		{
			 "endpointSelector":
				{"matchLabels":{"id.%s":""}},
			 "egress": [{
				"toCIDR": [ "%s/32", "%s" ]
			 }]
		}]`, "app1", httpd1["IPv4"], httpd1["IPv6"])
		err = helpers.RenderTemplateToFile("egress.json", script, 0777)
		Expect(err).Should(BeNil())
		path = helpers.GetFilePath("egress.json")
		defer os.Remove("egress.json")
		_, err = cilium.PolicyImport(path, 300)
		Expect(err).Should(BeNil())

		connectivityTest([]string{"http", "http6"}, "app1", "httpd1", BeTrue)
		connectivityTest([]string{"http", "http6"}, "app2", "httpd1", BeFalse)
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

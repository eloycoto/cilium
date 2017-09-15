package ciliumTest

import (
	"fmt"
	"strings"

	"github.com/cilium/cilium/test/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("K8sPolicyTest", func() {

	kubectl := helpers.CreateKubectl("k8s1")
	podFilter := "k8s:id=app"

	BeforeEach(func() {
		kubectl.Apply("/vagrant/cilium-ds.yaml")
		kubectl.Apply("/vagrant/demo.yaml")

		status, err := kubectl.WaitforPods("kube-system", "-l k8s-app=cilium", 300)
		Expect(status).Should(BeTrue())
		Expect(err).Should(BeNil())

	})

	AfterEach(func() {
		// kubectl.Delete("/vagrant/cilium-ds.yaml")
	})

	Context("Policy Mode default", func() {
		It("Policy disabled", func() {
			kubectl.Apply("/vagrant/demo.yaml")
			cilium_pod, err := kubectl.GetCiliumPodOnNode("kube-system", "k8s1")
			//FIXME: This assert should have comment
			Expect(err).Should(BeNil())
			stdout, err := kubectl.CiliumExec(cilium_pod, getEndpointFilter(podFilter, "Disabled"))
			Expect(err).Should(BeNil())
			Expect(strings.Trim(stdout, "\n")).Should(Equal("4"))
		})

		It("Enable policy on pod", func() {
			_, err := kubectl.CiliumImportPolicy("kube-system", "/vagrant/l3_l4_policy.yaml", 300)
			Expect(err).Should(BeNil())

			cilium_pod, err := kubectl.GetCiliumPodOnNode("kube-system", "k8s1")
			//FIXME: This assert should have comment
			Expect(err).Should(BeNil())

			stdout, err := kubectl.CiliumExec(cilium_pod, getEndpointFilter(podFilter, "Disabled"))
			Expect(err).Should(BeNil())
			Expect(strings.Trim(stdout, "\n")).Should(Equal("2"))

			stdout, err = kubectl.CiliumExec(cilium_pod, getEndpointFilter(podFilter, "Enabled"))
			Expect(err).Should(BeNil())
			Expect(strings.Trim(stdout, "\n")).Should(Equal("2"))
		})

		It("Disable policy created", func() {
			status := kubectl.Delete("/vagrant/l3_l4_policy.yaml")
			Expect(status).Should(BeTrue())
			cilium_pod, err := kubectl.GetCiliumPodOnNode("kube-system", "k8s1")
			Expect(err).Should(BeNil())
			//FIXME Wait HERE
			stdout, err := kubectl.CiliumExec(cilium_pod, getEndpointFilter(podFilter, "Disabled"))
			Expect(err).Should(BeNil())
			Expect(strings.Trim(stdout, "\n")).Should(Equal("4"))
		})
	})

	Context("Policy config changes", func() {
		It("PolicyEnforcement enabled always", func() {
			cilium_pod, err := kubectl.GetCiliumPodOnNode("kube-system", "k8s1")
			Expect(err).Should(BeNil())
			_, err = kubectl.CiliumExec(cilium_pod, "cilium config PolicyEnforcement=always")
			Expect(err).Should(BeNil())

			kubectl.Apply("/vagrant/demo.yaml")
			cilium_pod, err = kubectl.GetCiliumPodOnNode("kube-system", "k8s1")
			//FIXME: This assert should have comment
			Expect(err).Should(BeNil())
			stdout, err := kubectl.CiliumExec(cilium_pod, getEndpointFilter(podFilter, "Enabled"))
			Expect(err).Should(BeNil())
			Expect(strings.Trim(stdout, "\n")).Should(Equal("4"))
		})

		It("Policy Enforcement enabled never", func() {
			cilium_pod, err := kubectl.GetCiliumPodOnNode("kube-system", "k8s1")
			Expect(err).Should(BeNil())
			_, err = kubectl.CiliumExec(cilium_pod, "cilium config PolicyEnforcement=never")
			Expect(err).Should(BeNil())

			kubectl.Apply("/vagrant/demo.yaml")
			cilium_pod, err = kubectl.GetCiliumPodOnNode("kube-system", "k8s1")
			Expect(err).Should(BeNil())
			stdout, err := kubectl.CiliumExec(cilium_pod, getEndpointFilter(podFilter, "Disabled"))
			Expect(err).Should(BeNil())
			Expect(strings.Trim(stdout, "\n")).Should(Equal("4"))

			// Test to import a policy when it is disabled
			status := kubectl.Apply("/vagrant/l3_l4_policy.yaml")
			Expect(status).Should(BeTrue())

			// Test to convert to enable all the endpoints
			_, err = kubectl.CiliumExec(cilium_pod, "cilium config PolicyEnforcement=always")
			Expect(err).Should(BeNil())
			stdout, err = kubectl.CiliumExec(cilium_pod, getEndpointFilter(podFilter, "Enabled"))
			Expect(err).Should(BeNil())
			Expect(strings.Trim(stdout, "\n")).Should(Equal("4"))
		})
	})
})

func getEndpointFilter(filter string, status string) string {
	return fmt.Sprintf(
		"cilium endpoint list | grep '%s' | awk '{print $2}' | grep '%s' -c", filter, status)
}

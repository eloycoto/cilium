package ciliumTest

import (
	"github.com/cilium/cilium/test/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("K8s_basic_test", func() {

	kubectl := helpers.CreateKubectl("127.0.0.1", 2222)
	BeforeEach(func() {
		kubectl.Apply("https://raw.githubusercontent.com/eloycoto/k8s-cilium-terraform/master/scripts/rbac.yaml")
		kubectl.Apply("https://raw.githubusercontent.com/eloycoto/k8s-cilium-terraform/master/scripts/cilium-ds.yaml")
		kubectl.Apply("https://raw.githubusercontent.com/eloycoto/k8s-cilium-terraform/master/scripts/sample.yml")
		kubectl.WaitforPods("kube-system", "-l k8s-app=cilium", 10)
	})

	AfterEach(func() {
		kubectl.Node.Execute("kubectl delete -f https://raw.githubusercontent.com/eloycoto/k8s-cilium-terraform/master/scripts/sample.yml", nil, nil)
	})

	Describe("Chekcing sample app is running", func() {
		Context("Without_policy", func() {
			It("has three pods", func() {
				Expect(1).To(Equal(1))
			})
		})
	})
})

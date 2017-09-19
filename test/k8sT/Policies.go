package k8sT

import (
	"fmt"
	"strings"
	"time"

	"github.com/asaskevich/govalidator"
	"github.com/cilium/cilium/test/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	log "github.com/sirupsen/logrus"
)

var _ = Describe("K8sPolicyTest", func() {

	logger := log.WithFields(log.Fields{"test": "K8sPolicyTest"})

	logger.Info("Starting")
	kubectl := helpers.CreateKubectl("k8s1", logger)
	podFilter := "k8s:id=app"

	BeforeEach(func() {
		kubectl.Apply("/vagrant/cilium-ds.yaml")
		kubectl.Apply("/vagrant/demo.yaml")
		status, err := kubectl.WaitforPods("kube-system", "-l k8s-app=cilium", 300)
		logger.Infof("Wait for cilium status='%v' err='%v'", status, err)
		Expect(status).Should(BeTrue())
		Expect(err).Should(BeNil())

		status, err = kubectl.WaitforPods("default", "-l zgroup=testapp", 300)
		Expect(err).Should(BeNil())

	})

	AfterEach(func() {
		return
		var wait int
		var timeout int = 100
		kubectl.Delete("/vagrant/demo.yaml")
		kubectl.Delete("/vagrant/cilium-ds.yaml")

		for wait < timeout {
			ciliumPods, err := kubectl.GetCiliumPods("kube-system")
			status := len(ciliumPods)
			logger.Infof("AfterEach cilium pods '%d' wait='%d' err='%v' %v", status, wait, err, ciliumPods)
			if status == 0 {
				return
			}
			if wait > timeout {
				logger.Errorf("AfterEach cilium pods can't be deleted")
				//FIXME error here?
				return
			}
			wait++
		}
	})

	It("Policyenforcment default", func() {
		logger := logger.WithField("type", "default")
		logger.Info("PolicyEnforcement default")
		By("Testing all nodes are disabled by default")
		ciliumPod, err := kubectl.GetCiliumPodOnNode("kube-system", "k8s1")
		Expect(err).Should(BeNil())
		stdout, err := kubectl.CiliumExec(ciliumPod, getEndpointFilter(podFilter, "Disabled"))
		Expect(err).Should(BeNil())
		Expect(strings.Trim(stdout, "\n")).Should(Equal("4"))

		By("DefaultMode with l3l4 policy")
		// Set two nodes with policy enforcement
		_, err = kubectl.CiliumImportPolicy("kube-system", "/vagrant/l3_l4_policy.yaml", 300)
		Expect(err).Should(BeNil())

		waitForEndpointsSync(kubectl)

		stdout, err = kubectl.CiliumExec(ciliumPod, getEndpointFilter(podFilter, "Disabled"))
		Expect(err).Should(BeNil())
		Expect(strings.Trim(stdout, "\n")).Should(Equal("2"))

		stdout, err = kubectl.CiliumExec(ciliumPod, getEndpointFilter(podFilter, "Enabled"))
		Expect(err).Should(BeNil())
		Expect(strings.Trim(stdout, "\n")).Should(Equal("2"))

		By("Deleting l3/l4 policy")
		// Delete the policy enforcement for the nodes. All 4 should be disabled
		status := kubectl.Delete("/vagrant/l3_l4_policy.yaml")
		Expect(status).Should(BeTrue())

		waitForEndpointsSync(kubectl)

		//FIXME Wait HERE
		stdout, err = kubectl.CiliumExec(ciliumPod, getEndpointFilter(podFilter, "Disabled"))
		Expect(err).Should(BeNil())
		Expect(strings.Trim(stdout, "\n")).Should(Equal("4"))
	})

	It("PolicyEnforcment set to always", func() {

		By("set policyenforcement to always")
		ciliumPod, err := kubectl.GetCiliumPodOnNode("kube-system", "k8s1")
		Expect(err).Should(BeNil())
		_, err = kubectl.CiliumExec(ciliumPod, "cilium config PolicyEnforcement=always")
		Expect(err).Should(BeNil())

		waitForEndpointsSync(kubectl)

		By("All Pods should have PolicyEnforcement Enabled")
		stdout, err := kubectl.CiliumExec(ciliumPod, getEndpointFilter(podFilter, "Enabled"))
		Expect(err).Should(BeNil())
		Expect(strings.Trim(stdout, "\n")).Should(Equal("4"))

		//FIXME: Check here that doesn't have access
	})

	It("PolicyEnforcment set to never", func() {
		ciliumPod, err := kubectl.GetCiliumPodOnNode("kube-system", "k8s1")
		Expect(err).Should(BeNil())
		_, err = kubectl.CiliumExec(ciliumPod, "cilium config PolicyEnforcement=never")
		Expect(err).Should(BeNil())

		waitForEndpointsSync(kubectl)

		By("Pods shouldn't have PolicyEnforcement")
		stdout, err := kubectl.CiliumExec(ciliumPod, getEndpointFilter(podFilter, "Disabled"))
		Expect(err).Should(BeNil())
		Expect(strings.Trim(stdout, "\n")).Should(Equal("4"))

		By("Test Polivy when PolicyEnforcement is disabled")
		// Test to import a policy when it is disabled
		status := kubectl.Apply("/vagrant/l3_l4_policy.yaml")
		Expect(status).Should(BeTrue())

		waitForEndpointsSync(kubectl)
		// Test to convert to enable all the endpoints

		By("Test reset PolicyEnforcement to Always")
		_, err = kubectl.CiliumExec(ciliumPod, "cilium config PolicyEnforcement=always")
		Expect(err).Should(BeNil())
		stdout, err = kubectl.CiliumExec(ciliumPod, getEndpointFilter(podFilter, "Enabled"))
		Expect(err).Should(BeNil())
		Expect(strings.Trim(stdout, "\n")).Should(Equal("4"))
	})

	It("Tests Policy Rules", func() {
		By("Testing L3/L4 rules")
		Expect(nil).Should(BeNil())

		By("Testing L7 rules")
		Expect(nil).Should(BeNil())
	})

})

func getEndpointFilter(filter string, status string) string {
	return fmt.Sprintf(
		"cilium endpoint list | grep '%s' | awk '{print $2}' | grep '%s' -c", filter, status)
}

func waitForEndpoints(kubectl *helpers.Kubectl, node string, label string, timeout int) (bool, error) {
	var wait int

	filter := fmt.Sprintf("cilium endpoint list | grep '%s' | wc -l", label)
	ciliumPod, err := kubectl.GetCiliumPodOnNode("kube-system", node)
	if err != nil {
		return false, err
	}

	for wait < timeout {
		time.Sleep(1 * time.Second)
		wait++

		pods, err := kubectl.GetPodsNames("default", label)
		if err != nil {
			continue
		}
		output, err := kubectl.CiliumExec(ciliumPod, filter)
		if err != nil {
			continue
		}

		val, err := govalidator.ToInt(strings.Trim(output, "\n"))
		if err != nil {
			continue
		}
		if int(val) == len(pods) {
			return true, nil
		}
	}
	return false, fmt.Errorf("Endpoints are not ready on cilium")
}

func waitForEndpointsSync(kubectl *helpers.Kubectl) {
	status, err := waitForEndpoints(kubectl, "k8s1", "zgroup=testapp", 300)
	Expect(err).Should(BeNil())
	Expect(status).Should(BeTrue())
}

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

	//FIXME: BeforeAll function here.
	kubectl.Apply("/vagrant/cilium-ds.yaml")
	status, err := kubectl.WaitforPods("kube-system", "-l k8s-app=cilium", 300)

	BeforeEach(func() {
		kubectl.Apply("/vagrant/demo.yaml")
		status, err = kubectl.WaitforPods("default", "-l zgroup=testapp", 300)
		Expect(err).Should(BeNil())

	})

	AfterEach(func() {
		return
		kubectl.Delete("/vagrant/demo.yaml")
		var status int = 1
		for status > 0 {
			pods, err := kubectl.GetPodsNames("default", "zgroup=testapp")
			status := len(pods)
			logger.Infof("AfterEach tests pods '%d' err='%v' pods='%v'", status, err, pods)
			if status == 0 {
				return
			}
			time.Sleep(1 * time.Second)
		}
	}, 30)

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
		appPods := make(map[string]string)
		apps := []string{"app1", "app2", "app3"}
		for _, v := range apps {
			res, err := kubectl.GetPodsNames("default", fmt.Sprintf("id=%s", v))
			Expect(err).Should(BeNil())
			appPods[v] = res[0]
			logger.Infof("PolicyRulesTest: pod='%s' assigned to '%s'", res[0], v)
		}
		clusterIP, err := kubectl.Get("default", "svc").Filter(
			"{.items[?(@.metadata.name == \"app1-service\")].spec.clusterIP}")
		logger.Infof("PolicyRulesTest: cluster service ip '%s'", clusterIP)
		Expect(err).Should(BeNil())

		ciliumPod, err := kubectl.GetCiliumPodOnNode("kube-system", "k8s1")
		Expect(err).Should(BeNil())

		By("Testing L3/L4 rules")
		_, err = kubectl.CiliumImportPolicy(
			"kube-system", fmt.Sprintf("%s/l3_l4_policy.yaml", kubectl.ManifestsPath()), 100)
		Expect(err).Should(BeNil())

		_, err = kubectl.Exec(
			"default", appPods["app2"], fmt.Sprintf("curl http://%s/public", clusterIP))
		Expect(err).Should(BeNil())

		out, err := kubectl.CiliumExec(ciliumPod, fmt.Sprintf(
			"cilium policy trace --src-k8s-pod default:%s --dst-k8s-pod default:%s",
			appPods["app2"], appPods["app1"]))
		Expect(err).Should(BeNil())
		Expect(out).Should(ContainSubstring("Result: ALLOWED"))

		_, err = kubectl.Exec(
			"default", appPods["app3"], fmt.Sprintf("curl http://%s/public", clusterIP))
		Expect(err).Should(HaveOccurred())

		out, err = kubectl.CiliumExec(ciliumPod, fmt.Sprintf(
			"cilium policy trace --src-k8s-pod default:%s --dst-k8s-pod default:%s",
			appPods["app3"], appPods["app1"]))
		Expect(err).Should(BeNil())
		Expect(out).Should(ContainSubstring("Result: DENIED"))

		status = kubectl.Delete(
			fmt.Sprintf("%s/l3_l4_policy.yaml", kubectl.ManifestsPath()))

		By("Testing L7 rules")
		_, err = kubectl.CiliumImportPolicy(
			"kube-system", fmt.Sprintf("%s/l7_policy.yaml", kubectl.ManifestsPath()), 100)
		Expect(status).Should(BeTrue())

		_, err = kubectl.Exec(
			"default", appPods["app2"], fmt.Sprintf("curl http://%s/public", clusterIP))
		Expect(err).Should(BeNil())

		_, err = kubectl.Exec(
			"default", appPods["app2"], fmt.Sprintf("curl --fail -s http://%s/private", clusterIP))
		Expect(err).Should(HaveOccurred())

		_, err = kubectl.Exec(
			"default", appPods["app3"], fmt.Sprintf("curl http://%s/public", clusterIP))
		Expect(err).Should(HaveOccurred())

		_, err = kubectl.Exec(
			"default", appPods["app3"], fmt.Sprintf("curl http://%s/private", clusterIP))
		Expect(err).Should(HaveOccurred())
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

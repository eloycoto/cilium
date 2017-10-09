package k8sT

import (
	"fmt"

	"github.com/cilium/cilium/test/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	log "github.com/sirupsen/logrus"
)

var _ = Describe("K8sPolicyTest", func() {

	var demoPath string
	var initilized bool
	var kubectl *helpers.Kubectl
	var l3Policy, l7Policy string
	var logger *log.Entry
	var path string
	var podFilter string

	initilize := func() {
		if initilized == true {
			return
		}

		logger = log.WithFields(log.Fields{"test": "K8sPolicyTest"})
		logger.Info("Starting")
		kubectl = helpers.CreateKubectl("k8s1", logger)
		podFilter = "k8s:zgroup=testapp"

		//Manifest paths
		demoPath = fmt.Sprintf("%s/demo.yaml", kubectl.ManifestsPath())
		l3Policy = fmt.Sprintf("%s/l3_l4_policy.yaml", kubectl.ManifestsPath())
		l7Policy = fmt.Sprintf("%s/l7_policy.yaml", kubectl.ManifestsPath())

		path = fmt.Sprintf("%s/cilium_ds.yaml", kubectl.ManifestsPath())
		kubectl.Apply(path)
		status, err := kubectl.WaitforPods("kube-system", "-l k8s-app=cilium", 300)
		Expect(status).Should(BeTrue())
		Expect(err).Should(BeNil())
		initilized = true
	}

	BeforeEach(func() {
		initilize()
		kubectl.CiliumPolicyDeleteAll("kube-system")
		kubectl.Apply(demoPath)
		_, err := kubectl.WaitforPods("default", "-l zgroup=testapp", 300)
		Expect(err).Should(BeNil())
	})

	AfterEach(func() {
		return
		kubectl.Delete(demoPath)
	})

	It("PolicyEnforcement Changes", func() {
		//This is a small test that check that everything is working in k8s. Full monkey testing
		// is on runtime/Policies
		ciliumPod, err := kubectl.GetCiliumPodOnNode("kube-system", "k8s1")
		Expect(err).Should(BeNil())

		status := kubectl.CiliumExec(ciliumPod, "cilium config PolicyEnforcement=default")
		Expect(status.Correct()).Should(BeTrue())
		helpers.Sleep(5)
		kubectl.CiliumEndpointWait(ciliumPod)

		epsStatus := helpers.WithTimeout(func() bool {
			endpoints, err := kubectl.CiliumEndpointsGetByTag(ciliumPod, podFilter)
			if err != nil {
				return false
			}
			return endpoints.AreReady()
		}, "Couldn't get endpoints", &helpers.TimeoutConfig{Timeout: 100})
		Expect(epsStatus).Should(BeNil())

		endpoints, err := kubectl.CiliumEndpointsGetByTag(ciliumPod, podFilter)
		Expect(err).Should(BeNil())
		Expect(endpoints.AreReady()).Should(BeTrue())
		policyStatus := endpoints.GetPolicyStatus()
		Expect(policyStatus["enabled"]).Should(Equal(0))
		Expect(policyStatus["disabled"]).Should(Equal(4))

		By("Set PolicyEnforcement to always")

		status = kubectl.CiliumExec(ciliumPod, "cilium config PolicyEnforcement=always")
		Expect(status.Correct()).Should(BeTrue())
		kubectl.CiliumEndpointWait(ciliumPod)

		endpoints, err = kubectl.CiliumEndpointsGetByTag(ciliumPod, podFilter)
		Expect(err).Should(BeNil())
		Expect(endpoints.AreReady()).Should(BeTrue())
		policyStatus = endpoints.GetPolicyStatus()
		Expect(policyStatus["enabled"]).Should(Equal(4))
		Expect(policyStatus["disabled"]).Should(Equal(0))

		By("Return PolicyEnforcement to default")
		status = kubectl.CiliumExec(ciliumPod, "cilium config PolicyEnforcement=default")
		Expect(status.Correct()).Should(BeTrue())
		kubectl.CiliumEndpointWait(ciliumPod)

		endpoints, err = kubectl.CiliumEndpointsGetByTag(ciliumPod, podFilter)
		Expect(err).Should(BeNil())
		Expect(endpoints.AreReady()).Should(BeTrue())
		policyStatus = endpoints.GetPolicyStatus()
		Expect(policyStatus["enabled"]).Should(Equal(0))
		Expect(policyStatus["disabled"]).Should(Equal(4))
	}, 500)

	It("Policies", func() {
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

		status := kubectl.CiliumExec(ciliumPod, "cilium config PolicyEnforcement=default")
		Expect(status.Correct()).Should(BeTrue())
		kubectl.CiliumEndpointWait(ciliumPod)

		By("Testing L3/L4 rules")

		_, err = kubectl.CiliumImportPolicy("kube-system", l3Policy, 300)
		Expect(err).Should(BeNil())

		epsStatus := helpers.WithTimeout(func() bool {
			endpoints, err := kubectl.CiliumEndpointsGetByTag(ciliumPod, podFilter)
			if err != nil {
				return false
			}
			return endpoints.AreReady()
		}, "Couldn't get endpoints", &helpers.TimeoutConfig{Timeout: 100})

		Expect(epsStatus).Should(BeNil())

		endpoints, err := kubectl.CiliumEndpointsGetByTag(ciliumPod, podFilter)
		policyStatus := endpoints.GetPolicyStatus()
		Expect(policyStatus["enabled"]).Should(Equal(2))
		Expect(policyStatus["disabled"]).Should(Equal(2))

		_, err = kubectl.Exec(
			"default", appPods["app2"], fmt.Sprintf("curl http://%s/public", clusterIP))
		Expect(err).Should(BeNil())

		trace := kubectl.CiliumExec(ciliumPod, fmt.Sprintf(
			"cilium policy trace --src-k8s-pod default:%s --dst-k8s-pod default:%s --dport 80",
			appPods["app2"], appPods["app1"]))

		Expect(trace.Correct()).Should(BeTrue())
		Expect(trace.Output().String()).Should(ContainSubstring("Verdict: ALLOWED"))

		trace = kubectl.CiliumExec(ciliumPod, fmt.Sprintf(
			"cilium policy trace --src-k8s-pod default:%s --dst-k8s-pod default:%s",
			appPods["app3"], appPods["app1"]))
		Expect(trace.Correct()).Should(BeTrue())

		Expect(trace.Output().String()).Should(ContainSubstring("Verdict: DENIED"))
		_, err = kubectl.Exec(
			"default", appPods["app3"], fmt.Sprintf("curl --fail -s http://%s/public", clusterIP))
		Expect(err).Should(HaveOccurred())

		status = kubectl.Delete(l3Policy)
		Expect(status.Correct()).Should(BeTrue())
		kubectl.CiliumEndpointWait(ciliumPod)

		By("Testing L7 Policy")
		//All Monkey testing in this section is on runtime

		_, err = kubectl.CiliumImportPolicy("kube-system", l7Policy, 300)
		Expect(err).Should(BeNil())

		_, err = kubectl.Exec(
			"default", appPods["app2"], fmt.Sprintf("curl http://%s/public", clusterIP))
		Expect(err).Should(BeNil())

		msg, err := kubectl.Exec(
			"default", appPods["app2"], fmt.Sprintf("curl --fail -s http://%s/private", clusterIP))
		Expect(err).Should(HaveOccurred(), msg)

		_, err = kubectl.Exec(
			"default", appPods["app3"], fmt.Sprintf("curl -s --fail http://%s/public", clusterIP))
		Expect(err).Should(HaveOccurred())

		msg, err = kubectl.Exec(
			"default", appPods["app3"], fmt.Sprintf("curl --fail -s http://%s/private", clusterIP))
		Expect(err).Should(HaveOccurred(), msg)

		status = kubectl.Delete(l7Policy)
		Expect(status.Correct()).Should(BeTrue())
		kubectl.CiliumEndpointWait(ciliumPod)

		//After disable the policy, app3 can reach app1

		_, err = kubectl.Exec(
			"default", appPods["app3"], fmt.Sprintf("curl -s --fail http://%s/public", clusterIP))
		Expect(err).Should(BeNil())
	}, 500)
})

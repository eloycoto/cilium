package k8sT

import (
	"fmt"
	"strings"
	"time"

	"github.com/cilium/cilium/test/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	log "github.com/sirupsen/logrus"
)

var _ = Describe("K8sTunnelTest", func() {

	var kubectl *helpers.Kubectl
	var demoDSPath string
	var logger *log.Entry
	var initilized bool

	initilize := func() {
		if initilized == true {
			return
		}

		logger = log.WithFields(log.Fields{"test": "K8sTunnelTest"})
		logger.Info("Starting")

		kubectl = helpers.CreateKubectl("k8s1", logger)
		demoDSPath = fmt.Sprintf("%s/demo_ds.yaml", kubectl.ManifestsPath())
		initilized = true
	}

	BeforeEach(func() {
		initilize()
		//We use a DaemonSet to ensure that all the pods are in different node
		kubectl.Apply(demoDSPath)
	})

	AfterEach(func() {
		kubectl.Delete(demoDSPath)
	})

	It("Check VXLAN mode", func() {
		path := fmt.Sprintf("%s/cilium_ds.yaml", kubectl.ManifestsPath())
		kubectl.Apply(path)
		_, err := kubectl.WaitforPods("kube-system", "-l k8s-app=cilium", 300)
		Expect(err).Should(BeNil())

		ciliumPod, err := kubectl.GetCiliumPodOnNode("kube-system", "k8s1")
		Expect(err).Should(BeNil())

		//Check that cilium detects a
		By("Checking that BPF tunnels are in place")
		tunnelList, err := kubectl.CiliumExec(ciliumPod, "cilium bpf tunnel list | wc -l")
		Expect(err).Should(BeNil())
		Expect(strings.Trim(tunnelList, "\n")).Should(Equal("3"))

		By("Checking that BPF tunnels are working correctly")
		tunnStatus := isNodeNetworkingWorking(kubectl, "zgroup=testDS")
		Expect(tunnStatus).Should(BeTrue())
		kubectl.Delete(path)
		var status int = 1
		for status > 0 {
			pods, err := kubectl.GetCiliumPods("kube-system")
			status := len(pods)
			logger.Infof("VXLAN Mode pods termintating '%d' err='%v' pods='%v'", status, err, pods)
			if status == 0 {
				return
			}
			time.Sleep(1 * time.Second)
		}
	}, 600)

	It("Check Geneve mode", func() {
		path := fmt.Sprintf("%s/cilium_ds_geneve.yaml", kubectl.ManifestsPath())
		kubectl.Apply(path)
		_, err := kubectl.WaitforPods("kube-system", "-l k8s-app=cilium", 300)
		Expect(err).Should(BeNil())

		ciliumPod, err := kubectl.GetCiliumPodOnNode("kube-system", "k8s1")
		Expect(err).Should(BeNil())

		//Check that cilium detects a
		By("Checking that BPF tunnels are in place")
		tunnelList, err := kubectl.CiliumExec(ciliumPod, "cilium bpf tunnel list | wc -l")
		Expect(err).Should(BeNil())
		Expect(strings.Trim(tunnelList, "\n")).Should(Equal("3"))

		By("Checking that BPF tunnels are working correctly")
		tunnStatus := isNodeNetworkingWorking(kubectl, "zgroup=testDS")
		Expect(tunnStatus).Should(BeTrue())
		//FIXME: Maybe added here a cilium bpf tunnel status?
		kubectl.Delete(path)
		var status int = 1
		for status > 0 {
			pods, err := kubectl.GetCiliumPods("kube-system")
			status := len(pods)
			logger.Infof("VXLAN Mode pods termintating '%d' err='%v' pods='%v'", status, err, pods)
			if status == 0 {
				return
			}
			time.Sleep(1 * time.Second)
		}
	}, 600)
})

func isNodeNetworkingWorking(kubectl *helpers.Kubectl, filter string) bool {
	pods, err := kubectl.GetPodsNames("default", filter)
	Expect(err).Should(BeNil())
	podIP, err := kubectl.Get(
		"default",
		fmt.Sprintf("pod %s -o json", pods[1])).Filter("{.status.podIP}")
	Expect(err).Should(BeNil())
	_, err = kubectl.Exec("default", pods[0], fmt.Sprintf("ping -c 1 %s", podIP))
	if err != nil {
		return false
	}
	return true
}

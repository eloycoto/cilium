package k8sT

import (
	"fmt"
	"time"

	"github.com/cilium/cilium/test/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	log "github.com/sirupsen/logrus"
)

var _ = Describe("K8sTunnelTest", func() {
	logger := log.WithFields(log.Fields{"test": "K8sTunnelTest"})
	logger.Info("Starting")

	kubectl := helpers.CreateKubectl("k8s1", logger)
	demoDSPath := fmt.Sprintf("%s/demo_ds.yaml", kubectl.ManifestsPath())

	BeforeEach(func() {
		kubectl.Apply(demoDSPath)
	})

	AfterEach(func() {
		kubectl.Delete(demoDSPath)
	})

	FIt("Check VXLAN mode", func() {
		path := fmt.Sprintf("%s/cilium_ds.yaml", kubectl.ManifestsPath())
		kubectl.Apply(path)
		_, err := kubectl.WaitforPods("kube-system", "-l k8s-app=cilium", 300)
		Expect(err).Should(BeNil())

		//FIXME: Maybe added here a cilium bpf tunnel status?
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

	FIt("Check Geneve mode", func() {
		path := fmt.Sprintf("%s/cilium_geneve.yaml", kubectl.ManifestsPath())
		kubectl.Apply(path)
		_, err := kubectl.WaitforPods("kube-system", "-l k8s-app=cilium", 300)
		Expect(err).Should(BeNil())

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

package k8sT

import (
	"fmt"

	"github.com/cilium/cilium/test/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	log "github.com/sirupsen/logrus"
)

var _ = Describe("K8sTunnelTest", func() {

	var kubectl *helpers.Kubectl
	var logger *log.Entry
	var initilized bool

	initilize := func() {
		if initilized == true {
			return
		}
		logger = log.WithFields(log.Fields{"test": "K8sServiceTest"})
		logger.Info("Starting")

		kubectl = helpers.CreateKubectl("k8s1", logger)
		initilized = true
	}

	BeforeEach(func() {
		initilize()
	})
	// demoDSPath := fmt.Sprintf("%s/demo_ds.yaml", kubectl.ManifestsPath())

	FIt("Check Service", func() {
		fmt.Println(kubectl)
		Expect(nil).Should(BeNil())
	})

})

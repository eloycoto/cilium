package k8sT

import (
	"fmt"

	"github.com/asaskevich/govalidator"
	"github.com/cilium/cilium/test/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	log "github.com/sirupsen/logrus"
)

var _ = Describe("K8sTunnelTest", func() {

	var kubectl *helpers.Kubectl
	var logger *log.Entry
	var initilized bool
	var serviceName string = "app1-service"

	initilize := func() {
		if initilized == true {
			return
		}
		logger = log.WithFields(log.Fields{"test": "K8sServiceTest"})
		logger.Info("Starting")

		kubectl = helpers.CreateKubectl("k8s1", logger)
		path := fmt.Sprintf("%s/cilium_ds.yaml", kubectl.ManifestsPath())
		kubectl.Apply(path)
		_, err := kubectl.WaitforPods("kube-system", "-l k8s-app=cilium", 300)
		Expect(err).Should(BeNil())
		initilized = true
	}

	BeforeEach(func() {
		initilize()
		demoDSPath := fmt.Sprintf("%s/demo.yaml", kubectl.ManifestsPath())
		kubectl.Apply(demoDSPath)
		status, err := kubectl.WaitforPods("default", "-l zgroup=testapp", 300)
		Expect(status).Should(BeTrue())
		Expect(err).Should(BeNil())

	})

	It("Check Service", func() {
		svcIP, err := kubectl.Get(
			"default", fmt.Sprintf("service %s", serviceName)).Filter("{.spec.clusterIP}")
		Expect(err).Should(BeNil())
		Expect(govalidator.IsIP(svcIP.String())).Should(BeTrue())

		status := kubectl.Node.Execute(fmt.Sprintf("curl http://%s/", svcIP), nil, nil)
		Expect(status).Should(BeTrue())

		//FIXME: Validate here that the cilium service list -o jsonpath='{.svc.IP:80}"'
	})

	//FIXME: Check service with IPV6
	//FIXME: Check the service with cross-node
})

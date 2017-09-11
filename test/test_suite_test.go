package ciliumTest

import (
	"fmt"
	"os"

	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/reporters"
	. "github.com/onsi/gomega"

	"testing"

	ginkgoext "github.com/cilium/cilium/test/ginkgo-ext"
	"github.com/cilium/cilium/test/helpers"
	log "github.com/sirupsen/logrus"
)

var DefaultSettings map[string]string = map[string]string{
	"K8S_VERSION": "1.7",
}

var vagrant helpers.Vagrant

func init() {
	// log.SetOutput(os.Stdout)
	log.SetOutput(GinkgoWriter)
	log.SetLevel(log.DebugLevel)
	log.SetFormatter(&log.TextFormatter{
		DisableTimestamp: true,
	})

	// var filename string = "test.log"
	// f, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0755)
	// if err != nil {
	// 	fmt.Printf("Can't create the log file")
	// 	os.Exit(1)
	// }
	// log.SetOutput(f)

	for k, v := range DefaultSettings {
		getOrSetEnvVar(k, v)
	}
}

func TestTest(t *testing.T) {
	RegisterFailHandler(Fail)
	junitReporter := reporters.NewJUnitReporter(fmt.Sprintf(
		"%s.xml", ginkgoext.GetScopeWithVersion()))
	RunSpecsWithDefaultAndCustomReporters(
		t, ginkgoext.GetScopeWithVersion(), []Reporter{junitReporter})
}

var _ = BeforeSuite(func() {
	scope := ginkgoext.GetScope()
	switch scope {
	case "runtime":
		vagrant.Create("runtime", true)
	case "k8s":
		//FIXME: This should be:
		// Start k8s1 and provision kubernetes.
		// When finish, start to build cilium in background
		// Start k8s2
		// Wait until compilation finished, and pull cilium image on k8s2
		vagrant.Create(fmt.Sprintf("k8s1-%s", helpers.GetCurrentK8SEnv()), true)
		vagrant.Create(fmt.Sprintf("k8s2-%s", helpers.GetCurrentK8SEnv()), false)
	}
	return
})

var _ = AfterSuite(func() {
	return

	scope := ginkgoext.GetScope()
	log.Info("Running After Suite flag for scope='%s'", scope)
	switch scope {
	case "runtime":
		vagrant.Destroy("runtime")
	case "k8s":
		vagrant.Destroy(fmt.Sprintf("k8s1-%s", helpers.GetCurrentK8SEnv()))
		vagrant.Destroy(fmt.Sprintf("k8s2-%s", helpers.GetCurrentK8SEnv()))
	}
	return
})

func getOrSetEnvVar(key, value string) {
	if val := os.Getenv(key); val == "" {
		log.Infof("Init: Env var '%s' was not set, set default value '%s'", key, value)
		os.Setenv(key, value)
	}
}

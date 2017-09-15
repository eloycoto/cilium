package ciliumTest

import (
	"fmt"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"

	"github.com/cilium/cilium/test/helpers"
	log "github.com/sirupsen/logrus"
)

func init() {
	var filename string = "test.log"

	log.SetOutput(os.Stdout)
	log.SetLevel(log.DebugLevel)

	f, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE, 0755)
	if err != nil {
		fmt.Printf("Can't create the log file")
		os.Exit(1)
	}
	log.SetOutput(f)
}

func TestTest(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Cilium Test")
}

var vagrant helpers.Vagrant

var _ = BeforeSuite(func() {
	// This runs when the tests start, before all test
	log.Info("Running Before suite flag")
	vagrant.Create()
	return
})

var _ = AfterSuite(func() {
	// This runs when all the test finished
	// vagrant.Destroy()
})

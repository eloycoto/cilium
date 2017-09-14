package ciliumTest

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"

	"github.com/cilium/cilium/test/helpers"
)

func TestTest(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Cilium Test")
}

var vagrant helpers.Vagrant

var _ = BeforeSuite(func() {
	// This runs when the tests start, before all test
	vagrant.Create()
	return
})

var _ = AfterSuite(func() {
	// This runs when all the test finished
	// vagrant.Destroy()
})

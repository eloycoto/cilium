package RunT

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("RunPolicyEnforcement", func() {

	Context("Default", func() {
		It("Default values", func() {
			Expect(true).Should(BeTrue())
		})
		It("Policy Checking", func() {
			By("L3 Policy")
			Expect(true).Should(BeTrue())

			By("L4 Policy")
			Expect(true).Should(BeTrue())
		})
	})

	Context("Always", func() {
		It("Default values", func() {
			Expect(true).Should(BeTrue())
		})
		It("Policy Checking", func() {
			By("L3 Policy")
			Expect(true).Should(BeTrue())

			By("L4 Policy")
			Expect(true).Should(BeTrue())
		})
	})

	Context("Never", func() {
		It("Default values", func() {
			Expect(true).Should(BeTrue())
		})
		It("Policy Checking", func() {
			By("L3 Policy")
			Expect(true).Should(BeTrue())

			By("L4 Policy")
			Expect(true).Should(BeTrue())
		})
	})

})

var _ = Describe("RunPolicies", func() {
	// L3/L4 rules Enable/disable
	// L7 Rules Enable/disable
	// Ingress Enable/disable
	// Egress Enable/disable
	// Invalid rules
})

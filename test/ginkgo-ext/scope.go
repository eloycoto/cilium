package ginkgoext

import (
	"fmt"
	"strings"

	"github.com/onsi/ginkgo/config"

	"github.com/cilium/cilium/test/helpers"
)

//GetScope return scope for running test
func GetScope() string {
	focusString := strings.ToLower(config.GinkgoConfig.FocusString)
	switch {
	case strings.HasPrefix(focusString, "run"):
		return "runtime"
	case strings.HasPrefix(focusString, "k8s"):
		return "k8s"
	default:
		return "runtime"
	}
}

//GetScopeWithVersion return the scope and if it is k8s with the version
func GetScopeWithVersion() string {
	result := GetScope()
	if result != "k8s" {
		return result
	}
	return fmt.Sprintf("%s-%s", result, helpers.GetCurrentK8SEnv())
}

package ginkgoext

import (
	"strings"

	"github.com/onsi/ginkgo/config"
)

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

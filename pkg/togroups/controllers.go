package togroups

import (
	"time"

	cilium_v2 "github.com/cilium/cilium/pkg/k8s/apis/cilium.io/v2"
)

const (
	maxConcurrentUpdates = 4
)

func init() {
	go UpgradeCNPInformation()
}

func UpgradeCNPInformation() {
	for {
		cnpToUpdate := ToGroupsCNPCache.GetAllCNP()
		sem := make(chan bool, maxConcurrentUpdates)
		for _, cnp := range cnpToUpdate {
			sem <- true
			go func(cnp *cilium_v2.CiliumNetworkPolicy) {
				defer func() { <-sem }()
				addChildCNP(cnp)
			}(cnp)
		}
		time.Sleep(10 * time.Second)
	}
}

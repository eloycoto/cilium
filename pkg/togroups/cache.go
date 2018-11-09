package togroups

import (
	"sync"

	cilium_v2 "github.com/cilium/cilium/pkg/k8s/apis/cilium.io/v2"
)

var ToGroupsCNPCache = toGroupsCNPCache{}

type toGroupsCNPCache struct {
	sync.Map
}

func (cnpCache *toGroupsCNPCache) UpdateCNP(cnp *cilium_v2.CiliumNetworkPolicy) {
	cnpCache.Store(cnp.ObjectMeta.UID, cnp)
}

func (cnpCache *toGroupsCNPCache) DeleteCNP(cnp *cilium_v2.CiliumNetworkPolicy) {
	cnpCache.Delete(cnp.ObjectMeta.UID)
}

func (cnpCache *toGroupsCNPCache) GetAllCNP() []*cilium_v2.CiliumNetworkPolicy {
	result := []*cilium_v2.CiliumNetworkPolicy{}
	cnpCache.Range(func(k, v interface{}) bool {
		result = append(result, v.(*cilium_v2.CiliumNetworkPolicy))
		return true
	})
	return result
}

// Copyright 2017-2018 Authors of Cilium
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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

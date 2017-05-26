// Copyright 2016-2017 Authors of Cilium
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

package policy

import (
	"encoding/json"
	"net"
	"strconv"
	//	"github.com/cilium/cilium/api/v1/models"

	log "github.com/Sirupsen/logrus"
	"github.com/cilium/cilium/pkg/maps/cidrmap"
)

// L3PolicyMap is a list of CIDR filters indexable by address/prefixlen
// key format: "address/prefixlen", e.g., "10.1.1.0/24"
type L3PolicyMap struct {
	Map         map[string]net.IPNet // Allowed L3 prefixes
	IPv6Changed bool
	IPv6Count   int // Count of IPv6 prefixes in 'Map'
	IPv4Changed bool
	IPv4Count   int // Count of IPv4 prefixes in 'Map'
}

// Returns `1` if `cidr` is added to the map, `0` otherwise
func (m *L3PolicyMap) Insert(cidr string) int {
	_, ipnet, err := net.ParseCIDR(cidr)
	if err == nil {
		ones, _ := ipnet.Mask.Size()

		key := ipnet.IP.String() + "/" + strconv.Itoa(ones)

		if _, found := m.Map[key]; !found {
			m.Map[key] = *ipnet
			if ipnet.IP.To4() == nil {
				m.IPv6Count++
				m.IPv6Changed = true
			} else {
				m.IPv4Count++
				m.IPv4Changed = true
			}
			return 1
		}
	}

	return 0
}

func (m *L3PolicyMap) PopulateBPF(cidrmap *cidrmap.CIDRMap) error {
	for key, value := range m.Map {
		if value.IP.To4() == nil {
			if cidrmap.AddrSize != 16 {
				log.Warningf("Skipping IPv6 CIDR %s.", key)
				continue
			}
		} else {
			if cidrmap.AddrSize != 4 {
				log.Warningf("Skipping IPv4 CIDR %s.", key)
				continue
			}
		}
		log.Warningf("Allowing CIDR %s.", key)
		err := cidrmap.AllowCIDR(value)
		if err != nil {
			return err
		}
	}
	return nil
}

func (m *L3PolicyMap) GetModel() []string {
	str := []string{}
	for _, v := range m.Map {
		b, err := json.MarshalIndent(v, "", "  ")
		if err != nil {
			str = append(str, err.Error())
			return str
		}
		str = append(str, string(b))
	}

	return str
}

func (m L3PolicyMap) DeepCopy() L3PolicyMap {
	cpy := L3PolicyMap{
		Map:         make(map[string]net.IPNet, len(m.Map)),
		IPv6Changed: m.IPv6Changed,
		IPv6Count:   m.IPv6Count,
		IPv4Changed: m.IPv4Changed,
		IPv4Count:   m.IPv4Count,
	}
	for k, v := range m.Map {
		cpy.Map[k] = v
	}
	return cpy
}

type L3Policy struct {
	Ingress L3PolicyMap
	Egress  L3PolicyMap
}

func NewL3Policy() *L3Policy {
	return &L3Policy{
		Ingress: L3PolicyMap{Map: make(map[string]net.IPNet)},
		Egress:  L3PolicyMap{Map: make(map[string]net.IPNet)},
	}
}

//func (l3 *L3Policy) GetModel() *models.L3Policy {
//	if l3 == nil {
//		return nil
//	}
//
//	return &models.L3Policy{
//		Ingress: l3.Ingress.GetModel(),
//		Egress:  l3.Egress.GetModel(),
//	}
//}

func (l3 *L3Policy) DeepCopy() *L3Policy {
	return &L3Policy{
		Ingress: l3.Ingress.DeepCopy(),
		Egress:  l3.Egress.DeepCopy(),
	}
}

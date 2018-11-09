// Copyright 2018 Authors of Cilium
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

package api

import (
	"bytes"
	"fmt"
	"net"
	"sort"
)

type ProviderIntegration func(*ToGroups) ([]net.IP, error)

const (
	AWSPROVIDER = "AWS"
)

var (
	providers = map[string]ProviderIntegration{}
)

type ToGroupsActions interface {
	GetIPs(ToGroups) []net.IP
}

type ToGroups struct {
	Aws *AWSGroups `json:"aws,omitempty"`
}

type AWSGroups struct {
	Labels              map[string]string `json:"labels,omitempty"`
	SecurityGroupsIds   []string          `json:"securityGroupsIds,omitempty"`
	SecurityGroupsNames []string          `json:"securityGroupsNames,omitempty"`
	Region              string            `json:"region,omitempty"`
}

func RegisterToGroupsProvider(providerName string, callback ProviderIntegration) {
	providers[providerName] = callback
}

func (group *ToGroups) GetCidrSet() ([]CIDRRule, error) {

	var ips []net.IP
	emptyResult := []CIDRRule{}

	// Get per  provider CIDRSet
	if group.Aws != nil {
		callback, ok := providers[AWSPROVIDER]
		if !ok {
			return emptyResult, fmt.Errorf("Provider %s is not registered", AWSPROVIDER)
		}

		awsIPs, err := callback(group)
		if err != nil {
			return emptyResult, fmt.Errorf(
				"Cannot retrieve data from %s provider: %s",
				AWSPROVIDER, err)
		}
		ips = append(ips, awsIPs...)
	}

	// Sort IPS to have always the same result and do not update policies if it is not needed.
	sort.Slice(ips, func(i, j int) bool {
		return bytes.Compare(ips[i], ips[j]) < 0
	})
	return ipsToRules(ips), nil
}

// ipsToRules generates CIDRRules for the IPs passed in.
func ipsToRules(ips []net.IP) (cidrRules []CIDRRule) {
	for _, ip := range ips {
		rule := CIDRRule{ExceptCIDRs: make([]CIDR, 0)}
		rule.Generated = true
		if ip.To4() != nil {
			rule.Cidr = CIDR(ip.String() + "/32")
		} else {
			rule.Cidr = CIDR(ip.String() + "/128")
		}

		cidrRules = append(cidrRules, rule)
	}

	return cidrRules
}

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
	"fmt"
	"net"

	cilium_v2 "github.com/cilium/cilium/pkg/k8s/apis/cilium.io/v2"
	"github.com/cilium/cilium/pkg/policy/api"
	"k8s.io/apimachinery/pkg/types"

	. "gopkg.in/check.v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func getSamplePolicy(name, ns string) *cilium_v2.CiliumNetworkPolicy {
	cnp := &cilium_v2.CiliumNetworkPolicy{}

	cnp.ObjectMeta.Name = name
	cnp.ObjectMeta.Namespace = ns
	cnp.ObjectMeta.UID = types.UID("123")
	cnp.Spec = &api.Rule{
		EndpointSelector: api.EndpointSelector{
			LabelSelector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"test": "true",
				},
			},
		},
	}
	return cnp
}

func (s *GroupsTestSuite) TestCorrectChildName(c *C) {
	name := "test"
	cnp := getSamplePolicy(name, "testns")
	childCNP, err := getChildCNP(cnp)
	c.Assert(err, IsNil)
	c.Assert(
		childCNP.ObjectMeta.Name,
		Equals,
		fmt.Sprintf("%s-togroups-%s", name, cnp.ObjectMeta.UID))
}

func (s *GroupsTestSuite) TestChildPoliciesAreDeletedIfNoToGroups(c *C) {
	name := "test"
	cnp := getSamplePolicy(name, "testns")

	cnp.Spec.Egress = []api.EgressRule{
		api.EgressRule{
			ToPorts: []api.PortRule{
				api.PortRule{
					Ports: []api.PortProtocol{
						{Port: "5555"},
					},
				},
			},
		},
	}

	childCNP, err := getChildCNP(cnp)
	c.Assert(err, IsNil)
	c.Assert(childCNP.Spec.Egress, IsNil)
	c.Assert(len(childCNP.Specs), Equals, 0)
}

func (s *GroupsTestSuite) TestChildPoliciesAreInheritCorrectly(c *C) {

	cb := func(group *api.ToGroups) ([]net.IP, error) {
		return []net.IP{net.ParseIP("192.168.1.1")}, nil
	}

	api.RegisterToGroupsProvider(api.AWSPROVIDER, cb)

	name := "test"
	cnp := getSamplePolicy(name, "testns")

	cnp.Spec.Egress = []api.EgressRule{
		api.EgressRule{
			ToPorts: []api.PortRule{
				api.PortRule{
					Ports: []api.PortProtocol{
						{Port: "5555"},
					},
				},
			},
			ToGroups: []api.ToGroups{
				{
					Aws: &api.AWSGroups{
						Labels: map[string]string{
							"test": "a",
						},
					},
				},
			},
		},
	}

	childCNP, err := getChildCNP(cnp)
	c.Assert(err, IsNil)
	c.Assert(childCNP.Spec.Egress, IsNil)
	c.Assert(len(childCNP.Specs), Equals, 1)
	c.Assert(childCNP.Specs[0].Egress[0].ToPorts, DeepEquals, cnp.Spec.Egress[0].ToPorts)
	c.Assert(len(childCNP.Specs[0].Egress[0].ToGroups), Equals, 0)
}

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
	"sync"

	"github.com/cilium/cilium/pkg/k8s"
	cilium_v2 "github.com/cilium/cilium/pkg/k8s/apis/cilium.io/v2"
	clientset "github.com/cilium/cilium/pkg/k8s/client/clientset/versioned"
	"github.com/cilium/cilium/pkg/policy/api"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	parentCNP   = "parentCNP"
	CNPKindKey  = "CNPKind"
	CNPKindName = "child"
)

var (
	globalK8sClient *clientset.Clientset
	k8sMutex        = sync.Mutex{}
)

// getChildCNP will return a new CNP based on the given rule.
func getChildCNP(cnp *cilium_v2.CiliumNetworkPolicy) (*cilium_v2.CiliumNetworkPolicy, error) {
	childCNP := cnp.DeepCopy()
	childCNP.ObjectMeta.Name = fmt.Sprintf(
		"%s-togroups-%s",
		childCNP.ObjectMeta.Name,
		cnp.ObjectMeta.UID)
	childCNP.ObjectMeta.UID = ""
	childCNP.ObjectMeta.ResourceVersion = ""
	childCNP.ObjectMeta.Labels = map[string]string{
		parentCNP:  string(cnp.ObjectMeta.UID),
		CNPKindKey: CNPKindName,
	}
	childCNP.Spec = &api.Rule{}
	childCNP.Specs = api.Rules{}

	rules, err := cnp.Parse()
	if err != nil {
		return nil, fmt.Errorf("Cannot parse policies: %s", err)
	}

	for _, rule := range rules {
		if !rule.HasChildRule() {
			continue
		}
		newRule, err := rule.CreateChildRule()
		if err != nil {
			return childCNP, err
		}
		childCNP.Specs = append(childCNP.Specs, newRule)
	}

	return childCNP, nil
}

// getK8sClient return the kubernetes apiserver connection
func getK8sClient() (*clientset.Clientset, error) {
	k8sMutex.Lock()
	defer func() {
		k8sMutex.Unlock()
	}()
	if globalK8sClient != nil {
		return globalK8sClient, nil
	}

	restConfig, err := k8s.CreateConfig()
	if err != nil {
		return nil, fmt.Errorf("Unable to create rest configuration: %s", err)
	}
	k8sClient, err := clientset.NewForConfig(restConfig)
	if err != nil {
		return nil, fmt.Errorf("Unable to create Kubernetes configuration: %s", err)

	}
	globalK8sClient = k8sClient
	return globalK8sClient, nil
}

func updateCNPStatus(cnp *cilium_v2.CiliumNetworkPolicy) (*cilium_v2.CiliumNetworkPolicy, error) {
	k8sClient, err := getK8sClient()
	if err != nil {
		// @TODO change the error here
		return nil, err
	}
	return k8sClient.CiliumV2().CiliumNetworkPolicies(cnp.ObjectMeta.Namespace).UpdateStatus(cnp)
}

func updateOrCreateCNP(cnp *cilium_v2.CiliumNetworkPolicy) (*cilium_v2.CiliumNetworkPolicy, error) {
	k8sClient, err := getK8sClient()
	if err != nil {
		return nil, err
	}
	k8sCNP, err := k8sClient.CiliumV2().CiliumNetworkPolicies(cnp.ObjectMeta.Namespace).
		Get(cnp.ObjectMeta.Name, v1.GetOptions{})
	if err == nil {
		cnp.ObjectMeta.ResourceVersion = k8sCNP.ObjectMeta.ResourceVersion
		return k8sClient.CiliumV2().CiliumNetworkPolicies(cnp.ObjectMeta.Namespace).Update(cnp)
	}
	return k8sClient.CiliumV2().CiliumNetworkPolicies(cnp.ObjectMeta.Namespace).Create(cnp)
}

func updateChildrenStatus(cnp *cilium_v2.CiliumNetworkPolicy, childName string, err error) error {
	status := cilium_v2.CiliumChildPolicyStatus{
		LastUpdated: cilium_v2.NewTimestamp(),
		Enforcing:   false,
	}

	if err != nil {
		status.OK = false
		status.Error = fmt.Sprintf("%v", err)
	} else {
		status.OK = true
		status.Error = ""
	}

	k8sClient, clientErr := getK8sClient()
	if clientErr != nil {
		return fmt.Errorf("Cannot get Kubernetes apiserver client: %s", clientErr)
	}
	// This CNP can be modified by cilium agent or operator. To be able to push
	// the status correctly fetch the last version to avoid updates issues.
	k8sCNPStatus, clientErr := k8sClient.CiliumV2().
		CiliumNetworkPolicies(cnp.ObjectMeta.Namespace).
		Get(cnp.ObjectMeta.Name, v1.GetOptions{})
	if clientErr != nil {
		return fmt.Errorf("Cannot get Kubernetes policy: %s", clientErr)
	}
	if k8sCNPStatus.ObjectMeta.UID != cnp.ObjectMeta.UID {
		ToGroupsCNPCache.DeleteCNP(k8sCNPStatus)
		return fmt.Errorf("Policy UID mistmatch")
	}

	k8sCNPStatus.SetChildrenPolicy(childName, status)
	ToGroupsCNPCache.UpdateCNP(k8sCNPStatus)
	_, err = updateCNPStatus(k8sCNPStatus)
	return err
}

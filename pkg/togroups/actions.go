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
	"time"

	cilium_v2 "github.com/cilium/cilium/pkg/k8s/apis/cilium.io/v2"
	"github.com/cilium/cilium/pkg/logging/logfields"
	"github.com/sirupsen/logrus"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	MaxNumberOfAttempts = 5
	SleepDuration       = 5 * time.Second
)

// AddChildrenCNPIfNeeded will create a new CNP if the given CNP has any rule
// that need to create a new child policy.
func AddChildrenCNPIfNeeded(cnp *cilium_v2.CiliumNetworkPolicy) bool {
	if !cnp.HasChildrenPolicies() {
		log.WithFields(logrus.Fields{
			logfields.CiliumNetworkPolicyName: cnp.ObjectMeta.Name,
			logfields.K8sNamespace:            cnp.ObjectMeta.Namespace,
		}).Debug("CNP does not have children policies, skipped")
		return true
	}
	return addChildCNP(cnp)
}

// UpdateChildrenCNPIfNeeded will update or create a  CNP if the given CNP has
// any rule that need to create a new child policy.  In case that the newCNP
// will not have any children policy and the old one had one, it'll delete the
// old policy.
func UpdateChildrenCNPIfNeeded(newCNP *cilium_v2.CiliumNetworkPolicy, oldCNP *cilium_v2.CiliumNetworkPolicy) bool {
	if !newCNP.HasChildrenPolicies() && oldCNP.HasChildrenPolicies() {
		log.WithFields(logrus.Fields{
			logfields.CiliumNetworkPolicyName: newCNP.ObjectMeta.Name,
			logfields.K8sNamespace:            newCNP.ObjectMeta.Namespace,
		}).Info("New CNP does not have child policy, but old had. Deleted old policies")
		DeleteChildrenCNP(oldCNP)
		return false
	}

	if !newCNP.HasChildrenPolicies() {
		return false
	}
	return addChildCNP(newCNP)
}

// DeleteChildrenCNP if the given policy has any children policy will be
// deleted from the repo and the cache.
func DeleteChildrenCNP(cnp *cilium_v2.CiliumNetworkPolicy) error {

	scopedLog := log.WithFields(logrus.Fields{
		logfields.CiliumNetworkPolicyName: cnp.ObjectMeta.Name,
		logfields.K8sNamespace:            cnp.ObjectMeta.Namespace,
	})

	if !cnp.HasChildrenPolicies() {
		scopedLog.Debug("CNP does not have children policies, skipped")
		return nil
	}

	k8sClient, err := getK8sClient()
	if err != nil {
		scopedLog.WithError(err).Error("Cannot get kubertenes configuration")
	}

	err = k8sClient.CiliumV2().CiliumNetworkPolicies(cnp.ObjectMeta.Namespace).DeleteCollection(
		&v1.DeleteOptions{},
		v1.ListOptions{LabelSelector: fmt.Sprintf("%s=%s", parentCNP, cnp.ObjectMeta.UID)})

	ToGroupsCNPCache.DeleteCNP(cnp)
	return err
}

func addChildCNP(cnp *cilium_v2.CiliumNetworkPolicy) bool {

	scopedLog := log.WithFields(logrus.Fields{
		logfields.CiliumNetworkPolicyName: cnp.ObjectMeta.Name,
		logfields.K8sNamespace:            cnp.ObjectMeta.Namespace,
	})

	var childCNP *cilium_v2.CiliumNetworkPolicy
	var err error
	for numAttempts := 0; numAttempts <= MaxNumberOfAttempts; numAttempts++ {
		childCNP, err = getChildCNP(cnp)
		if err == nil {
			break
		}
		scopedLog.WithError(err).Error("Cannot create child")
		statusErr := updateChildrenStatus(cnp, childCNP.ObjectMeta.Name, err)
		if statusErr != nil {
			log.WithError(err).Error("Cannot update CNP status on invalid child")
		}
		if numAttempts == MaxNumberOfAttempts {
			return false
		}
		time.Sleep(SleepDuration)
	}
	ToGroupsCNPCache.UpdateCNP(cnp)
	_, err = updateOrCreateCNP(childCNP)
	if err != nil {
		statusErr := updateChildrenStatus(cnp, childCNP.ObjectMeta.Name, err)
		if statusErr != nil {
			scopedLog.WithError(err).Error("Cannot update CNP status on invalid child")
		}
		return false
	}

	err = updateChildrenStatus(cnp, childCNP.ObjectMeta.Name, nil)
	if err != nil {
		scopedLog.WithError(err).Error("Cannot update CNP status on valid child policy")
		return false
	}

	return true
}

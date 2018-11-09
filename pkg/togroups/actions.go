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
	parentCNP           = "parentCNP"
	CNPKindKey          = "CNPKind"
	CNPKindName         = "child"
	MaxNumberOfAttempts = 5
	SleepDuration       = 5 * time.Second
)

func AddChildrenCNPIfNeeded(cnp *cilium_v2.CiliumNetworkPolicy) bool {
	if !cnp.HasChildrenPolicies() {
		log.WithFields(logrus.Fields{
			logfields.CiliumNetworkPolicyName: cnp.ObjectMeta.Name,
			logfields.K8sNamespace:            cnp.ObjectMeta.Namespace,
		}).Debug("CNP does not have children policies, skipped")
		return true
	}
	addChildCNP(cnp)
	return true
}

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
	addChildCNP(newCNP)

	return true
}

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
		statusErr := UpdateChildrenStatus(cnp, childCNP.ObjectMeta.Name, err)
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
		statusErr := UpdateChildrenStatus(cnp, childCNP.ObjectMeta.Name, err)
		if statusErr != nil {
			scopedLog.WithError(err).Error("Cannot update CNP status on invalid child")
		}
		return false
	}

	err = UpdateChildrenStatus(cnp, childCNP.ObjectMeta.Name, nil)
	if err != nil {
		scopedLog.WithError(err).Error("Cannot update CNP status on valid child policy")
		return false
	}

	return true
}

func UpdateChildrenStatus(cnp *cilium_v2.CiliumNetworkPolicy, childName string, err error) error {
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

func UpdateChildrenCNPData(cnp *cilium_v2.CiliumNetworkPolicy) {
	return
}

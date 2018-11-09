package togroups

import (
	"fmt"

	"github.com/cilium/cilium/pkg/k8s"
	cilium_v2 "github.com/cilium/cilium/pkg/k8s/apis/cilium.io/v2"
	clientset "github.com/cilium/cilium/pkg/k8s/client/clientset/versioned"
	"github.com/cilium/cilium/pkg/policy/api"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	globalK8sClient *clientset.Clientset
)

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
		if !rule.HasChildrenPolicy() {
			continue
		}
		newRule, err := rule.CreateChildrenPolicy()
		if err != nil {
			return childCNP, err
		}
		childCNP.Specs = append(childCNP.Specs, newRule)
	}

	return childCNP, nil
}

func getK8sClient() (*clientset.Clientset, error) {
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
	// @TODO lock here!
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

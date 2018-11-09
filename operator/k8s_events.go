package main

import (
	"fmt"
	"time"

	"github.com/cilium/cilium/pkg/k8s"
	cilium_v2 "github.com/cilium/cilium/pkg/k8s/apis/cilium.io/v2"
	clientset "github.com/cilium/cilium/pkg/k8s/client/clientset/versioned"
	k8sUtils "github.com/cilium/cilium/pkg/k8s/utils"
	"github.com/cilium/cilium/pkg/metrics"
	"github.com/cilium/cilium/pkg/togroups"
)

const (
	reSyncPeriod = 1 * time.Second
)

func EnableK8sWatcher() error {
	// @TODO uncomment this
	// if !k8s.IsEnabled() {
	// 	log.Debug("Not enabling k8s event listener because k8s is not enabled")
	// 	return nil
	// }

	restConfig, err := k8s.CreateConfig()
	if err != nil {
		return fmt.Errorf("Unable to create rest configuration: %s", err)
	}

	ciliumNPClient, err := clientset.NewForConfig(restConfig)
	if err != nil {
		return fmt.Errorf("Unable to create cilium network policy client: %s", err)
	}

	k8sUtils.ResourceEventHandlerFactory(
		func(i interface{}) func() error {
			return func() error {
				cnp := i.(*cilium_v2.CiliumNetworkPolicy)
				togroups.AddChildrenCNPIfNeeded(cnp)
				return nil
			}
		},
		func(i interface{}) func() error {
			return func() error {
				cnp := i.(*cilium_v2.CiliumNetworkPolicy)
				togroups.DeleteChildrenCNP(cnp)
				return nil
			}
		},
		func(old, new interface{}) func() error {
			return func() error {
				newCNP := new.(*cilium_v2.CiliumNetworkPolicy)
				oldCNP := old.(*cilium_v2.CiliumNetworkPolicy)
				togroups.UpdateChildrenCNPIfNeeded(newCNP, oldCNP)
				return nil
			}
		},
		nil,
		&cilium_v2.CiliumNetworkPolicy{},
		ciliumNPClient,
		reSyncPeriod,
		metrics.EventTSK8s,
	)
	return nil
}

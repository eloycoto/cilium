package togroups

import (
	"fmt"
	"time"

	"github.com/cilium/cilium/pkg/controller"
	"github.com/cilium/cilium/pkg/k8s"
	clientset "github.com/cilium/cilium/pkg/k8s/client/clientset/versioned"
	"github.com/cilium/cilium/pkg/logging"
	"github.com/cilium/cilium/pkg/logging/logfields"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
)

var log = logging.DefaultLogger.WithField(logfields.LogSubsys, "eloy")

func Initialize() {
	fmt.Println("Eloy Initialize")
}

func init() {
	controller.NewManager().UpdateController("La puta estampa",
		controller.ControllerParams{
			RunInterval: 1 * time.Second,
			DoFunc: func() error {
				startController()
				return nil
			},
		})
}

func startController() {

	var (
		restConfig *rest.Config
		k8sClient  *clientset.Clientset
		err        error
	)

	restConfig, err = k8s.CreateConfig()
	if err != nil {
		log.Errorf("Eloy--Invalid Config")
		return
	}

	k8sClient, err = clientset.NewForConfig(restConfig)
	if err != nil {
		return
	}
	merda, err := k8sClient.CiliumV2().CiliumNetworkPolicies(meta_v1.NamespaceAll).List(meta_v1.ListOptions{})
	log.Errorf("Eloy List of ceps %v", merda)
	log.Errorf("Eloy List of ceps err := %v", err)

	// ceps, err := ciliumClient.CiliumEndpoints(meta_v1.NamespaceAll).List(meta_v1.ListOptions{})
}

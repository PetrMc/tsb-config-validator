package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/PetrMc/tsb-config-validator/apis"
	cp "github.com/PetrMc/tsb-config-validator/apis/install/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

var kubeconfig string

func init() {
	flag.StringVar(&kubeconfig, "kubeconfig", "/home/petr/.kube/config", "path to Kubernetes config file")
	flag.Parse()
}

func main() {
	var config *rest.Config
	var err error

	if kubeconfig == "" {
		log.Printf("using in-cluster configuration")
		config, err = rest.InClusterConfig()
	} else {
		log.Printf("using configuration from '%s'", kubeconfig)
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
	}

	if err != nil {
		panic(err)
	}

	// cfg, err := kConfig.GetConfig()
	// if err != nil {
	// 	fmt.Printf("Could not get apiserver config: %v", err)
	// }

	// Create a new Cmd to provide shared dependencies and start components
	syncPeriod := time.Hour

	mgr, err := manager.New(config, manager.Options{
		SyncPeriod:              &syncPeriod,
		Namespace:               "istio-system",
		LeaderElection:          true,
		LeaderElectionNamespace: "istio-system",
		LeaderElectionID:        "tsb-operator-mgmtplane-lock",
		Port:                    443, // the default port has changed from 443 (back in v0.6.3) to 9443 (since v0.7.0)
	})

	//cfg, err := kubeConfig.GetConfig()
	// if err != nil {
	// 	fmt.Println(err)
	// 	return
	// }

	// // establish connection to k8s cluster
	// clientset, err := kubernetes.NewForConfig(cfg)
	// // client, err := clientV1alpha1.NewForConfig(cfg)
	// if err != nil {
	// 	fmt.Println(err, clientset)
	// 	return
	// }

	// controlplaneResource := schema.GroupVersionKind{Group: v1alpha1.SchemeGroupVersion.Group, Version: v1alpha1.SchemeGroupVersion.Version, Kind: "ControlPlane"}

	if err = apis.AddToScheme(mgr.GetScheme()); err != nil {
		fmt.Printf("Could not add ManagementPlane Custom Resource to scheme: %v", err)
	}
	// // controlplaneResource := schema.GroupVersionKind{Group: cp.SchemeGroupVersion.Group, Version: cp.SchemeGroupVersion.Version, Kind: "ControlPlane"}

	crdconfig := *config
	crdconfig.ContentConfig.GroupVersion = &schema.GroupVersion{Group: cp.SchemeGroupVersion.Group, Version: cp.SchemeGroupVersion.Version}
	crdconfig.APIPath = "/apis"
	crdconfig.NegotiatedSerializer = serializer.NewCodecFactory(scheme.Scheme)
	crdconfig.UserAgent = rest.DefaultKubernetesUserAgent()

	// client, err := rest.RESTClientFor(&config)
	// if err != nil {
	// 	fmt.Println(err)
	// } else {
	// 	fmt.Println(client)
	// }

	exampleRestClient, err := rest.UnversionedRESTClientFor(&crdconfig)

	if err != nil {
		panic(err)
	}
	// controlplanes, err := client.Controlplane("istio-system").List(metav1.ListOptions{})
	// if err != nil {
	// 	panic(err)
	// }
	// Create a context

	// Setup context
	ctx := context.Background()

	result := cp.ControlPlaneList{}
	err = exampleRestClient.Get().
		Resource("controlplanes").
		Namespace("istio-system").
		Do(ctx).Into(&result)

	fmt.Printf("projects found: %+v\n", result)
	fmt.Printf("%v\n", exampleRestClient)

	fmt.Printf("%v", crdconfig.ContentConfig.GroupVersion)
	fmt.Println(" ")
	fmt.Printf("%v", crdconfig.NegotiatedSerializer)

}

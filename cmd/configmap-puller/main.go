// Note: the example only works with the code within the same release/branch.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	//
	// Uncomment to load all auth plugins
	// _ "k8s.io/client-go/plugin/pkg/client/auth"
	//
	// Or uncomment to load specific auth plugins
	// _ "k8s.io/client-go/plugin/pkg/client/auth/azure"
	// _ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	// _ "k8s.io/client-go/plugin/pkg/client/auth/oidc"
	// _ "k8s.io/client-go/plugin/pkg/client/auth/openstack"
)

var (
	kubeconfig         *string
	configmapName      *string
	configmapNamespace *string
	configmapKey       *string
	outfileName        *string
)

func main() {
	kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	configmapName = flag.String("configmap-name", "traefik-config", "name of the configmap to watch")
	configmapNamespace = flag.String("configmap-namespace", "default", "namespace of the configmap to watch")
	configmapKey = flag.String("configmap-key", "rules.toml", "key of the configmap to read")
	outfileName = flag.String("outfile-name", "/watched/rules.toml", "name of the file to write")

	flag.Parse()

	// load config depending if we are outside or inside a cluster
	var config *rest.Config
	if len(*kubeconfig) > 0 {
		var err error
		config, err = clientcmd.BuildConfigFromFlags("", *kubeconfig)
		if err != nil {
			panic(err.Error())
		}
	} else {
		var err error
		config, err = rest.InClusterConfig()
		if err != nil {
			panic(err.Error())
		}
	}

	// create the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	namespace := *configmapNamespace
	name := *configmapName

	for {
		cm, err := clientset.CoreV1().ConfigMaps(namespace).Get(context.TODO(), name, metav1.GetOptions{})
		if errors.IsNotFound(err) {
			panic(fmt.Sprintf("Configmap %s in namespace %s not found\n", name, namespace))
		} else if statusError, isStatus := err.(*errors.StatusError); isStatus {
			panic(fmt.Sprintf("Error getting configmap %s in namespace %s: %v\n",
				name, namespace, statusError.ErrStatus.Message))
		} else if err != nil {
			panic(err.Error())
		}

		fmt.Printf("Found configmap %s in namespace %s\n", name, namespace)
		data := cm.Data[*configmapKey]
		fmt.Printf("Data: %+v", data)

		// TODO: only write if content changed)
		if err := writeStringFile(*outfileName, data); err != nil {
			panic(err)
		}

		fmt.Println("Wrote data to file", *outfileName)

		time.Sleep(10 * time.Second)
	}
}

func writeStringFile(filename, data string) error {
	f, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return fmt.Errorf("could not open file for writing: %w", err)
	}
	defer f.Close()

	if _, err := f.WriteString(data); err != nil {
		return fmt.Errorf("could not open file for writing: %w", err)
	}

	return nil
}

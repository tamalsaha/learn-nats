package main

import (
	"context"
	"fmt"
	"log"
	"path/filepath"

	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	corev1 "k8s.io/client-go/applyconfigurations/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

func main() {
	masterURL := ""
	kubeconfigPath := filepath.Join(homedir.HomeDir(), ".kube", "config")

	config, err := clientcmd.BuildConfigFromFlags(masterURL, kubeconfigPath)
	if err != nil {
		log.Fatalf("Could not get Kubernetes config: %s", err)
	}

	client := kubernetes.NewForConfigOrDie(config)

	p := corev1.Pod("busybox", "default").
		WithLabels(map[string]string{
			"app": "busybox",
		}).WithSpec(corev1.PodSpec().
		WithRestartPolicy(core.RestartPolicyAlways).
		WithContainers(corev1.Container().
			WithImage("ubuntu:18.04").
			WithImagePullPolicy(core.PullIfNotPresent).
			WithName("busybox").
			WithCommand("sleep", "3600").
			WithResources(corev1.ResourceRequirements().
				WithLimits(core.ResourceList{
					core.ResourceCPU:    resource.MustParse("500m"),
					core.ResourceMemory: resource.MustParse("1Gi"),
				}))))

	p2, err := client.CoreV1().Pods("default").Apply(context.Background(), p, metav1.ApplyOptions{
		Force:        true,
		FieldManager: "tamal",
	})
	if err != nil {
		panic(err)
	}
	fmt.Printf("%+v", p2)
}

package main

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"sync"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/tamalsaha/nats-hop-demo/shared"
	"github.com/tamalsaha/nats-hop-demo/transport"
	core "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	corev1 "k8s.io/client-go/applyconfigurations/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"k8s.io/klog/v2"
	"kmodules.xyz/client-go/tools/clusterid"
)

func main() {
	http.Get("")

	nc, err := nats.Connect(shared.NATS_URL)
	if err != nil {
		klog.Fatalln(err)
	}
	defer nc.Close()

	masterURL := ""
	kubeconfigPath := filepath.Join(homedir.HomeDir(), ".kube", "config")
	// kubeconfigPath = "/home/tamal/Downloads/config"

	config, err := clientcmd.BuildConfigFromFlags(masterURL, kubeconfigPath)
	if err != nil {
		log.Fatalf("Could not get Kubernetes config: %s", err)
	}

	// func RESTClientFor(config *Config) (*RESTClient, error)
	// k8s.io/client-go/rest/config.go
	// k8s.io/client-go/transport/transport.go # TLSConfigFor

	uid, err := clusterid.ClusterUID(kubernetes.NewForConfigOrDie(config).CoreV1().Namespaces())
	if err != nil {
		panic(err)
	}

	c2 := rest.CopyConfig(config)
	cfg, err := c2.TransportConfig()
	if err != nil {
		panic(err)
	}
	c2.Transport, err = transport.New(cfg, nc, "cluster."+uid+".proxy", 10000*time.Second)
	if err != nil {
		panic(err)
	}

	//// needs to be updated for each GVR
	//if err := setConfigDefaults(c2); err != nil {
	//	panic(err)
	//}
	//rc, err := transport.RESTClientFor(c2)
	//if err != nil {
	//	panic(err)
	//}
	//kc := kubernetes.New(rc)

	kc := kubernetes.NewForConfigOrDie(c2)

	nodes, err := kc.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		panic(err)
	}
	for _, n := range nodes.Items {
		fmt.Println(n.Name)
	}

	deploys, err := kc.AppsV1().Deployments("kube-system").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		panic(err)
	}
	for _, n := range deploys.Items {
		fmt.Println(n.Name)
	}

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

	p2, err := kc.CoreV1().Pods("default").Apply(context.Background(), p, metav1.ApplyOptions{
		Force:        true,
		FieldManager: "tamal",
	})
	if err != nil {
		panic(err)
	}
	fmt.Printf("%+v", p2)
}

func setConfigDefaults(config *rest.Config) error {
	gv := v1.SchemeGroupVersion
	config.GroupVersion = &gv
	config.APIPath = "/api"
	config.NegotiatedSerializer = scheme.Codecs.WithoutConversion()

	if config.UserAgent == "" {
		config.UserAgent = rest.DefaultKubernetesUserAgent()
	}

	return nil
}

var pool = sync.Pool{
	New: func() interface{} {
		return new(bytes.Buffer)
	},
}

type natshop struct {
	nc *nats.Conn
}

func (rt *natshop) RoundTrip(r *http.Request) (*http.Response, error) {
	buf := pool.Get().(*bytes.Buffer)
	defer pool.Put(buf)
	buf.Reset()

	if err := r.Write(buf); err != nil { // WriteProxy
		return nil, err
	}
	fmt.Println(buf.String())

	msg, err := rt.nc.RequestMsg(&nats.Msg{
		Subject: "k8s",
		Data:    buf.Bytes(),
	}, 5*time.Second)
	if err != nil {
		fmt.Println("-----------------", err.Error())
		return nil, err
	}
	buf.Reset()
	return http.ReadResponse(bufio.NewReader(bytes.NewReader(msg.Data)), r)
}

var _ http.RoundTripper = &natshop{}

package main

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"github.com/nats-io/nats.go"
	"github.com/tamalsaha/nats-hop-demo/shared"
	"github.com/tamalsaha/nats-hop-demo/transport"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"k8s.io/klog/v2"
	"log"
	"net/http"
	"path/filepath"
	"sync"
	"time"
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

	config, err := clientcmd.BuildConfigFromFlags(masterURL, kubeconfigPath)
	if err != nil {
		log.Fatalf("Could not get Kubernetes config: %s", err)
	}

	// func RESTClientFor(config *Config) (*RESTClient, error)
	// k8s.io/client-go/rest/config.go
	// k8s.io/client-go/transport/transport.go # TLSConfigFor

	c2 := rest.CopyConfig(config)
	cfg, err := c2.TransportConfig()
	if err != nil {
		panic(err)
	}
	c2.Transport, err = transport.New(cfg, nc, "k8s", 10000*time.Second)
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

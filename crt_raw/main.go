package main

import (
	"context"
	"fmt"
	"kubedb.dev/apimachinery/apis/kubedb/v1alpha2"

	//"time"

	// 	"github.com/nats-io/nats.go"
	// 	"github.com/tamalsaha/nats-hop-demo/shared"
	// 	"github.com/tamalsaha/nats-hop-demo/transport"
	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	//"k8s.io/klog/v2"
	"k8s.io/klog/v2/klogr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
	kubedbscheme "kubedb.dev/apimachinery/client/clientset/versioned/scheme"
	kubevaultscheme "kubevault.dev/apimachinery/client/clientset/versioned/scheme"
	schemav1alpha1 "kubedb.dev/schema-manager/apis/schema/v1alpha1"
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func main() {
	if err := run(); err != nil {
		panic(err)
	}
}

func run() error {
	_ = clientgoscheme.AddToScheme(scheme)
	_ = kubedbscheme.AddToScheme(scheme)
	_ = kubevaultscheme.AddToScheme(scheme)
	_ = schemav1alpha1.AddToScheme(scheme)

	// 	nc, err := nats.Connect(shared.NATS_URL)
	// 	if err != nil {
	// 		klog.Fatalln(err)
	// 	}
	// 	defer nc.Close()

	ctrl.SetLogger(klogr.New())
	cfg := ctrl.GetConfigOrDie()

	// 	tr, err := cfg.TransportConfig()
	// 	if err != nil {
	// 		panic(err)
	// 	}
	// 	cfg.Transport, err = transport.New(tr, nc, "k8s", 10000*time.Second)
	// 	if err != nil {
	// 		panic(err)
	// 	}

	mapper, err := apiutil.NewDynamicRESTMapper(cfg)
	if err != nil {
		return err
	}

	c, err := client.New(cfg, client.Options{
		Scheme: scheme,
		Mapper: mapper,
		Opts: client.WarningHandlerOptions{
			SuppressWarnings:   false,
			AllowDuplicateLogs: false,
		},
	})
	if err != nil {
		return err
	}

	var nodes core.NodeList
	err = c.List(context.TODO(), &nodes)
	if err != nil {
		panic(err)
	}
	for _, n := range nodes.Items {
		fmt.Println(n.Name)
	}

	var mysqls v1alpha2.MySQLList
	err = c.List(context.TODO(), &mysqls)
	if err != nil {
		panic(err)
	}
	for _, n := range mysqls.Items {
		fmt.Println(n.Name)
	}

	return nil
}

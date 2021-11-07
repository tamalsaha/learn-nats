package main

import (
	"encoding/json"
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"time"

	"github.com/google/uuid"
	flag "github.com/spf13/pflag"
	"gomodules.xyz/blobfs/testing"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
	"kubepack.dev/kubepack/apis/kubepack/v1alpha1"
	"kubepack.dev/kubepack/pkg/lib"
)

var (
	file = "artifacts/kubedb-community/order.yaml"
)
var (
	url     = "https://charts.appscode.com/stable"
	name    = "cert-manager-csi-driver-cacerts"
	version = "" // "v0.1.0"
)

const (
	YAMLBucket                   = "gs://connect.bytebuilders.link"
	YAMLHost                     = "https://connect.bytebuilders.link"
	GoogleApplicationCredentials = "/Users/tamal/AppsCode/credentials/license-issuer@appscode-domains.json"
)

func main() {
	flag.StringVar(&file, "file", file, "Path to Order file")
	flag.Parse()

	bs, err := NewTestBlobStore()
	if err != nil {
		klog.Fatal(err)
	}

	order := newOrder(url, name, version)
	order.UID = types.UID(uuid.New().String())

	scripts, err := lib.GenerateYAMLScript(bs, lib.DefaultRegistry, order)
	if err != nil {
		klog.Fatal(err)
	}
	data, err := json.MarshalIndent(scripts, "", "  ")
	if err != nil {
		klog.Fatal(err)
	}
	fmt.Println(string(data))
}

func newOrder(url, name, version string) v1alpha1.Order {
	return v1alpha1.Order{
		TypeMeta: metav1.TypeMeta{
			APIVersion: v1alpha1.SchemeGroupVersion.String(),
			Kind:       v1alpha1.ResourceKindOrder,
		}, ObjectMeta: metav1.ObjectMeta{
			Name:              name,
			UID:               types.UID(uuid.New().String()),
			CreationTimestamp: metav1.NewTime(time.Now()),
		},
		Spec: v1alpha1.OrderSpec{
			Packages: []v1alpha1.PackageSelection{
				{
					Chart: &v1alpha1.ChartSelection{
						ChartRef: v1alpha1.ChartRef{
							URL:  url,
							Name: name,
						},
						Version:     version,
						ReleaseName: name,
						Namespace:   metav1.NamespaceDefault, // change to kubeops or bytebuilders?
						Bundle:      nil,
						ValuesFile:  "values.yaml",
						ValuesPatch: nil,
						Resources:   nil,
						WaitFors:    nil,
					},
				},
			},
			KubeVersion: "",
		},
	}
}

func main_helm3() {
	flag.StringVar(&file, "file", file, "Path to Order file")
	flag.Parse()

	bs, err := NewTestBlobStore()
	if err != nil {
		klog.Fatal(err)
	}

	order := newOrder(url, name, version)
	order.UID = types.UID(uuid.New().String())

	scripts, err := lib.GenerateHelm3Script(bs, lib.DefaultRegistry, order)
	if err != nil {
		klog.Fatal(err)
	}
	data, err := json.MarshalIndent(scripts, "", "  ")
	if err != nil {
		klog.Fatal(err)
	}
	fmt.Println(string(data))
}

func NewTestBlobStore() (*lib.BlobStore, error) {
	fs, err := testing.NewTestGCS(YAMLBucket, GoogleApplicationCredentials)
	if err != nil {
		return nil, err
	}
	return &lib.BlobStore{
		BlobFS: fs,
		Host:   YAMLHost,
		Bucket: YAMLBucket,
	}, nil
}

package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/google/uuid"
	flag "github.com/spf13/pflag"
	"gomodules.xyz/blobfs/testing"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
	"kubepack.dev/kubepack/apis"
	"kubepack.dev/kubepack/apis/kubepack/v1alpha1"
	"kubepack.dev/kubepack/pkg/lib"
	"sigs.k8s.io/yaml"
)

var (
	file = "artifacts/kubedb-community/order.yaml"
)

const (
	YAMLBucket                   = "gs://connect.bytebuilders.link"
	YAMLHost                     = "https://connect.bytebuilders.link"
	GoogleApplicationCredentials = "/Users/tamal/AppsCode/credentials/license-issuer@appscode-domains.json"
)

func main() {
	flag.StringVar(&file, "file", file, "Path to Order file")
	flag.Parse()

	bs, err := lib.NewTestBlobStore()
	if err != nil {
		klog.Fatal(err)
	}

	data, err := ioutil.ReadFile(file)
	if err != nil {
		klog.Fatal(err)
	}
	var order v1alpha1.Order
	err = yaml.Unmarshal(data, &order)
	if err != nil {
		klog.Fatal(err)
	}
	order.UID = types.UID(uuid.New().String())

	scripts, err := lib.GenerateYAMLScript(bs, lib.DefaultRegistry, order)
	if err != nil {
		klog.Fatal(err)
	}
	data, err = json.MarshalIndent(scripts, "", "  ")
	if err != nil {
		klog.Fatal(err)
	}
	fmt.Println(string(data))
}

func main_helm3() {
	flag.StringVar(&file, "file", file, "Path to Order file")
	flag.Parse()

	bs, err := lib.NewTestBlobStore()
	if err != nil {
		klog.Fatal(err)
	}

	data, err := ioutil.ReadFile(file)
	if err != nil {
		klog.Fatal(err)
	}
	var order v1alpha1.Order
	err = yaml.Unmarshal(data, &order)
	if err != nil {
		klog.Fatal(err)
	}
	order.UID = types.UID(uuid.New().String())

	scripts, err := lib.GenerateHelm3Script(bs, lib.DefaultRegistry, order)
	if err != nil {
		klog.Fatal(err)
	}
	data, err = json.MarshalIndent(scripts, "", "  ")
	if err != nil {
		klog.Fatal(err)
	}
	fmt.Println(string(data))
}

func NewTestBlobStore() (*lib.BlobStore, error) {
	fs, err := testing.NewTestGCS(apis.YAMLBucket, GoogleApplicationCredentials)
	if err != nil {
		return nil, err
	}
	return &lib.BlobStore{
		BlobFS: fs,
		Host:   apis.YAMLHost,
		Bucket: apis.YAMLBucket,
	}, nil
}

package main

import (
	"encoding/json"
	"fmt"
	"gomodules.xyz/jsonpatch/v2"
	"gomodules.xyz/ulids"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"time"

	"gomodules.xyz/blobfs/testing"
	"k8s.io/klog/v2"
	"kubepack.dev/kubepack/apis/kubepack/v1alpha1"
	"kubepack.dev/kubepack/pkg/lib"
)

var (
	url     = "https://charts.appscode.com/stable"
	name    = "cluster-connector"
	version = "" // "v0.1.0"
)

const (
	YAMLBucket                   = "gs://connect.bytebuilders.link"
	YAMLHost                     = "https://connect.bytebuilders.link"
	// GoogleApplicationCredentials = "/Users/tamal/AppsCode/credentials/license-issuer@appscode-domains.json"
	GoogleApplicationCredentials = "/personal/AppsCode/credentials/license-issuer@appscode-domains.json"
)

func main() {
	bs, err := NewTestBlobStore()
	if err != nil {
		klog.Fatal(err)
	}

	order, err := newOrder(url, name, version)
	if err != nil {
		panic(err)
	}

	result, err := GenerateScripts(bs, order)
	if err != nil {
		panic(err)
	}
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		klog.Fatal(err)
	}
	fmt.Println(string(data))
}

type UserValues struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	Token string `json:"token"`
}

type ChartValues struct {
	User UserValues `json:"user"`
}

func generatePatch() ([]byte, error) {
	cv := ChartValues{
		User: UserValues{
			Name:  "Tamal Saha",
			Email: "tamal@appscode.com",
			Token: "****-****-***",
		},
	}

	data, err := json.Marshal(cv)
	if err != nil {
		return nil, err
	}

	empty, err := json.Marshal(ChartValues{})
	if err != nil {
		return nil, err
	}
	ops, err := jsonpatch.CreatePatch(empty, data)
	if err != nil {
		return nil, err
	}
	return json.MarshalIndent(ops, "", "  ")
}

func newOrder(url, name, version string) (*v1alpha1.Order, error) {
	patch, err := generatePatch()
	if err != nil {
		return nil, err
	}

	return &v1alpha1.Order{
		TypeMeta: metav1.TypeMeta{
			APIVersion: v1alpha1.SchemeGroupVersion.String(),
			Kind:       v1alpha1.ResourceKindOrder,
		}, ObjectMeta: metav1.ObjectMeta{
			Name: name,
			UID:               types.UID(ulids.MustNew().String()), // using ulids instead of UUID
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
						Namespace:   "kubeops", // change to kubeops or bytebuilders?
						Bundle:      nil,
						// ValuesFile:  "values.yaml",
						ValuesPatch: &runtime.RawExtension{Raw: patch},
						Resources:   nil,
						WaitFors:    nil,
					},
				},
			},
			KubeVersion: "",
		},
	}, nil
}

func GenerateScripts(bs *lib.BlobStore, order *v1alpha1.Order) (map[string]string, error) {
	scriptsYAML, err := lib.GenerateYAMLScript(bs, lib.DefaultRegistry, *order, lib.DisableApplicationCRD, lib.OsIndependentScript)
	if err != nil {
		return nil, err
	}

	scriptsHelm3, err := lib.GenerateHelm3Script(bs, lib.DefaultRegistry, *order, lib.DisableApplicationCRD, lib.OsIndependentScript)
	if err != nil {
		return nil, err
	}

	return map[string]string{
		"yaml":  scriptsYAML[0].Script,
		"helm3": scriptsHelm3[0].Script,
	}, nil
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

package main

import (
	"context"

	"gomodules.xyz/blobfs/testing"
	"k8s.io/klog/v2"
	"kubepack.dev/kubepack/pkg/lib"
)

const (
	YAMLBucket                   = "gs://connect.bytebuilders.link"
	YAMLHost                     = "https://connect.bytebuilders.link"
	GoogleApplicationCredentials = "/Users/tamal/Downloads/appscode-domains-01b8640dee04.json"
)

func main() {
	bs, err := NewTestBlobStore()
	if err != nil {
		klog.Fatal(err)
	}
	err = bs.BlobFS.WriteFile(context.Background(), "test.txt", []byte("This is a test"))
	if err != nil {
		klog.Fatal(err)
	}
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

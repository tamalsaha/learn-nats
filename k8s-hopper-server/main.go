package main

import (
	"fmt"
	"github.com/nats-io/nats.go"
	"github.com/tamalsaha/nats-hop-demo/shared"
	"k8s.io/klog/v2"
	"time"
)

func main() {
	fmt.Println(shared.NATS_URL)
	nc, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		klog.Fatalln(err)
	}
	defer nc.Close()

	msg := nats.NewMsg("k8s")
	msg.Data = []byte("x")

	if m, err := nc.RequestMsg(msg, 10*time.Second); err != nil {
		fmt.Println("***", err.Error())
	} else {
		fmt.Println(string(m.Data))
	}

	select {}
}

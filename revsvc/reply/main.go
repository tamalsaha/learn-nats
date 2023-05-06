package main

import (
	"fmt"
	"path/filepath"

	"github.com/nats-io/nats.go"
	"github.com/tamalsaha/nats-hop-demo/revsvc/backend"
	"k8s.io/client-go/util/homedir"
	"k8s.io/klog/v2"
)

// B replies
func main() {
	// NKEYS_PATH=$HOME/.local/share/nats/nsc/keys
	// $NKEYS_PATH/creds/appscode/B/y.creds

	credFile := filepath.Join(homedir.HomeDir(), ".local/share/nats/nsc/keys", "creds/appscode/B/y.creds")

	nc, err := backend.NewConnection("B", "nats://localhost:4222", credFile)
	if err != nil {
		panic(err)
	}
	defer nc.Drain()

	_, err = nc.QueueSubscribe("k8s.proxy.handler", "cc", func(msg *nats.Msg) {
		fmt.Println("REQ:", string(msg.Data), "REPLY_TO:", msg.Reply)

		resp := "echo>>>" + string(msg.Data)
		if err = msg.Respond([]byte(resp)); err != nil {
			klog.Error(err)
		}
	})

	select {}
}

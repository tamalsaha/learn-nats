package main

import (
	"github.com/rs/xid"
	"github.com/tamalsaha/nats-hop-demo/revsvc/backend"
	"k8s.io/client-go/util/homedir"
	"log"
	"path/filepath"
	"time"
)

// A requests
func main() {
	// NKEYS_PATH=$HOME/.local/share/nats/nsc/keys
	// $NKEYS_PATH/creds/appscode/A/x.creds

	credFile := filepath.Join(homedir.HomeDir(), ".local/share/nats/nsc/keys", "creds/appscode/A/x.creds")

	nc, err := backend.NewConnection("A", "nats://localhost:4222", credFile)
	if err != nil {
		panic(err)
	}
	defer nc.Drain()

	// fmt.Println(xid.New().String())

	//msg := &nats.Msg{
	//	Subject: "k8s.proxy.handler.cid_b",
	//	Reply:   "k8s.proxy.resp.1",
	//	Header:  nil,
	//	Data:    []byte("hello"),
	//	Sub:     nil,
	//}
	//reply, err := nc.RequestMsg(msg, 10*time.Second)
	//if err != nil {
	//	panic(err)
	//}
	//fmt.Println(string(reply.Data))

	// Create a unique subject name for replies.
	uid := xid.New().String()
	// uniqueReplyTo := "k8s.proxy.resp." + uid

	// Listen for a single response
	sub, err := nc.SubscribeSync("k8s.proxy.resp.cid_b." + uid)
	if err != nil {
		log.Fatal(err)
	}

	// Send the request.
	// If processing is synchronous, use Request() which returns the response message.
	if err := nc.PublishRequest("k8s.proxy.handler.cid_b", "k8s.proxy.resp."+uid, []byte("hello")); err != nil {
		log.Fatal(err)
	}

	// Read the reply
	msg, err := sub.NextMsg(time.Second)
	if err != nil {
		log.Fatal(err)
	}

	// Use the response
	log.Printf("Reply: %s", msg.Data)
}

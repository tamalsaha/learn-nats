package main

import (
	"errors"
	"fmt"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/tamalsaha/learn-nats/natsclient"
	"gomodules.xyz/oneliners"
)

func main() {
	nc, err := natsclient.NewConnection(nats.DefaultURL, "")
	if err != nil {
		panic(err)
	}
	defer nc.Drain()

	for {
		fmt.Println("sleeping....")
		time.Sleep(30 * time.Second)
		fmt.Println("woke up from sleep....")

		err := nc.Publish("greet.joe", []byte("hello"))
		if err != nil {
			oneliners.FILE(err)
			continue
		}

		sub, err := nc.SubscribeSync("greet.*")
		if err != nil {
			oneliners.FILE(err)
			continue
		}

		msg, err := sub.NextMsg(10 * time.Millisecond)
		if err != nil && !errors.Is(err, nats.ErrTimeout) {
			oneliners.FILE(err)
			continue
		}
		fmt.Println("subscribed after a publish...")
		fmt.Printf("msg is nil? %v\n", msg == nil)

		err = nc.Publish("greet.joe", []byte("hello"))
		if err != nil {
			oneliners.FILE(err)
			continue
		}
		err = nc.Publish("greet.pam", []byte("hello"))
		if err != nil {
			oneliners.FILE(err)
			continue
		}

		msg, err = sub.NextMsg(10 * time.Millisecond)
		if err != nil {
			oneliners.FILE(err)
			continue
		}
		fmt.Printf("msg data: %q on subject %q\n", string(msg.Data), msg.Subject)

		msg, err = sub.NextMsg(10 * time.Millisecond)
		if err != nil {
			oneliners.FILE(err)
			continue
		}
		fmt.Printf("msg data: %q on subject %q\n", string(msg.Data), msg.Subject)

		err = nc.Publish("greet.bob", []byte("hello"))
		if err != nil {
			oneliners.FILE(err)
			continue
		}

		msg, err = sub.NextMsg(10 * time.Millisecond)
		if err != nil {
			oneliners.FILE(err)
			continue
		}
		fmt.Printf("msg data: %q on subject %q\n", string(msg.Data), msg.Subject)
	}
}

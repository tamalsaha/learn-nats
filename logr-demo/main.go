package main

import (
	"fmt"
	"github.com/nats-io/nats.go"
	"log"
	"os"
	"time"

	"github.com/go-logr/logr"
	"github.com/go-logr/logr/funcr"
)

var LogNatsError = true

// prefix will be always "", since the NewFormatter() is used without any prefix

func NewAsync(nc *nats.Conn, subj string) logr.Logger {
	return NewAsyncWithOptions(nc, subj, funcr.Options{})
}

func NewAsyncWithOptions(nc *nats.Conn, subj string, opts funcr.Options) logr.Logger {
	return funcr.New(func(_, args string) {
		if err := nc.Publish(subj, []byte(args)); err != nil && LogNatsError {
			_, _ = fmt.Fprintln(os.Stderr, "failed to publish to nats", err)
		}
	}, opts)
}

func NewSync(nc *nats.Conn, subj string, timeout time.Duration) logr.Logger {
	return NewSyncWithOptions(nc, subj, timeout, funcr.Options{})
}

func NewSyncWithOptions(nc *nats.Conn, subj string, timeout time.Duration, opts funcr.Options) logr.Logger {
	return funcr.New(func(_, args string) {
		if _, err := nc.Request(subj, []byte(args), timeout); err != nil && LogNatsError {
			_, _ = fmt.Fprintln(os.Stderr, "failed to publish to nats", err)
		}
	}, opts)
}

/*
nats-server -p 4222 -m 8222 -js --user myname --pass password

nats sub msg.test --user myname --password password --connection-name=demo
*/
func main() {
	// Set a user and plain text password
	nc, err := nats.Connect("127.0.0.1", nats.UserInfo("myname", "password"))
	if err != nil {
		log.Fatal(err)
	}
	// defer nc.Close()
	defer nc.Drain()

	l := NewAsync(nc, "msg.test")
	l.Info("default info log", "stringVal", "value", "intVal", 12345)
	l.V(0).Info("V(0) info log", "stringVal", "value", "intVal", 12345)
	l.Error(fmt.Errorf("an error"), "error log", "stringVal", "value", "intVal", 12345)

	select {}
}

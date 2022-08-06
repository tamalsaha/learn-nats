package main

import (
	"bytes"
	"fmt"
	"github.com/nats-io/nats.go"
	"os"
	"sync"
	"time"

	"github.com/go-logr/logr"
	"github.com/go-logr/logr/funcr"
)

var LogNatsError = true

var pool = sync.Pool{
	New: func() interface{} {
		return new(bytes.Buffer)
	},
}

func NewAsync(nc *nats.Conn, subj string) logr.Logger {
	return NewAsyncWithOptions(nc, subj, funcr.Options{})
}

func NewAsyncWithOptions(nc *nats.Conn, subj string, opts funcr.Options) logr.Logger {
	return funcr.New(func(prefix, args string) {
		data := []byte(args)
		if prefix != "" {
			buf := pool.Get().(*bytes.Buffer)
			buf.Reset()
			buf.WriteString(prefix)
			buf.WriteString(": ")
			buf.WriteString(args)
			data = buf.Bytes()
			pool.Put(buf)
		}
		if err := nc.Publish(subj, data); err != nil && LogNatsError {
			_, _ = fmt.Fprintln(os.Stderr, err)
		}
	}, opts)
}

func NewSync(nc *nats.Conn, subj string, timeout time.Duration) logr.Logger {
	return NewSyncWithOptions(nc, subj, timeout, funcr.Options{})
}

func NewSyncWithOptions(nc *nats.Conn, subj string, timeout time.Duration, opts funcr.Options) logr.Logger {
	return funcr.New(func(prefix, args string) {
		data := []byte(args)
		if prefix != "" {
			buf := pool.Get().(*bytes.Buffer)
			buf.Reset()
			buf.WriteString(prefix)
			buf.WriteString(": ")
			buf.WriteString(args)
			data = buf.Bytes()
			pool.Put(buf)
		}
		if _, err := nc.Request(subj, data, timeout); err != nil && LogNatsError {
			_, _ = fmt.Fprintln(os.Stderr, err)
		}
	}, opts)
}

// NewStdoutLogger returns a logr.Logger that prints to stdout.
func NewStdoutLogger() logr.Logger {
	return funcr.New(func(prefix, args string) {
		if prefix != "" {
			_ = fmt.Sprintf("%s: %s\n", prefix, args)
		} else {
			fmt.Println(args)
		}
	}, funcr.Options{})
}

func main() {
	l := NewStdoutLogger()
	l.Info("default info log", "stringVal", "value", "intVal", 12345)
	l.V(0).Info("V(0) info log", "stringVal", "value", "intVal", 12345)
	l.Error(fmt.Errorf("an error"), "error log", "stringVal", "value", "intVal", 12345)
}

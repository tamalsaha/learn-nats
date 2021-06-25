package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/nats-io/nats.go"
	"github.com/tamalsaha/nats-hop-demo/shared"
	"github.com/tamalsaha/nats-hop-demo/transport"
	"github.com/unrolled/render"
	"go.wandrs.dev/binding"
	utilnet "k8s.io/apimachinery/pkg/util/net"
	"k8s.io/klog/v2"
	"net"
	"net/http"
	"os"
	"os/signal"
	"time"
	"k8s.io/apiserver/pkg/endpoints/handlers/responsewriters"
)

func main() {
	fmt.Println(shared.NATS_URL)
	nc, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer nc.Close()

	_, err = nc.QueueSubscribe("k8s", "NATS-RPLY-22", func(msg *nats.Msg) {
		resp, err := respond(msg.Data)
		if err != nil {
			responsewriters.ErrorToAPIStatus(err)
		}


		var r transport.R
		err := json.Unmarshal(msg.Data, &r)
		if err != nil {
			klog.ErrorS(err, "failed to parse message")
			return
		}

		req, err := http.ReadRequest(bufio.NewReader(bytes.NewReader(r.Request)))
		if err != nil {
			klog.ErrorS(err, "failed to parse request")
			return
		}

		resp := nats.NewMsg(msg.Reply)
		resp.Data = []byte("response_from_go")
		if err := msg.RespondMsg(resp); err != nil {
			fmt.Println("----", err)
		}
	})
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	ic := make(chan os.Signal, 1)
	signal.Notify(ic, os.Interrupt)
	<-ic
}

// k8s.io/client-go/transport/cache.go
const idleConnsPerHost = 25

func respond(in []byte) (*http.Response, error) {
	var r transport.R
	err := json.Unmarshal(in, &r)
	if err != nil {
		return nil, err
	}

	req, err := http.ReadRequest(bufio.NewReader(bytes.NewReader(r.Request)))
	if err != nil {
		return nil, err
	}

	// cache transport
	rt := http.DefaultTransport
	if r.TLS != nil {
		dial := (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext

		tlsconfig, err := r.TLS.TLSConfigFor()
		if err != nil {
			return nil, err
		}
		rt = utilnet.SetTransportDefaults(&http.Transport{
			Proxy:               http.ProxyFromEnvironment,
			TLSHandshakeTimeout: 10 * time.Second,
			TLSClientConfig:     tlsconfig,
			MaxIdleConnsPerHost: idleConnsPerHost,
			DialContext:         dial,
			DisableCompression:  r.DisableCompression,
		})

	}

	httpClient := &http.Client{
		Transport: rt,
		Timeout:   r.Timeout,
	}
	return httpClient.Do(req)
}

func main_() {
	nc, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		klog.Fatalln(err)
	}
	defer nc.Close()

	i := 0
	for {
		msg := nats.NewMsg("k8s")
		msg.Data = []byte("x")

		if m, err := nc.RequestMsg(msg, 10*time.Second); err != nil {
			fmt.Println("***", err.Error())
		} else {
			fmt.Println(string(m.Data))
		}
		i++
		if i == 3 {
			break
		}
	}

	m := chi.NewRouter()
	m.Use(middleware.RequestID)
	m.Use(middleware.RealIP)
	m.Use(middleware.Logger) // middlewares.NewLogger()
	m.Use(middleware.Recoverer)
	m.Use(binding.Injector(render.New()))

	m.HandleFunc("/*", func(w http.ResponseWriter, r *http.Request) {
		var buf bytes.Buffer
		if err := r.Write(&buf); err != nil { // WriteProxy
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if _, err := w.Write(buf.Bytes()); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		msg, err := nc.RequestMsg(&nats.Msg{
			Subject: "k8s.hopper.nodes",
			Reply:   "",
			Header:  nil,
			Data:    buf.Bytes(),
			Sub:     nil,
		}, 5*time.Second)
		if err != nil {
			fmt.Println("-----------------", err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		fmt.Println(string(msg.Data))
		// w.Write(msg.Data)



		// this is the last in the chain, no not calling next.ServeHTTP()
		return
	})

	klog.Infoln()
	klog.Infoln("Listening on :4000")
	if err := http.ListenAndServe(":4000", m); err != nil {
		klog.Fatalln(err)
	}
}

package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/nats-io/nats.go"
	"github.com/tamalsaha/nats-hop-demo/shared"
	"github.com/tamalsaha/nats-hop-demo/transport"
	"github.com/unrolled/render"
	"go.wandrs.dev/binding"
	utilnet "k8s.io/apimachinery/pkg/util/net"
	"k8s.io/apiserver/pkg/endpoints/handlers/responsewriters"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/klog/v2"
	"k8s.io/klog/v2/klogr"
	cu "kmodules.xyz/client-go/client"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
)

var pool = sync.Pool{
	New: func() interface{} {
		return new(bytes.Buffer)
	},
}

func main() {
	nc, err := nats.Connect(shared.NATS_URL)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer nc.Close()

	ctrl.SetLogger(klogr.New())
	cfg := ctrl.GetConfigOrDie()

	mapper, err := apiutil.NewDynamicRESTMapper(cfg)
	if err != nil {
		panic(err)
	}

	c, err := client.New(cfg, client.Options{
		Scheme: clientgoscheme.Scheme,
		Mapper: mapper,
		Opts: client.WarningHandlerOptions{
			SuppressWarnings:   false,
			AllowDuplicateLogs: false,
		},
	})
	if err != nil {
		panic(err)
	}

	uid, err := cu.ClusterUID(c)

	_, err = nc.QueueSubscribe("proxy."+uid, "proxy."+uid, func(msg *nats.Msg) {
		r2, req, resp, err := respond(msg.Data)
		if err != nil {
			status := responsewriters.ErrorToAPIStatus(err)
			data, _ := json.Marshal(status)

			resp = &http.Response{
				Status:           "", // status.Status,
				StatusCode:       int(status.Code),
				Proto:            "",
				ProtoMajor:       0,
				ProtoMinor:       0,
				Header:           nil,
				Body:             io.NopCloser(bytes.NewReader(data)),
				ContentLength:    int64(len(data)),
				TransferEncoding: nil,
				Close:            true,
				Uncompressed:     false,
				Trailer:          nil,
				Request:          nil,
				TLS:              nil,
			}
			if req != nil {
				resp.Proto = req.Proto
				resp.ProtoMajor = req.ProtoMajor
				resp.ProtoMinor = req.ProtoMinor

				resp.TransferEncoding = req.TransferEncoding
				resp.Request = req
				resp.TLS = req.TLS
			}
			if r2 != nil {
				resp.Uncompressed = r2.DisableCompression
			}
		}

		buf := pool.Get().(*bytes.Buffer)
		defer pool.Put(buf)
		buf.Reset()

		respMsg := &nats.Msg{
			Subject: msg.Reply,
		}
		if err := resp.Write(buf); err != nil { // WriteProxy
			respMsg.Data = []byte(err.Error())
		} else {
			respMsg.Data = buf.Bytes()
		}

		if err := msg.RespondMsg(respMsg); err != nil {
			klog.ErrorS(err, "failed to respond to message")
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

func respond(in []byte) (*transport.R, *http.Request, *http.Response, error) {
	var r transport.R
	err := json.Unmarshal(in, &r)
	if err != nil {
		return nil, nil, nil, err
	}

	req, err := http.ReadRequest(bufio.NewReader(bytes.NewReader(r.Request)))
	if err != nil {
		return &r, nil, nil, err
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
			return &r, req, nil, err
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

	//req.URL = nil
	req.RequestURI = ""

	httpClient := &http.Client{
		Transport: rt,
		Timeout:   r.Timeout,
	}
	resp, err := httpClient.Do(req)
	return &r, req, resp, err
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

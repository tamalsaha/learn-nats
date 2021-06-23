package main

import (
	"bytes"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/nats-io/nats.go"
	"github.com/unrolled/render"
	"go.wandrs.dev/binding"
	"k8s.io/klog/v2"
	"net/http"
	"time"
)

func main() {
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

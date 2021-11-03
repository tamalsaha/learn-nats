package main

import (
	"context"
	"github.com/hashicorp/go-cleanhttp"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptrace"
)

func main() {
	// client trace to log whether the request's underlying tcp connection was re-used
	clientTrace := &httptrace.ClientTrace{
		GotConn: func(info httptrace.GotConnInfo) { log.Printf("conn was reused: %t", info.Reused) },
	}
	traceCtx := httptrace.WithClientTrace(context.Background(), clientTrace)

	// 1st request
	req, err := http.NewRequestWithContext(traceCtx, http.MethodGet, "http://google.com", nil)
	if err != nil {
		log.Fatal(err)
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	if _, err := io.Copy(ioutil.Discard, res.Body); err != nil {
		log.Fatal(err)
	}
	res.Body.Close()
	// 2nd request
	req, err = http.NewRequestWithContext(traceCtx, http.MethodGet, "http://google.com", nil)
	if err != nil {
		log.Fatal(err)
	}
	_, err = http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal(err)
	}

	cleanhttp.DefaultPooledClient().Do(req)
}

package main

import (
	"bytes"
	"crypto/tls"
	"flag"
	"golang.org/x/net/http2"
	"io"
	"net/http"
	"sync"

	log "github.com/sirupsen/logrus"
)

func main() {

	flag.Parse()
	urls := flag.Args()

	log.Info("HTTP2.0 CLIENT %s", urls)

	tr := &http2.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	hclient := &http.Client{Transport: tr}

	var wg sync.WaitGroup
	wg.Add(len(urls))
	for _, addr := range urls {
		log.Info("GET %s", addr)
		go func(addr string) {
			rsp, err := hclient.Get(addr)
			if err != nil {
				panic(err)
			}
			log.Info("Got response for %s: %#v", addr, rsp)

			body := &bytes.Buffer{}
			_, err = io.Copy(body, rsp.Body)
			if err != nil {
				panic(err)
			}
			log.Info("Request Body:")
			log.Info("%s", body.Bytes())
			wg.Done()
		}(addr)
	}
	wg.Wait()
	log.Info("HTTP2.0 CLIENT BITTI")
}

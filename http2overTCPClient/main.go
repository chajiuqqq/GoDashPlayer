package main

import (
	"bytes"
	"crypto/tls"
	"flag"
	"golang.org/x/net/http2"
	"io"
	"net"
	"net/http"
	"sync"

	log "github.com/sirupsen/logrus"
)

func main() {

	flag.Parse()
	urls := flag.Args()

	log.Info("HTTP2.0 CLIENT %s", urls)

	hclient := http.Client{
		// Skip TLS dial
		Transport: &http2.Transport{
			DialTLS: func(netw, addr string, cfg *tls.Config) (net.Conn, error) {
				return net.Dial(netw, addr)
			},
			AllowHTTP: true,
		},
	}

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
}

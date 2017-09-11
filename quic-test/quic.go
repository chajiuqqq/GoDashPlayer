package main

import (
	"bytes"
	"crypto/tls"
	"flag"
	"github.com/lucas-clemente/quic-go/h2quic"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/http2"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

func main() {
	quic := flag.Bool("quic", false, "quic")
	h2 := flag.Bool("h2", false, "h2")
	flag.Parse()
	args := flag.Args()

	segment_limit, _ := strconv.Atoi(args[0])
	url_base := args[1]
	//url_base = "https://caddy.quic/media/BigBuckBunny/4sec/bunny_3936261bps/BigBuckBunny_4s$SEGMENT$.m4s"
	var hclient *http.Client

	urls := make([]string, segment_limit)

	for i := 0; i < len(urls); i++ {
		urls[i] = strings.Replace(url_base, "$SEGMENT$", strconv.Itoa(i+1), -1)
	}

	if *quic {
		log.Info("QUIC CLIENT")
		log.Info(url_base)

		hclient = &http.Client{
			Transport: &h2quic.RoundTripper{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}},
		}
	} else if *h2 {
		log.Info("HTTP2 CLIENT")
		log.Info(url_base)

		tr := &http2.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		hclient = &http.Client{Transport: tr}

	} else {
		log.Info("HTTP1.1 CLIENT")
		log.Info(url_base)

		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		hclient = &http.Client{Transport: tr}

	}
	var wg sync.WaitGroup
	wg.Add(len(urls))
	startTime := GetNow()
	for _, addr := range urls {
		log.Info("GET %s", addr)
		go func(addr string) {
			log.Info("Downloading %s", addr)
			rsp, err := hclient.Get(addr)
			if err != nil {
				panic(err)
			}
			defer rsp.Body.Close()

			body := &bytes.Buffer{}
			_, err = io.Copy(body, rsp.Body)
			if err != nil {
				panic(err)
			}
			//utils.Infof("Request Body:")
			log.Info("%s", body.Bytes())
			log.Info("Finished %s:", addr)
			wg.Done()
		}(addr)
		//	time.Sleep(100 * time.Millisecond)
	}
	wg.Wait()
	log.Info("TOTAL DURATION: %s", FloatToString((GetNow() - startTime)))
}

func FloatToString(input_num float64) string {

	// to convert a float number to a string, precision 2 digits
	return strconv.FormatFloat(input_num, 'f', 2, 64)
}

func GetNow() float64 {

	now := time.Now()
	secs := now.Unix()
	nanos := now.UnixNano()

	// Note that there is no `UnixMillis`, so to get the
	// milliseconds since epoch you'll need to manually
	// divide from nanoseconds.
	millis := nanos / 10000000
	str := strconv.FormatInt(secs, 10) + "." + strconv.FormatInt(millis-secs*100, 10)
	f, err := strconv.ParseFloat(str, 64)
	if err != nil {
		panic(err)
	}
	return f
}

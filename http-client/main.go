package main

import (
	"bytes"
	"crypto/tls"
	"flag"
	"io"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/sevketarisu/GoDashPlayer/utils"
)

func main() {
	verbose := flag.Bool("v", false, "verbose")
	flag.Parse()
	urls := flag.Args()

	if *verbose {
		utils.SetLogLevel(utils.LogLevelDebug)
	} else {
		utils.SetLogLevel(utils.LogLevelInfo)
	}
	utils.SetLogTimeFormat("")

	utils.Infof("HTTP1.1 CLIENT %s", urls)

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	hclient := &http.Client{Transport: tr}

	var wg sync.WaitGroup
	wg.Add(len(urls))
	startTime := GetNow()
	for _, addr := range urls {
		utils.Infof("GET %s", addr)
		go func(addr string) {
			rsp, err := hclient.Get(addr)
			if err != nil {
				panic(err)
			}
			utils.Infof("Got response for %s: %#v", addr, rsp)

			body := &bytes.Buffer{}
			_, err = io.Copy(body, rsp.Body)
			if err != nil {
				panic(err)
			}
			utils.Infof("Response Body of:", addr)
			utils.Infof("%s", body.Bytes())
			wg.Done()
		}(addr)
	}
	wg.Wait()
	utils.Infof("TOTAL DURATION: %s", FloatToString((GetNow() - startTime)))
	utils.Infof("HTTP1.1 CLIENT BITTI")
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

func FloatToString(input_num float64) string {

	// to convert a float number to a string, precision 2 digits
	return strconv.FormatFloat(input_num, 'f', 2, 64)
}

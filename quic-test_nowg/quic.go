package main

import (
	log "github.com/sirupsen/logrus"
	//"bytes"
	"crypto/tls"
	"flag"
	"github.com/lucas-clemente/quic-go/h2quic"
	//"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func main() {

	quic := flag.Bool("quic", false, "quic")
	flag.Parse()
	args := flag.Args()

	//	urls := flag.Args()
	//urls := []string{"a", "b", "c", "d"}

	//url_base := "https://caddy.quic/media/BigBuckBunny/4sec/bunny_3936261bps/BigBuckBunny_4s$SEGMENT$.m4s"

	segment_limit, _ := strconv.Atoi(args[0])
	url_base := args[1]
	//url_base = "https://caddy.quic/media/BigBuckBunny/4sec/bunny_3936261bps/BigBuckBunny_4s$SEGMENT$.m4s"
	var hclient *http.Client

	var urls []string
	urls = make([]string, segment_limit)

	for i := 0; i < len(urls); i++ {
		urls[i] = strings.Replace(url_base, "$SEGMENT$", strconv.Itoa(i+1), -1)
	}

	if *quic {
		log.Info("QUIC CLIENT")
		log.Info(url_base)
		hclient = &http.Client{
			Transport: &h2quic.RoundTripper{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}},
		}
	} else {
		log.Info("HTTP CLIENT")
		log.Info(url_base)
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		hclient = &http.Client{Transport: tr}
	}
	//var wg sync.WaitGroup
	//wg.Add(len(urls))
	startTime := GetNow()
	for _, addr := range urls {
		//utils.Infof("GET %s", addr)

		log.Info("Downloading %s", addr)
		response, err := hclient.Get(addr)
		log.Info("get finished %s", addr)
		if err != nil {
			panic(err)
		}
		defer response.Body.Close()

		/*	body := &bytes.Buffer{}
			_, err = io.Copy(body, response.Body)
			if err != nil {
				panic(err)
			}
		*/

		//	utils.Infof("Request Body:")
		//	utils.Infof("%s", body.Bytes())
		log.Info("Finished %s:", addr)

	}

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

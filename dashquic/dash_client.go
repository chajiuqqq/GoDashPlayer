package main

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"dashquic/config"
	"flag"
	"fmt"
	h2quic "github.com/quic-go/quic-go/http3"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/http2"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Command Line  Parameters
var segmentLimitParameter int
var list bool
var download bool
var playbackType string
var mpdFullUrl string
var quic bool
var h2 bool
var h1 bool
var LOCAL_TEMP_DIR string
var jumpSecondsArgs []string
var proto bool
var java bool
var protoClientPort string
var protoClientHostName string
var repeatCount int
var sumOfTotalDownloaded float64
var sumOfPlaybackTime float64

type DashPlayback struct {
	minBufferTime    float64                  // as seconds
	playbackDuration float64                  // as seconds
	video            map[float64]*MediaObject // MediaObject objesi MPD dosyasinda representationa karsilik geliyor
}

// MediaObject objesi MPD dosyasinda representationa karsilik geliyor
type MediaObject struct {
	id              string
	start           int
	timeScale       float64
	segmentDuration float64
	initialization  string
	baseUrl         string
	urlList         []string
	segmentSizes    []float64
}

func init() {

}

func main() {
	fmt.Println("###dash_client.main() STARTING###")

	// get verbose flag from program arguments
	verbose := flag.Bool("v", false, "verbose")
	jump := flag.Bool("j", false, "jump")

	// log.SetLevel(log.ErrorLevel)

	// get other program arguments
	parse_arguments()
	print_arguments()

	// logger
	// 创建日志文件
	file, err := os.OpenFile("logs_20mins.json", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	// 设置日志格式为JSON
	log.SetFormatter(&log.JSONFormatter{})

	// 创建多个日志输出器
	multiWriter := io.MultiWriter(os.Stdout, file)

	// 设置输出器
	log.SetOutput(multiWriter)

	// set log level
	if *verbose {
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetLevel(log.ErrorLevel)
	}
	if *jump {
		jumpSecondsArgs = flag.Args()
	}

	log.WithFields(log.Fields{
		"jumpSecondsArgs": jumpSecondsArgs,
	}).Info("DASH_CLIENT:")

	// download mpd file and return a http client
	log.WithFields(log.Fields{
		"mpdFile": mpdFullUrl,
	}).Info("DASH_CLIENT: Downloading MPD file")

	mpdFile, hclient, cmdWriter, scanner, localTempFolder := get_mpd(mpdFullUrl)

	log.WithFields(log.Fields{
		"mpdFile": mpdFile,
	}).Info("DASH_CLIENT: Downloaded MPD file")

	// get domain name from mpd full url
	domainName := get_domain_name(mpdFullUrl)

	sumOfTotalDownloaded = 0.0
	sumOfPlaybackTime = 0.0

	for runNo := 1; runNo <= repeatCount; runNo++ {

		// create a new DashPlayback object and hold pointer
		dpObject := &DashPlayback{} // &DashPlayback{} = new(DashPlayback)

		log.WithFields(log.Fields{
			"mpdFile": mpdFile,
		}).Info("DASH_CLIENT: Reading MPD file")

		// read mpd file and get videoSegmentDuration
		videoSegmentDuration := read_mpd(mpdFile, dpObject)

		log.WithFields(log.Fields{
			"representation count": len(dpObject.video),
		}).Info("DASH_CLIENT: MPD file DASH media video")

		// if -l is provided in program arguments only list presentations and exit immediately

		if list {
			log.Info("DASH_CLIENT: Printing representations")
			print_representations(dpObject)
			log.Info("DASH_CLIENT: Printed representations")
		}

		log.Info("DASH_CLIENT: Starting BASIC playback")

		// start playback with BASIC algorith

		fmt.Println("###dash_client.main() starting  start_playback_smart: ### Run No:", runNo, "/", repeatCount)
		playbackTime, totalDownloaded := start_playback_smart(hclient, cmdWriter, scanner, dpObject, domainName, "BASIC", download, videoSegmentDuration, jumpSecondsArgs, localTempFolder, runNo)

		fmt.Println("###dash_client.main() playbackTime:       ### Run No:", runNo, "->", playbackTime)
		fmt.Println("###dash_client.main() totalDownloaded KB: ### Run No:", runNo, "->", totalDownloaded/1024)
		fmt.Println("###dash_client.main() Bytes per Second:   ### Run No:", runNo, "->", totalDownloaded/playbackTime)

		sumOfTotalDownloaded = sumOfTotalDownloaded + totalDownloaded
		sumOfPlaybackTime = sumOfPlaybackTime + playbackTime
	}

	fmt.Println("###dash_client.main() TOTAL playbackTime: ###  :", sumOfPlaybackTime)
	fmt.Println("###dash_client.main() TOTAL totalDownloaded KB: ###  :", sumOfTotalDownloaded/1024)
	fmt.Println("###dash_client.main() TOTAL Bytes per Second: ###  :", sumOfTotalDownloaded/sumOfPlaybackTime)

	if proto {
		io.WriteString(*cmdWriter, "exit\n")
	}

	fmt.Println("###dash_client.main() FINISHED###")

}

func start_playback_smart(hclient *http.Client, cmdWriter *io.WriteCloser, scanner *bufio.Scanner, dpObject *DashPlayback, domainName string,
	playbackType string, download bool, videoSegmentDuration float64, jumpSeconds []string, localTempFolder string, runNo int) (float64, float64) {

	// create a new DashPlayer object and hold pointer
	dp := &DashPlayer{}

	// initiliaze player
	dp.__init__(dpObject.playbackDuration, videoSegmentDuration)

	// start the player with another thread
	dp.start()

	// temp folder name to saving downloded segment files
	//	var fileIdentifier string = id_generator()

	log.WithFields(log.Fields{
		"localTempFolder": localTempFolder,
	}).Info("DASH_CLIENT: The segments are stored in folder")

	// presentations list  format:  map[segmentNumber]map[bitrate]segment_url
	var dpList = make(map[int]map[float64]string)

	// calculate segment count :  ((596+4)/4) + 1 (init) =151
	var segmentCountPerBitrate = int((dpObject.playbackDuration+videoSegmentDuration)/videoSegmentDuration) + 1 // +1 for init segment

	// initilizse segmentNumbers map    segmentNumber 1 = init segment
	for i := 1; i <= segmentCountPerBitrate; i++ {
		dpList[i] = make(map[float64]string)
	}

	// Loop for each bitrate in mpd file
	for bitrate := range dpObject.video {

		// # Getting the URL list for each bitrate
		get_url_list(dpObject.video[bitrate], videoSegmentDuration, dpObject.playbackDuration, bitrate)

		// replace $Bandwidth$ in segment url with releated bitrate value
		if strings.Contains(dpObject.video[bitrate].initialization, "$RepresentationID$") {
			dpObject.video[bitrate].initialization = strings.Replace(dpObject.video[bitrate].initialization, "$RepresentationID$", dpObject.video[bitrate].id, -1)
		}

		// create a temp string[] that contains only initialization url for releated bitrate
		initUrl := []string{dpObject.video[bitrate].initialization}

		// append urlList to initUrl
		mediaUrls := append(initUrl[:1], dpObject.video[bitrate].urlList...)

		// starting from dpObject.video[bitrate].start loop to len(mediaUrls) :  1 to 151
		for segmentCount := dpObject.video[bitrate].start; segmentCount <= len(mediaUrls); segmentCount++ {
			dpList[segmentCount][bitrate] = mediaUrls[segmentCount-1]
		}
	}

	var bitrates []float64

	// make distict bitrates list and sort ascending
	for keys, _ := range dpObject.video {
		bitrates = append(bitrates, keys)
	}
	sort.Float64s(bitrates)

	// variables for basic adaptation
	var averageDwnTime float64 = 0.0
	var previousSegmentTimes []float64
	var recentDownloadSizes []float64
	var totalDownloaded float64 = 0.0
	var delay float64 = 0.0 // Delay in terms of the number of segments
	var segmentDuration float64 = 0.0
	var segmentDownloadTime float64 = 0.0
	var segmentInfo SegmentInfo
	var currentBitrate = bitrates[0] // start with smallest bitrate in bitrates list
	var previousBitrate float64 = -1 // there's no previousBitrate before not playing any segment
	var jumpCounter int = 0

	for segmentNumber := dpObject.video[currentBitrate].start; segmentNumber <= len(dpList); segmentNumber++ { // segment=map[float64]string   segment[bitrate]=mediaUrl

		log.WithFields(log.Fields{
			"segmentNumber": segmentNumber,
			"playbackType":  playbackType,
		}).Info("DASH_CLIENT: Processing the segment")

		// if is it first segmnet set previousBitrate = currentBitrate
		if previousBitrate != -1 {
			previousBitrate = currentBitrate
		}

		// if segment limit parameter (-n) is given in program arguments
		if segmentLimitParameter > 0 {
			if dp.SegmentLimit == -1 { // if it is unSet in player, set it to segmentLimitParameter once
				dp.SegmentLimit = segmentLimitParameter
			}
			// if segment limit reached exit loop
			if segmentNumber > dp.SegmentLimit {
				log.WithFields(log.Fields{
					"SegmentLimit": dp.SegmentLimit,
				}).Info("DASH_CLIENT: Segment limit reached. Downloading segments will stop... ")
				break
			}
		}

		// if it is starting segment choose smallest bitrate, otherwise determine it with basic algorithm
		if segmentNumber == dpObject.video[currentBitrate].start {
			currentBitrate = bitrates[0]
		} else {
			if playbackType == "BASIC" {
				currentBitrate, averageDwnTime = basic_dash2(segmentNumber, bitrates, averageDwnTime, recentDownloadSizes, previousSegmentTimes, currentBitrate)
				if dp.Buffer.len() > config.BASIC_THRESHOLD {
					delay = float64(dp.Buffer.len() - config.BASIC_THRESHOLD)
				}
				log.WithFields(log.Fields{
					"Next Bitrate":        FloatToString(currentBitrate),
					"Next Segment Number": segmentNumber,
				}).Info("DASH_CLIENT: Basic-DASH: Selected bitrate for the next segment")

			} else {
				log.WithFields(log.Fields{
					"playbackType": playbackType,
				}).Error("DASH_CLIENT: Unknown playback type ")
			}
		}

		// segmnet url to download
		// todo fix url
		var segmentUrl string = domainName + "/www/video/20mins/" + dpList[segmentNumber][currentBitrate]

		// if delay occurred wait for a while before downloading next segment
		if delay > 0 {
			delayStart := GetNow()
			log.WithFields(log.Fields{
				"delay":                  delay,
				"delay*segment_duration": delay * segmentDuration,
			}).Info("DASH_CLIENT: SLEEPING...")
			for GetNow()-delayStart < (delay * segmentDuration) {
				time.Sleep(1 * time.Second)
			}
			log.Info("DASH_CLIENT: SLEPT")
			delay = 0
		}

		// download startTime
		var startTime float64 = GetNow()

		log.WithFields(log.Fields{
			"segmentUrl": segmentUrl,
		}).Info("DASH_CLIENT: Segment URL")

		// download segment file and get its size and name
		segmentSize, segmentFileName := download_segment(hclient, cmdWriter, scanner, segmentUrl, localTempFolder)

		log.WithFields(log.Fields{
			"segmentFileName": segmentFileName,
			"segmentSize":     FloatToString(segmentSize),
		}).Info("DASH_CLIENT: Downloaded segment")

		// calculate segmentDownloadTime
		segmentDownloadTime = GetNow() - startTime
		if segmentDownloadTime < 0.0 {
			segmentDownloadTime = 0.0
		}

		// add to previousSegmentTimes list and recentDownloadSizes size
		previousSegmentTimes = append(previousSegmentTimes, segmentDownloadTime)
		recentDownloadSizes = append(recentDownloadSizes, segmentSize)

		// downloaded total segment sizes so far
		totalDownloaded += segmentSize

		log.WithFields(log.Fields{
			"totalDownloaded": FloatToString(totalDownloaded),
			"segmentSize":     FloatToString(segmentSize),
			"segmentNumber":   segmentNumber,
		}).Info("DASH_CLIENT: The total downloaded, segment_size, segment_number")

		// initialize segment info before writing it to buffer
		segmentInfo = make(map[string]string)

		// assign segmentInfo properties
		segmentInfo["playback_length"] = strconv.FormatFloat(videoSegmentDuration, 'f', -1, 64)
		segmentInfo["size"] = strconv.FormatFloat(segmentSize, 'f', -1, 64)
		segmentInfo["bitrate"] = strconv.FormatFloat(currentBitrate, 'f', -1, 64)
		segmentInfo["data"] = segmentFileName
		segmentInfo["URI"] = segmentUrl
		segmentInfo["segment_number"] = strconv.Itoa(segmentNumber)

		// write current segment to player's buffer
		dp.writeToBuffer(segmentInfo)

		log.WithFields(log.Fields{
			"segmentSize":         FloatToString(segmentSize),
			"segmentNumber":       segmentNumber,
			"segmentDownloadTime": FloatToString(segmentDownloadTime),
		}).Info("DASH_CLIENT: DOWNLOADED:")

		// determine up_shift or down_shift
		if previousBitrate == -1 { // None
			if previousBitrate < currentBitrate {
				// fmt.Println("up_shift")

			} else if previousBitrate < currentBitrate {
				//	fmt.Println("down_shift")
			}
			previousBitrate = currentBitrate
		}

		// if player's buffer is full wait for player reads and deletes some segments
		for !(dp.BufferLength < dp.MaxBufferSize) {

			/*	log.WithFields(log.Fields{
				"len(dp.Buffer)":   len(dp.Buffer),
				"dp.MaxBufferSize": dp.MaxBufferSize,
				"dp.BufferLength":  dp.BufferLength,
			}).Info("Player Buffer is full, waiting for 0.5 sec")*/
			time.Sleep(100 * time.Millisecond)
		}

		// seek control

		/*	log.WithFields(log.Fields{
			"jumpCounter":          jumpCounter,
			"len(jumpSecondsArgs)": len(jumpSecondsArgs),
		}).Info("DASH_JUMP: Jump control")*/

		if jumpCounter < len(jumpSecondsArgs) {
			j := strings.Split(jumpSecondsArgs[jumpCounter], ",")
			jumpAtSecond, err := strconv.ParseFloat(j[0], 64)
			jumpToSecond, err := strconv.ParseFloat(j[1], 64)
			if err != nil {
				panic(err)
			}

			log.WithFields(log.Fields{
				"dp.CurrentSegment": dp.getCurrentSegment(),
				"jumpAtSecond":      jumpAtSecond,
				"(dp.CurrentSegment * videoSegmentDuration)": int(dp.getCurrentSegment() * videoSegmentDuration),
			}).Info("DASH_JUMP: Checking Jump")

			// if dp.PlaybackTimer.time() >= jumpAtSecond {
			if (dp.getCurrentSegment() * videoSegmentDuration) >= jumpAtSecond {
				log.WithFields(log.Fields{
					"dp.CurrentSegment": dp.getCurrentSegment(),
					"jumpAtSecond":      jumpAtSecond,
					"jumpToSecond":      jumpToSecond,
					"BufferLength":      dp.BufferLength,
					"dp.Buffer.Len()":   dp.Buffer.len(),
					"jumpCounter":       jumpCounter,
				}).Info("DASH_JUMP: Clearing Buffer:")

				dp.clearBuffer()
				jumpCounter++
				segmentNumber = int(jumpToSecond / videoSegmentDuration)
				dp.setCurrentSegment(float64(segmentNumber))

				log.WithFields(log.Fields{
					"jumpAtSecond":    jumpAtSecond,
					"jumpToSecond":    jumpToSecond,
					"BufferLength":    dp.BufferLength,
					"dp.Buffer.Len()": dp.Buffer.len(),
					"segmentNumber":   segmentNumber,
					"jumpCounter":     jumpCounter,
				}).Info("DASH_JUMP: Jumped To Segment:")

			}
		}
		//	log.Info("DASH_CLIENT: jumpCounter:", jumpCounter)

		if int(segmentNumber)%3 == 0 {
			fmt.Println("###dash_client.main().startPlaybackSmart dp.PlaybackTimer  :##", runNo, "->", segmentNumber, "->", dp.PlaybackTimer.time())
			fmt.Println("###dash_client.main().startPlaybackSmart totalDownloaded KB:##", runNo, "->", segmentNumber, "->", totalDownloaded/1024)
			fmt.Println("###dash_client.main().startPlaybackSmart Bytes per Second  :##", runNo, "->", segmentNumber, "->", totalDownloaded/dp.PlaybackTimer.time())
		}
	} // main for loop for downloading segments

	// after downloading all segments wait for player to stop
	for !(dp.PlaybackState == "STOP" || dp.PlaybackState == "END") { // while player is  not stopped
		time.Sleep(1 * time.Second)
		log.Info("DASH_CLIENT: Client is waiting for player stop, current player state:", dp.PlaybackState)
	}
	return dp.PlaybackTimer.time(), totalDownloaded
} // end of start_playback_smart

/*
download_segment: downloads and saves requested segment file and returns it's size and saved local file name
*/
func download_segment(hclient *http.Client, cmdWriter *io.WriteCloser, scanner *bufio.Scanner, segmentUrl string, dashFolder string) (float64, string) {

	splitedBySlash := strings.SplitAfterN(segmentUrl, "/", strings.Count(segmentUrl, "/")+1)
	segmentLocalFileName := dashFolder + splitedBySlash[strings.Count(segmentUrl, "/")]

	if !proto {

		response, err := hclient.Get(segmentUrl)
		if err != nil {
			panic(err)
		}
		defer response.Body.Close()

		// Create directory if not exits
		if _, err := os.Stat(dashFolder); os.IsNotExist(err) {
			os.Mkdir(dashFolder, os.ModePerm)
		}

		// open a file for writing
		file, err := os.Create(segmentLocalFileName)
		if err != nil {
			panic(err)
		}
		// Use io.Copy to just dump the response body to the file. This supports huge files
		_, err = io.Copy(file, response.Body)
		if err != nil {
			panic(err)
		}
		file.Close()
		return float64(response.ContentLength), segmentLocalFileName
	} else {

		done := make(chan bool)

		var fileSize float64

		io.WriteString(*cmdWriter, segmentUrl+"\n")
		go func() {
			fileSize = download_with_console(done, segmentUrl, scanner)
		}()
		//	download_with_console(segmentUrl, scanner)
		// fmt.Println("waiting for download... ", segmentUrl)
		// wait for dowload finish
		// Block until we receive a notification from the
		// worker on the channel.

		<-done

		return fileSize, segmentLocalFileName

	}

}

func basic_dash2(segmentNumber int, bitrates []float64, averageDwnTime float64,
	recentDownloadSizes []float64, previousSegmentTimes []float64, currentBitrate float64) (float64, float64) {
	// slice pop ornekleri
	// https://github.com/golang/go/wiki/SliceTricks

	for len(previousSegmentTimes) > config.BASIC_DELTA_COUNT {
		previousSegmentTimes = previousSegmentTimes[:len(previousSegmentTimes)-1] // POP from slice
	}
	for len(recentDownloadSizes) > config.BASIC_DELTA_COUNT {
		recentDownloadSizes = recentDownloadSizes[:len(recentDownloadSizes)-1] // POP from slice
	}

	if len(previousSegmentTimes) == 0 || len(recentDownloadSizes) == 0 {
		return bitrates[0], -1 // -1=None
	}

	updatedDwnTime := sumFloat64(previousSegmentTimes) / float64(len(previousSegmentTimes))

	// Calculate the running download_rate in Kbps for the most recent segments*
	downloadRate := sumFloat64(recentDownloadSizes) * 8 / (updatedDwnTime * float64(len(previousSegmentTimes)))

	sort.Float64s(bitrates)
	nextRate := bitrates[0]

	// Check if we need to increase or decrease bitrate
	if downloadRate > currentBitrate*config.BASIC_UPPER_THRESHOLD { // # Increase rate only if  download_rate is higher by a certain margin
		// Check if the bitrate is already at max
		if currentBitrate == bitrates[len(bitrates)-1] {
			nextRate = currentBitrate
		} else {
			// if the bitrate is not at maximum then select the next higher bitrate
			currentIndex := getIndexOf(bitrates, currentBitrate)
			nextRate = bitrates[currentIndex+1]
		}
	} else {
		// If the download_rate is lower than the current bitrate then pick the most suitable bitrate
		for i := 1; i < len(bitrates); i++ {
			if downloadRate > bitrates[i]*config.BASIC_UPPER_THRESHOLD {
				nextRate = bitrates[i]
			} else {
				nextRate = bitrates[i-1]
				break
			}
		}
	}

	log.WithFields(log.Fields{
		"downloadRate": FloatToString(downloadRate),
		"nextRate":     FloatToString(nextRate),
	}).Info("DASH_CLIENT: Basic Adaptation")

	return nextRate, updatedDwnTime
}

func getIndexOf(slice []float64, value float64) int {
	for p, v := range slice {
		if v == value {
			return p
		}
	}
	return -1
}

func sumFloat64(floats []float64) float64 {
	sum := 0.0
	for _, float := range floats {
		sum += float
	}
	return sum
}

func get_url_list(media *MediaObject, segmentDuration float64, playbackDuration float64, bitrate float64) { // *MediaObject {
	//  # Counting the init file
	totalPlayback := segmentDuration
	segmentCount := media.start
	// # Get the Base URL string
	baseUrl := media.baseUrl

	if strings.Contains(baseUrl, "$RepresentationID$") {
		baseUrl = strings.Replace(baseUrl, "$RepresentationID$", media.id, -1)
	}
	// if strings.Contains(baseUrl, "$Number$") {
	// 	splitedBaseUrl := strings.Split(baseUrl, "$")
	// 	splitedBaseUrl[1] = strings.Replace(splitedBaseUrl[1], "$", "", -1)
	// 	splitedBaseUrl[1] = strings.Replace(splitedBaseUrl[1], "Number", "", -1)
	// 	baseUrl = strings.Join(splitedBaseUrl, "")
	// }
	for {
		media.urlList = append(media.urlList, strings.Replace(baseUrl, "$Number$", strconv.Itoa(segmentCount), -1))
		segmentCount += 1
		if totalPlayback > playbackDuration {
			break
		}
		totalPlayback += segmentDuration
	}
	return
}

func read_mpd(mpdFile string, dpObject *DashPlayback) float64 {

	var videoSegmentDuration float64

	mpdData, err := ioutil.ReadFile(mpdFile)
	if err != nil {
		panic(err)
	}

	mpd := MPD{} // new(MPD) //&MPD{}=new(MPD)

	err = mpd.Decode(mpdData)
	if err != nil {

		panic(err)
	}

	dpObject.playbackDuration = get_playback_duration(mpd.MediaPresentationDuration)
	dpObject.minBufferTime = 1.5 // 1.5 olmali get_playback_duration(mpd.MinBufferTime)  format: PT1.500000S

	moMap := make(map[float64]*MediaObject)
	for _, as := range mpd.Period.AdaptationSets {
		if !strings.Contains(as.MimeType, "video") {
			continue
		}
		for _, rep := range as.Representations {

			mediaObject := &MediaObject{}
			mediaObject.baseUrl = as.SegmentTemplate.Media
			mediaObject.start = as.SegmentTemplate.StartNumber
			mediaObject.timeScale = as.SegmentTemplate.Timescale
			mediaObject.initialization = as.SegmentTemplate.Initialization
			mediaObject.id = rep.ID

			for _, seg := range rep.SegmentSizes {
				mediaObject.segmentSizes = append(mediaObject.segmentSizes, seg.Size*seg.getSegmentScaleAsFloat(seg.Scale))
			}

			videoSegmentDuration = float64(as.SegmentTemplate.Duration / as.SegmentTemplate.Timescale)
			moMap[rep.Bandwidth] = mediaObject
		}
	}
	dpObject.video = moMap

	return videoSegmentDuration
}

/*
get_playback_duration: parses playback duration, format: PT0H9M56.46S=>9 Minutes 56 Seconds 46 miliseconds => 596.46 secondes
*/
func get_playback_duration(MediaPresentationDuration string) float64 {
	re := regexp.MustCompile(`[PTHM.S]`)
	splited := re.Split(MediaPresentationDuration, -1)
	var d = []int{}
	for _, i := range splited {
		if len(i) > 0 { // skip "" words in splited array
			j, err := strconv.Atoi(i)
			if err != nil {
				panic(err)
			}
			d = append(d, j)
		}
	}
	return float64(d[0]*60*60 + d[1]*60 + d[2])
}

/*
get_domain_name: finds domain name and protocol from mpdFileName
*/
func get_domain_name(mpdFileName string) string {
	u, err := url.Parse(mpdFileName)
	if err != nil {
		panic(err)
	}
	return u.Scheme + "://" + u.Host
}

/*
print_presentations: prints presentations in mpd file and exits
*/
func print_representations(dpObject *DashPlayback) {

	for k, m := range dpObject.video {

		log.WithFields(log.Fields{
			"key":              FloatToString(k),
			"playbackDuration": dpObject.playbackDuration,
			"baseUrl":          m.baseUrl,
			"initialization":   m.initialization,
			"minBufferTime":    dpObject.minBufferTime,
			"segmentDuration":  m.segmentDuration,
			"start":            m.start,
			"timeScale":        m.timeScale,
		}).Info("DASH_CLIENT: Representations")
	}
	os.Exit(0)
}

/*
get_mpd: downloads and saves mpd file and returns quic or http2 or http client pointer
*/
func get_mpd(mpdFullUrl string) (string, *http.Client, *io.WriteCloser, *bufio.Scanner, string) {

	var mpdLocalFileName string
	var hclient *http.Client
	var adrCmdWriter *io.WriteCloser
	//	var cmdWriter *io.WriteCloser
	var scanner *bufio.Scanner
	var localTempFolder string
	var cmdName string
	var cmdArgs []string

	localTempFolder = get_random_folder_path()

	splitedBySlash := strings.SplitAfterN(mpdFullUrl, "/", strings.Count(mpdFullUrl, "/")+1)
	mpdLocalFileName = localTempFolder + splitedBySlash[strings.Count(mpdFullUrl, "/")]

	_ = os.Mkdir(localTempFolder, os.FileMode(0777))

	if quic { // use QUIC
		hclient = &http.Client{
			Transport: &h2quic.RoundTripper{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}},
		}
	} else if h2 { // use HTTP2
		tr := &http2.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		hclient = &http.Client{Transport: tr}
	} else if h1 { // use HTTP1.1
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		hclient = &http.Client{Transport: tr}
	}

	if quic || h2 || h1 {

		response, err := hclient.Get(mpdFullUrl)
		if err != nil {
			panic(err)
		}
		defer response.Body.Close()

		// utils.Infof("Got response for %s: %#v", mpdFullUrl, response)
		body := &bytes.Buffer{}
		_, err = io.Copy(body, response.Body)
		if err != nil {
			panic(err)
		}

		// utils.Infof("Writing mpd file: %s", mpdLocalFileName)
		err = ioutil.WriteFile(mpdLocalFileName, body.Bytes(), 0644)
		//	utils.Infof("Saved mpd file: %s", mpdLocalFileName)

		if err != nil {
			panic(err)
		}
	} else {
		if proto {
			cmdName = config.PROTO_QUIC_CLIENT_EXEC
			cmdArgs = []string{"--host=" + protoClientHostName, // ec2-13-56-87-190.us-west-1.compute.amazonaws.com",
				"--port=" + protoClientPort,
				"--v=0",
				"--disable-certificate-verification",
				"--folder=" + get_random_folder_path(),
				"x"}

			log.WithFields(log.Fields{
				"cmdArgs": cmdArgs,
				"cmdName": cmdName,
			}).Info("DASH_CLIENT: cmdArgs")
		}

		if java {
			cmdName = "/usr/lib/jvm/java-8-oracle/bin/java"
			/*	cmdArgs = []string{"-classpath",
				"/home/sevket/eclipse_workspace/javahttpclient/target/classes:" +
					"/home/sevket/.m2/repository/org/apache/httpcomponents/httpclient/4.5.2/httpclient-4.5.2.jar:" +
					"/home/sevket/.m2/repository/org/apache/httpcomponents/httpcore/4.4.4/httpcore-4.4.4.jar:" +
					"/home/sevket/.m2/repository/commons-logging/commons-logging/1.2/commons-logging-1.2.jar:" +
					"/home/sevket/.m2/repository/commons-codec/commons-codec/1.9/commons-codec-1.9.jar",
				"javahttpclient.DownloadFileHttpCilent" + " " + get_random_folder_path()
			}*/
			cmdArgs = []string{"java -version"}

			log.WithFields(log.Fields{
				"cmdArgs": cmdArgs,
				"cmdName": cmdName,
			}).Info("DASH_CLIENT: cmdArgs")
		}

		cmd := exec.Command(cmdName, cmdArgs...)
		cmdReader, err1 := cmd.StdoutPipe()
		if err1 != nil {
			fmt.Fprintln(os.Stderr, "Error creating StdoutPipe for Cmd", err1)
			os.Exit(1)
		}

		cmdWriter, err2 := cmd.StdinPipe()
		adrCmdWriter = &cmdWriter
		if err2 != nil {
			fmt.Fprintln(os.Stderr, "Error creating StdinPipe for Cmd", err2)
			os.Exit(1)
		}
		scanner = bufio.NewScanner(cmdReader)
		err1 = cmd.Start()
		if err1 != nil {
			fmt.Fprintln(os.Stderr, "Error starting Cmd", err1)
			os.Exit(1)
		}

		done := make(chan bool)
		io.WriteString(cmdWriter, mpdFullUrl+"\n")
		go download_with_console(done, mpdFullUrl, scanner)
		//	download_with_console(mpdFullUrl, scanner)
		//	fmt.Println("waiting for download... ", mpdFullUrl)
		// wait for dowload finish
		// Block until we receive a notification from the worker on the channel.
		<-done

	}

	return mpdLocalFileName, hclient, adrCmdWriter, scanner, localTempFolder
}

func download_with_console(done chan bool, segmentUrl string, scanner *bufio.Scanner) float64 {
	// func download_with_console(segmentUrl string, scanner *bufio.Scanner) {

	var fileSize float64

	// fmt.Println("working..for url: ", segmentUrl)
	for scanner.Scan() {
		// fmt.Printf("console out | %s\n", scanner.Text())

		if strings.HasPrefix(scanner.Text(), "file_size") {

			s := strings.Split(scanner.Text(), ":")
			fileSize, _ = strconv.ParseFloat(s[1], 64)

		} else if strings.HasPrefix(scanner.Text(), "Request succeeded (200).") {
			break
		}
	}
	// fmt.Println(" done for url: ", segmentUrl)
	// Send a value to notify that we're done.
	done <- true
	// fmt.Println("fileSize", fileSize)
	return fileSize
}

/*
parse_arguments: parses and assings command line parameters
*/
func parse_arguments() {
	flag.StringVar(&protoClientPort, "port", "443", "Port number for proto-client ")
	flag.IntVar(&segmentLimitParameter, "n", 200, "The Segment number limit ")
	flag.BoolVar(&list, "l", false, "List all the representations")
	flag.BoolVar(&download, "d", false, "Keep the video files after playback")
	flag.BoolVar(&quic, "quic", false, "Enable quic with quic-go library")
	flag.BoolVar(&proto, "proto", false, "Enable proto-quic client console for downloading segments")
	flag.BoolVar(&h2, "h2", false, "Enable http2")
	flag.BoolVar(&h1, "h1", false, "Enable http1.1")
	flag.StringVar(&playbackType, "p", "basic", "Playback type (basic, sara, netflix, or all)")
	flag.StringVar(&mpdFullUrl, "m", "https://caddy.quic/BigBuckBunny_4s.mpd", "Url to the MPD File")
	flag.StringVar(&LOCAL_TEMP_DIR, "f", "/Users/chajiu/Desktop/GoDashPlayer/dashquic/_tmp/DOWNLOADED/", "Temp folder for downloading segments")
	flag.BoolVar(&java, "java", false, "Enable java-client console for downloading segments")
	flag.StringVar(&protoClientHostName, "host", "127.0.0.1", "Hostname for proto client")
	flag.IntVar(&repeatCount, "r", 10, "Repeat count for playing file ")
	flag.Parse()
}

/*
print_arguments: prints  command line parameters
*/
func print_arguments() {
	log.WithFields(log.Fields{
		"segmentLimitParameter": segmentLimitParameter,
		"list":                  list,
		"download":              download,
		"quic":                  quic,
		"http2":                 h2,
		"playbackType":          playbackType,
		"mpdFullUrl":            mpdFullUrl,
		"LOCAL_TEMP_DIR":        LOCAL_TEMP_DIR,
	}).Infoln("DASH_CLIENT: Command Line Parameters:")
}

/*
get_random_folder_path: returns random folder
*/
func get_random_folder_path() string {
	return LOCAL_TEMP_DIR + config.LOCAL_SEGMENT_FOLDER_PREFIX + strconv.Itoa(random(1000, 100000)) + "/"
}

/*
random: returns a random integer between given intervals
*/
func random(min, max int) int {
	rand.Seed(time.Now().Unix())
	return rand.Intn(max-min) + min
}

/*
GetNow: returns the current time in seconds.miliseconds(2 digits) format
*/
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

/***********************************************************
******************UNUSED CODES:*****************************
************************************************************/
type Int64arr []int64

func (s Int64arr) Len() int {
	return len(s)
}
func (s Int64arr) Less(i, j int) bool {
	return s[i] < s[j]
}
func (s Int64arr) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s Int64arr) Sort() {
	sort.Sort(s)
}
func download_segment_old(hclient *http.Client, segmentUrl string, dashFolder string) (float64, string) {

	/*https://stackoverflow.com/questions/22417283/save-an-image-from-url-to-file*/

	rsp, err := hclient.Get(segmentUrl)
	if err != nil {
		panic(err)
	}
	defer rsp.Body.Close()
	// utils.Infof("Got response for %s: %#v", segmentUrl, rsp)

	body := &bytes.Buffer{}
	_, err = io.Copy(body, rsp.Body)
	if err != nil {
		panic(err)
	}

	splitedBySlash := strings.SplitAfterN(segmentUrl, "/", strings.Count(segmentUrl, "/")+1)
	segmentLocalFileName := dashFolder + splitedBySlash[strings.Count(segmentUrl, "/")]

	// utils.Infof("writing segment file: %s", segmentLocalFileName)
	err = ioutil.WriteFile(segmentLocalFileName, body.Bytes(), 0644)

	if err != nil {
		panic(err)
	}

	return float64(body.Len()), segmentLocalFileName
}
func get_mpd_with_wg(mpdFullUrl string) string {

	hclient := &http.Client{
		Transport: &h2quic.RoundTripper{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}},
	}

	var urls [1]string
	urls[0] = mpdFullUrl

	var mpdLocalFileName string
	//	var lastAddr string

	var wg sync.WaitGroup
	wg.Add(len(urls))
	for _, addr := range urls {

		go func(addr string) {
			rsp, err := hclient.Get(addr)
			if err != nil {
				panic(err)
			}
			//	utils.Infof("Got response for %s: %#v", addr, rsp)

			body := &bytes.Buffer{}
			_, err = io.Copy(body, rsp.Body)
			if err != nil {
				panic(err)
			}

			splitedBySlash := strings.SplitAfterN(addr, "/", strings.Count(addr, "/")+1)
			mpdLocalFileName = "/home/sevket/go/src/bola/BolaClient/dash_quic/tmp/" + splitedBySlash[strings.Count(addr, "/")]

			// utils.Infof("writing mpd file: %s", mpdLocalFileName)

			// write whole the body
			err = ioutil.WriteFile(mpdLocalFileName, body.Bytes(), 0644)

			if err != nil {
				panic(err)
			}
			//		lastAddr = addr
			wg.Done()
		}(addr)
	}
	wg.Wait()

	return mpdLocalFileName
}

/**
	DashPlayback{} //Notice we didnt usee DashPlayback{}. we need a pointer to the DashPlayback
	The new(DashPlayback) was just a syntactic sugar for &DashPlayback{} and we  need a pointer to the DashPlayback
**/

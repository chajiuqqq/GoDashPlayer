package main

import (
	"github.com/sevketarisu/GoDashPlayer/config"
	//"bola/BolaClient/utils"
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	"strconv"
	"sync"
)

type DashPlayer struct {
	PlaybackStartTime float64
	PlaybackDuration  float64
	SegmentDuration   float64
	PlaybackTimer     StopWatch
	ActualStartTime   float64
	PlaybackState     string
	PlaybackStateLock sync.Mutex
	MaxBufferSize     float64
	BufferLength      float64
	BufferLengthLock  sync.Mutex
	InitialBuffer     float64
	Alpha             float64
	Beta              float64
	SegmentLimit      int
	Buffer            SegmentQueue
	BufferLock        sync.Mutex
	CurrentSegment    float64
}

type SegmentInfo map[string]string
type SegmentQueue []*SegmentInfo

func (dp *DashPlayer) start() {
	dp.setState("INITIAL_BUFFERING")

	log.Info("Starting the Player")
	go dp.start_player()
}

func (dp *DashPlayer) setState(state string) {
	if state == "INITIALIZED" || state == "INITIAL_BUFFERING" || state == "PLAY" ||
		state == "PAUSE" || state == "BUFFERING" || state == "STOP" || state == "END" {
		dp.PlaybackStateLock.Lock()

		log.WithFields(log.Fields{
			"from": dp.PlaybackState,
			"to":   state,
		}).Info("Changing Player State")

		dp.PlaybackState = state
		dp.PlaybackStateLock.Unlock()
	} else {
		err := errors.New("Unidentified Player State")
		if err != nil {
			panic(err)
		}
	}
}

func (dp *DashPlayer) writeToBuffer(segment SegmentInfo) {

	if dp.ActualStartTime > -1 {
		dp.ActualStartTime = GetNow()
	}

	dp.BufferLock.Lock()

	log.WithFields(log.Fields{
		"segment": segment["segment_number"],
		"time":    FloatToString(GetNow() - dp.ActualStartTime),
	}).Info("Writing segment ")

	dp.Buffer.Push(&segment)
	dp.BufferLock.Unlock()

	tmpBufferLength, err := strconv.ParseFloat(segment["playback_length"], 64)
	if err != nil {
		panic(err)
	}

	dp.BufferLengthLock.Lock()
	dp.BufferLength = dp.BufferLength + float64(int(tmpBufferLength))
	dp.BufferLengthLock.Unlock()
}

func (dp *DashPlayer) readFromBuffer() SegmentInfo {

	dp.BufferLock.Lock()
	nextSegment := *dp.Buffer.Pop()
	dp.BufferLock.Unlock()

	tmpBufferLength, err := strconv.ParseFloat((nextSegment)["playback_length"], 64)
	if err != nil {
		panic(err)
	}

	dp.BufferLengthLock.Lock()
	dp.BufferLength = dp.BufferLength - float64(int(tmpBufferLength))
	dp.BufferLengthLock.Unlock()

	return nextSegment
}

func (dp *DashPlayer) __init__(videoLength float64, segmentDuration float64) {

	log.Info("Initializing the Buffer")
	dp.PlaybackStartTime = -1 //None
	dp.PlaybackDuration = videoLength
	dp.SegmentDuration = segmentDuration
	// Timers to keep track of playback time and the actual time
	dp.PlaybackTimer = StopWatch{}
	dp.ActualStartTime = -1 //None
	//  # Playback State
	dp.PlaybackState = "INITIALIZED"
	dp.PlaybackStateLock = sync.Mutex{}

	if config.MAX_BUFFER_SIZE > -1 {
		dp.MaxBufferSize = config.MAX_BUFFER_SIZE
	} else {
		dp.MaxBufferSize = videoLength
	}
	// Duration of the current buffer
	dp.BufferLength = 0
	dp.BufferLengthLock = sync.Mutex{}
	//# Buffer Constants
	dp.InitialBuffer = 1
	dp.Alpha = config.ALPHA_BUFFER_COUNT
	dp.Beta = config.BETA_BUFFER_COUNT
	dp.SegmentLimit = -1 //None
	dp.Buffer = SegmentQueue{}
	dp.BufferLock = sync.Mutex{}
	dp.CurrentSegment = -1 //None

	log.WithFields(log.Fields{
		"PlaybackDuration": dp.PlaybackDuration,
		"SegmentDuration":  dp.SegmentDuration,
		"MaxBufferSize ":   dp.MaxBufferSize,
		"InitialBuffer":    dp.InitialBuffer,
		"Alpha":            dp.Alpha,
		"Beta":             dp.Beta,
	}).Info("Video Info")

}

func (dp *DashPlayer) start_player() { // PLAYER THREAD START

	var startTime float64 = GetNow()
	var initialWait float64 = 0.0
	var paused bool = false
	var buffering bool = false
	var interruptionStart float64 = -1.0 //None
	var noBreak bool = false

	log.WithFields(log.Fields{
		"PlaybackDuration": dp.PlaybackDuration,
	}).Info("Initialized player with video length")

	for {
		if dp.PlaybackState == "END" {
			dp.PlaybackTimer.pause()
			return
		}
		if dp.PlaybackState == "STOP" {
			dp.PlaybackTimer.pause()
			return
		}
		if dp.PlaybackState == "PAUSE" {
			if !(paused) {
				dp.PlaybackTimer.pause()
				paused = true
			}
			continue
		}

		if dp.PlaybackState == "BUFFERING" {
			if !(buffering) {
				dp.PlaybackTimer.pause()
				buffering = true
				interruptionStart = GetNow()
			} else {
				//# If the RE_BUFFERING_DURATION is greater than the remaing length of the video then do not wait
				remainingPlaybackTime := dp.PlaybackDuration - dp.PlaybackTimer.time()
				if dp.Buffer.Len() >= config.RE_BUFFERING_COUNT ||
					(config.RE_BUFFERING_COUNT*dp.SegmentDuration >= remainingPlaybackTime && dp.Buffer.Len() > 0) {

					buffering = false
					if interruptionStart > -1 {
						interruptionEnd := GetNow()
						interruption := interruptionEnd - interruptionStart
						interruptionStart = -1 //None
						fmt.Println("interruption seconds:", interruption)
					}
					dp.setState("PLAY")
				}
			}
		} //END IF BUFFERING

		if dp.PlaybackState == "INITIAL_BUFFERING" {
			if dp.Buffer.Len() < config.INITIAL_BUFFERING_COUNT {
				initialWait = GetNow() - startTime
				continue
			} else {

				log.WithFields(log.Fields{
					"initialWait": initialWait,
				}).Info("Initial Waiting Time")

				dp.setState("PLAY")
			}
		} //END IF INITIAL_BUFFERING

		if dp.PlaybackState == "PLAY" {

			// Check of the buffer has any segments
			if dp.PlaybackTimer.time() == dp.PlaybackDuration {
				dp.setState("END")
			}
			if dp.Buffer.Len() == 0 {
				dp.PlaybackTimer.pause()
				dp.setState("BUFFERING")
				continue
			}

			// Read one the segment from the buffer
			playSegment := dp.readFromBuffer()

			log.WithFields(log.Fields{
				"Segment":       playSegment["segment_number"],
				"Playback Time": dp.PlaybackTimer.time(),
				"Bitrate":       playSegment["bitrate"],
			}).Info("Reading the segment from the buffer at playtime")

			// Calculate time playback when the segment finishes
			playbackLength, err := strconv.ParseFloat(playSegment["playback_length"], 64)
			if err != nil {
				panic(err)
			}

			future := dp.PlaybackTimer.time() + playbackLength

			// # Start the playback timer
			dp.PlaybackTimer.start()

			noBreak = true
			for dp.PlaybackTimer.time() < future {
				//  If playback hasn't started yet, set the playback_start_time
				if dp.PlaybackStartTime != -1 {
					dp.PlaybackStartTime = GetNow()
				}
				// Duration for which the video was played in seconds (integer)
				if dp.PlaybackTimer.time() >= dp.PlaybackDuration {

					log.WithFields(log.Fields{
						"PlaybackDuration": dp.PlaybackDuration,
					}).Info("Completed the video playback seconds")

					dp.PlaybackTimer.pause()
					dp.setState("END")
					log.Info("PLAYER ENDED")
					return //returns from this function (start_player)
				}
				noBreak = false
			}

			if noBreak { //while else
				playbackLength, err := strconv.ParseFloat(playSegment["playback_length"], 64)
				if err != nil {
					panic(err)
				}
				dp.BufferLengthLock.Lock()
				dp.BufferLength = dp.BufferLength - playbackLength
				dp.BufferLengthLock.Unlock()
			}

			if dp.SegmentLimit != -1 { //None
				segmentNumber, _ := strconv.Atoi(playSegment["segment_number"])
				if segmentNumber >= dp.SegmentLimit {

					log.WithFields(log.Fields{
						"SegmentLimit":  dp.SegmentLimit,
						"segmentNumber": segmentNumber,
					}).Info("Segment limit reached, Player will stop ")

					dp.setState("STOP")
				}
			}

		} //END IF STATE=PLAY
	} //FOR INFITITE LOOP
} // END start_player

func (q *SegmentQueue) Push(n *SegmentInfo) {
	*q = append(*q, n)
}

func (q *SegmentQueue) Pop() (n *SegmentInfo) {
	n = (*q)[0]
	*q = (*q)[1:]
	return
}

func (q *SegmentQueue) Len() int {
	return len(*q)
}

package main

import (
	"strconv"
	//"sync"
	"fmt"
	"time"
)

type StopWatch2 struct {
	StartTime   float64
	ElapsedTime float64
	Running     bool
	//StateLock   sync.Mutex
}

func (sw *StopWatch2) start() {
	if !sw.Running {
		sw.StartTime = GetNow() - sw.ElapsedTime
		//sw.StateLock.Lock()
		sw.Running = true
		//	sw.StateLock.Unlock()
	}
}

func (sw *StopWatch2) pause() {
	if sw.Running {
		sw.ElapsedTime = GetNow() - sw.StartTime
		//sw.StateLock.Lock()
		sw.Running = false
		//sw.StateLock.Unlock()
	}
}

func (sw *StopWatch2) time() float64 {
	if sw.Running {
		sw.ElapsedTime = GetNow() - sw.StartTime
	}
	return float64(int(sw.ElapsedTime))
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
	fmt.Println("GetNow ici str: ", str)
	f, err := strconv.ParseFloat(str, 64)
	if err != nil {
		panic(err)
	}
	return f
}

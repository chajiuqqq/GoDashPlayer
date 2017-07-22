package main

type StopWatch struct {
	StartTime   float64
	ElapsedTime float64
	Running     bool
}

func (sw *StopWatch) start() {
	if !sw.Running {
		sw.StartTime = GetNow() - sw.ElapsedTime
		sw.Running = true
	}
}

func (sw *StopWatch) pause() {
	if sw.Running {
		sw.ElapsedTime = GetNow() - sw.StartTime
		sw.Running = false
	}
}

func (sw *StopWatch) time() float64 {
	if sw.Running {
		sw.ElapsedTime = GetNow() - sw.StartTime
	}
	return float64(int(sw.ElapsedTime))
}

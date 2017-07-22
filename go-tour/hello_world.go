package main

import (
	//"fmt"
	log "github.com/sirupsen/logrus"
	//"time"
)

func main() {

	log.Infoln("deneme")
	log.Info("denem2")

	var a int = 5
	var b string = "test"
	//rand.Int
	log.WithFields(log.Fields{
		"animail": b,
		"number":  a,
	}).Info("A walrus appears")

	/*	sw2 := new(StopWatch2)
		//fmt.Println("getNow: %s", GetNow())
		fmt.Println("time:", sw2.time())
		sw2.start()
		fmt.Println("goint to sleep")
		time.Sleep(5 * time.Second)
		fmt.Println("woke up")
		//fmt.Println("getNow: %s", GetNow())
		fmt.Println("time:", sw2.time())

		sw2.pause()
		fmt.Println("paused")
		fmt.Println("goint to sleep")
		time.Sleep(5 * time.Second)
		fmt.Println("woke up")
		//fmt.Println("getNow: %s", GetNow())
		fmt.Println("time:", sw2.time())
		fmt.Println("starting again")
		sw2.start()
		fmt.Println("goint to sleep")
		time.Sleep(5 * time.Second)
		fmt.Println("woke up")
		//fmt.Println("getNow: %s", GetNow())
		fmt.Println("time:", sw2.time())

		fmt.Println("***********************")

		sw := new(Stopwatch)
		sw.Log("helloworld")
		sw.Start(0)
		fmt.Println("goint to sleep")
		time.Sleep(5 * time.Second)
		fmt.Println("woke up")
		sw.Log("helloworld")
		sw.Stop()
		fmt.Println("stopped")
		fmt.Println("goint to sleep")
		time.Sleep(5 * time.Second)
		fmt.Println("woke up")
		fmt.Println("starting again")
		sw.Start(0)
		fmt.Println("goint to sleep")
		time.Sleep(5 * time.Second)
		fmt.Println("woke up")
		sw.Log("helloworld")*/

}

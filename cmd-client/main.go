package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"time"
)

func main() {

	cmdName := "/home/sevket/proto-quic/src/out/Default/quic_client"
	cmdArgs := []string{"--host=ec2-13-56-87-190.us-west-1.compute.amazonaws.com",
		"--port=53",
		"--v=1",
		"--disable-certificate-verification",
		"--folder=/home/sevket/tmp/ec2_down/",
		"x"}
	segmentUrl := []string{"https://ec2-13-56-87-190.us-west-1.compute.amazonaws.com/media/BigBuckBunny/4sec/bunny_2409742bps/BigBuckBunny_4s1.m4s",
		"https://ec2-13-56-87-190.us-west-1.compute.amazonaws.com/media/BigBuckBunny/4sec/bunny_2409742bps/BigBuckBunny_4s2.m4s",
		"https://ec2-13-56-87-190.us-west-1.compute.amazonaws.com/media/BigBuckBunny/4sec/bunny_2409742bps/BigBuckBunny_4s3.m4s"}

	//--host=ec2-13-56-87-190.us-west-1.compute.amazonaws.com --port=53 --v=0 --disable-certificate-verification x --folder=/home/sevket/tmp/ec2_down/

	cmd := exec.Command(cmdName, cmdArgs...)
	cmdReader, err1 := cmd.StdoutPipe()
	if err1 != nil {
		fmt.Fprintln(os.Stderr, "Error creating StdoutPipe for Cmd", err1)
		os.Exit(1)
	}

	cmdWriter, err2 := cmd.StdinPipe()
	if err2 != nil {
		fmt.Fprintln(os.Stderr, "Error creating StdoutPipe for Cmd", err2)
		os.Exit(1)
	}
	defer io.WriteString(cmdWriter, "exit\n")

	scanner := bufio.NewScanner(cmdReader)
	done := make(chan bool, 1)

	go download_with_console(done, scanner)

	/*go func(segmnentUrl string, done chan bool) {
		for scanner.Scan() {
			fmt.Printf("docker build out | %s\n", scanner.Text())
		}
	}()*/

	err1 = cmd.Start()
	if err1 != nil {
		fmt.Fprintln(os.Stderr, "Error starting Cmd", err1)
		os.Exit(1)
	}

	for i := 0; i < len(segmentUrl); i++ {
		io.WriteString(cmdWriter, segmentUrl[i]+"\n")
	}

	io.WriteString(cmdWriter, "exit\n")
	err1 = cmd.Wait()
	if err1 != nil {
		fmt.Fprintln(os.Stderr, "Error waiting for Cmd", err1)
		os.Exit(1)
	}
}

func download_with_console(done chan bool, scanner *bufio.Scanner) {

	for scanner.Scan() {
		fmt.Printf("console out | %s\n", scanner.Text())
	}
	done <- true

}

// We can use channels to synchronize execution
// across goroutines. Here's an example of using a
// blocking receive to wait for a goroutine to finish.

// This is the function we'll run in a goroutine. The
// `done` channel will be used to notify another
// goroutine that this function's work is done.
func worker(done chan bool) {
	fmt.Print("working...")
	time.Sleep(time.Second)
	fmt.Println("done")

	// Send a value to notify that we're done.
	done <- true
}

func main2() {

	// Start a worker goroutine, giving it the channel to
	// notify on.
	done := make(chan bool, 1)
	go worker(done)

	// Block until we receive a notification from the
	// worker on the channel.
	<-done
}

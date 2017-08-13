// We can use channels to synchronize execution
// across goroutines. Here's an example of using a
// blocking receive to wait for a goroutine to finish.

package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	//	"time"
)

func main() { //Player Client

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

	err1 = cmd.Start()
	if err1 != nil {
		fmt.Fprintln(os.Stderr, "Error starting Cmd", err1)
		os.Exit(1)
	}

	// Start a worker goroutine, giving it the channel to
	// notify on.

	for i := 0; i < len(segmentUrl); i++ { //150 total segment size
		done := make(chan bool)
		//done <- false
		io.WriteString(cmdWriter, segmentUrl[i]+"\n")
		go download_with_console(done, segmentUrl[i], scanner)
		fmt.Println("waiting for download... ", segmentUrl[i])
		//wait for dowload finish
		// Block until we receive a notification from the
		// worker on the channel.
		<-done
	}
	io.WriteString(cmdWriter, "exit\n")

}

// This is the function we'll run in a goroutine. The
// `done` channel will be used to notify another
// goroutine that this function's work is done.
func download_with_console(done chan bool, segmentUrl string, scanner *bufio.Scanner) {

	fmt.Println("working..for url: ", segmentUrl)

	for scanner.Scan() {
		//fmt.Printf("console out | %s\n", scanner.Text())

		if strings.HasSuffix(scanner.Text(), "Request succeeded (200).") {
			break
		}

	}

	fmt.Println(" done for url: ", segmentUrl)

	// Send a value to notify that we're done.
	done <- true
}

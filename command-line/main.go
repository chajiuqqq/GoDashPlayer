package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	//	"time"
	//	"bytes"
	//"strings"
)

func main() {

	consoleClient := get_console_client("/home/sevket/proto-quic/src/out/Default/quic_client", "https://www.example.org") //Just for testing, replace with your subProcess
	get_file_with_console_client(consoleClient, "https://www.example.org/BigBuckBunny_4s.mpd")
}

func get_console_client(commandStr string, domainUrl string) *exec.Cmd {
	var consoleClient *exec.Cmd
	consoleClient = exec.Command(commandStr, domainUrl)
	return consoleClient
}
func get_file_with_console_client(consoleClient *exec.Cmd, segmentUrl string) {
	//var consoleReader int = 0

	//stdout, err := consoleClient.StdoutPipe()
	stdin, err := consoleClient.StdinPipe()
	if err != nil {
		log.Println(err) //replace with logger, or anything you want
	}
	defer stdin.Close() // the doc says subProcess.Wait will close it, but I'm not sure, so I kept this line

	consoleClient.Stdout = os.Stdout
	//consoleClient.Stdin = os.Stdin

	fmt.Println("START")                         //for debug
	if err = consoleClient.Start(); err != nil { //Use start, not run
		log.Println("An error occured: ", err) //replace with logger, or anything you want
	}

	//os.Stdin.WriteString("test")
	//ConsoleDeneme.StdinPipe()
	io.WriteString(stdin, segmentUrl+"\n")
	//consoleClient.Wait()
	fmt.Println("END") //for debug

}

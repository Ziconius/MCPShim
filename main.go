package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"log/slog"
	"os"
	"os/exec"
	"sync"
)

type Config struct {
	ProxyAddr string // Need to check the type for IP/Port.
	ProxyPort string
}

func NewConfig() Config {
	// TODO - Not returning an error as this should just panic if a failed config exists.
	return Config{}
}

var CFG Config

func init() {
	// Create config object
	CFG = NewConfig()

}

func main() {
	// Send logging to a sensible location.
	f, err := os.OpenFile("/tmp/mcp_shim_log.txt", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}

	opts := &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}
	// logger := slog.New(slog.NewJSONHandler(os.Stdout, opts))
	logger := slog.New(slog.NewTextHandler(f, opts))
	slog.SetDefault(logger)

	parentIn := make(chan string)  // STDOUT from parent/Our STDIN
	parentOut := make(chan string) // STDOUT to parent

	childIn := make(chan string)  // STDIN to child
	childOut := make(chan string) // STDOUT from child

	var wg sync.WaitGroup
	wg.Add(1)

	go ParentReciever(parentIn)
	go ParentSender(parentOut)

	// TODO: pull CLI args and recreate the initial MCP args.
	args := GetMCPServerArgs()
	go ChildProcess(args, childOut, childIn)

	// This is where we'll build out proxy tooling.
	/*
		Use config to direct web proxy for flexible inspections
		JSON-RPC 2.0 messages contain an ID field for requests which require a response, for those without an
			ID field this is a notification and the server *SHOULD NOT* return a message. This measn that we
			do not need to implement a bi-directional proxy.

		This means we may to have less basic shim format.
	*/
	go ParentShim(parentIn, childOut)
	go ChildShim(childIn, parentOut)
	wg.Wait()

}

func GetMCPServerArgs() []string {
	argsWithoutProg := os.Args[1:]
	slog.Debug("MCP Server + args", "cmd", argsWithoutProg)
	return argsWithoutProg
}

func ParentReciever(PI chan string) {
	slog.Info("Starting parent reciever.")
	reader := bufio.NewReader(os.Stdin)
	for {
		text, err := reader.ReadString('\n')
		if err != nil {
			if err.Error() != "EOF" {
				slog.Error("Failed to read string", "error", err)
			}
		} else {
			PI <- text
			slog.Debug("Message recieved from parent", "msg", text)
		}
	}
}

func ParentSender(PO chan string) {
	slog.Info("Starting Parent Sender")
	for {
		msg := <-PO
		slog.Debug("Sending to parent via our stdout", "msg", msg)
		fmt.Printf("%v", msg)
	}
}

func ChildProcess(args []string, CO, CI chan string) {

	/*
		This function will take the args and start the process.

		This will need the ChildOut (Messages to be send OUT to the subprocess)
		and it will need to take child IN (messages recieved from the subprocess)

	*/

	// cmd := exec.Command("npx", "-y", "@modelcontextprotocol/server-filesystem", "/tmp/mcp")
	// TODO: implement & test args.
	cmd := exec.Command("npx", "@playwright/mcp@latest")
	stdin, err := cmd.StdinPipe()
	if err != nil {
		// log.Fatal(err)
		slog.Error("StdinPipe failed", "error", err)
	}
	defer stdin.Close()

	out, err := cmd.StdoutPipe()
	if err != nil {
		slog.Error("StdoutPipe failed", "error", err)
	}

	slog.Debug("About to enter sub process loop.")
	err = cmd.Start()
	if err != nil {
		slog.Error("failed to exec command", "error", err)
	}

	var wg sync.WaitGroup
	wg.Add(1)
	// Child sender (or Parent Out)
	go ChildSender(CO, stdin)
	go ChildReciever(CI, out)
	wg.Wait()
}

func ChildReciever(CI chan string, stdout io.ReadCloser) {
	slog.Debug("Starting child reciever")
	// FUCK KNOWS

	reader := bufio.NewReader(stdout)
	reader.Size()
	for {

		line, err := reader.ReadString('\n')
		// slog.Debug("we've got a line", "msg", line)
		if len(line) != 0 {
			slog.Debug("Out line", "msg", line)
			CI <- line
		}
		for err == nil {
			// fmt.Println(line)
			// slog.Error("CR: err== nil")
			line, err = reader.ReadString('\n')
			if len(line) != 0 {
				slog.Debug("Out line", "msg", line)
				CI <- line
			}
		}

	}
}

// Takens PO and send them to child process
func ChildSender(CO chan string, stdin io.WriteCloser) {
	slog.Debug("Starting child sender.")
	for {
		tString := <-CO
		slog.Debug("tString", "msg", tString)
		_, err := stdin.Write([]byte(tString))
		// _, err := io.WriteString(stdin, tString)
		if err != nil {
			slog.Error("Failed to with message to child process", "error", err)
		}
	}
}

func ParentShim(PI, CO chan string) {
	for {
		v := <-PI
		slog.Debug("We're got a parent shim", "shim", v)
		CO <- v
	}
}

func ChildShim(CI, PO chan string) {
	for {
		v := <-CI
		slog.Debug("We're got a child shim", "shim", v)
		PO <- v
	}
}

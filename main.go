package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"log/slog"
	"os"
	"os/exec"
	"sync"
)

var CFG Config

func init() {
	CFG = NewConfig()
}

func main() {
	f, err := os.OpenFile(CFG.LogFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal("error creating or opening log file file", "error", err)
	}

	opts := &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}
	// logger := slog.New(slog.NewJSONHandler(os.Stdout, opts))
	logger := slog.New(slog.NewTextHandler(f, opts))
	slog.SetDefault(logger)

	// Output current config for debugging
	slog.Debug("Current config", "Logfile", CFG.LogFile, "Intercept", CFG.Intercept)

	// Parent channels with with the MCP client stdio interface
	parentIn := make(chan string) 
	parentOut := make(chan string)

	// Child channels interface with the MCP Server launched by the shim.
	childIn := make(chan string)
	childOut := make(chan string)

	var wg sync.WaitGroup
	wg.Add(1)

	go ParentReciever(parentIn)
	go ParentSender(parentOut)
	args, err := GetMCPServerArgs()
	if err != nil {
		slog.Error("No MCP server defined", "error", err)
		panic("error no MCP server configured via args.")
	}
	go ChildProcess(args, childOut, childIn)

	// This is where we'll build out proxy tooling.
	/*
		Use config to direct web proxy for flexible inspections
		JSON-RPC 2.0 messages contain an ID field for requests which require a response, for those without an
			ID field this is a notification and the server *SHOULD NOT* return a message. This measn that we
			do not need to implement a bi-directional proxy.

		This means we may to have less basic shim format.
	*/
	if CFG.Intercept.Enabled {
		go HTTPParentShim(CFG.Intercept.Address, parentIn, childOut)
		go HTTPChildShim(CFG.Intercept.Address, childIn, parentOut)
	} else {
		go ParentShim(parentIn, childOut)
		go ChildShim(childIn, parentOut)
	}

	wg.Wait()

}

func GetMCPServerArgs() ([]string, error) {
	argsWithoutProg := os.Args[1:]
	slog.Debug("MCP Server + args", "cmd", argsWithoutProg)
	if len(argsWithoutProg) == 0 {
		return argsWithoutProg, errors.New("no args provided to the process")
	}
	return argsWithoutProg, nil
}

func ParentReciever(PI chan string) {
	slog.Debug("Starting parent reciever.")
	reader := bufio.NewReader(os.Stdin)
	for {
		text, err := reader.ReadString('\n')
		if err != nil {
			if err.Error() != "EOF" {
				slog.Error("Failed to read string", "error", err)
			}
		} else {
			PI <- text
			slog.Debug("Message recieved from parent", "request", text)
		}
	}
}

func ParentSender(PO chan string) {
	slog.Info("Starting Parent Sender")
	for {
		data := <-PO
		slog.Debug("Sending to parent via our stdout", "response", data)
		fmt.Printf("%v", data)
	}
}

func ChildProcess(args []string, CO, CI chan string) {

	/*
		This function will take the args and start the process.

		This will need the ChildOut (Messages to be send OUT to the subprocess)
		and it will need to take child IN (messages recieved from the subprocess)
	*/

	// cmd := exec.Command("npx", "-y", "@modelcontextprotocol/server-filesystem", "/tmp/mcp")
	// cmd := exec.Command("npx", "@playwright/mcp@latest")

	cmd := &exec.Cmd{}
	if len(args) == 1 {
		cmd = exec.Command(args[0])
	} else {
		slog.Debug("executed command", "command", args[0], "args", args[1:])
		cmd = exec.Command(args[0], args[1:]...)
	}
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

	err = cmd.Start()
	if err != nil {
		slog.Error("failed to exec command", "error", err)
	}

	var wg sync.WaitGroup
	wg.Add(1)
	go ChildSender(CO, stdin)
	go ChildReciever(CI, out)
	wg.Wait()
}

func ChildReciever(CI chan string, stdout io.ReadCloser) {
	slog.Debug("Starting child reciever")
	reader := bufio.NewReader(stdout)
	reader.Size()
	for {

		line, err := reader.ReadString('\n')
		if len(line) != 0 {
			slog.Debug("Recieved message from MCP Server", "response", line)
			CI <- line
		}
		for err == nil {
			line, err = reader.ReadString('\n')
			if len(line) != 0 {
				slog.Debug("Recieved message from MCP Server", "response", line)
				CI <- line
			}
		}

	}
}

// Takens PI and send them to child process
func ChildSender(CO chan string, stdin io.WriteCloser) {
	slog.Debug("Starting child sender")
	for {
		req := <-CO
		slog.Debug("Sending request to MCP Server", "request", req)
		_, err := stdin.Write([]byte(req))
		if err != nil {
			slog.Error("Failed to with message to child process", "error", err)
		}
	}
}

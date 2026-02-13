package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"reflect"
	"strconv"
)

// Legacy shim
// func _HTTPParentShim(proxyAddr string, PI, CO chan string, responses *map[string]string) {
// 	serverPort := "15001"
// 	go startServer(serverPort, CO, responses)
// 	slog.Debug("HTTP Parent shim enabled")
// 	for {
// 		v := <-PI
// 		slog.Info("Message to MCP Server", "request", v)

// 		// Post Request

// 		if IsNotification(v) {
// 			SendNotification(serverPort, proxyAddr, v)
// 		} else {
// 			SendRequest(serverPort, proxyAddr, v, responses)
// 		}
// 	}
// }

func HTTPParentShim(proxyAddr string, PI, CO chan string, out, in map[string]string) {
	serverPort := "15001"
	serv := NewHttpIntercept(serverPort, proxyAddr, CO, out, in)

	go serv.StartServer()
	slog.Debug("HTTP Parent shim enabled")
	for {
		v := <-PI
		slog.Info("Message to MCP Server", "request", v)

		// New implementations
		go serv.HandleNewMessage(v)
	}
}

func HTTPChildShim(proxyAddr string, CI, PO chan string, out, in map[string]string) {
	serverPort := "15002"
	serv := NewHttpIntercept(serverPort, proxyAddr, PO, out, in)

	go serv.StartServer()
	slog.Debug("HTTP Child shim enabled")
	for {
		v := <-CI
		slog.Info("Response from MCP server", "response", v)
		// PO <- v

		// New implementations
		go serv.HandleNewMessage(v)
	}
}

// Log only mode.
// These are the same function but the logging is different - should be squashed.
func ParentShim(PI, CO chan string) {
	for {
		v := <-PI
		slog.Info("Message to MCP Server", "request", v)
		CO <- v
	}
}

func ChildShim(CI, PO chan string) {
	for {
		v := <-CI
		slog.Info("Response from MCP server", "response", v)
		PO <- v
	}
}

// Initial parsing function to determine if the message is a JSON-RPC 2.0 notification.
func IsNotification(raw string) bool {
	// TODO: Improve the checking as ID can be either a string or int. JSON RPC requests the value is not a null, or decimal so we can ignore those edge
	//	cases for now.
	var msg map[string]any
	json.Unmarshal([]byte(raw), &msg)
	_, ok := msg["id"]
	if !ok {
		return true
	}
	return false
}

func IsResponse(raw string) bool {
	var msg map[string]any
	json.Unmarshal([]byte(raw), &msg)
	_, ok := msg["result"]
	if ok {
		return true
	}
	_, ok = msg["error"]
	if ok {
		return true
	}
	return false
}

func ExtractID(raw string) (string, error) {
	var msg map[string]any
	json.Unmarshal([]byte(raw), &msg)
	// _, ok := msg["id"]
	// if !ok {
	// 	return "", errors.New("no ID found in message. One was expected.")
	// }

	// fmt.Printf("V: %#v\n", msg)

	// id could be int or string
	v, ok := msg["id"].(string)
	fmt.Printf("Reflect: %v\n", reflect.TypeOf(msg["id"]))

	if reflect.TypeOf(msg["id"]) == reflect.TypeOf("") {
		fmt.Printf("We have a string")
	}
	if reflect.TypeOf(msg["id"]).Kind() == reflect.Float64 {
		fmt.Printf("We have a string")
	}		
	
	if !ok {
		v, ok := msg["id"].(int)
		fmt.Printf("Int: %#v: OK: %v\n", v, ok)
		if !ok {
			return "", errors.New("ID was not int or string. Error in JSON-RPC message")
		}
		return strconv.Itoa(v), nil
	}

	return v, nil

}

// func NewJSONRPCMessage(raw string) JSONRPCMessage {
// 	n := JSONRPCMessage{
// 		Message: raw,
// 	}
// 	n.parse()
// 	return n
// }

// // TODO: Cleanup
// type JSONRPCMessage struct {
// 	Id           string
// 	Message      string
// 	Notification bool
// 	Response     bool
// }

// func (j *JSONRPCMessage) parse() {
// 	// TODO: Clean up type asserions
// 	var msg map[string]any
// 	json.Unmarshal([]byte(j.Message), &msg)
// 	s, ok := msg["id"]
// 	if !ok {
// 		j.Notification = true
// 		// No ID to parse so we return.
// 		return
// 	}
// 	if a, ok := s.(string); ok {
// 		j.Id = a
// 	}
// 	if b, ok := s.(int); ok {
// 		j.Id = strconv.Itoa(b)
// 	}
// 	j.Response = IsResponse(msg)
// 	j.Notification = false
// }

// func IsResponse(msg map[string]any) bool {
// 	_, ok := msg["result"]
// 	if !ok {
// 		return true
// 	}
// 	_, ok = msg["error"]
// 	if !ok {
// 		return true
// 	}
// 	return false
// }

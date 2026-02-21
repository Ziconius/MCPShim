package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"reflect"
	"strconv"
)

func HTTPParentShim(proxyAddr string, PI, CO, PO chan string, out, in map[string]string) {
	serverPort := "15001"
	serv := NewHttpIntercept(serverPort, proxyAddr, CO, PO, out, in)

	go serv.StartServer()
	slog.Debug("HTTP Parent shim enabled")
	for {
		v := <-PI
		slog.Info("Message to MCP Server", "request", v)

		// New implementations
		go serv.HandleNewMessage(v)
	}
}

func HTTPChildShim(proxyAddr string, CI, PO, CO chan string, out, in map[string]string) {
	serverPort := "15002"
	serv := NewHttpIntercept(serverPort, proxyAddr, PO, CO, out, in)

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

// TODO: Remove slow Sprintf usage.
func ExtractID(raw string) (string, error) {
	var msg map[string]any
	json.Unmarshal([]byte(raw), &msg)
	id, ok := msg["id"]
	if !ok {
		return "", errors.New("no ID found in message. One was expected.")
	}

	switch reflect.TypeOf(msg["id"]).Kind() {
	case reflect.String:
		return id.(string), nil
	case reflect.Float64:
		return fmt.Sprintf("%v", id.(float64)), nil
	case reflect.Int:
		return strconv.Itoa(id.(int)), nil
	default:
		err := fmt.Sprintf("unmanaged type used in ID field: %v\n", reflect.TypeOf(msg["id"]).Kind())
		return "", errors.New(err)
	}
}

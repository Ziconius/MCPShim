package main

import "log/slog"

func HTTPParentShim(proxyAddr string, PI, CO chan string) {
	go startServer("15001", CO)
	slog.Debug("HTTP Parent shim enabled")
	for {
		v := <-PI
		slog.Info("Message to MCP Server", "request", v)

		// Post Request

		SendRequest("15001", proxyAddr, v)
	}
}

func HTTPChildShim(proxyAddr string, CI, PO chan string) {
	go startServer("15002", PO)
	slog.Debug("HTTP Child shim enabled")
	for {
		v := <-CI
		slog.Info("Response from MCP server", "response", v)
		// PO <- v
		SendRequest("15002", proxyAddr, v)
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

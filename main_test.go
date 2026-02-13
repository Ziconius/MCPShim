package main

import (
	"fmt"
	"testing"
)

func TestJSONRPCParse(t *testing.T) {
	raw := `{"method":"initialize","params":{"protocolVersion":"2025-11-25","capabilities":{},"clientInfo":{"name":"opencode","version":"1.1.48"}},"jsonrpc":"2.0","id":0}`
	if IsNotification(raw) {
		fmt.Println("PASS")
	}

	raw2 := `{"method":"initialize","params":{"protocolVersion":"2025-11-25","capabilities":{},"clientInfo":{"name":"opencode","version":"1.1.48"}},"jsonrpc":"2.0","id":"0"}`
	if IsNotification(raw2) {
		fmt.Println("PASS")
	}
}

func TestIsReponse(t *testing.T) {
	raw := `{"jsonrpc":"2.0","method":"notifications/cancelled","params":{"requestId":0,"reason":"McpError: MCP error -32001: Request timed out"}}`
	raw2 := `{"method":"initialize","params":{"protocolVersion":"2025-11-25","capabilities":{},"clientInfo":{"name":"opencode","version":"1.1.48"}},"jsonrpc":"2.0","id":0}`

	if IsResponse(raw) {
		fmt.Println("PASS")
	}
	if IsResponse(raw2) {
		fmt.Println("PASS")
	}
}

func TestExtractID(t *testing.T) {
	tv := `{"result":{"protocolVersion":"2025-11-25","capabilities":{"tools":{}},"serverInfo":{"name":"Playwright","version":"0.0.64"}},"jsonrpc":"2.0","id":0}`
	o, err := ExtractID(tv)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		t.Fail()
	}
	if o != "0" {
		t.Fail()
	}

}

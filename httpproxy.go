package main

import (
	"bytes"
	"io"
	"log/slog"
	"net/http"
	"net/url"
)

/*

Dev work - looking to create two servers which will be created for each direction.
	We will need to merge these for request/response manipulation in tools liek ZAP and burp.
*/

// This is run as a go routine, and will recievbe proxied messages
func startServer(port string, onward chan string) {

	httpintercept := HTTPIntercept{}
	httpintercept.Target = onward

	mux := http.NewServeMux()
	mux.HandleFunc("/", httpintercept.ServeHTTP)
	err := http.ListenAndServe("0.0.0.0:"+port, mux)
	if err != nil {
		panic("server crash")
	}
}

type HTTPIntercept struct {
	Target chan string
}

func (a *HTTPIntercept) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	data, err := io.ReadAll(r.Body)

	if err != nil {
		slog.Error("Failed to read body", "len", len(data), "error", err)
		w.WriteHeader(500)
		return
	}

	a.Target <- string(data)
	w.WriteHeader(200)
}

func SendRequest(shimPort, proxyAddr, message string) {
	proxyURL, _ := url.Parse(proxyAddr)
	proxy := http.ProxyURL(proxyURL)
	transport := &http.Transport{Proxy: proxy}
	client := &http.Client{Transport: transport}
	req, _ := http.NewRequest("POST", "http://127.0.0.1:"+shimPort, bytes.NewReader([]byte(message)))
	_, err := client.Do(req)
	if err != nil {
		slog.Error("failed to send request", "error", err)
	}
}

package main

import (
	"bytes"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"time"
)

func NewHttpIntercept(srvport, proxyAddr string, CO, PO chan string, outbound, inbound map[string]string) HTTPIntercept {
	return HTTPIntercept{
		ServerPort:       srvport,
		ProxyAddr:        proxyAddr,
		Target:           CO,
		Origin: PO,
		OutboundResponse: outbound,
		InboundResponse:  inbound,
	}
}

type HTTPIntercept struct {
	ServerPort string
	ProxyAddr  string
	Target     chan string
	Origin     chan string
	// Old
	// Response *map[string]string
	// New
	OutboundResponse map[string]string
	InboundResponse  map[string]string
}

func (h *HTTPIntercept) StartServer() {
	mux := http.NewServeMux()
	mux.HandleFunc("/request", h.HttpRequest)
	mux.HandleFunc("/notification", h.HtttpNotification)
	err := http.ListenAndServe("0.0.0.0:"+h.ServerPort, mux)
	if err != nil {
		panic("server crash")
	}

}

// JSON-RPC Request handler - Maybe rename
func (h *HTTPIntercept) HttpRequest(w http.ResponseWriter, r *http.Request) {
	data, err := io.ReadAll(r.Body)

	if err != nil {
		slog.Error("Failed to read body", "len", len(data), "error", err)
		w.WriteHeader(500)
		return
	}

	// TODO:
	// Loop inbound messages until we hit.
	// extract ID

	// Loop until we hit the response from the inbound.

	slog.Debug("recieved request message")
	h.Target <- string(data)

	id, err := ExtractID(string(data))
	if err != nil {
		w.WriteHeader(500)
		return
	}
	for {
		resp, ok := h.InboundResponse[id]
		if ok {
			// add to body
			w.Write([]byte(resp))
			// w.WriteHeader(200)
			time.Sleep(100 * time.Millisecond)
			slog.Debug("Returning Inbound response.", "msg", resp)

			// Remove the item from the list.
			delete(h.InboundResponse, id)

			return
		}

	}
}

// JSON-RPC Notification handler - Maybe rename
func (h *HTTPIntercept) HtttpNotification(w http.ResponseWriter, r *http.Request) {

	data, err := io.ReadAll(r.Body)

	if err != nil {
		slog.Error("Failed to read body", "len", len(data), "error", err)
		w.WriteHeader(500)
		return
	}

	slog.Debug("recieved notification message", "msg", string(data))
	h.Target <- string(data)
	w.WriteHeader(200)

}

// Send JSON-RPC Notification request, called from the HandleNewMessage
func (h *HTTPIntercept) sendNotification(raw string) {
	proxyURL, _ := url.Parse(h.ProxyAddr)
	proxy := http.ProxyURL(proxyURL)
	transport := &http.Transport{Proxy: proxy}
	client := &http.Client{Transport: transport}
	req, _ := http.NewRequest("POST", "http://127.0.0.1:"+h.ServerPort+"/notification", bytes.NewReader([]byte(raw)))
	_, err := client.Do(req)
	if err != nil {
		slog.Error("failed to send request", "error", err)
	}
}

// TODO: Recieve message
func (h *HTTPIntercept) sendRequest(raw string) {
	proxyURL, _ := url.Parse(h.ProxyAddr)
	proxy := http.ProxyURL(proxyURL)
	transport := &http.Transport{Proxy: proxy}
	client := &http.Client{Transport: transport}
	req, _ := http.NewRequest("POST", "http://127.0.0.1:"+h.ServerPort+"/request", bytes.NewReader([]byte(raw)))
	resp, err := client.Do(req)
	if err != nil {
		slog.Error("failed to send request", "error", err)
	}
	slog.Debug("sendRequest completed.")
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		slog.Error("Failed to read the request response body", "error", err)
		return
	}
	h.Origin <- string(body)

}

// Send message
func (h *HTTPIntercept) HandleNewMessage(raw string) {
	// Rewrite the notification issues
	if IsNotification(raw) {
		slog.Debug("message type notification.")
		h.sendNotification(raw)
	} else {
		if IsResponse(raw) {
			slog.Debug("IsResponse returned true", "raw", raw)
			// TODO:
			//	 Add to the correct queue.
			//	 h.OutboundResponse get ID + raw
			v, err := ExtractID(raw)
			if err != nil {
				slog.Error("skipping adding msg to outbound response", "error", err)
			} else {
				h.OutboundResponse[v] = raw
			}
		} else {
			h.sendRequest(raw)
		}

	}
}

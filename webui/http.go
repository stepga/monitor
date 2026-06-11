package webui

import (
	"embed"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/stepga/monitor/bus"
	"github.com/stepga/monitor/subsystems"
)

//go:embed assets/*
var assetsFS embed.FS

var webUiReporter *subsystems.WebUiReporter

func rootHandler(w http.ResponseWriter, r *http.Request) {
	index_path := "assets/index.html"
	index_data, err := assetsFS.ReadFile(index_path)
	if err != nil {
		slog.Error("ReadFile() failed", "error", err, "index_path", index_path)
		return
	}
	w.Write(index_data)
}

func sseSendData(w http.ResponseWriter, data any) {
	_, err := fmt.Fprintf(w, "data: %s\n\n", data)
	if err != nil {
		slog.Error("sseSendData() writing response failed", "error", err)
	}
}

func sseHandler(w http.ResponseWriter, r *http.Request) {
	// http headers required for SSE
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	// may be needed (locally) for CORS requests
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// create a channel for client disconnection
	clientGone := r.Context().Done()

	responseController := http.NewResponseController(w)
	for {
		select {
		case <-clientGone:
			slog.Info("client disconnected")
			return
		case msg := <-webUiReporter.RelevantMessages:
			data, err := json.Marshal(msg)
			if err != nil {
				slog.Error("failed to json encode the report, will send as plaintext", "msg", msg)
				sseSendData(w, msg)
			} else {
				sseSendData(w, data)
			}

			err = responseController.Flush()
			if err != nil {
				slog.Error("sseHandler() flushing response failed", "error", err)
				return
			}
		}
	}
}

func assetsFileHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		// TODO: further security measures as path sanitizing
		http.ServeFileFS(w, r, assetsFS, "assets"+r.URL.Path)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		fmt.Fprintf(w, "Only GET supported")
	}
}

func InitHttpHandlers(address string) {
	webUiReporter = &subsystems.WebUiReporter{
		RelevantMessages: make(chan any, bus.BusMsgSize),
	}
	webUiReporter.Init()

	http.HandleFunc("/events", sseHandler)
	http.HandleFunc("/static/", assetsFileHandler)
	http.HandleFunc("/", rootHandler)

	go func() {
		slog.Info("web user interface listens on ", "address", address)
		err := http.ListenAndServe(address, nil)
		slog.Error("ListenAndServe failed", "address", address, "error", err)
	}()
}

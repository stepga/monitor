package webui

import (
	"embed"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/stepga/monitor/bus"
	"github.com/stepga/monitor/config"
	"github.com/stepga/monitor/store"
)

// Report format for WebUI messages sent from daemon to browser
type WebUiMessage struct {
	Summary    string
	IsCritical bool
	Details    string
}

//go:embed assets/*
var assetsFS embed.FS

func rootHandler(w http.ResponseWriter, r *http.Request) {
	indexPath := "assets/index.html"
	indexData, err := assetsFS.ReadFile(indexPath)
	if err != nil {
		slog.Error("ReadFile() failed", "error", err, "indexPath", indexPath)
		return
	}
	w.Write(indexData)
}

func sseSendData(w http.ResponseWriter, data any) {
	_, err := fmt.Fprintf(w, "data: %s\n\n", data)
	if err != nil {
		slog.Error("sseSendData() writing response failed", "error", err)
	}
}

func (s *Server) notificationHandler(w http.ResponseWriter, r *http.Request) {
	// http headers required for SSE
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	// may be needed (locally) for CORS requests
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// create a channel for client disconnection
	clientGone := r.Context().Done()

	ch := bus.Subscribe()
	defer bus.Unsubscribe(ch)

	responseController := http.NewResponseController(w)
	for {
		select {
		case <-clientGone:
			slog.Info("client disconnected")
			return
		case m := <-ch:
			switch msg := m.(type) {
			case bus.Info:
				webUiMessage := infoToWebUiMessage(msg)
				webUiMessageJson, err := json.Marshal(webUiMessage)
				if err != nil {
					slog.Error("json encoding of WebUiMessage failed", "error", err, "webUiMessage", webUiMessage)
					continue
				}
				sseSendData(w, string(webUiMessageJson))

				err = responseController.Flush()
				if err != nil {
					slog.Error("notificationHandler() flushing response failed", "error", err)
				}
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

func (s *Server) criticalHandler(w http.ResponseWriter, _ *http.Request) {
	var messages []WebUiMessage
	w.Header().Set("Content-Type", "application/json")
	for _, critical := range store.FetchCritical() {
		messages = append(messages, infoToWebUiMessage(critical))
	}
	err := json.NewEncoder(w).Encode(messages)
	if err != nil {
		slog.Error("criticalHandler() json encoding failed", "error", err)
	}
}

func infoToWebUiMessage(info bus.Info) WebUiMessage {
	isCritical := false
	if _, ok := info.(bus.Critical); ok {
		isCritical = true
	}

	return WebUiMessage{
		Summary:    info.Summary(),
		IsCritical: isCritical,
		Details:    info.Details(),
	}
}

type Server struct{}

func (s *Server) Init() error {
	address := config.Cfg.WebUiAddress

	http.HandleFunc("/notifications", s.notificationHandler)
	http.HandleFunc("/static/", assetsFileHandler)
	http.HandleFunc("/critical", s.criticalHandler)
	http.HandleFunc("/", rootHandler)

	go func() {
		slog.Info("webui listens on", "address", address)
		err := http.ListenAndServe(address, nil)
		slog.Error("ListenAndServe failed", "address", address, "error", err)
	}()

	return nil
}

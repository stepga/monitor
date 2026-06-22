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

// Serialize bus.Info for sending from daemon to browser
type WebUiInfo struct {
	Summary   string `json:"summary"`
	Details   string `json:"details"`
	Timestamp string `json:"timestamp"`
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
				webUiInfo := infoToWebUiInfo(msg)
				webUiInfoJson, err := json.Marshal(webUiInfo)
				if err != nil {
					slog.Error("json encoding of WebUiInfo failed", "error", err, "webUiInfo", webUiInfo)
					continue
				}
				sseSendData(w, string(webUiInfoJson))

				err = responseController.Flush()
				if err != nil {
					slog.Error("notificationHandler() flushing response failed", "error", err)
				}
			}

		}
	}
}

func assetsFileHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		fmt.Fprintf(w, "Only GET supported")
		return
	}
	switch path := r.URL.Path; path {
	case "/index.html":
		fallthrough
	case "/static/style.css":
		fallthrough
	case "/static/script.js":
		http.ServeFileFS(w, r, assetsFS, "assets"+path)
	default:
		slog.Error("assetsFileHandler: requested asset url not whitelisted", "url", path)
		w.WriteHeader(http.StatusNotFound)
	}
}

func (s *Server) criticalHandler(w http.ResponseWriter, _ *http.Request) {
	var infos []WebUiInfo
	w.Header().Set("Content-Type", "application/json")
	for _, critical := range store.FetchCritical() {
		infos = append(infos, infoToWebUiInfo(critical))
	}
	err := json.NewEncoder(w).Encode(infos)
	if err != nil {
		slog.Error("criticalHandler() json encoding failed", "error", err)
	}
}

func infoToWebUiInfo(info bus.Info) WebUiInfo {
	return WebUiInfo{
		Summary:   info.Summary(),
		Details:   info.Details(),
		Timestamp: info.Timestamp(),
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

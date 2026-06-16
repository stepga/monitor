package webui

import (
	"embed"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"sync"

	"github.com/stepga/monitor/bus"
	"github.com/stepga/monitor/config"
)

//go:embed assets/*
var assetsFS embed.FS

type WebUiMessager interface {
	WebUiMessage() bus.WebUiMessage
}

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
			case WebUiMessager:
				webUiMessage := msg.WebUiMessage()
				webUiMessageJson, err := json.Marshal(webUiMessage)
				if err != nil {
					slog.Error("json encoding of WebUiMessage failed", "error", err, "webUiMessage", webUiMessage)
					continue
				}
				s.addWebUiMessage(webUiMessage)
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

func (s *Server) unresolvedHandler(w http.ResponseWriter, _ *http.Request) {
	slog.Info("XXX unresolvedHandler")
	w.Header().Set("Content-Type", "application/json")
	unresolved := make([]bus.WebUiMessage, 0)
	for _, msg := range s.Unresolved {
		slog.Info("XXX unresolvedHandler", "msg", msg)
		unresolved = append(unresolved, *msg)
	}
	err := json.NewEncoder(w).Encode(unresolved)
	if err != nil {
		slog.Error("unresolvedHandler() json encoding failed", "error", err)
	}
}

type Server struct {
	WebUiMessages []bus.WebUiMessage
	Unresolved    []*bus.WebUiMessage
	lock          sync.RWMutex
}

func (s *Server) addWebUiMessage(msg bus.WebUiMessage) {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.WebUiMessages = append(s.WebUiMessages, msg)
	if msg.IsCritical {
		msgPtr := &s.WebUiMessages[len(s.WebUiMessages)-1]
		s.Unresolved = append(s.Unresolved, msgPtr)
		return
	}

	stillUnresolved := []*bus.WebUiMessage{}
	for _, unresolvedMsg := range s.Unresolved {
		if msg.Source == unresolvedMsg.Source {
			// NodeInfo resolves NodeTimeout
			if msg.SubSystemName == "node" && unresolvedMsg.SubSystemName == "heartbeat" {
				continue
			}
			// DiskFineAgain resolves DiskGettingFull
			if msg.SubSystemName == "diskmon" && unresolvedMsg.SubSystemName == "diskmon" {
				continue
			}
			// TODO not yet implemented -> CertError
			// TODO not yet implemented -> CertExpiresSoon
		}
		stillUnresolved = append(stillUnresolved, unresolvedMsg)
	}
	s.Unresolved = stillUnresolved
}

func (s *Server) Init() error {
	address := config.Cfg.WebUiAddress

	http.HandleFunc("/notifications", s.notificationHandler)
	http.HandleFunc("/static/", assetsFileHandler)
	http.HandleFunc("/unresolved", s.unresolvedHandler)
	http.HandleFunc("/", rootHandler)

	go func() {
		slog.Info("webui listens on", "address", address)
		err := http.ListenAndServe(address, nil)
		slog.Error("ListenAndServe failed", "address", address, "error", err)
	}()

	return nil
}

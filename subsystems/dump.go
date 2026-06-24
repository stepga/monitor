package subsystems

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"time"

	"path/filepath"
	"strings"

	"github.com/stepga/monitor/bus"
	"github.com/stepga/monitor/config"
	"github.com/stepga/monitor/store"
)

type Dump struct{}
type DumpObject struct {
	SummaryField    string `json:"summary"`
	IdentifierField string `json:"identifier"`
	DetailsField    string `json:"details"`
	TimestampField  string `json:"timestamp"`
}

func (d DumpObject) Summary() string {
	now := time.Now().Format(time.DateTime)
	return fmt.Sprintf("%s (Restored: %s)", d.SummaryField, now)
}
func (d DumpObject) Identifier() string { return d.IdentifierField }
func (d DumpObject) Details() string    { return d.DetailsField }
func (d DumpObject) Timestamp() string  { return d.TimestampField }
func (d DumpObject) Critical()          {}

func dump(path string) error {
	dumps := []DumpObject{}
	for _, critical := range store.FetchCritical() {
		if _, ok := critical.(bus.NodeTimeout); !ok {
			continue
		}
		dumps = append(
			dumps,
			DumpObject{
				SummaryField:    critical.Summary(),
				IdentifierField: critical.Identifier(),
				DetailsField:    critical.Details(),
				TimestampField:  critical.Timestamp(),
			})
	}
	data, err := json.MarshalIndent(dumps, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func restoreData(data []byte) error {
	var dumps []DumpObject

	if err := json.Unmarshal(data, &dumps); err != nil {
		return err
	}

	for _, dump := range dumps {
		bus.Publish(dump)
	}

	return nil
}

func restore(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	err = restoreData(data)
	if err != nil {
		return err
	}
	return nil
}

func (r *Dump) Init() error {
	path, err := pathCheck(config.Cfg.DumpPath)
	if err != nil {
		return fmt.Errorf("Dump: path error: %v", err)
	}

	err = restore(path)
	if err != nil {
		slog.Error("Dump: initial dump restore failed", "error", err)
	}

	ch := bus.Subscribe()
	go func() {
		defer bus.Unsubscribe(ch)
		for msg := range ch {
			if _, ok := msg.(bus.CriticalListChanged); !ok {
				continue
			}
			err := dump(path)
			if err != nil {
				slog.Error("Dump: dump() failed", "error", err)
			}
		}
	}()

	return nil
}

func expandHome(path string) (string, error) {
	if strings.HasPrefix(path, "~") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(home, strings.TrimPrefix(path, "~")), nil
	}
	return path, nil
}

func pathCheck(path string) (string, error) {
	expandedPath, err := expandHome(path)
	if err != nil {
		return "", err
	}
	_, err = os.Stat(expandedPath)
	if err != nil && !os.IsNotExist(err) {
		return "", err
	}
	return expandedPath, nil
}

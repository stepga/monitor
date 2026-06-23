package subsystems

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"

	"path/filepath"
	"strings"

	"github.com/stepga/monitor/bus"
	"github.com/stepga/monitor/config"
	"github.com/stepga/monitor/store"
)

type Dump struct{}

type DumpObject struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

func dumpObject(typeName string, obj any) (*DumpObject, error) {
	data, err := json.MarshalIndent(obj, "", "  ")
	if err != nil {
		return nil, err
	}
	return &DumpObject{
		Type: typeName,
		Data: data,
	}, nil
}

func dump(path string) error {
	dumps := []DumpObject{}
	for _, critical := range store.FetchCritical() {
		if _, ok := critical.(bus.NodeTimeout); !ok {
			continue
		}
		dump, err := dumpObject("NodeTimeout", critical)
		if err != nil {
			slog.Error("dump() failed", "error", err, "critical message", critical)
			continue
		}
		dumps = append(dumps, *dump)
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
		switch dump.Type {
		case "NodeTimeout":
			obj := bus.NodeTimeout{}
			if err := json.Unmarshal(dump.Data, &obj); err != nil {
				return err
			}
			bus.Publish(obj)
		default:
			slog.Error("restoreData: type not implemented", "type", dump.Type)
		}
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

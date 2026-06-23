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
	if os.IsNotExist(err) {
		slog.Info("Dump: path does not exist yet", "expandedPath", expandedPath)
	}
	return expandedPath, nil
}

func dumpCritical(path string) error {
	dumps := []bus.CriticalDump{}
	for _, critical := range store.FetchCritical() {
		dump, err := critical.Dump()
		if err != nil {
			return fmt.Errorf("Dump() of '%s' failed: %v", critical.Identifier(), err)
		}
		dumps = append(dumps, *dump)
	}
	data, err := json.MarshalIndent(dumps, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func restoreData(data []byte) ([]bus.Critical, error) {
	var dumps []bus.CriticalDump

	if err := json.Unmarshal(data, &dumps); err != nil {
		return nil, err
	}

	result := make([]bus.Critical, 0, len(dumps))

	for _, dump := range dumps {
		factory, ok := bus.CriticalRegistry[dump.Type]
		if !ok {
			return nil, fmt.Errorf("unknown type %q", dump.Type)
		}
		obj := factory()
		if err := json.Unmarshal(dump.Data, obj); err != nil {
			return nil, err
		}
		result = append(result, obj)
	}

	return result, nil
}

func restoreCritical(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	criticals, err := restoreData(data)
	if err != nil {
		return err
	}
	for _, critical := range criticals {
		store.Add(critical)
	}
	return nil
}

func (r *Dump) Init() error {
	path, err := pathCheck(config.Cfg.DumpPath)
	if err != nil {
		return fmt.Errorf("Dump: path error: %v", err)
	}

	err = restoreCritical(path)
	if err != nil {
		slog.Error("Dump: restoring dump file failed", "error", err)
	}

	ch := bus.Subscribe()
	go func() {
		defer bus.Unsubscribe(ch)
		for msg := range ch {
			if _, ok := msg.(bus.CriticalListChanged); !ok {
				continue
			}
			err := dumpCritical(path)
			if err != nil {
				slog.Error("Dump: dumpCritical() failed", "error", err)
			}
		}
	}()

	return nil
}

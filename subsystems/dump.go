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

func pathExists(path string) (bool, error) {
	path, err := expandHome(path)
	if err != nil {
		return false, err
	}

	_, err = os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func (r *Dump) Init() error {
	dumpPath, err := expandHome(config.Cfg.DumpPath)
	if err != nil {
		return fmt.Errorf("Dump: path error: %v", err)
	}
	_, err = os.Stat(dumpPath)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("Dump: path error: %v", err)
	}
	if os.IsNotExist(err) {
		slog.Info("Dump: path does not exist yet", "path", dumpPath)
	}

	// TODO: restore file

	ch := bus.Subscribe()
	go func() {
		defer bus.Unsubscribe(ch)
		for msg := range ch {
			switch msg.(type) {
			case bus.CriticalListChanged:
				slog.Error("XXX fetch")
				infos := store.FetchCritical()
				b, err := json.MarshalIndent(infos, "", "  ")
				if err != nil {
					slog.Error("XXX", "error", err)
					continue
				}
				os.WriteFile(dumpPath, b, 0644)
			default:
			}
		}
	}()

	return nil
}

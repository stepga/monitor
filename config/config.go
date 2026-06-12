package config

import (
	"encoding/json"
	"os"
	"slices"
	"time"
)

var (
	Cfg Config
)

func LoadConfig(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	if err := json.NewDecoder(file).Decode(&Cfg); err != nil {
		return err
	}

	return nil
}

// Given a list of subsystems returns true if all of them are
// configured to be started, returns false otherwise
func AllSubsystemsConfigured(subsystems []string) bool {
	for _, s := range subsystems {
		if !slices.Contains(Cfg.Subsystems, s) {
			return false
		}
	}
	return true
}

type ListenerConfig struct {
	Address      string `json:"address"`
	MaxMsgSizeMB int    `json:"maxMsgSizeMB"`
}

type CertConfig struct {
	MinimumDaysLeft      int           `json:"minimum_days_left"`
	CheckIntervalInHours time.Duration `json:"check_interval_in_hours"`
	Urls                 []string      `json:"urls"`
}

type HeartbeatConfig struct {
	NodeTimeoutInMinutes   time.Duration `json:"node_timeout_in_minutes"`
	CheckIntervalInMinutes time.Duration `json:"check_interval_in_minutes"`
}

type Config struct {
	Subsystems    []string        `json:"subsystems"`
	DiskThreshold float64         `json:"diskThreshold"`
	Cert          CertConfig      `json:"cert"`
	Listener      ListenerConfig  `json:"listener"`
	WebUiAddress  string          `json:"webui_address"`
	Heartbeat     HeartbeatConfig `json:"heartbeat"`
}

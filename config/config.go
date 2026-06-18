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

// Loads Config from given file and sets Cfg on success, returns
// error otherwise
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
	MaxMsgSizeMB int    `json:"max_msg_size_mb"`
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

type RedmineConfig struct {
	Url     string `json:"url"`
	IssueId string `json:"issue_id"`
}

type Config struct {
	Subsystems    []string        `json:"subsystems"`
	DiskThreshold float64         `json:"disk_threshold"`
	Cert          CertConfig      `json:"cert"`
	Listener      ListenerConfig  `json:"listener"`
	WebUiAddress  string          `json:"webui_address"`
	Heartbeat     HeartbeatConfig `json:"heartbeat"`
	Redmine       RedmineConfig   `json:"redmine"`
}

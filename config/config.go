package config

import (
	"encoding/json"
	"os"
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

type ListenerConfig struct {
	Address      string `json:"address"`
	MaxMsgSizeMB int    `json:"maxMsgSizeMB"`
}

type CertConfig struct {
	MinimumDaysLeft int      `json:"minimum_days_left"`
	Urls            []string `json:"urls"`
}

type Config struct {
	Reporter      []string       `json:"reporter"`
	Collectors    []string       `json:"collectors"`
	DiskThreshold float64        `json:"diskThreshold"`
	Cert          CertConfig     `json:"cert"`
	Listener      ListenerConfig `json:"listener"`
}

package main

import (
	"encoding/json"
	"flag"
	"os"
	"time"
	"fmt"
)

type CertConfig struct {
	MinimumDaysLeft int      `json:"minimum_days_left"`
	Urls            []string `json:"urls"`
}

type Config struct {
	Cert CertConfig `json:"cert"`
}

func LoadConfig(path string) (*Config, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var cfg Config
	if err := json.NewDecoder(file).Decode(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func printCertInfo(info []CertInfo, minimumDaysLeft int) {
	threshold := time.Duration(minimumDaysLeft * 24) * time.Hour
	for _, info := range info {
		fmt.Printf("%s (%dms): ", info.Url, info.Took.Milliseconds())
		if info.Error != nil {
			fmt.Printf("ERROR: %s\n", info.Error)
			continue
		}
		remaining := time.Until(*info.Expiry)

		if remaining < threshold {
			fmt.Printf(
				"EXPIRES SOON %v remaining, expires %s\n",
				remaining.Round(time.Hour),
				info.Expiry.Format(time.UnixDate),
			)
		} else {
			fmt.Printf(
				"OK %v remaining, expires %s\n",
				remaining.Round(time.Hour),
				info.Expiry.Format(time.UnixDate),
			)
		}
	}
}

func main() {
	config_file := flag.String("config", "config.json", "Path to config.json file")
	flag.Parse()

	cfg, err := LoadConfig(*config_file)
	if err != nil {
		panic(err)
	}

	info := check_certs(cfg.Cert.Urls)
	printCertInfo(info, cfg.Cert.MinimumDaysLeft)
}

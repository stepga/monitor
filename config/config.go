package config

type ListenerConfig struct {
	Address string `json:"address"`
}

type CertConfig struct {
	MinimumDaysLeft int      `json:"minimum_days_left"`
	Urls            []string `json:"urls"`
}

type Config struct {
	Reporter   []string       `json:"reporter"`
	Collectors []string       `json:"collectors"`
	Cert       CertConfig     `json:"cert"`
	Listener   ListenerConfig `json:"listener"`
}

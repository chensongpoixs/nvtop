package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type ServerConfig struct {
	Host string `yaml:"host" json:"host"`
	Port int    `yaml:"port" json:"port"`
}

type MonitorConfig struct {
	PollIntervalSeconds int `yaml:"poll_interval_seconds" json:"poll_interval_seconds"`
	HistorySize         int `yaml:"history_size" json:"history_size"`
}

type LogConfig struct {
	Level string `yaml:"level" json:"level"`
}

type Config struct {
	Server  ServerConfig  `yaml:"server" json:"server"`
	Monitor MonitorConfig `yaml:"monitor" json:"monitor"`
	Log     LogConfig     `yaml:"log" json:"log"`
}

var Default = Config{
	Server: ServerConfig{
		Host: "0.0.0.0",
		Port: 8080,
	},
	Monitor: MonitorConfig{
		PollIntervalSeconds: 1,
		HistorySize:         60,
	},
	Log: LogConfig{
		Level: "info",
	},
}

func Load(path string) (*Config, error) {
	cfg := Default

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &cfg, nil
		}
		return nil, err
	}

	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

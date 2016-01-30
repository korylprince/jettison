package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/kelseyhightower/envconfig"
)

//Config stores configuration from the environment
type Config struct {
	GroupStr     string `envconfig:"GROUPS"`
	Groups       []string
	Serial       string
	HardwareAddr string
	Location     string

	ReportInterval time.Duration //in seconds
	CheckInterval  time.Duration //in seconds

	HTTPServerAddr string
	RPCServerAddr  string
	CachePath      string
}

//ParseEnv parses a Config from the environment, returning an error if one occurred
func ParseEnv() (*Config, error) {
	config := &Config{}
	err := envconfig.Process("JETTISON", config)
	if err != nil {
		return nil, fmt.Errorf("Error reading configuration from environment: %v", err)
	}

	config.Groups = strings.Split(config.GroupStr, ",")
	if len(config.Groups) == 0 {
		return nil, fmt.Errorf("JETTISON_GROUPS must be configured")
	}
	if config.Location == "" {
		return nil, fmt.Errorf("JETTISON_LOCATION must be configured")
	}
	if config.Serial == "" {
		return nil, fmt.Errorf("JETTISON_SERIAL must be configured")
	}
	if config.HardwareAddr == "" {
		return nil, fmt.Errorf("JETTISON_HARDWAREADDR must be configured")
	}
	if config.ReportInterval == 0 {
		config.ReportInterval = 60

	}
	if config.CheckInterval == 0 {
		config.CheckInterval = 10 * 60

	}
	if config.HTTPServerAddr == "" {
		return nil, fmt.Errorf("JETTISON_HTTPSERVERADDR must be configured")
	}
	if config.RPCServerAddr == "" {
		return nil, fmt.Errorf("JETTISON_RPCSERVERADDR must be configured")
	}
	if config.CachePath == "" {
		return nil, fmt.Errorf("JETTISON_CACHEPATH must be configured")
	}

	return config, nil
}

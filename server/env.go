package main

import (
	"fmt"

	"github.com/kelseyhightower/envconfig"
)

//Config stores configuration from the environment
type Config struct {
	HTTPListenAddr string
	RPCListenAddr  string
	DefinitionPath string
	CachePath      string
}

//ParseEnv parses a Config from the environment, returning an error if one occurred
func ParseEnv() (*Config, error) {
	config := &Config{}
	err := envconfig.Process("JETTISON", config)
	if err != nil {
		return nil, fmt.Errorf("Error reading configuration from environment: %v", err)
	}
	if config.HTTPListenAddr == "" {
		config.HTTPListenAddr = ":50080"
	}
	if config.RPCListenAddr == "" {
		config.RPCListenAddr = ":50081"
	}
	if config.DefinitionPath == "" {
		return nil, fmt.Errorf("JETTISON_DEFINITIONPATH must be configured")
	}
	if config.CachePath == "" {
		return nil, fmt.Errorf("JETTISON_CACHEPATH must be configured")
	}

	return config, nil
}

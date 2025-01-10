package gateway

import (
	"encoding/json"
	"os"
)

type Downstream struct {
	Host  string `json:"host"`
	Port  int    `json:"port"`
	Route string `json:"route"`
}

type AuthConfig struct {
	Key           string   `json:"key"`
	RoleClaimKey  string   `json:"roleClaimKey"`
	AllowedValues []string `json:"allowedValues"`
}

type Route struct {
	Downstream Downstream     `json:"downstream"`
	Upstream   string         `json:"upstream"`
	Methods    []string       `json:"methods"`
	Middleware []string       `json:"middleware"`
	Metadata   map[string]any `json:"metadata"`
	Auth       *AuthConfig    `json:"auth"`
}

type Config struct {
	Routes []Route `json:"routes"`
}

func LoadConfig(configPath string) (*Config, error) {
	file, err := os.Open(configPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var config Config
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&config); err != nil {
		return nil, err
	}
	return &config, nil
}

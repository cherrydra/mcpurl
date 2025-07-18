package config

import (
	"encoding/json"
	"fmt"
	"os"
)

type KV map[string]string

func (kv KV) Encode() []string {
	var ret []string
	for k, v := range kv {
		ret = append(ret, fmt.Sprintf("%s=%s", k, v))
	}
	return ret
}

type Server struct {
	Type    string   `json:"type"`
	Command string   `json:"command"`
	Args    []string `json:"args"`
	Env     KV       `json:"env"`
	URL     string   `json:"url"`
	Headers KV       `json:"headers"`
}

type Config struct {
	Servers map[string]Server `json:"servers"`
}

func Parse(file string) (*Config, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, fmt.Errorf("failed to open config file: %w", err)
	}
	var conf Config
	if err := json.NewDecoder(f).Decode(&conf); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}
	return &conf, nil
}
